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
  onActiveSpeakers: (uids: string[]) => void;
  onPeerEvent: (uid: string, kind: 'joined' | 'left') => void;
  onError: (key: string) => void;
}

interface StreamingMsg {
  type: string;
  room_id?: string;
  user_id?: string;
  data?: Record<string, unknown>;
}

interface CandidatePairStats extends RTCStats {
  nominated: boolean;
  state: string;
  localCandidateId: string;
  remoteCandidateId: string;
  packetsSent?: number;
  packetsReceived?: number;
  bytesSent?: number;
  bytesReceived?: number;
  currentRoundTripTime?: number;
}

interface CandidateStats extends RTCStats {
  candidateType?: string;
}

interface InboundVideoStats extends RTCStats {
  kind?: string;
  mediaType?: string;
  trackIdentifier?: string;
  codecId?: string;
  bytesReceived?: number;
  packetsReceived?: number;
  packetsLost?: number;
  jitter?: number;
  frameWidth?: number;
  frameHeight?: number;
  framesDecoded?: number;
  framesDropped?: number;
  framesPerSecond?: number;
  keyFramesDecoded?: number;
  freezeCount?: number;
}

interface CodecStats extends RTCStats {
  mimeType?: string;
}

const DEFAULT_ICE: RTCIceServer[] = [{ urls: 'stun:stun.l.google.com:19302' }];
const ICE_RESTART_DELAY_MS = 1500;
const ICE_RESTART_RETRY_MS = 5000;
const MAX_ICE_RESTARTS = 2;
const MAX_MEDIA_RECONNECTS = 3;
const MEDIA_RECONNECT_COOLDOWN_MS = 30000;
const VIDEO_STATS_INTERVAL_MS = 5000;
const VIDEO_STALL_THRESHOLD_MS = 10000;

/** Pion 的同一发布者音视频 ontrack 事件可能携带不同 MediaStream 实例，按 track 聚合后再交给 UI。 */
export function mergeRemoteStream(target: MediaStream, incoming: MediaStream, eventTrack: MediaStreamTrack) {
  const known = new Set(target.getTracks().map(track => track.id));
  for (const track of [...incoming.getTracks(), eventTrack]) {
    if (!known.has(track.id)) {
      target.addTrack(track);
      known.add(track.id);
    }
  }
}

/** Phase 0 统一使用软件可并发编码的 VP8，避免多标签页耗尽硬件 H264 encoder 后只输出约 1 FPS。 */
export function preferredVideoCodecs(codecs: RTCRtpCodec[]) {
  return codecs.filter(codec => codec.mimeType.toLowerCase() === 'video/vp8');
}

export function parseActiveSpeakers(data: Record<string, unknown>) {
  if (!Array.isArray(data.speakers)) return [];
  return [...new Set(data.speakers.filter((uid): uid is string => typeof uid === 'string' && uid.length > 0))];
}

export class SFUGroupEngine {
  private ws: WebSocket | null = null;
  private pc: RTCPeerConnection | null = null;
  private localStream: MediaStream | null = null;
  private remoteUids = new Set<string>();
  private remoteStreams = new Map<string, MediaStream>();
  private hadPeer = false;
  private joinTimer: ReturnType<typeof setTimeout> | null = null;
  private callId = '';
  private mediaType: CallMediaType = 'voice';
  private phase: CallPhase = 'idle';
  private iceServers: RTCIceServer[] = DEFAULT_ICE;
  private remoteDescSet = false;
  private pendingIce: RTCIceCandidateInit[] = [];
  private negotiation = Promise.resolve();
  private iceRestartTimer: ReturnType<typeof setTimeout> | null = null;
  private iceRestartAttempts = 0;
  private mediaReconnectAttempts = 0;
  private mediaReconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private videoStatsTimer: ReturnType<typeof setInterval> | null = null;
  private videoTrackOwners = new Map<string, string>();
  private remoteVideoEnabled = new Map<string, boolean>();
  private videoSnapshots = new Map<string, { framesDecoded: number; stalledSince: number | null; reportedStall: boolean }>();
  private readonly sessionId = typeof crypto !== 'undefined' && 'randomUUID' in crypto
    ? crypto.randomUUID()
    : `${Date.now()}-${Math.random().toString(36).slice(2)}`;

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
    this.diagnose('cleanup', { reason: 'hangup' });
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
    this.sendMediaState();
  }
  toggleCamera(on: boolean) {
    this.localStream?.getVideoTracks().forEach(tr => (tr.enabled = on));
    this.sendMediaState();
  }

  /** 本端当前开关麦/摄像头状态 */
  private myMediaState() {
    const audio = this.localStream?.getAudioTracks().some(t => t.enabled) ?? true;
    const video = this.localStream?.getVideoTracks().some(t => t.enabled) ?? false;
    return { audio, video };
  }
  /** 广播本端媒体状态给同通话其余参与者（经 streaming ws 控制面，SFU 不经手） */
  private sendMediaState() {
    this.sendWs('sfu_media_state', { call_id: this.callId, ...this.myMediaState() });
  }

  // ==================== streaming ws（控制面 + SFU 协商） ====================

  private async connectWs(): Promise<void> {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) return;
    await this.loadIceServers();
    return new Promise((resolve, reject) => {
      const url = `${this.opts.wsUrl}?token=${encodeURIComponent(this.opts.token)}`;
      const ws = new WebSocket(url);
      this.ws = ws;
      ws.onopen = () => {
        this.diagnose('ws_opened');
        resolve();
      };
      ws.onerror = () => {
        this.diagnose('ws_error');
        reject(new Error('group signaling ws connect failed'));
      };
      ws.onclose = (event) => {
        console.warn('[sfu-gcall] ws closed', event.code, event.reason);
      };
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
        this.diagnose('group_ready', {
          message_type: msg.type,
          participant_count: Array.isArray(data.participants) ? data.participants.length : 0,
        });
        this.publish().catch(() => this.cb.onError('call.err.connect'));
        break;
      case 'sfu_publish_answer':
        this.diagnose('sfu_answer_received', { sdp_bytes: String(data.sdp ?? '').length });
        this.queueNegotiation(async () => {
          if (!this.pc) return;
          await this.pc.setRemoteDescription({ type: 'answer', sdp: String(data.sdp ?? '') });
          this.remoteDescSet = true;
          await this.flushPendingIce();
        });
        break;
      case 'sfu_offer':
        // 服务端发起的 renegotiation（新增/移除下行轨）：应答回去
        this.diagnose('sfu_offer_received', { sdp_bytes: String(data.sdp ?? '').length });
        this.queueNegotiation(() => this.handleServerOffer(String(data.sdp ?? '')));
        break;
      case 'sfu_ice':
        this.diagnose('ice_candidate_received', this.candidateSummary(data));
        void this.handleRemoteIce(data);
        break;
      case 'sfu_reconnect_ready':
        this.diagnose('media_reconnect_ready', { reconnect_attempt: this.mediaReconnectAttempts });
        this.publish().catch((error) => {
          this.diagnose('negotiation_error', { error: this.errorSummary(error), stage: 'media_reconnect_publish' });
        });
        break;
      case 'sfu_peer_left': {
        const uid = data.uid as string;
        this.diagnose('peer_left', { peer_uid: uid || '', ended: data.ended === true });
        if (uid) this.onParticipantGone(uid);
        if (data.ended === true) this.cleanup();
        break;
      }
      case 'sfu_media_state': {
        const uid = data.uid as string;
        if (uid && uid !== this.opts.selfId) {
          const video = data.video === true;
          this.remoteVideoEnabled.set(uid, video);
          this.cb.onParticipantMedia(uid, { audio: data.audio !== false, video });
        }
        break;
      }
      case 'active_speakers':
        this.cb.onActiveSpeakers(parseActiveSpeakers(data));
        break;
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

  private diagnose(event: string, fields: Record<string, unknown> = {}) {
    if (this.ws?.readyState !== WebSocket.OPEN) return;
    this.ws.send(JSON.stringify({
      type: 'sfu_diagnostic',
      room_id: this.callId,
      data: { call_id: this.callId, session_id: this.sessionId, event, fields },
      timestamp: new Date().toISOString(),
    }));
  }

  // ==================== SFU PeerConnection ====================

  /** 建本端与 SFU 的 PeerConnection，发布本地轨并通过 sfu_ice trickle candidate。 */
  private async publish() {
    if (this.pc) return; // 已发布
    const pc = new RTCPeerConnection({ iceServers: this.iceServers });
    this.pc = pc;
    pc.ontrack = (ev) => this.onRemoteTrack(ev);
    pc.onicecandidate = (ev) => {
      if (!ev.candidate) return;
      this.diagnose('ice_candidate_sent', {
        candidate_type: ev.candidate.type,
        protocol: ev.candidate.protocol,
        tcp_type: ev.candidate.tcpType,
      });
      this.sendWs('sfu_ice', {
        call_id: this.callId,
        candidate: ev.candidate.candidate,
        sdpMid: ev.candidate.sdpMid,
        sdpMLineIndex: ev.candidate.sdpMLineIndex,
        usernameFragment: ev.candidate.usernameFragment,
      });
    };
    pc.oniceconnectionstatechange = () => {
      this.log('ice:', pc.iceConnectionState);
      this.diagnose('ice_state', { state: pc.iceConnectionState });
      if (pc.iceConnectionState === 'disconnected' || pc.iceConnectionState === 'failed') {
        void this.reportCandidatePair();
        this.scheduleIceRestart();
      }
    };
    pc.onicegatheringstatechange = () => {
      this.log('gather:', pc.iceGatheringState);
      this.diagnose('ice_gathering_state', { state: pc.iceGatheringState });
    };
    pc.onconnectionstatechange = () => {
      this.log('conn:', pc.connectionState);
      this.diagnose('peer_state', { state: pc.connectionState });
      if (pc.connectionState === 'connected') {
        this.clearIceRestartTimer();
        this.iceRestartAttempts = 0;
        this.setPhase('connected');
      }
      if (pc.connectionState === 'disconnected' || pc.connectionState === 'failed') {
        this.scheduleIceRestart();
      }
    };

    this.localStream?.getTracks().forEach(track => {
      const sender = pc.addTrack(track, this.localStream!);
      if (track.kind !== 'video') return;
      const codecs = RTCRtpSender.getCapabilities?.('video')?.codecs ?? [];
      const preferred = preferredVideoCodecs(codecs);
      if (preferred.length > 0) {
        pc.getTransceivers().find(transceiver => transceiver.sender === sender)?.setCodecPreferences(preferred);
      }
    });
    this.diagnose('publish_started', {
      audio_tracks: this.localStream?.getAudioTracks().length ?? 0,
      video_tracks: this.localStream?.getVideoTracks().length ?? 0,
    });
    if (this.mediaType === 'video') this.startVideoDiagnostics();

    const offer = await pc.createOffer();
    await pc.setLocalDescription(offer);
    this.sendWs('sfu_publish', { call_id: this.callId, sdp: pc.localDescription?.sdp ?? '' });
  }

  /** 处理服务端 renegotiation/ICE restart offer：setRemote -> answer -> setLocal -> 回 sfu_answer。 */
  private async handleServerOffer(sdp: string) {
    if (!this.pc || !sdp) return;
    await this.pc.setRemoteDescription({ type: 'offer', sdp });
    this.remoteDescSet = true;
    await this.flushPendingIce();
    const answer = await this.pc.createAnswer();
    await this.pc.setLocalDescription(answer);
    this.sendWs('sfu_answer', { call_id: this.callId, sdp: this.pc.localDescription?.sdp ?? '' });
  }

  private scheduleIceRestart() {
    if (this.iceRestartTimer) return;
    const delay = this.iceRestartAttempts === 0 ? ICE_RESTART_DELAY_MS : ICE_RESTART_RETRY_MS;
    this.iceRestartTimer = setTimeout(() => {
      this.iceRestartTimer = null;
      if (!this.pc || (this.pc.connectionState !== 'disconnected' && this.pc.connectionState !== 'failed')) return;
      if (this.iceRestartAttempts >= MAX_ICE_RESTARTS) {
        this.diagnose('ice_restart_exhausted', { attempts: this.iceRestartAttempts });
        this.rebuildMediaConnection();
        return;
      }
      this.iceRestartAttempts += 1;
      this.remoteDescSet = false;
      this.diagnose('ice_restart_requested', { attempt: this.iceRestartAttempts });
      this.sendWs('sfu_restart', { call_id: this.callId });
      this.scheduleIceRestart();
    }, delay);
  }

  private rebuildMediaConnection() {
    if (this.mediaReconnectAttempts >= MAX_MEDIA_RECONNECTS) {
      this.cb.onError('call.err.connect');
      this.diagnose('media_reconnect_paused', { reconnect_attempt: this.mediaReconnectAttempts });
      if (!this.mediaReconnectTimer) {
        this.mediaReconnectTimer = setTimeout(() => {
          this.mediaReconnectTimer = null;
          this.mediaReconnectAttempts = 0;
          this.rebuildMediaConnection();
        }, MEDIA_RECONNECT_COOLDOWN_MS);
      }
      return;
    }
    this.mediaReconnectAttempts += 1;
    this.clearIceRestartTimer();
    this.iceRestartAttempts = 0;
    try { this.pc?.close(); } catch { /* ignore */ }
    this.pc = null;
    this.remoteDescSet = false;
    this.pendingIce = [];
    this.negotiation = Promise.resolve();
    for (const uid of this.remoteUids) this.cb.onParticipantLeft(uid);
    this.remoteUids.clear();
    this.remoteStreams.clear();
    this.stopVideoDiagnostics();
    this.diagnose('media_reconnect_requested', { reconnect_attempt: this.mediaReconnectAttempts });
    this.sendWs('sfu_reconnect', { call_id: this.callId });
  }

  private async reportCandidatePair() {
    if (!this.pc) return;
    try {
      const report = await this.pc.getStats();
      let selected: CandidatePairStats | undefined;
      report.forEach(stat => {
        const pair = stat as CandidatePairStats;
        if (pair.type === 'candidate-pair' && pair.nominated && pair.state === 'succeeded') selected = pair;
      });
      if (!selected) return;
      const local = report.get(selected.localCandidateId) as CandidateStats | undefined;
      const remote = report.get(selected.remoteCandidateId) as CandidateStats | undefined;
      this.diagnose('candidate_pair', {
        local_candidate_type: local?.candidateType ?? 'unknown',
        remote_candidate_type: remote?.candidateType ?? 'unknown',
        packets_sent: selected.packetsSent ?? 0,
        packets_received: selected.packetsReceived ?? 0,
        bytes_sent: selected.bytesSent ?? 0,
        bytes_received: selected.bytesReceived ?? 0,
        rtt_ms: Math.round((selected.currentRoundTripTime ?? 0) * 1000),
      });
    } catch (error) {
      console.warn('[sfu-gcall] get selected candidate pair failed', error);
    }
  }

  private clearIceRestartTimer() {
    if (!this.iceRestartTimer) return;
    clearTimeout(this.iceRestartTimer);
    this.iceRestartTimer = null;
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
      this.sendMediaState(); // 让新出现的参与者立即知道我当前的开关麦/摄像头状态
    }
    let aggregate = this.remoteStreams.get(uid);
    if (!aggregate) {
      aggregate = new MediaStream();
      this.remoteStreams.set(uid, aggregate);
    }
    mergeRemoteStream(aggregate, stream, ev.track);
    this.cb.onParticipantStream(uid, aggregate);
    this.diagnose('remote_track', { peer_uid: uid, kind: ev.track.kind, track_state: ev.track.readyState });
    if (ev.track.kind === 'video') this.videoTrackOwners.set(ev.track.id, uid);
  }

  private startVideoDiagnostics() {
    if (this.videoStatsTimer) return;
    this.videoStatsTimer = setInterval(() => void this.reportVideoStats(), VIDEO_STATS_INTERVAL_MS);
  }

  private stopVideoDiagnostics() {
    if (this.videoStatsTimer) clearInterval(this.videoStatsTimer);
    this.videoStatsTimer = null;
    this.videoTrackOwners.clear();
    this.remoteVideoEnabled.clear();
    this.videoSnapshots.clear();
  }

  private async reportVideoStats() {
    if (!this.pc || this.pc.connectionState !== 'connected') return;
    try {
      const report = await this.pc.getStats();
      const now = Date.now();
      report.forEach(raw => {
        const stat = raw as InboundVideoStats;
        if (stat.type !== 'inbound-rtp' || (stat.kind ?? stat.mediaType) !== 'video') return;
        const peerUid = this.videoTrackOwners.get(stat.trackIdentifier ?? '') ?? 'unknown';
        const framesDecoded = stat.framesDecoded ?? 0;
        const previous = this.videoSnapshots.get(stat.id);
        let stalledSince = previous?.stalledSince ?? null;
        let reportedStall = previous?.reportedStall ?? false;
        if (this.remoteVideoEnabled.get(peerUid) === false || !previous || framesDecoded > previous.framesDecoded) {
          stalledSince = null;
          reportedStall = false;
        } else if (stalledSince === null) {
          stalledSince = now;
        }
        const stalledMs = stalledSince === null ? 0 : now - stalledSince;
        const codec = report.get(stat.codecId ?? '') as CodecStats | undefined;
        const fields = {
          peer_uid: peerUid,
          codec: codec?.mimeType ?? 'unknown',
          bytes_received: stat.bytesReceived ?? 0,
          packets_received: stat.packetsReceived ?? 0,
          packets_lost: stat.packetsLost ?? 0,
          jitter_ms: Math.round((stat.jitter ?? 0) * 1000),
          frame_width: stat.frameWidth ?? 0,
          frame_height: stat.frameHeight ?? 0,
          frames_decoded: framesDecoded,
          frames_dropped: stat.framesDropped ?? 0,
          frames_per_second: stat.framesPerSecond ?? 0,
          key_frames_decoded: stat.keyFramesDecoded ?? 0,
          freeze_count: stat.freezeCount ?? 0,
          stalled_ms: stalledMs,
        };
        this.diagnose('video_inbound_stats', fields);
        if (!reportedStall && stalledMs >= VIDEO_STALL_THRESHOLD_MS) {
          this.diagnose('video_stalled', fields);
          reportedStall = true;
        }
        this.videoSnapshots.set(stat.id, { framesDecoded, stalledSince, reportedStall });
      });
    } catch (error) {
      this.diagnose('video_inbound_stats', { error: this.errorSummary(error), stage: 'get_stats' });
    }
  }

  /** 某参与者离开（服务端 sfu_peer_left）：只移除 tile；通话是否结束由服务端权威状态决定。 */
  private onParticipantGone(uid: string) {
    if (!this.remoteUids.delete(uid)) return;
    this.remoteStreams.delete(uid);
    this.cb.onParticipantLeft(uid);
    this.cb.onPeerEvent(uid, 'left');
  }

  private queueNegotiation(task: () => Promise<void>) {
    this.negotiation = this.negotiation.then(task).catch((error) => {
      console.warn('[sfu-gcall] negotiation failed', error);
      this.diagnose('negotiation_error', { error: this.errorSummary(error), signaling_state: this.pc?.signalingState ?? 'none' });
      this.cb.onError('call.err.connect');
    });
  }

  private async handleRemoteIce(data: Record<string, unknown>) {
    const candidate: RTCIceCandidateInit = {
      candidate: String(data.candidate ?? ''),
      sdpMid: (data.sdpMid as string) ?? null,
      sdpMLineIndex: (data.sdpMLineIndex as number) ?? null,
      usernameFragment: (data.usernameFragment as string) ?? null,
    };
    if (!this.pc || !this.remoteDescSet) {
      this.pendingIce.push(candidate);
      return;
    }
    try {
      await this.pc.addIceCandidate(candidate);
    } catch (error) {
      console.warn('[sfu-gcall] add ICE candidate failed', candidate.usernameFragment, error);
      this.diagnose('ice_candidate_error', { error: this.errorSummary(error), stage: 'direct', ...this.candidateSummary(data) });
    }
  }

  private async flushPendingIce() {
    if (!this.pc) return;
    const pending = this.pendingIce;
    this.pendingIce = [];
    for (const candidate of pending) {
      try {
        await this.pc.addIceCandidate(candidate);
      } catch (error) {
        console.warn('[sfu-gcall] flush ICE candidate failed', candidate.usernameFragment, error);
        this.diagnose('ice_candidate_error', { error: this.errorSummary(error), stage: 'flush' });
      }
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
    console.log('[sfu-gcall]', ...args);
  }

  private candidateSummary(data: Record<string, unknown>) {
    try {
      const candidate = new RTCIceCandidate({
        candidate: String(data.candidate ?? ''),
        sdpMid: (data.sdpMid as string) ?? null,
        sdpMLineIndex: (data.sdpMLineIndex as number) ?? null,
      });
      return { candidate_type: candidate.type, protocol: candidate.protocol, tcp_type: candidate.tcpType };
    } catch {
      return { candidate_type: 'invalid' };
    }
  }

  private errorSummary(error: unknown) {
    if (error instanceof Error) return `${error.name}: ${error.message}`;
    return String(error);
  }

  private cleanup() {
    if (this.joinTimer) { clearTimeout(this.joinTimer); this.joinTimer = null; }
    this.clearIceRestartTimer();
    this.iceRestartAttempts = 0;
    this.mediaReconnectAttempts = 0;
    this.stopVideoDiagnostics();
    if (this.mediaReconnectTimer) {
      clearTimeout(this.mediaReconnectTimer);
      this.mediaReconnectTimer = null;
    }
    try { this.pc?.close(); } catch { /* ignore */ }
    this.pc = null;
    this.remoteDescSet = false;
    this.pendingIce = [];
    this.negotiation = Promise.resolve();
    this.remoteUids.clear();
    this.remoteStreams.clear();
    this.hadPeer = false;
    this.localStream?.getTracks().forEach(tr => tr.stop());
    this.diagnose('cleanup', { reason: 'engine_cleanup', phase: this.phase });
    try { this.ws?.close(); } catch { /* ignore */ }
    this.ws = null;
    this.localStream = null;
    this.cb.onLocalStream(null);
    this.setPhase('ended');
    this.callId = '';
    this.phase = 'idle';
  }
}
