/**
 * GroupCallEngine —— 群组（Mesh）通话引擎。
 *
 * 与 1:1 引擎并存、互不影响。群组 = N 条 PeerConn（两两 P2P）：
 *  - 振铃走 im ws call.signal（event=group.invite），由 call-store 喂入
 *  - 媒体协商走本引擎自建的 streaming ws（offer/answer/ice 帧带 to=对端uid，由服务端定向 relay）
 *  - 防 glare：老人(已在房间)向新人(peer_joined)发 offer，新人答复
 */

import type { CallMediaType, CallPhase } from './call-engine';
import { PeerConn } from './peer-conn';

export interface GroupCallEngineCallbacks {
  onPhase: (phase: CallPhase, info?: { reason?: string }) => void;
  onLocalStream: (s: MediaStream | null) => void;
  onParticipantStream: (uid: string, stream: MediaStream | null) => void;
  onParticipantMedia: (uid: string, media: { audio: boolean; video: boolean }) => void;
  onParticipantLeft: (uid: string) => void;
  onError: (key: string) => void;
}

interface StreamingMsg {
  type: string;
  room_id?: string;
  user_id?: string;
  data?: Record<string, unknown>;
}

const DEFAULT_ICE: RTCIceServer[] = [{ urls: 'stun:stun.l.google.com:19302' }];

export class GroupCallEngine {
  private ws: WebSocket | null = null;
  private localStream: MediaStream | null = null;
  private peers = new Map<string, PeerConn>();
  private hadPeer = false; // 是否曾有其他参与者（用于「只剩自己时结束」判断）
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
    this.broadcastMediaState();
  }
  toggleCamera(on: boolean) {
    this.localStream?.getVideoTracks().forEach(tr => (tr.enabled = on));
    this.broadcastMediaState();
  }

  /** 本端当前媒体开关状态（默认音开、视频按通话类型） */
  private mediaStatePayload() {
    const audio = this.localStream?.getAudioTracks().some(t => t.enabled) ?? true;
    const video = this.localStream?.getVideoTracks().some(t => t.enabled) ?? false;
    return { audio, video };
  }
  /** 向全部对端广播本端开关麦/摄像头状态 */
  private broadcastMediaState() {
    const st = this.mediaStatePayload();
    this.peers.forEach((_, uid) => this.sendWs('media_state', { to: uid, ...st }));
  }
  /** 向单个对端发送本端状态（新人加入时让其立即看到我已静音/关摄像头） */
  private sendMediaStateTo(uid: string) {
    this.sendWs('media_state', { to: uid, ...this.mediaStatePayload() });
  }

  // ==================== streaming ws ====================

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
    const from = msg.user_id || '';
    this.log('ws recv:', msg.type, 'from=', from);

    switch (msg.type) {
      case 'group_created':
        this.callId = (data.call_id as string) || this.callId;
        this.setPhase('connected'); // 发起人已在房间，等其他人加入
        break;
      case 'group_roster':
        this.callId = (data.call_id as string) || this.callId;
        // 现有参与者会主动向我 offer；这里仅切到 connected（tile 随流到达出现）
        this.setPhase('connected');
        break;
      case 'peer_joined': {
        // 我是老人 → 向新人发 offer
        const uid = data.uid as string;
        if (uid && uid !== this.opts.selfId) {
          const pc = this.ensurePeer(uid);
          pc.makeOffer().catch(() => this.cb.onError('call.err.connect'));
        }
        break;
      }
      case 'peer_left': {
        const uid = data.uid as string;
        this.removePeer(uid);
        break;
      }
      case 'offer':
        if (from) this.ensurePeer(from).handleOffer(String(data.sdp ?? '')).catch(() => { /* ignore */ });
        break;
      case 'answer':
        if (from) this.peers.get(from)?.handleAnswer(String(data.sdp ?? '')).catch(() => { /* ignore */ });
        break;
      case 'ice_candidate':
        if (from) this.peers.get(from)?.addRemoteIce(data);
        break;
      case 'media_state':
        if (from) this.cb.onParticipantMedia(from, { audio: data.audio !== false, video: data.video === true });
        break;
      case 'error':
        this.cb.onError('call.err.generic');
        break;
    }
  }

  private ensurePeer(uid: string): PeerConn {
    let pc = this.peers.get(uid);
    if (!pc) {
      pc = new PeerConn(
        uid,
        this.localStream,
        this.iceServers,
        (type, d) => this.sendWs(type, d),
        (id, stream) => this.cb.onParticipantStream(id, stream),
        (id, state) => { if (state === 'failed') this.removePeer(id); },
      );
      this.peers.set(uid, pc);
      this.hadPeer = true;
      if (this.joinTimer) { clearTimeout(this.joinTimer); this.joinTimer = null; }
      this.sendMediaStateTo(uid); // 让新对端立即看到我当前的开关麦/摄像头状态
    }
    return pc;
  }

  private removePeer(uid: string) {
    const pc = this.peers.get(uid);
    if (pc) { pc.close(); this.peers.delete(uid); }
    this.cb.onParticipantLeft(uid);
    // 群通话只剩自己（曾有人、现已 0 人）→ 直接结束
    if (this.hadPeer && this.peers.size === 0 && (this.phase === 'connected' || this.phase === 'connecting')) {
      this.hangup();
    }
  }

  private sendWs(type: string, data: Record<string, unknown>) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type, room_id: this.callId, data, timestamp: new Date().toISOString() }));
    }
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
    console.log('[gcall]', ...args);
  }

  private cleanup() {
    if (this.joinTimer) { clearTimeout(this.joinTimer); this.joinTimer = null; }
    this.peers.forEach(p => p.close());
    this.peers.clear();
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
