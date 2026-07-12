/**
 * SFUGroupEngine —— 群组通话引擎（SFU / 自建 pion 版，取代 mesh 的 GroupCallEngine）。
 *
 * 与 1:1 引擎并存、互不影响。群组媒体不再两两 P2P，而是每人与 streaming 内的 SFU 建**一条**
 * PeerConnection：上行发布本地轨、下行由 SFU 转发其他人的轨；服务端作 offerer 驱动 renegotiation。
 *  - 呼叫控制仍走 streaming ws（group_invite/join/leave）与 im ws 振铃（call.signal）。
 *  - 拿到 group_created/group_roster 后发布（sfu_publish -> sfu_publish_answer）。
 *  - 服务端每次为本端增/减下行轨会发 sfu_offer，本端回 sfu_answer。
 *  - 远端轨经 pc.ontrack 到达，按 stream.id(=发布者 uid) 分组渲染。
 *
 * 回调契约与旧 GroupCallEngine 完全一致（GroupCallEngineCallbacks），故 call-store/CallOverlay 改动最小。
 */

import type { CallMediaType, CallPhase } from './call-engine';

/** 群通话引擎回调契约（UI/store 侧消费，来源从 mesh 换成 SFU 后保持不变）。 */
export interface GroupCallEngineCallbacks {
  onPhase: (phase: CallPhase, info?: { reason?: string }) => void;
  onLocalStream: (s: MediaStream | null) => void;
  onParticipantStream: (uid: string, stream: MediaStream | null) => void;
  onParticipantMedia: (uid: string, media: { audio: boolean; video: boolean }) => void;
  onParticipantLeft: (uid: string) => void;
  onPeerEvent: (uid: string, kind: 'joined' | 'left') => void;
  onError: (key: string) => void;
}

interface StreamingMsg {
  type: string;
  room_id?: string;
  user_id?: string;
  data?: Record<string, unknown>;
}

const DEFAULT_ICE: RTCIceServer[] = [{ urls: 'stun:stun.l.google.com:19302' }];

export class SFUGroupEngine {
  private ws: WebSocket | null = null;
  private pc: RTCPeerConnection | null = null;
  private localStream: MediaStream | null = null;
  private remoteUids = new Set<string>();
  private hadPeer = false;
  private joinTimer: ReturnType<typeof setTimeout> | null = null;
  private callId = '';
  private mediaType: CallMediaType = 'voice';
  private phase: CallPhase = 'idle';
  private iceServers: RTCIceServer[] = DEFAULT_ICE;

  constructor(
    private opts: { wsUrl: string; httpBase: string; token: string; selfId: string },
    private cb: GroupCallEngineCallbacks,
  ) {}

  get currentPhase() { return this.phase; }

  // ==================== 发起 / 接听 ====================

  /** 发起群通话（发起人） */
  async startGroupCall(groupId: string, members: string[], type: CallMediaType) {
    if (this.phase !== 'idle') return;
    this.mediaType = type;
    this.setPhase('outgoing');
    try {
      await this.getLocalMedia(type);
      await this.connectWs();
      this.sendWs('group_invite', { group_id: groupId, members, call_type: type });
      // 兜底：45s 内无人加入则结束（发起人空等）
      this.joinTimer = setTimeout(() => {
        if (!this.hadPeer) {
          this.cb.onError('call.err.noAnswer');
          this.hangup();
        }
      }, 45000);
    } catch (e) {
      this.cb.onError(this.mediaErr(e));
      this.cleanup();
    }
  }

  /** 接听群通话来电（被邀者）。callId/type 来自先前 setIncoming。 */
  async acceptGroup() {
    if (this.phase !== 'incoming' || !this.callId) return;
    this.setPhase('connecting');
    try {
      await this.getLocalMedia(this.mediaType);
      await this.connectWs();
      this.sendWs('group_join', { call_id: this.callId });
    } catch (e) {
      this.cb.onError(this.mediaErr(e));
      this.cleanup();
    }
  }

  /** 主动加入一通正在进行的群通话（点「加入通话」，非被邀振铃）。 */
  async joinExisting(callId: string, type: CallMediaType) {
    if (this.phase !== 'idle') return;
    this.callId = callId;
    this.mediaType = type;
    this.setPhase('connecting');
    try {
      await this.getLocalMedia(type);
      await this.connectWs();
      this.sendWs('group_join', { call_id: callId });
    } catch (e) {
      this.cb.onError(this.mediaErr(e));
      this.cleanup();
    }
  }

  /** 标记来电（call-store 收到 im ws group.invite 时调用） */
  setIncoming(callId: string, type: CallMediaType) {
    if (this.phase !== 'idle') return;
    this.callId = callId;
    this.mediaType = type;
    this.setPhase('incoming');
  }

  /** 挂断 / 离开群通话 */
  hangup() {
    if (this.callId && this.ws?.readyState === WebSocket.OPEN) {
      this.sendWs('group_leave', { call_id: this.callId });
    }
    this.cleanup();
  }

  reject() {
    // 拒接群通话：直接结束本端（MVP 不通知发起方）
    this.cleanup();
  }

  toggleMute(muted: boolean) {
    this.localStream?.getAudioTracks().forEach(tr => (tr.enabled = !muted));
  }
  toggleCamera(on: boolean) {
    this.localStream?.getVideoTracks().forEach(tr => (tr.enabled = on));
  }

  // ==================== streaming ws（控制面 + SFU 协商） ====================

  private async connectWs(): Promise<void> {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) return;
    await this.loadIceServers();
    return new Promise((resolve, reject) => {
      const url = `${this.opts.wsUrl}?token=${encodeURIComponent(this.opts.token)}`;
      const ws = new WebSocket(url);
      this.ws = ws;
      ws.onopen = () => resolve();
      ws.onerror = () => reject(new Error('group signaling ws connect failed'));
      ws.onmessage = (e) => this.handleStreamingMsg(e);
    });
  }

  private handleStreamingMsg(e: MessageEvent) {
    let msg: StreamingMsg;
    try { msg = JSON.parse(e.data); } catch { return; }
    const data = msg.data || {};
    this.log('ws recv:', msg.type);

    switch (msg.type) {
      case 'group_created':
      case 'group_roster':
        // 服务端已建好本端 SFU peer；开始发布本地轨
        this.callId = (data.call_id as string) || this.callId;
        this.setPhase('connected');
        this.publish().catch(() => this.cb.onError('call.err.connect'));
        break;
      case 'sfu_publish_answer':
        this.pc?.setRemoteDescription({ type: 'answer', sdp: String(data.sdp ?? '') })
          .catch(() => { /* ignore */ });
        break;
      case 'sfu_offer':
        // 服务端发起的 renegotiation（新增/移除下行轨）：应答回去
        this.handleServerOffer(String(data.sdp ?? '')).catch(() => { /* ignore */ });
        break;
      case 'sfu_peer_left': {
        const uid = data.uid as string;
        if (uid) this.onParticipantGone(uid);
        break;
      }
      case 'error':
        this.cb.onError('call.err.generic');
        break;
    }
  }

  private sendWs(type: string, data: Record<string, unknown>) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type, room_id: this.callId, data, timestamp: new Date().toISOString() }));
    }
  }

  // ==================== SFU PeerConnection ====================

  /** 建本端与 SFU 的 PeerConnection，发布本地轨并发 sfu_publish（非 trickle：等 ICE 收集完再发）。 */
  private async publish() {
    if (this.pc) return; // 已发布
    const pc = new RTCPeerConnection({ iceServers: this.iceServers });
    this.pc = pc;
    pc.ontrack = (ev) => this.onRemoteTrack(ev);
    pc.onconnectionstatechange = () => {
      if (pc.connectionState === 'failed') this.cb.onError('call.err.connect');
    };

    this.localStream?.getTracks().forEach(tr => pc.addTrack(tr, this.localStream!));

    const offer = await pc.createOffer();
    await pc.setLocalDescription(offer);
    await this.waitIceGathering(pc);
    this.sendWs('sfu_publish', { call_id: this.callId, sdp: pc.localDescription?.sdp ?? '' });
  }

  /** 处理服务端 renegotiation offer：setRemote -> answer -> setLocal -> 等收集 -> 回 sfu_answer。 */
  private async handleServerOffer(sdp: string) {
    if (!this.pc || !sdp) return;
    await this.pc.setRemoteDescription({ type: 'offer', sdp });
    const answer = await this.pc.createAnswer();
    await this.pc.setLocalDescription(answer);
    await this.waitIceGathering(this.pc);
    this.sendWs('sfu_answer', { call_id: this.callId, sdp: this.pc.localDescription?.sdp ?? '' });
  }

  /** 远端轨到达：stream.id = 发布者 uid（SFU 下行轨 streamID=pubUID），按 uid 分组渲染。 */
  private onRemoteTrack(ev: RTCTrackEvent) {
    const stream = ev.streams[0];
    if (!stream) return;
    const uid = stream.id;
    if (!uid || uid === this.opts.selfId) return;

    if (!this.remoteUids.has(uid)) {
      this.remoteUids.add(uid);
      this.hadPeer = true;
      if (this.joinTimer) { clearTimeout(this.joinTimer); this.joinTimer = null; }
      this.cb.onPeerEvent(uid, 'joined');
    }
    this.cb.onParticipantStream(uid, stream);
  }

  /** 某参与者离开（服务端 sfu_peer_left）：移除 tile；只剩自己则结束。 */
  private onParticipantGone(uid: string) {
    if (!this.remoteUids.delete(uid)) return;
    this.cb.onParticipantLeft(uid);
    this.cb.onPeerEvent(uid, 'left');
    if (this.hadPeer && this.remoteUids.size === 0 && (this.phase === 'connected' || this.phase === 'connecting')) {
      this.hangup();
    }
  }

  private waitIceGathering(pc: RTCPeerConnection): Promise<void> {
    if (pc.iceGatheringState === 'complete') return Promise.resolve();
    return new Promise(resolve => {
      const done = () => { pc.removeEventListener('icegatheringstatechange', check); resolve(); };
      const check = () => { if (pc.iceGatheringState === 'complete') done(); };
      pc.addEventListener('icegatheringstatechange', check);
      setTimeout(done, 3000); // 兜底：本地/局域网通常秒级完成
    });
  }

  private async loadIceServers() {
    try {
      const ctrl = new AbortController();
      const timer = setTimeout(() => ctrl.abort(), 2500);
      const res = await fetch(`${this.opts.httpBase}/v1/streaming/ice-servers?token=${encodeURIComponent(this.opts.token)}`, { signal: ctrl.signal });
      clearTimeout(timer);
      if (res.ok) {
        const body = await res.json();
        if (Array.isArray(body?.iceServers) && body.iceServers.length) this.iceServers = body.iceServers as RTCIceServer[];
      }
    } catch {
      this.iceServers = DEFAULT_ICE;
    }
  }

  // ==================== 媒体 / 状态 ====================

  private async getLocalMedia(type: CallMediaType) {
    const stream = await navigator.mediaDevices.getUserMedia({
      audio: true,
      video: type === 'video' ? { width: 1280, height: 720 } : false,
    });
    this.localStream = stream;
    this.cb.onLocalStream(stream);
  }

  private mediaErr(e: unknown) {
    const name = (e as { name?: string })?.name;
    if (name === 'NotAllowedError' || name === 'NotFoundError') return 'call.err.media';
    return 'call.err.init';
  }

  private setPhase(p: CallPhase) {
    this.phase = p;
    this.cb.onPhase(p);
  }

  private log(...args: unknown[]) {
    console.log('[sfu-gcall]', ...args);
  }

  private cleanup() {
    if (this.joinTimer) { clearTimeout(this.joinTimer); this.joinTimer = null; }
    try { this.pc?.close(); } catch { /* ignore */ }
    this.pc = null;
    this.remoteUids.clear();
    this.hadPeer = false;
    this.localStream?.getTracks().forEach(tr => tr.stop());
    try { this.ws?.close(); } catch { /* ignore */ }
    this.ws = null;
    this.localStream = null;
    this.cb.onLocalStream(null);
    this.setPhase('ended');
    this.callId = '';
    this.phase = 'idle';
  }
}
