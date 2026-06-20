/**
 * PeerConn —— 封装「与某一个对端」的单条 WebRTC 连接（Mesh 群通话用）。
 * 群组 = N 条并行的 PeerConn；每条逻辑与 1:1 一致：offer/answer/ice + 一路远端流。
 * 信令通过 send 回调发出（带 to=peerId，由 streaming 定向 relay）。
 */

export type SendSignal = (type: string, data: Record<string, unknown>) => void;

export class PeerConn {
  private pc: RTCPeerConnection;
  readonly remoteStream = new MediaStream();
  private remoteDescSet = false;
  private pendingIce: RTCIceCandidateInit[] = [];

  constructor(
    readonly peerId: string,
    localStream: MediaStream | null,
    iceServers: RTCIceServer[],
    private send: SendSignal,
    private onRemoteStream: (uid: string, stream: MediaStream) => void,
    private onState: (uid: string, state: RTCPeerConnectionState) => void,
  ) {
    this.pc = new RTCPeerConnection({ iceServers });
    localStream?.getTracks().forEach(tr => this.pc.addTrack(tr, localStream));

    this.pc.onicecandidate = (ev) => {
      if (ev.candidate) {
        this.send('ice_candidate', {
          to: peerId,
          candidate: ev.candidate.candidate,
          sdpMid: ev.candidate.sdpMid,
          sdpMLineIndex: ev.candidate.sdpMLineIndex,
        });
      }
    };
    this.pc.ontrack = (ev) => {
      ev.streams[0]?.getTracks().forEach(tr => this.remoteStream.addTrack(tr));
      this.onRemoteStream(peerId, this.remoteStream);
    };
    this.pc.onconnectionstatechange = () => this.onState(peerId, this.pc.connectionState);
  }

  /** 作为 offerer（老人向新人）发起协商 */
  async makeOffer() {
    const offer = await this.pc.createOffer();
    await this.pc.setLocalDescription(offer);
    this.send('offer', { to: this.peerId, sdp: offer.sdp });
  }

  /** 收到对端 offer（新人答复老人）→ 回 answer */
  async handleOffer(sdp: string) {
    await this.pc.setRemoteDescription({ type: 'offer', sdp });
    this.remoteDescSet = true;
    await this.flushIce();
    const answer = await this.pc.createAnswer();
    await this.pc.setLocalDescription(answer);
    this.send('answer', { to: this.peerId, sdp: answer.sdp });
  }

  async handleAnswer(sdp: string) {
    await this.pc.setRemoteDescription({ type: 'answer', sdp });
    this.remoteDescSet = true;
    await this.flushIce();
  }

  async addRemoteIce(data: Record<string, unknown>) {
    const cand: RTCIceCandidateInit = {
      candidate: String(data.candidate ?? ''),
      sdpMid: (data.sdpMid as string) ?? null,
      sdpMLineIndex: (data.sdpMLineIndex as number) ?? null,
    };
    if (!this.remoteDescSet) { this.pendingIce.push(cand); return; }
    try { await this.pc.addIceCandidate(cand); } catch { /* 忽略个别失败候选 */ }
  }

  private async flushIce() {
    const pending = this.pendingIce;
    this.pendingIce = [];
    for (const c of pending) {
      try { await this.pc.addIceCandidate(c); } catch { /* ignore */ }
    }
  }

  close() {
    try { this.pc.close(); } catch { /* ignore */ }
  }
}
