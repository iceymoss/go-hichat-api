/**
 * 通话状态 store —— UI 与 CallEngine 的桥接
 *
 * 收呼叫控制信令：在 chat-store 的 ws.on('call.signal') 里调 useCallStore.getState().onSignal(sig)
 * 发动作 + 媒体：本 store 持有的 CallEngine（自建 streaming ws）
 */

import { create } from 'zustand';
import { CallEngine, type CallMediaType, type CallPhase, type CallPeer, type CallSignal } from './call-engine';
import { useIMStore } from './im-store';
import { useChatStore } from './chat-store';

/** 通话结束原因 -> 聊天记录状态；返回 null 表示不记录（忙线/失败/异常） */
function mapCallStatus(reason?: string): string | null {
  switch (reason) {
    case 'completed':
      return 'completed';
    case 'canceled':
      return 'canceled';
    case 'rejected':
      return 'rejected';
    case 'no_answer':
      return 'no_answer';
    default:
      return null;
  }
}

/** 计算 streaming 服务地址：开发直连 :10093，生产经反代 /streaming */
function streamingEndpoints(): { wsUrl: string; httpBase: string } {
  if (typeof window === 'undefined') return { wsUrl: 'ws://localhost:10093/ws', httpBase: 'http://localhost:10093' };
  const host = window.location.hostname;
  const isDev = host === 'localhost' || host === '127.0.0.1';
  if (isDev) {
    return { wsUrl: `ws://${host}:10093/ws`, httpBase: `http://${host}:10093` };
  }
  const wsProto = window.location.protocol === 'https:' ? 'wss' : 'ws';
  return {
    wsUrl: `${wsProto}://${window.location.host}/streaming/ws`,
    httpBase: `${window.location.protocol}//${window.location.host}/streaming`,
  };
}

interface CallState {
  phase: CallPhase;
  peer: CallPeer | null;
  mediaType: CallMediaType;
  localStream: MediaStream | null;
  remoteStream: MediaStream | null;
  muted: boolean;
  cameraOff: boolean;
  remoteMedia: { audio: boolean; video: boolean };
  startedAt: number | null; // 接通时间戳（毫秒），用于计时
  isCaller: boolean;        // 本端是否主叫（决定由谁投递通话记录）
  errorMsg: string | null;

  // actions
  startCall: (peer: CallPeer, type: CallMediaType) => void;
  accept: () => void;
  reject: () => void;
  hangup: () => void;
  toggleMute: () => void;
  toggleCamera: () => void;
  onSignal: (sig: CallSignal) => void;
  clearError: () => void;
}

let engine: CallEngine | null = null;

export const useCallStore = create<CallState>((set, get) => {
  // 用 store 自身的 set 接 engine 回调
  const callbacks = {
    onPhase: (phase: CallPhase, info?: { reason?: string; duration?: number }) => {
      const prev = get();
      if (phase === 'connected' && !prev.startedAt) set({ startedAt: Date.now() });

      // 通话结束：由主叫端投递一条通话记录到聊天（双方会话内展示，可回拨）
      if (phase === 'ended' && prev.isCaller && prev.peer?.id) {
        const status = mapCallStatus(info?.reason);
        if (status) {
          const duration = prev.startedAt ? Math.max(0, Math.floor((Date.now() - prev.startedAt) / 1000)) : 0;
          try {
            useChatStore.getState().sendCallRecord(prev.peer.id, prev.mediaType, status, duration);
          } catch { /* ignore */ }
        }
      }

      set({ phase });
      if (phase === 'idle' || phase === 'ended') {
        set({
          startedAt: null,
          muted: false,
          cameraOff: false,
          remoteMedia: { audio: true, video: true },
        });
      }
    },
    onLocalStream: (s: MediaStream | null) => set({ localStream: s }),
    onRemoteStream: (s: MediaStream | null) => set({ remoteStream: s }),
    onRemoteMedia: (state: { audio: boolean; video: boolean }) => set({ remoteMedia: state }),
    onError: (msg: string) => set({ errorMsg: msg }),
  };

  const ensureEngine = (): CallEngine | null => {
    const me = useIMStore.getState().currentUser;
    if (!me?.token || !me.id) return null;
    if (!engine) {
      const { wsUrl, httpBase } = streamingEndpoints();
      engine = new CallEngine({ wsUrl, httpBase, token: me.token, selfId: me.id }, callbacks);
    }
    return engine;
  };

  return {
    phase: 'idle',
    peer: null,
    mediaType: 'voice',
    localStream: null,
    remoteStream: null,
    muted: false,
    cameraOff: false,
    remoteMedia: { audio: true, video: true },
    startedAt: null,
    isCaller: false,
    errorMsg: null,

    startCall: (peer, type) => {
      const eng = ensureEngine();
      if (!eng) return;
      set({ peer, mediaType: type, isCaller: true, errorMsg: null });
      eng.startCall(peer, type);
    },

    accept: () => {
      get().clearError();
      engine?.accept();
    },

    reject: () => {
      engine?.reject();
    },

    hangup: () => {
      engine?.hangup();
    },

    toggleMute: () => {
      const next = !get().muted;
      set({ muted: next });
      engine?.toggleMute(next);
    },

    toggleCamera: () => {
      const next = !get().cameraOff;
      set({ cameraOff: next });
      engine?.toggleCamera(!next);
    },

    onSignal: (sig) => {
      const eng = ensureEngine();
      if (!eng) return;
      if (sig.event === 'invite') {
        // 来电：昵称/头像优先用本地好友资料解析（后端不再发 RPC 取资料）
        const fromUid = sig.fromUid || '';
        const friend = useIMStore.getState().friends?.find(f => f.friend_uid === fromUid || f.id === fromUid);
        const name = sig.fromName || friend?.remark || friend?.name || friend?.nickname || fromUid || '对方';
        const avatar = sig.fromAvatar || friend?.avatar;
        set({
          peer: { id: fromUid, name, avatar },
          mediaType: sig.callType || 'voice',
          isCaller: false,
          errorMsg: null,
        });
      }
      eng.onControlSignal(sig);
    },

    clearError: () => set({ errorMsg: null }),
  };
});
