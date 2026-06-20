/**
 * WebRTC 通话引擎 —— 对接 streaming 服务
 *
 * 信令双通道（见 docs/specs/streaming-audio-video-call.md）：
 *   - 收呼叫控制（来电/接听/拒接/挂断/超时）：im ws 的 method=call.signal，由 call-store 喂进 onControlSignal()
 *   - 发动作 + 媒体协商：本引擎自建的 streaming ws（:10093/ws?token=）
 *
 * 1:1 为 P2P：streaming 仅转发 offer/answer/ice，媒体在两端浏览器之间直连。
 */

export type CallMediaType = 'voice' | 'video';
export type CallPhase = 'idle' | 'outgoing' | 'incoming' | 'connecting' | 'connected' | 'ended';

export interface CallPeer {
  id: string;
  name: string;
  avatar?: string;
}

/** im ws call.signal 帧（与后端 ws.CallSignal 对应） */
export interface CallSignal {
  event: 'invite' | 'cancel' | 'accept' | 'reject' | 'busy' | 'timeout' | 'end' | 'group.invite' | 'group.state';
  callId: string;
  callType?: CallMediaType;
  mediaMode?: string;
  scope?: string;
  fromUid?: string;
  fromName?: string;
  fromAvatar?: string;
  groupId?: string;
  members?: string[];
  reason?: string;
  duration?: number;
}

export interface CallEngineCallbacks {
  onPhase: (phase: CallPhase, info?: { reason?: string; duration?: number }) => void;
  onLocalStream: (s: MediaStream | null) => void;
  onRemoteStream: (s: MediaStream | null) => void;
  onRemoteMedia: (state: { audio: boolean; video: boolean }) => void;
  onError: (msg: string) => void;
}

interface StreamingMsg {
  type: string;
  room_id?: string;
  user_id?: string;
  data?: Record<string, unknown>;
}

const DEFAULT_ICE: RTCIceServer[] = [{ urls: 'stun:stun.l.google.com:19302' }];

export class CallEngine {
  private ws: WebSocket | null = null;
  private pc: RTCPeerConnection | null = null;
  private localStream: MediaStream | null = null;
  private remoteStream: MediaStream | null = null;

  private callId = '';
  private peer: CallPeer | null = null;
  private mediaType: CallMediaType = 'voice';
  private isCaller = false;
  private phase: CallPhase = 'idle';

  private iceServers: RTCIceServer[] = DEFAULT_ICE;
  private pendingIce: RTCIceCandidateInit[] = [];
  private remoteDescSet = false;

  constructor(
    private opts: { wsUrl: string; httpBase: string; token: string; selfId: string },
    private cb: CallEngineCallbacks,
  ) {}

  // ==================== 公开状态 ====================

  get currentPhase() { return this.phase; }
  get currentPeer() { return this.peer; }
  get currentMediaType() { return this.mediaType; }
  get currentCallId() { return this.callId; }

  // ==================== 主叫 ====================

  /** 发起呼叫 */
  async startCall(peer: CallPeer, type: CallMediaType) {
    if (this.phase !== 'idle') return;
    this.isCaller = true;
    this.peer = peer;
    this.mediaType = type;
    this.setPhase('outgoing');

    try {
      await this.getLocalMedia(type);
      await this.connectWs();
      this.sendWs('call_invite', { callee_id: peer.id, call_type: type });
    } catch (e) {
      this.cb.onError(this.mediaErr(e));
      this.cleanup('failed');
    }
  }

  // ==================== 被叫 ====================

  /** 接听来电 */
  async accept() {
    if (this.phase !== 'incoming' || !this.callId) return;
    this.setPhase('connecting');
    try {
      await this.getLocalMedia(this.mediaType);
      await this.connectWs();
      this.createPeerConnection(); // 被叫先备好 PC，等主叫 offer
      this.sendWs('call_accept', { call_id: this.callId });
    } catch (e) {
      this.cb.onError(this.mediaErr(e));
      this.sendWsBestEffort('call_reject', { call_id: this.callId });
      this.cleanup('failed');
    }
  }

  /** 拒接来电 */
  async reject() {
    if (this.phase !== 'incoming' || !this.callId) {
      this.cleanup();
      return;
    }
    try {
      await this.connectWs();
      this.sendWs('call_reject', { call_id: this.callId });
    } catch {
      /* ignore */
    }
    this.cleanup();
  }

  // ==================== 通用 ====================

  /** 挂断 / 取消 */
  hangup() {
    // 本端主动挂断也要带结束原因，否则主叫端无法据此投递通话记录：
    // 已接通=completed；响铃/连接中挂断=canceled（主叫放弃）
    const reason = this.phase === 'connected' ? 'completed' : 'canceled';
    if (this.callId && this.ws?.readyState === WebSocket.OPEN) {
      this.sendWs('call_end', { call_id: this.callId });
    }
    this.cleanup(reason);
  }

  toggleMute(muted: boolean) {
    this.localStream?.getAudioTracks().forEach(t => (t.enabled = !muted));
    this.sendWsBestEffort('media_state', { call_id: this.callId, audio: !muted, video: this.isVideoOn() });
  }

  toggleCamera(on: boolean) {
    this.localStream?.getVideoTracks().forEach(t => (t.enabled = on));
    this.sendWsBestEffort('media_state', { call_id: this.callId, audio: this.isAudioOn(), video: on });
  }

  /** im ws call.signal -> 引擎 */
  onControlSignal(sig: CallSignal) {
    this.log('control signal:', sig.event, 'isCaller=', this.isCaller, 'phase=', this.phase);
    switch (sig.event) {
      case 'invite':
        // 来电：仅置 incoming，展示振铃界面；接听才连 ws/取媒体
        if (this.phase !== 'idle') {
          // 已在通话中，理论上后端 busy 已挡；前端兜底忽略
          return;
        }
        this.isCaller = false;
        this.callId = sig.callId;
        this.mediaType = sig.callType || 'voice';
        this.peer = { id: sig.fromUid || '', name: sig.fromName || '', avatar: sig.fromAvatar };
        this.setPhase('incoming');
        break;

      case 'accept':
        // 主叫收到被叫接听 -> 开始媒体协商（主叫造 offer）
        if (this.isCaller && this.phase === 'outgoing') {
          this.setPhase('connecting');
          this.startNegotiationAsCaller();
        }
        break;

      case 'reject':
        this.cb.onError('call.err.rejected');
        this.cleanup('rejected');
        break;

      case 'busy':
        this.cb.onError('call.err.busy');
        this.cleanup('busy');
        break;

      case 'timeout':
        // 主叫：对方未接听；被叫：你有一个未接来电
        this.cb.onError(this.isCaller ? 'call.err.noAnswer' : this.missedKey());
        this.cleanup('no_answer');
        break;

      case 'cancel':
        // 被叫收到主叫响铃中取消 -> 未接来电提示
        this.cb.onError(this.missedKey());
        this.cleanup('canceled');
        break;

      case 'end':
        this.cleanup(sig.reason, sig.duration);
        break;
    }
  }

  // ==================== 内部：媒体协商 ====================

  private async startNegotiationAsCaller() {
    try {
      this.createPeerConnection();
      const offer = await this.pc!.createOffer();
      await this.pc!.setLocalDescription(offer);
      this.sendWs('offer', { call_id: this.callId, sdp: offer.sdp });
    } catch (e) {
      this.cb.onError('call.err.connect');
      this.hangup();
    }
  }

  private createPeerConnection() {
    if (this.pc) return;
    this.log('createPeerConnection iceServers=', JSON.stringify(this.iceServers));
    const pc = new RTCPeerConnection({ iceServers: this.iceServers });
    this.remoteStream = new MediaStream();
    this.cb.onRemoteStream(this.remoteStream);

    this.localStream?.getTracks().forEach(t => pc.addTrack(t, this.localStream!));

    pc.onicecandidate = (ev) => {
      if (ev.candidate) {
        this.log('local ICE candidate:', ev.candidate.type, ev.candidate.protocol, ev.candidate.address);
        this.sendWs('ice_candidate', {
          call_id: this.callId,
          candidate: ev.candidate.candidate,
          sdpMid: ev.candidate.sdpMid,
          sdpMLineIndex: ev.candidate.sdpMLineIndex,
        });
      } else {
        this.log('local ICE gathering complete');
      }
    };

    pc.ontrack = (ev) => {
      this.log('ontrack: got remote track', ev.track.kind);
      ev.streams[0]?.getTracks().forEach(t => this.remoteStream?.addTrack(t));
      this.cb.onRemoteStream(this.remoteStream);
    };

    pc.oniceconnectionstatechange = () => this.log('iceConnectionState=', pc.iceConnectionState);

    pc.onconnectionstatechange = () => {
      this.log('connectionState=', pc.connectionState);
      if (pc.connectionState === 'connected') {
        this.setPhase('connected');
      } else if (pc.connectionState === 'failed') {
        this.cb.onError('call.err.mediaFailed');
        this.hangup();
      }
    };

    this.pc = pc;
  }

  private async handleOffer(sdp: string) {
    if (!this.pc) this.createPeerConnection();
    await this.pc!.setRemoteDescription({ type: 'offer', sdp });
    this.remoteDescSet = true;
    await this.flushPendingIce();
    const answer = await this.pc!.createAnswer();
    await this.pc!.setLocalDescription(answer);
    this.sendWs('answer', { call_id: this.callId, sdp: answer.sdp });
  }

  private async handleAnswer(sdp: string) {
    if (!this.pc) return;
    await this.pc.setRemoteDescription({ type: 'answer', sdp });
    this.remoteDescSet = true;
    await this.flushPendingIce();
  }

  private async handleRemoteIce(data: Record<string, unknown>) {
    const cand: RTCIceCandidateInit = {
      candidate: String(data.candidate ?? ''),
      sdpMid: (data.sdpMid as string) ?? null,
      sdpMLineIndex: (data.sdpMLineIndex as number) ?? null,
    };
    if (!this.pc || !this.remoteDescSet) {
      this.pendingIce.push(cand);
      return;
    }
    try {
      await this.pc.addIceCandidate(cand);
    } catch {
      /* 忽略个别失败的候选 */
    }
  }

  private async flushPendingIce() {
    if (!this.pc) return;
    const pending = this.pendingIce;
    this.pendingIce = [];
    for (const c of pending) {
      try { await this.pc.addIceCandidate(c); } catch { /* ignore */ }
    }
  }

  // ==================== 内部：streaming ws ====================

  private async connectWs(): Promise<void> {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) return;
    await this.loadIceServers();
    return new Promise((resolve, reject) => {
      const url = `${this.opts.wsUrl}?token=${encodeURIComponent(this.opts.token)}`;
      const ws = new WebSocket(url);
      this.ws = ws;
      ws.onopen = () => resolve();
      ws.onerror = () => reject(new Error('signaling ws connect failed'));
      ws.onclose = () => { if (this.phase !== 'idle' && this.phase !== 'ended') { /* 由控制信令收尾 */ } };
      ws.onmessage = (e) => this.handleStreamingMsg(e);
    });
  }

  private handleStreamingMsg(e: MessageEvent) {
    let msg: StreamingMsg;
    try { msg = JSON.parse(e.data); } catch { return; }
    const data = msg.data || {};
    this.log('ws recv:', msg.type);
    switch (msg.type) {
      case 'call_signal':
        // 服务端经 streaming ws 直推的通话控制信令（accept/reject/cancel/end/timeout），
        // 稳定通道，不受 im ws 顶号影响；与 im ws call.signal 走同一处理。
        this.onControlSignal(data as unknown as CallSignal);
        break;
      case 'call_created':
        this.callId = (data.call_id as string) || msg.room_id || this.callId;
        break;
      case 'offer':
        this.handleOffer(String(data.sdp ?? ''));
        break;
      case 'answer':
        this.handleAnswer(String(data.sdp ?? ''));
        break;
      case 'ice_candidate':
        this.handleRemoteIce(data);
        break;
      case 'media_state':
        this.cb.onRemoteMedia({ audio: Boolean(data.audio), video: Boolean(data.video) });
        break;
      case 'call_reject':
        // 主叫忙线回执（后端 invite 时 callee 忙线）
        if ((data.reason as string) === 'busy') {
          this.cb.onError('call.err.busy');
          this.cleanup('busy');
        }
        break;
      case 'error':
        this.cb.onError('call.err.generic');
        break;
    }
  }

  private sendWs(type: string, data: Record<string, unknown>) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.log('ws send:', type);
      this.ws.send(JSON.stringify({ type, room_id: this.callId, data, timestamp: new Date().toISOString() }));
    } else {
      this.log('ws send SKIPPED (not open):', type, 'state=', this.ws?.readyState);
    }
  }

  private sendWsBestEffort(type: string, data: Record<string, unknown>) {
    try { this.sendWs(type, data); } catch { /* ignore */ }
  }

  private async loadIceServers() {
    try {
      // 加超时：ICE 配置拉取绝不能卡住建连主流程，失败/超时一律用默认 STUN
      const ctrl = new AbortController();
      const timer = setTimeout(() => ctrl.abort(), 2500);
      const res = await fetch(
        `${this.opts.httpBase}/v1/streaming/ice-servers?token=${encodeURIComponent(this.opts.token)}`,
        { signal: ctrl.signal },
      );
      clearTimeout(timer);
      if (res.ok) {
        const body = await res.json();
        if (Array.isArray(body?.iceServers) && body.iceServers.length) {
          this.iceServers = body.iceServers as RTCIceServer[];
        }
      }
    } catch {
      // 拉取失败/超时用默认 STUN（同机/同局域网靠 host 候选也能连）
      this.iceServers = DEFAULT_ICE;
    }
  }

  // ==================== 内部：媒体设备 ====================

  private async getLocalMedia(type: CallMediaType) {
    const stream = await navigator.mediaDevices.getUserMedia({
      audio: true,
      video: type === 'video' ? { width: 1280, height: 720 } : false,
    });
    this.localStream = stream;
    this.cb.onLocalStream(stream);
  }

  private isAudioOn() {
    return this.localStream?.getAudioTracks().some(t => t.enabled) ?? false;
  }

  private isVideoOn() {
    return this.localStream?.getVideoTracks().some(t => t.enabled) ?? false;
  }

  private mediaErr(e: unknown) {
    const name = (e as { name?: string })?.name;
    if (name === 'NotAllowedError' || name === 'NotFoundError') return 'call.err.media';
    return 'call.err.init';
  }

  // ==================== 内部：状态 / 清理 ====================

  private setPhase(p: CallPhase, info?: { reason?: string; duration?: number }) {
    this.log('phase ->', p, info?.reason ? `(${info.reason})` : '');
    this.phase = p;
    this.cb.onPhase(p, info);
  }

  private missedKey(): string {
    return this.mediaType === 'video' ? 'call.missed.video' : 'call.missed.voice';
  }

  private log(...args: unknown[]) {
    // 通话诊断日志，前缀便于在控制台过滤
    console.log('[call]', ...args);
  }

  private cleanup(reason?: string, duration?: number) {
    // 幂等：已是 idle（无活跃通话）则不重复收尾，避免重复投递通话记录
    if (this.phase === 'idle' && !this.callId) return;
    this.pc?.getSenders().forEach(s => s.track?.stop());
    this.localStream?.getTracks().forEach(t => t.stop());
    try { this.pc?.close(); } catch { /* ignore */ }
    try { this.ws?.close(); } catch { /* ignore */ }
    this.pc = null;
    this.ws = null;
    this.localStream = null;
    this.remoteStream = null;
    this.pendingIce = [];
    this.remoteDescSet = false;
    this.cb.onLocalStream(null);
    this.cb.onRemoteStream(null);

    this.setPhase('ended', { reason, duration });
    // 复位
    this.callId = '';
    this.peer = null;
    this.isCaller = false;
    this.phase = 'idle';
  }
}
