/**
 * 通话状态 store —— UI 与 CallEngine 的桥接
 *
 * 收呼叫控制信令：在 chat-store 的 ws.on('call.signal') 里调 useCallStore.getState().onSignal(sig)
 * 发动作 + 媒体：本 store 持有的 CallEngine（自建 streaming ws）
 */

import { create } from 'zustand';
import { toast } from 'sonner';
import { CallEngine, type CallMediaType, type CallPhase, type CallPeer, type CallSignal } from './call-engine';
import { SFUGroupEngine } from './sfu-group-engine';
import { useIMStore } from './im-store';
import { useChatStore } from './chat-store';
import { t as translate } from './i18n';
import { useSettingsStore } from './settings-store';

/** 非组件上下文下的翻译（toast 等用） */
const tt = (key: string) => translate(key, useSettingsStore.getState().language);

/** 群通话参与者（每人一路远端流） */
export interface GroupParticipant {
  id: string;
  name: string;
  avatar?: string;
  stream: MediaStream | null;
  media: { audio: boolean; video: boolean }; // 对端开关麦/摄像头状态（实时同步）
}

/** 某群当前活跃通话（用于群聊顶部「通话中」横幅 + 会话列表标识 + 加入通话） */
export interface ActiveGroupCall {
  callId: string;
  callType: CallMediaType;
  participants: string[];
}

/** 按 uid 解析展示名/头像：群通话参与者可能不是好友，故优先查群成员资料，再查好友，最后回退 uid */
function resolveContact(uid: string): { name: string; avatar?: string } {
  // 1) 群成员资料（不要求互为好友）
  const groups = useChatStore.getState().groupMembers as Record<string, Array<{ user_id?: string; nickname?: string; group_nickname?: string; user_avatar_url?: string }>>;
  for (const gid in groups) {
    const m = groups[gid]?.find(x => x.user_id === uid);
    if (m) return { name: m.group_nickname || m.nickname || uid, avatar: m.user_avatar_url || undefined };
  }
  // 2) 好友资料
  const friend = useIMStore.getState().friends?.find(f => f.friend_uid === uid || f.id === uid);
  if (friend) return { name: friend.remark || friend.name || friend.nickname || uid, avatar: friend.avatar };
  // 3) 回退 uid
  return { name: uid };
}

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
  minimized: boolean;       // 通话窗口是否最小化为悬浮球
  isGroup: boolean;         // 是否群组通话（mesh）
  participants: Record<string, GroupParticipant>; // 群组：每个对端一路流
  activeSpeakerIds: string[];
  activeGroupCalls: Record<string, ActiveGroupCall>; // groupId -> 该群当前活跃通话（横幅/列表标识）
  errorMsg: string | null;

  // actions
  startCall: (peer: CallPeer, type: CallMediaType) => void;
  startGroupCall: (groupId: string, members: string[], type: CallMediaType, groupName?: string) => void;
  joinGroupCall: (groupId: string, callId: string, type: CallMediaType, groupName?: string) => void;
  fetchGroupCallState: (groupId: string) => void;
  accept: () => void;
  reject: () => void;
  hangup: () => void;
  toggleMute: () => void;
  toggleCamera: () => void;
  setMinimized: (v: boolean) => void;
  setVisibleGroupVideos: (uids: string[]) => void;
  onSignal: (sig: CallSignal) => void;
  clearError: () => void;
}

let engine: CallEngine | null = null;
let groupEngine: SFUGroupEngine | null = null;

export const useCallStore = create<CallState>((set, get) => {
  // 用 store 自身的 set 接 engine 回调
  const callbacks = {
    onPhase: (phase: CallPhase, info?: { reason?: string; duration?: number }) => {
      const prev = get();
      if (phase === 'connected' && !prev.startedAt) set({ startedAt: Date.now() });

      // 通话结束：由主叫端投递一条通话记录到聊天（双方会话内展示，可回拨）
      if (phase === 'ended' && prev.isCaller && prev.peer?.id) {
        const status = mapCallStatus(info?.reason);
        console.log('[call] record check: reason=', info?.reason, 'status=', status, 'peer=', prev.peer?.id);
        if (status) {
          const duration = prev.startedAt ? Math.max(0, Math.floor((Date.now() - prev.startedAt) / 1000)) : 0;
          try {
            useChatStore.getState().sendCallRecord(prev.peer.id, prev.mediaType, status, duration);
            console.log('[call] record sent:', status, duration);
          } catch (e) { console.warn('[call] record failed', e); }
        }
      }

      set({ phase });
      if (phase === 'idle' || phase === 'ended') {
        set({
          startedAt: null,
          muted: false,
          cameraOff: false,
          minimized: false,
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

  // 群组（mesh）引擎回调
  const groupCallbacks = {
    onPhase: (phase: CallPhase) => {
      if (phase === 'connected' && !get().startedAt) set({ startedAt: Date.now() });

      // 群通话结束：由发起人往群聊投一条通话记录（接通过才记）
      if (phase === 'ended') {
        const prev = get();
        if (prev.isCaller && prev.peer?.id && prev.startedAt) {
          const duration = Math.max(0, Math.floor((Date.now() - prev.startedAt) / 1000));
          try {
            useChatStore.getState().sendGroupCallRecord(prev.peer.id, prev.mediaType, 'completed', duration);
          } catch { /* ignore */ }
        }
      }

      set({ phase, isGroup: true });
      if (phase === 'idle' || phase === 'ended') {
        set({
          startedAt: null, muted: false, cameraOff: false, minimized: false,
          isGroup: false, participants: {}, activeSpeakerIds: [],
        });
      }
    },
    onLocalStream: (s: MediaStream | null) => set({ localStream: s }),
    onParticipantStream: (uid: string, stream: MediaStream | null) => {
      set(state => {
        const c = resolveContact(uid);
        const prev = state.participants[uid];
        return { participants: { ...state.participants, [uid]: { id: uid, name: c.name, avatar: c.avatar, stream, media: prev?.media ?? { audio: true, video: true } } } };
      });
    },
    onParticipantMedia: (uid: string, media: { audio: boolean; video: boolean }) => {
      set(state => {
        const c = resolveContact(uid);
        const prev = state.participants[uid];
        return { participants: { ...state.participants, [uid]: { id: uid, name: prev?.name ?? c.name, avatar: prev?.avatar ?? c.avatar, stream: prev?.stream ?? null, media } } };
      });
    },
    onParticipantLeft: (uid: string) => {
      set(state => {
        const next = { ...state.participants };
        delete next[uid];
        return { participants: next, activeSpeakerIds: state.activeSpeakerIds.filter(id => id !== uid) };
      });
    },
    onActiveSpeakers: (activeSpeakerIds: string[]) => set({ activeSpeakerIds }),
    onRoster: (uids: string[]) => {
      set(state => {
        const next: Record<string, GroupParticipant> = {};
        for (const uid of uids) {
          const previous = state.participants[uid];
          const contact = resolveContact(uid);
          next[uid] = previous ?? { id: uid, name: contact.name, avatar: contact.avatar, stream: null, media: { audio: true, video: true } };
        }
        return { participants: next, activeSpeakerIds: state.activeSpeakerIds.filter(uid => uids.includes(uid)) };
      });
    },
    onPeerEvent: (uid: string, kind: 'joined' | 'left') => {
      const c = resolveContact(uid);
      toast(`${c.name} ${kind === 'joined' ? tt('call.toast.joined') : tt('call.toast.left')}`);
    },
    onError: (key: string) => set({ errorMsg: key }),
  };

  const ensureGroupEngine = (): SFUGroupEngine | null => {
    const me = useIMStore.getState().currentUser;
    if (!me?.token || !me.id) return null;
    if (!groupEngine) {
      const { wsUrl, httpBase } = streamingEndpoints();
      groupEngine = new SFUGroupEngine({ wsUrl, httpBase, token: me.token, selfId: me.id }, groupCallbacks);
    }
    return groupEngine;
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
    minimized: false,
    isGroup: false,
    participants: {},
    activeSpeakerIds: [],
    activeGroupCalls: {},
    errorMsg: null,

    startCall: (peer, type) => {
      const eng = ensureEngine();
      if (!eng) return;
      set({ peer, mediaType: type, isCaller: true, isGroup: false, minimized: false, errorMsg: null });
      eng.startCall(peer, type);
    },

    startGroupCall: (groupId, members, type, groupName) => {
      const eng = ensureGroupEngine();
      if (!eng) return;
      set({
        isGroup: true, isCaller: true, mediaType: type, minimized: false, errorMsg: null,
        participants: {}, activeSpeakerIds: [], peer: { id: groupId, name: groupName || '群通话' },
      });
      eng.startGroupCall(groupId, members, type);
    },

    // 主动加入一通正在进行的群通话（点「加入通话」）。非发起人，不投通话记录。
    joinGroupCall: (groupId, callId, type, groupName) => {
      if (get().phase !== 'idle' && get().phase !== 'ended') return; // 已在通话中
      const eng = ensureGroupEngine();
      if (!eng) return;
      set({
        isGroup: true, isCaller: false, mediaType: type, minimized: false, errorMsg: null,
        participants: {}, activeSpeakerIds: [], peer: { id: groupId, name: groupName || '群通话' },
      });
      eng.joinExisting(callId, type);
    },

    // 进群聊时补拉该群当前通话状态（错过 group.state 广播时）。
    fetchGroupCallState: (groupId) => {
      const me = useIMStore.getState().currentUser;
      if (!me?.token || !groupId) return;
      const { httpBase } = streamingEndpoints();
      fetch(`${httpBase}/v1/streaming/group-call?group_id=${encodeURIComponent(groupId)}&token=${encodeURIComponent(me.token)}`)
        .then(r => (r.ok ? r.json() : null))
        .then((body: { active?: boolean; callId?: string; callType?: string; participants?: string[] } | null) => {
          if (!body) return;
          set(state => {
            const next = { ...state.activeGroupCalls };
            if (body.active && body.callId) {
              next[groupId] = { callId: body.callId, callType: (body.callType as CallMediaType) || 'voice', participants: body.participants || [] };
            } else {
              delete next[groupId];
            }
            return { activeGroupCalls: next };
          });
        })
        .catch(() => { /* ignore */ });
    },

    accept: () => {
      get().clearError();
      if (get().isGroup) groupEngine?.acceptGroup();
      else engine?.accept();
    },

    reject: () => {
      if (get().isGroup) groupEngine?.reject();
      else engine?.reject();
    },

    hangup: () => {
      if (get().isGroup) groupEngine?.hangup();
      else engine?.hangup();
    },

    toggleMute: () => {
      const next = !get().muted;
      set({ muted: next });
      if (get().isGroup) groupEngine?.toggleMute(next);
      else engine?.toggleMute(next);
    },

    toggleCamera: () => {
      const next = !get().cameraOff;
      set({ cameraOff: next });
      if (get().isGroup) groupEngine?.toggleCamera(!next);
      else engine?.toggleCamera(!next);
    },

    setMinimized: (v) => set({ minimized: v }),

    setVisibleGroupVideos: (uids) => groupEngine?.setVisibleVideoParticipants(uids),

    onSignal: (sig) => {
      // 群通话活跃状态广播（全体群成员收）：更新横幅/列表标识，不弹通话 UI
      if (sig.event === 'group.state') {
        const gid = sig.groupId || '';
        if (!gid) return;
        set(state => {
          const next = { ...state.activeGroupCalls };
          const parts = sig.members || [];
          if (parts.length > 0 && sig.callId) {
            next[gid] = { callId: sig.callId, callType: sig.callType || 'voice', participants: parts };
          } else {
            delete next[gid];
          }
          return { activeGroupCalls: next };
        });
        return;
      }

      // 群通话来电
      if (sig.event === 'group.invite') {
        const eng = ensureGroupEngine();
        if (!eng) return;
        const initiator = resolveContact(sig.fromUid || '');
        set({
          isGroup: true, isCaller: false, mediaType: sig.callType || 'voice',
          minimized: false, errorMsg: null, participants: {}, activeSpeakerIds: [],
          peer: { id: sig.groupId || '', name: initiator.name, avatar: initiator.avatar },
        });
        eng.setIncoming(sig.callId, sig.callType || 'voice');
        return;
      }

      // 1:1
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
          isGroup: false,
          minimized: false,
          errorMsg: null,
        });
      }
      eng.onControlSignal(sig);
    },

    clearError: () => set({ errorMsg: null }),
  };
});
