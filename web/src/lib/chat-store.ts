/**
 * Chat Store — 管理会话列表、消息、WebSocket 生命周期
 *
 * 替代 mock-data 中的 conversations / conversationMessagesMap，
 * 通过 IM REST API 获取数据 + WebSocket 实时推送。
 */

import { create } from 'zustand';
import { createLatestRequest } from './latest-request';
import {
  createIMWs,
  getIMWs,
  ChatType,
  MsgType,
  ContentType,
  type WsChatData,
} from './ws-client';

import {
  getChatLog,
  getConversations,
  setupConversation,
  updateConversations,
  setConversationSettings as apiSetConversationSettings,
  recallMsg,
  getAtMeMessages,
  type ChatLogItem,
  type ConversationItem,
  type BackendUser,
} from './api-client';
import { useIMStore } from './im-store';
import { playMessageSound, vibrate } from './notification';
import { useSettingsStore } from './settings-store';
import { mediaPreview } from './media-message';
import { useCallStore } from './call-store';
import type { CallSignal } from './call-engine';
import type { Message, Conversation } from './types';
import { toast } from 'sonner';
import { sendFriendRequest } from './friend-group-api';

const conversationsRequest = createLatestRequest();

// 私聊被后端鉴权拦截（对方已删好友）时，顶部弹"重新添加好友"通知。
// 与红感叹号（消息标 failed）并行：感叹号是消息级反馈，这条是关系级引导。
// 文案硬编码中文，与本模块其余 toast（AddFriendPanel/GroupList 等）保持一致。
function notifyFriendBlocked(peerId: string) {
  toast('你们已不是好友，是否重新添加好友', {
    id: `friend-block-${peerId}`, // 稳定 id：连发多条只弹一个，不堆叠
    duration: 6000,
    action: {
      label: '重新添加',
      onClick: async () => {
        const token = useIMStore.getState().currentUser?.token;
        if (!token) return;
        const ok = await sendFriendRequest(token, peerId);
        if (ok) toast.success('好友请求已发送');
        else toast.error('发送失败，请重试');
      },
    },
  });
}

// ========== 消息类型映射 ==========

const backMsgTypeMap: Record<number, Message['type']> = {
  [MsgType.Text]: 'text',
  [MsgType.File]: 'file',
  [MsgType.Voice]: 'voice',
  [MsgType.Image]: 'image',
  [MsgType.Memes]: 'memes',
  [MsgType.Video]: 'video',
  [MsgType.Call]: 'call',
};

const frontMsgTypeMap: Record<string, number> = {
  text: MsgType.Text,
  image: MsgType.Image,
  voice: MsgType.Voice,
  file: MsgType.File,
  video: MsgType.Video,
  memes: MsgType.Memes,
  call: MsgType.Call,
};

/** 解析引用消息 JSON（{id,uid,name,preview,mType,thumb}）为 Message.replyTo */
function quoteToReplyTo(quote?: string): Message['replyTo'] {
  if (!quote) return undefined;
  try {
    const q = JSON.parse(quote);
    if (q && (q.name || q.preview)) {
      return { senderName: q.name || '', content: q.preview || '', msgId: q.id, senderId: q.uid, mType: q.mType, thumbUrl: q.thumb };
    }
  } catch { /* ignore */ }
  return undefined;
}

// ========== Store 接口 ==========

/** 用户资料缓存 */
interface UserProfile {
  id: string;
  nickname: string;
  avatar: string;
}

export interface GroupMember {
  user_id: string;
  nickname: string;
  user_avatar_url: string;
  role_level: number; // 与后端 GroupRoleLevel 一致：0=普通成员 1=管理员 2=群主
  group_nickname: string;
}

interface ChatState {
  // 连接状态
  wsState: 'disconnected' | 'connecting' | 'connected';

  // 会话列表 (从 API 获取 + 实时更新)
  conversations: Conversation[];

  // 消息 (按 conversationId 分组)
  messagesMap: Record<string, Message[]>;

  // 用户资料缓存 (userId → profile)
  userProfiles: Record<string, UserProfile>;

  // 群成员缓存 (groupId → members[])
  groupMembers: Record<string, GroupMember[]>;

  // 加载状态
  loadingConversations: boolean;
  loadingMessages: Record<string, boolean>;

  // Actions
  initWs: (token: string, userId: string, wsUrl?: string) => void;
  destroyWs: () => void;
  resetAuthState: () => void;
  fetchConversations: (token: string) => Promise<void>;
  fetchMessages: (token: string, conversationId: string, oldestMsgId?: string) => Promise<void>;
  /** 哪些会话处于"浏览历史"态（跳转到了非最新窗口） */
  anchoredConvs: Record<string, boolean>;
  /** 跳转到某条消息的上下文窗口（替换当前列表），返回是否命中目标 */
  jumpToContext: (token: string, conversationId: string, msgId: string) => Promise<boolean>;
  /** 向下增量加载更新的消息（浏览历史态用），返回是否已到最新 */
  fetchNewer: (token: string, conversationId: string) => Promise<boolean>;
  /** 回到最新页（退出浏览历史态） */
  backToLatest: (token: string, conversationId: string) => Promise<void>;  sendMessage: (token: string, userId: string, conversationId: string, content: string, msgType?: string, quote?: string, mentions?: { atUsers?: string[]; atAll?: boolean }) => void;
  /** 通话结束后由主叫端投递一条通话记录消息（mType=call），双方会话内展示，可点击回拨 */
  sendCallRecord: (peerId: string, callType: 'voice' | 'video', status: string, duration: number) => void;
  /** 群通话结束由发起人投递一条群聊通话记录 */
  sendGroupCallRecord: (groupId: string, callType: 'voice' | 'video', status: string, duration: number) => void;
  resendMessage: (token: string, userId: string, conversationId: string, msgId: string) => void;
  markRead: (userId: string, conversationId: string, msgIds: string[]) => void;
  /** 撤回消息：调后端校验，成功后原位置为撤回态（ws 事件会同步其它端） */
  recallMessage: (token: string, conversationId: string, msgId: string) => Promise<void>;
  getOrCreateConversation: (token: string, userId: string, targetId: string) => Promise<Conversation>;
  deleteConversation: (token: string, conversationId: string) => void;
  setConversationSettings: (token: string, conversationId: string, settings: { pinned?: boolean; muted?: boolean }) => Promise<void>;
  clearUnread: (conversationId: string) => void;
  /** 各会话中"@我且未读"的消息 id 列表（按时间升序），供"有人@我"横幅逐条跳转 */
  atMeMap: Record<string, string[]>;
  /** 拉取某会话 @我未读消息列表（进会话时调用，先于 markRead 生效以避免被标已读清空） */
  fetchAtMe: (token: string, conversationId: string) => Promise<void>;
  /** 消费（移除）一条已跳转的 @我消息 */
  consumeAtMe: (conversationId: string, msgId: string) => void;
  /** 清空某会话的 @我列表（关闭横幅 / 离开会话） */
  clearAtMe: (conversationId: string) => void;
  fetchGroupMembers: (token: string, groupId: string) => Promise<void>;
  ensureUserProfiles: (token: string, userIds: string[]) => void;
  /** 已播放语音消息 id 集合（控制未读红点），本地持久化 */
  playedVoices: Record<string, true>;
  markVoicePlayed: (msgId: string) => void;
  /** 已失效的会话（被踢/退群/解散/删好友），值含事件类型，前端据此禁用输入 */
  disabledConversations: Record<string, { eventType: string; operatorId?: string }>;
  /** 标记某群会话为"已被移出/解散"：禁用输入框 + 插入系统消息（按稳定 id 去重）。relation.changed 与打开会话成员校验共用。 */
  markGroupRemoved: (conversationId: string, eventType: string) => void;
}

export const useChatStore = create<ChatState>()((set, get) => ({
  wsState: 'disconnected',
  groupMembers: {},
  conversations: [],
  messagesMap: {},
  userProfiles: {},
  loadingConversations: false,
  loadingMessages: {},
  resetAuthState: () => {
    conversationsRequest.invalidate();
    get().destroyWs();
    set({ wsState: 'disconnected', conversations: [], messagesMap: {}, userProfiles: {}, groupMembers: {}, loadingConversations: false, loadingMessages: {}, anchoredConvs: {}, atMeMap: {}, disabledConversations: {} });
  },
  disabledConversations: {},
  playedVoices: (() => {
    if (typeof window === 'undefined') return {};
    try { return JSON.parse(localStorage.getItem('hichat_played_voices') || '{}'); }
    catch { return {}; }
  })(),

  markVoicePlayed: (msgId: string) => {
    if (get().playedVoices[msgId]) return;
    set(s => {
      const next = { ...s.playedVoices, [msgId]: true as const };
      if (typeof window !== 'undefined') {
        try { localStorage.setItem('hichat_played_voices', JSON.stringify(next)); } catch { /* ignore */ }
      }
      return { playedVoices: next };
    });
  },

  // ==================== WebSocket 生命周期 ====================

  initWs: (token, userId, wsUrl) => {
    // 开发环境直连 WS 服务 (10090)，生产环境通过 Caddy/Nginx 代理
    const host = typeof window !== 'undefined' ? window.location.hostname : '';
    const isDev = host === 'localhost' || host === '127.0.0.1';
    const defaultWsUrl = isDev
      ? `ws://${host}:10090/ws`
      : `${window?.location?.protocol === 'https:' ? 'wss' : 'ws'}://${window?.location?.host}/ws`;

    const ws = createIMWs({
      url: wsUrl || defaultWsUrl,
      token,
      onStateChange: (state) => {
        set({ wsState: state });
        if (state === 'connected') {
          const imStore = useIMStore.getState();
          void imStore.refreshSocialRequestUnread();
          void imStore.refreshNotificationUnread();
          imStore.bumpNotificationVersion();
          imStore.invalidateFriendRequests();
          imStore.invalidateGroupRequests();
          imStore.invalidateFriends();
          imStore.invalidateGroups();
          void get().fetchConversations(token);
        }
      },
      onError: (err, msgId) => {
        // 业务错误帧（发送被鉴权拦截）：ACK 已先 resolve 了 send promise，故在此按消息 id 把对应消息标记失败（红感叹号）。
        if (msgId) {
          // 先定位消息所在会话（只读），再标失败 + 私聊场景弹"重新添加好友"
          const state = get();
          let foundCid: string | undefined;
          for (const cid of Object.keys(state.messagesMap)) {
            if (state.messagesMap[cid].some(m => m.id === msgId)) { foundCid = cid; break; }
          }
          if (!foundCid) return;

          set(s => {
            const arr = s.messagesMap[foundCid!];
            const idx = arr.findIndex(m => m.id === msgId);
            if (idx < 0) return {};
            const copy = arr.slice();
            copy[idx] = { ...copy[idx], status: 'failed' };
            return { messagesMap: { ...s.messagesMap, [foundCid!]: copy } };
          });

          // 仅私聊被拦才引导重加好友；群被移出已有横幅 + 系统消息闭环，跳过。
          const conv = state.conversations.find(c => c.id === foundCid);
          if (conv?.type === 'private') {
            const peerId = foundCid.split('_').find(p => p !== userId);
            if (peerId) notifyFriendBlocked(peerId);
          }
          return;
        }
        console.warn('[ChatStore] ws error:', err);
      },
    });

    // 服务端推送消息 — push.go NewMessage 不设 method，所以 method 为 ""
    ws.on('', (data, raw) => {
      const chat = data as WsChatData | null;
      if (!chat?.conversationId) return;
      handlePush(chat, raw?.id, userId);
    });

    ws.on('chat.ping', () => { /* pong */ });

    // 动态消息通知（赞/评论/回复/@）：实时 +1 未读并 bump 版本，供 MomentsFeed 刷新列表
    ws.on('trend.notify', () => {
      const imStore = useIMStore.getState();
      imStore.setMomentsUnreadCount(imStore.momentsUnreadCount + 1);
      imStore.bumpTrendNotifyVersion();
    });

    // 关系变更：好友删除走隐式（不禁用输入框，发送时红感叹号闭环）；仅群事件（被踢/解散）显式禁用 + 通知
    ws.on('relation.changed', (data) => {
      const evt = data as { conversationId?: string; eventType?: string; operatorId?: string } | null;
      if (!evt?.eventType) return;
      const imStore = useIMStore.getState();
      if (evt.eventType === 'friend.added' || evt.eventType === 'friend.deleted') imStore.invalidateFriends();
      if (evt.eventType.startsWith('group.')) {
        imStore.invalidateGroups();
        void get().fetchConversations(token);
      }
      if (evt.conversationId && (evt.eventType === 'group.member.removed' || evt.eventType === 'group.disbanded')) {
        get().markGroupRemoved(evt.conversationId, evt.eventType);
      }
    });

    // 音视频通话控制信令：来电/接听/拒接/挂断/超时 -> 通话 store 驱动来电与通话界面
    ws.on('call.signal', (data) => {
      const sig = data as CallSignal | null;
      if (!sig?.event) return;
      useCallStore.getState().onSignal(sig);
    });

    // 公共通知（好友/群申请等）：按 notifyType 分发 —— 实时红点 + 气泡提示（点击跳到对应入口）。
    // 文案硬编码中文，与本模块其余 toast 保持一致；历史列表/已读由通知中心走 REST 拉取。
    ws.on('notify', (data) => {
      const n = data as { notifyType?: string; bizId?: string } | null;
      if (!n?.notifyType) return;
      const imStore = useIMStore.getState();
      const friendNotification = ['friend.apply', 'friend.accept', 'friend.reject'].includes(n.notifyType);
      const groupNotification = ['group.apply', 'group.accept', 'group.reject', 'group.invalidated', 'group.invite'].includes(n.notifyType);
      if (friendNotification) {
        imStore.invalidateFriendRequests();
        void imStore.refreshFriendRequestUnread();
      }
      if (groupNotification) {
        imStore.invalidateGroupRequests();
        void imStore.refreshGroupRequestUnread();
      }
      if (n.notifyType === 'friend.accept') imStore.invalidateFriends();
      if (n.notifyType === 'group.accept') {
        imStore.invalidateGroups();
        void get().fetchConversations(token);
      }
      // 点击气泡跳到来源 + 子 tab（好友→新的朋友；群→群申请）
      const go = { label: '查看', onClick: () => useIMStore.getState().navigateToNotificationSource(n.notifyType!, n.bizId) };
      switch (n.notifyType) {
        case 'friend.apply':
          toast('有人申请添加你为好友', { action: go });
          break;
        case 'friend.accept':
          toast.success('对方通过了你的好友申请', { action: go });
          break;
        case 'friend.reject':
          toast('对方拒绝了你的好友申请', { action: go });
          break;
        case 'group.apply':
          toast('有人申请加入你管理的群聊', { action: go });
          break;
        case 'group.accept':
          toast.success('你的入群申请已通过', { action: go });
          break;
        case 'group.reject':
          toast('你的入群申请被拒绝', { action: go });
          break;
        case 'group.invalidated':
          toast('入群申请已失效', { action: go });
          break;
        case 'group.invite':
          toast('你收到了一条群聊邀请', { action: go });
          break;
        case 'group.removed':
          toast('你已被移出群聊');
          break;
        case 'group.admin.set':
          toast.success('你已被设为群管理员');
          break;
        case 'group.admin.unset':
          toast('你已被取消群管理员');
          break;
        case 'group.owner.transferred':
          toast.success('你已成为新群主');
          break;
      }
      // 公共通知未读以 REST 为真相，避免 reconnect GET 与本地 +1 互相覆盖。
      void imStore.refreshNotificationUnread();
      imStore.bumpNotificationVersion();
    });

    ws.connect();

    // 定时刷新好友在线状态（每 30 秒）
    if (typeof window !== 'undefined') {
      const onlineTimer = setInterval(async () => {
        const currentUser = useIMStore.getState().currentUser;
        if (!currentUser?.token) return;
        try {
          const res = await fetch('/api/social/friends/online', {
            headers: { Authorization: `Bearer ${currentUser.token}` },
          });
          const d = await res.json();
          if (d.success && d.data?.onLineList) {
            const onlineMap: Record<string, boolean> = d.data.onLineList;
            set(s => ({
              conversations: s.conversations.map(c => {
                if (c.type !== 'private') return c;
                const parts = c.id.split('_');
                const peerId = parts.find(p => p !== currentUser.id);
                return { ...c, online: peerId ? !!onlineMap[peerId] : false };
              }),
            }));
          }
        } catch { /* silent */ }
      }, 30000);
      // Store timer for cleanup
      (window as any).__hc_online_timer = onlineTimer;
    }
  },

  destroyWs: () => {
    getIMWs().disconnect();
    set({ wsState: 'disconnected' });
    // Clear online polling timer
    if (typeof window !== 'undefined' && (window as any).__hc_online_timer) {
      clearInterval((window as any).__hc_online_timer);
      delete (window as any).__hc_online_timer;
    }
  },

  // ==================== REST API ====================

  fetchConversations: async (token) => {
    const isLatest = conversationsRequest.begin();
    set({ loadingConversations: true });
    try {
      const resp = await getConversations(token);
      const map = resp?.conversationList || {};
      if (!isLatest() || useIMStore.getState().currentUser?.token !== token) return;
      // 后端已附带 targetName/targetAvatar，mapConversation 直接使用
      const list: Conversation[] = Object.values(map).map(mapConversation);
      set({ conversations: list });

      const currentUserId = useIMStore.getState().currentUser?.id;
      if (currentUserId) {
        // 查询好友在线状态
        try {
          const onlineRes = await fetch('/api/social/friends/online', {
            headers: { Authorization: `Bearer ${token}` },
          });
          const onlineData = await onlineRes.json();
          if (!isLatest() || useIMStore.getState().currentUser?.token !== token) return;
          if (onlineData.success && onlineData.data?.onLineList) {
            const onlineMap: Record<string, boolean> = onlineData.data.onLineList;
            set(s => ({
              conversations: s.conversations.map(c => {
                if (c.type !== 'private') return c;
                const parts = c.id.split('_');
                const peerId = parts.find(p => p !== currentUserId);
                return { ...c, online: peerId ? !!onlineMap[peerId] : false };
              }),
            }));
          }
        } catch { /* silent */ }

        // 预加载群成员
        const groupConvs = list.filter(c => c.type === 'group');
        if (!isLatest() || useIMStore.getState().currentUser?.token !== token) return;
        for (const gc of groupConvs) {
          get().fetchGroupMembers(token, gc.id);
        }

        // 预加载私聊对方的用户资料（昵称/头像），保证会话标题、引用归属、回复名稳定可解析
        const peerIds = new Set<string>();
        for (const c of list) {
          if (c.type !== 'private') continue;
          const peerId = c.id.split('_').find(p => p !== currentUserId);
          if (peerId && !get().userProfiles[peerId]) peerIds.add(peerId);
        }
        if (peerIds.size > 0) resolveUserProfiles(token, Array.from(peerIds));
      }
    } catch (e) {
      console.error('[ChatStore] fetch conversations error:', e);
    } finally {
      if (isLatest()) set({ loadingConversations: false });
    }
  },

  fetchMessages: async (token, conversationId, oldestMsgId) => {
    set(s => ({ loadingMessages: { ...s.loadingMessages, [conversationId]: true } }));
    try {
      const resp = await getChatLog(token, conversationId, oldestMsgId || '', 30);
      // API 返回倒序（最新在前），前端需要正序（最早在前）
      const list = (resp?.list || []).map(mapChatLog).reverse();

      set(s => {
        const existing = s.messagesMap[conversationId] || [];
        // 加载更早的历史：放在前面；首次加载：替换
        const merged = oldestMsgId ? [...list, ...existing] : list;
        // 全量加载（无游标）= 回到最新，清除浏览历史态
        const anchoredConvs = oldestMsgId ? s.anchoredConvs : { ...s.anchoredConvs, [conversationId]: false };
        return { messagesMap: { ...s.messagesMap, [conversationId]: merged }, anchoredConvs };
      });
    } catch (e) {
      console.error('[ChatStore] fetch messages error:', e);
    } finally {
      set(s => ({ loadingMessages: { ...s.loadingMessages, [conversationId]: false } }));
    }
  },

  anchoredConvs: {},

  jumpToContext: async (token, conversationId, msgId) => {
    try {
      // around：目标消息 + 其前若干条（含目标），API 返回倒序 → reverse 成正序
      const resp = await getChatLog(token, conversationId, msgId, 30, 'around');
      const list = (resp?.list || []).map(mapChatLog).reverse();
      const hit = list.some(m => m.id === msgId);
      if (!hit) return false;
      set(s => ({
        messagesMap: { ...s.messagesMap, [conversationId]: list },
        anchoredConvs: { ...s.anchoredConvs, [conversationId]: true },
      }));
      return true;
    } catch (e) {
      console.error('[ChatStore] jumpToContext error:', e);
      return false;
    }
  },

  fetchNewer: async (token, conversationId) => {
    const existing = get().messagesMap[conversationId] || [];
    const newest = existing[existing.length - 1];
    if (!newest) return true;
    try {
      // newer：严格晚于游标的若干条，API 已按时间升序返回 → 不 reverse，直接追加
      const resp = await getChatLog(token, conversationId, newest.id, 30, 'newer');
      const list = (resp?.list || []).map(mapChatLog);
      // 去重（防止边界重复）
      const seen = new Set(existing.map(m => m.id));
      const fresh = list.filter(m => !seen.has(m.id));
      const reachedLatest = list.length < 30; // 不足一页 → 已到最新
      set(s => ({
        messagesMap: { ...s.messagesMap, [conversationId]: [...(s.messagesMap[conversationId] || []), ...fresh] },
        anchoredConvs: reachedLatest
          ? { ...s.anchoredConvs, [conversationId]: false }
          : s.anchoredConvs,
      }));
      return reachedLatest;
    } catch (e) {
      console.error('[ChatStore] fetchNewer error:', e);
      return false;
    }
  },

  backToLatest: async (token, conversationId) => {
    set(s => ({ anchoredConvs: { ...s.anchoredConvs, [conversationId]: false } }));
    await get().fetchMessages(token, conversationId);
  },

  atMeMap: {},

  fetchAtMe: async (token, conversationId) => {
    try {
      const resp = await getAtMeMessages(token, conversationId);
      const ids = (resp?.list || []).map(m => m.id).filter(Boolean) as string[];
      set(s => ({ atMeMap: { ...s.atMeMap, [conversationId]: ids } }));
    } catch (e) {
      console.error('[ChatStore] fetchAtMe error:', e);
    }
  },

  consumeAtMe: (conversationId, msgId) => {
    set(s => {
      const cur = s.atMeMap[conversationId];
      if (!cur || cur.length === 0) return s;
      const next = cur.filter(id => id !== msgId);
      return { atMeMap: { ...s.atMeMap, [conversationId]: next } };
    });
  },

  clearAtMe: (conversationId) => {
    set(s => {
      if (!s.atMeMap[conversationId]?.length) return s;
      return { atMeMap: { ...s.atMeMap, [conversationId]: [] } };
    });
  },

  // ==================== 发送消息 ====================

  sendMessage: (token, userId, conversationId, content, msgType = 'text', quote, mentions) => {
    const conv = get().conversations.find(c => c.id === conversationId);
    const chatType = conv?.type === 'group' ? ChatType.Group : ChatType.Single;

    // @ 仅群聊生效
    const atUsers = chatType === ChatType.Group ? (mentions?.atUsers || undefined) : undefined;
    const atAll = chatType === ChatType.Group ? !!mentions?.atAll : false;

    // 解析接收者 ID
    let recvId = '';
    if (chatType === ChatType.Single) {
      const parts = conversationId.split('_');
      recvId = parts.find(p => p !== userId) || parts[0] || '';
    } else {
      recvId = conversationId;
    }

    // 乐观更新 UI
    const localMsgId = `local_${Date.now()}`;
    const localMsg: Message = {
      id: localMsgId,
      senderId: userId,
      content,
      timestamp: new Date(),
      type: (msgType as Message['type']) || 'text',
      status: 'sending',
      replyTo: quoteToReplyTo(quote),
      atUsers,
      atAll,
    };

    set(s => {
      const msgs = [...(s.messagesMap[conversationId] || []), localMsg];
      const convs = s.conversations.map(c =>
        c.id === conversationId
          ? { ...c, lastMessage: mediaPreview(localMsg.type, content), lastMessageTime: new Date() }
          : c,
      );
      return { messagesMap: { ...s.messagesMap, [conversationId]: msgs }, conversations: convs };
    });

    // 通过 WebSocket 发送 (RigorAck)
    const ws = getIMWs();
    const wsData: WsChatData = {
      conversationId,
      chatType,
      sendId: userId,
      recvId,
      sendTime: Date.now() * 1e6,
      msg: {
        mType: frontMsgTypeMap[msgType] || MsgType.Text,
        content,
        quote,
        readRecords: {},
        atUsers,
        atAll: atAll || undefined,
      },
    };

    const updateMsgStatus = (status: Message['status']) => {
      set(s => {
        const msgs = (s.messagesMap[conversationId] || []).map(m =>
          m.id === localMsgId ? { ...m, status } : m,
        );
        return { messagesMap: { ...s.messagesMap, [conversationId]: msgs } };
      });
    };

    ws.send('chat.user', wsData, localMsgId)
      .then(() => updateMsgStatus('sent'))
      .catch(err => {
        // 被鉴权闸门拦截 / 超时等：标记失败（红感叹号）即可，不 console.error 以免开发模式错误浮层
        console.warn('[ChatStore] send failed:', err);
        updateMsgStatus('failed');
      });
  },

  // ==================== 通话记录 ====================

  sendCallRecord: (peerId, callType, status, duration) => {
    const me = useIMStore.getState().currentUser;
    if (!me?.token || !me.id) return;
    // 私聊会话 id：优先用已存在的会话，否则按 uid 排序拼（与后端一致的稳定顺序）
    const existing = get().conversations.find(
      c => c.type === 'private' && c.id.split('_').includes(peerId) && c.id.split('_').includes(me.id),
    );
    const conversationId = existing?.id || [me.id, peerId].sort().join('_');
    const content = JSON.stringify({ callType, status, duration });
    get().sendMessage(me.token, me.id, conversationId, content, 'call');
  },

  sendGroupCallRecord: (groupId, callType, status, duration) => {
    const me = useIMStore.getState().currentUser;
    if (!me?.token || !me.id) return;
    const content = JSON.stringify({ callType, status, duration, scope: 'group' });
    // 群聊会话 id 即 groupId
    get().sendMessage(me.token, me.id, groupId, content, 'call');
  },

  // ==================== 重发失败消息 ====================

  resendMessage: (token, userId, conversationId, msgId) => {
    const msgs = get().messagesMap[conversationId] || [];
    const failedMsg = msgs.find(m => m.id === msgId && m.status === 'failed');
    if (!failedMsg) return;
    // 删除旧消息，重新发送
    set(s => ({
      messagesMap: {
        ...s.messagesMap,
        [conversationId]: (s.messagesMap[conversationId] || []).filter(m => m.id !== msgId),
      },
    }));
    get().sendMessage(token, userId, conversationId, failedMsg.content, failedMsg.type);
  },

  // ==================== 标记已读 ====================

  markRead: (userId, conversationId, msgIds) => {
    // Filter out local/push IDs that aren't real MongoDB ObjectIDs
    msgIds = msgIds.filter(id => !id.startsWith('local_') && !id.startsWith('push_'));
    if (!msgIds.length) return;
    const conv = get().conversations.find(c => c.id === conversationId);
    const chatType = conv?.type === 'group' ? ChatType.Group : ChatType.Single;
    let recvId = '';
    if (chatType === ChatType.Single) {
      const parts = conversationId.split('_');
      recvId = parts.find(p => p !== userId) || '';
    } else {
      recvId = conversationId;
    }

    const readRecords: Record<string, string> = {};
    msgIds.forEach(id => { readRecords[id] = '1'; });

    getIMWs().send('chat.markChat', {
      chatType, recvId, conversationId, sendId: userId, msgIds, readRecords,
    }).catch(err => console.error('[ChatStore] markRead failed:', err));
  },

  recallMessage: async (token, conversationId, msgId) => {
    const conv = get().conversations.find(c => c.id === conversationId);
    const chatType = conv?.type === 'group' ? ChatType.Group : ChatType.Single;
    const userId = useIMStore.getState().currentUser?.id;
    // 调后端校验+撤回；成功后乐观置态（ws 撤回事件会同步本人其它端及会话各端）
    await recallMsg(token, conversationId, msgId, chatType);
    applyRecall(conversationId, msgId, userId);
  },

  // ==================== 建立会话 ====================

  getOrCreateConversation: async (token, userId, targetId) => {
    // 生成会话 ID (与后端 wuid.CombineId 一致: 按数值从小到大排序)
    // Go 端用 strconv.ParseUint(id, 0, 64) 做数值比较
    const parseId = (s: string) => {
      try { return BigInt(s); } catch { return BigInt(0); }
    };
    const ids = [userId, targetId].sort((a, b) => {
      const na = parseId(a), nb = parseId(b);
      return na < nb ? -1 : na > nb ? 1 : 0;
    });
    const convId = `${ids[0]}_${ids[1]}`;

    const existing = get().conversations.find(c => c.id === convId);
    if (existing) return existing;

    try {
      await setupConversation(token, userId, targetId, ChatType.Single);
    } catch { /* 会话可能已存在 */ }

    // 查询对方用户资料
    if (!get().userProfiles[targetId]) {
      await resolveUserProfiles(token, [targetId]);
    }
    const profile = get().userProfiles[targetId];

    const newConv: Conversation = {
      id: convId,
      type: 'private',
      name: profile?.nickname || targetId,
      avatar: profile?.avatar || '',
      lastMessage: '',
      lastMessageTime: new Date(),
      unreadCount: 0,
      pinned: false,
      muted: false,
    };
    set(s => ({ conversations: [newConv, ...s.conversations] }));
    return newConv;
  },

  // ==================== 删除会话 ====================

  deleteConversation: (token, conversationId) => {
    // 先从 UI 移除（乐观更新）
    // 先读取会话类型，再从 state 中删除
    const conv = get().conversations.find(c => c.id === conversationId);
    const chatType = conv?.type === 'group' ? ChatType.Group : ChatType.Single;

    set(s => ({
      conversations: s.conversations.filter(c => c.id !== conversationId),
      messagesMap: Object.fromEntries(
        Object.entries(s.messagesMap).filter(([k]) => k !== conversationId)
      ),
    }));

    // 调后端 PUT /v1/im/conversation 设置 isShow=false
    updateConversations(token, {
      [conversationId]: {
        conversationId,
        chatType,
        isShow: false,
      } as any,
    }).catch(err => console.error('[ChatStore] deleteConversation failed:', err));
  },

  // ==================== 置顶 / 免打扰 ====================

  setConversationSettings: async (token, conversationId, settings) => {
    const prev = get().conversations;
    const target = prev.find(c => c.id === conversationId);
    if (!target) return;
    const nextPinned = settings.pinned ?? target.pinned;
    const nextMuted = settings.muted ?? target.muted;

    // 乐观更新
    set({
      conversations: prev.map(c =>
        c.id === conversationId ? { ...c, pinned: nextPinned, muted: nextMuted } : c
      ),
    });

    try {
      await apiSetConversationSettings(token, conversationId, nextPinned, nextMuted);
    } catch (e) {
      console.error('[ChatStore] setConversationSettings failed:', e);
      set({ conversations: prev }); // 失败回滚
    }
  },

  // ==================== 清除未读 ====================

  clearUnread: (conversationId) => {
    const conv = get().conversations.find(c => c.id === conversationId);
    const unreadCount = conv?.unreadCount || 0;

    // 1. 清除前端未读计数（含 @我 / 未接来电 角标）
    set(s => ({
      conversations: s.conversations.map(c =>
        c.id === conversationId ? { ...c, unreadCount: 0, hasAtMe: false, hasMissedCall: false } : c,
      ),
    }));

    // 2. 同步后端：PUT conversation 更新已读数
    if (unreadCount > 0) {
      const token = useIMStore.getState().currentUser?.token;
      if (token) {
        const chatType = conv?.type === 'group' ? ChatType.Group : ChatType.Single;
        updateConversations(token, {
          [conversationId]: {
            conversationId,
            chatType,
            isShow: true,
            read: unreadCount,
          } as any,
        }).catch(err => console.error('[ChatStore] clearUnread sync failed:', err));
      }
    }
  },

  fetchGroupMembers: async (token, groupId) => {
    if (get().groupMembers[groupId]?.length) return; // already cached
    try {
      const res = await fetch(`/api/social/group/users?group_id=${groupId}`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      const d = await res.json();
      if (useIMStore.getState().currentUser?.token !== token) return;
      if (d.success && d.data?.List) {
        set(s => ({
          groupMembers: { ...s.groupMembers, [groupId]: d.data.List },
        }));
        // Also cache user profiles from group members
        const profiles: Record<string, UserProfile> = {};
        for (const m of d.data.List) {
          profiles[m.user_id] = {
            id: m.user_id,
            nickname: m.group_nickname || m.nickname || m.user_id,
            avatar: m.user_avatar_url || '',
          };
        }
        set(s => ({ userProfiles: { ...s.userProfiles, ...profiles } }));
      }
    } catch (e) {
      console.error('[ChatStore] fetchGroupMembers error:', e);
    }
  },

  ensureUserProfiles: (token, userIds) => {
    const missing = userIds.filter(id => id && !get().userProfiles[id]);
    if (missing.length > 0) resolveUserProfiles(token, missing);
  },

  markGroupRemoved: (conversationId, eventType) => {
    set(s => {
      const sysId = `system_removed_${conversationId}`;
      const existing = s.messagesMap[conversationId] || [];
      const next: Partial<ChatState> = {
        disabledConversations: {
          ...s.disabledConversations,
          [conversationId]: { eventType },
        },
      };
      // 插入"你已被移出群聊 / 该群聊已解散"系统消息（稳定 id 去重，避免实时帧 + 打开校验重复插）
      if (!existing.some(m => m.id === sysId)) {
        const sysMsg: Message = {
          id: sysId,
          senderId: 'system',
          content: eventType === 'group.disbanded' ? '该群聊已解散' : '你已被移出群聊',
          timestamp: new Date(),
          type: 'system',
        };
        next.messagesMap = { ...s.messagesMap, [conversationId]: [...existing, sysMsg] };
      }
      return next;
    });
  },
}));

if (typeof window !== 'undefined') {
  window.addEventListener('hichat:auth-reset', () => useChatStore.getState().resetAuthState());
}

// ========== 内部辅助函数 ==========

/**
 * 智能解析后端时间戳 — 自动判断纳秒/微秒/毫秒/秒
 * 后端 conversation.go 用 UnixNano (纳秒)，但 MongoDB 存储后可能被截断为毫秒
 */
/** 计算未读数，防止后端 seq 数据异常（如存了时间戳）导致天文数字 */
function calcUnread(seq: number | undefined, read: number | undefined): number {
  // 后端 read 字段实际是 toRead（未读消息数），直接使用
  const unread = read || 0;
  if (unread < 0 || unread > 10000) return 0;
  return unread;
}

function parseTimestamp(ts: number): Date {
  if (!ts) return new Date();
  if (ts > 1e18) return new Date(ts / 1e6);  // 纳秒 → 毫秒
  if (ts > 1e15) return new Date(ts / 1e3);  // 微秒 → 毫秒
  if (ts > 1e12) return new Date(ts);         // 毫秒
  return new Date(ts * 1000);                 // 秒
}

function mapConversation(raw: ConversationItem): Conversation {
  const msg = raw.message;
  const unread = calcUnread(raw.seq, raw.read);
  // 历史未接来电：最后一条是对方发来的通话记录、状态为超时/取消、且未读 → 红标
  const me = useIMStore.getState().currentUser?.id;
  const missedCall = !!msg && msg.msgType === MsgType.Call && msg.sendId !== me && unread > 0 && (() => {
    try {
      const c = JSON.parse(msg.msgContent) as { status?: string };
      return c.status === 'no_answer' || c.status === 'canceled';
    } catch { return false; }
  })();
  return {
    id: raw.conversationId,
    type: raw.chatType === ChatType.Group ? 'group' : 'private',
    name: (raw as any).targetName || raw.conversationId,
    avatar: (raw as any).targetAvatar || '',
    lastMessage: msg ? mediaPreview(backMsgTypeMap[msg.msgType] || 'text', msg.msgContent) : '',
    lastMessageTime: msg?.sendTime ? parseTimestamp(msg.sendTime) : new Date(),
    unreadCount: unread,
    pinned: !!raw.isTop,
    muted: !!raw.isMute,
    members: (raw as any).memberCount || undefined,
    hasAtMe: !!(raw as any).hasAtMe,
    hasMissedCall: missedCall,
  };
}

function mapChatLog(log: ChatLogItem): Message {
  const base: Message = {
    id: log.id,
    senderId: log.sendId,
    content: log.msgContent,
    timestamp: log.sendTime ? parseTimestamp(log.sendTime) : new Date(),
    type: backMsgTypeMap[log.msgType] || 'text',
    replyTo: quoteToReplyTo(log.quote),
    recalled: log.status === 1,
    recalledBy: log.recalledBy,
    atUsers: log.atUsers,
    atAll: log.atAll,
  };
  // 从 REST 返回的 readRecords 恢复已读状态。
  // 服务端语义：
  // - 私聊：初始写库时 ReadRecords 为 256 字节 bitmap（仅 sender 的 hash 位置），
  //   接收方读后会被 msg_read_transfer 覆盖为 []byte{1}（base64="AQ=="）。
  //   所以只有 readRecords === "AQ==" 才代表对方已读。
  // - 群聊：ReadRecords 一直是 256 字节 bitmap，发送者在 addChatLog 时就 Set，
  //   其他读者陆续 Set 自己的位。前端用"总 bit 数 - 1"估算已读人数（扣除自己）。
  if (log.readRecords) {
    if (log.chatType === ChatType.Group) {
      const bits = countBitsSet(log.readRecords);
      const readCount = Math.max(0, bits - 1);
      if (readCount > 0) {
        base.readCount = readCount;
        base.isRead = true;
      }
    } else if (log.readRecords === 'AQ==') {
      base.isRead = true;
    }
  }
  return consumePendingReceipt(base);
}

/**
 * 批量查询用户资料并写入 userProfiles 缓存
 */
async function resolveUserProfiles(token: string, userIds: string[]) {
  if (userIds.length === 0) return;
  try {
    const resp = await fetch(`/api/user/search?ids=${userIds.join(',')}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    const json = await resp.json();
    if (useIMStore.getState().currentUser?.token !== token) return;
    const users: BackendUser[] = json.data?.users || [];
    const profiles: Record<string, UserProfile> = {};
    for (const u of users) {
      profiles[u.id] = {
        id: u.id,
        nickname: u.nickname || u.id,
        avatar: u.avatar || '',
      };
    }
    useChatStore.setState(s => ({
      userProfiles: { ...s.userProfiles, ...profiles },
    }));
  } catch (e) {
    console.error('[ChatStore] resolveUserProfiles error:', e);
  }
}

/** 从缓存获取用户昵称，无则返回 userId */
function getProfileName(userId: string): string {
  return useChatStore.getState().userProfiles[userId]?.nickname || userId;
}

function getProfileAvatar(userId: string): string {
  return useChatStore.getState().userProfiles[userId]?.avatar || '';
}

/**
 * 乱序回执缓存：回执（mType=6）可能先于其引用的消息到达。
 * 按 msgId 暂存 readRecords，稍后收到对应消息或拉取历史记录时合并。
 */
const pendingReadReceipts: Record<string, { readCount: number; isRead: boolean }> = {};

/**
 * 把某条 local_<timestamp> 占位消息替换为真实的 MongoDB ObjectID。
 * 策略：在该会话中找内容一致、id 为 local_ 前缀、senderId=自己、且 status!=='failed' 的最近一条，替换 id。
 * 若同内容消息有多条，按"最后一条"匹配——符合连续发相同内容的常见场景。
 */
function reconcileLocalMessageId(convId: string, content: string, realId: string) {
  useChatStore.setState(s => {
    const list = s.messagesMap[convId];
    if (!list || list.length === 0) return s;
    // 反向找最近一条 local_ 占位消息，内容一致则替换
    for (let i = list.length - 1; i >= 0; i--) {
      const m = list[i];
      if (!m.id?.startsWith('local_')) continue;
      if (m.content !== content) continue;
      if (m.status === 'failed') continue;
      const updated: Message = { ...m, id: realId };
      const nextList = [...list];
      nextList[i] = consumePendingReceipt(updated);
      return { messagesMap: { ...s.messagesMap, [convId]: nextList } };
    }
    return s;
  });
}

/** 解析 base64 bitmap 里的 set 位数，用于群聊「X/N 人已读」 */
function countBitsSet(base64Str: string): number {
  if (!base64Str) return 0;
  try {
    const bin = atob(base64Str);
    let count = 0;
    for (let i = 0; i < bin.length; i++) {
      let v = bin.charCodeAt(i);
      while (v) { v &= v - 1; count++; }
    }
    return count;
  } catch { return 0; }
}

/** 将回执里的 readRecords 合并进某条消息 */
function applyReceiptToMsg(m: Message, rec: string, isGroup: boolean): Message {
  if (isGroup) {
    // 服务端 bitmap 里 sender 自己的位置也被 addChatLog 置为 1（表示"发送人自动已读"），
    // 但 UI 上的 X/N 指的是"除自己外的已读人数"，所以要减去 1
    const readCount = Math.max(m.readCount || 0, Math.max(0, countBitsSet(rec) - 1));
    return { ...m, readCount, isRead: readCount > 0 };
  }
  return { ...m, isRead: true };
}

/** 处理已读回执：更新消息；未匹配到的缓存起来等消息稍后到达 */
function handleReadReceipt(chat: WsChatData) {
  const convId = chat.conversationId;
  const readRecords = chat.msg?.readRecords || {};
  const isGroup = chat.chatType === ChatType.Group;
  const msgIds = Object.keys(readRecords);
  if (msgIds.length === 0) return;

  useChatStore.setState(s => {
    const existing = s.messagesMap[convId];
    if (!existing) {
      // 整个会话都还没加载，全部缓存；群聊的 bitmap 里 sender 自己那一位也算 1，所以要减 1
      for (const id of msgIds) {
        const count = isGroup ? Math.max(0, countBitsSet(readRecords[id]) - 1) : 1;
        const prev = pendingReadReceipts[id];
        pendingReadReceipts[id] = {
          readCount: Math.max(prev?.readCount || 0, count),
          isRead: true,
        };
      }
      return s;
    }
    const matchedIds = new Set<string>();
    const nextList = existing.map(m => {
      const rec = readRecords[m.id];
      if (!rec) return m;
      matchedIds.add(m.id);
      return applyReceiptToMsg(m, rec, isGroup);
    });
    // 未匹配的 msgId 先缓存；群聊减去 sender 自己那一位
    for (const id of msgIds) {
      if (matchedIds.has(id)) continue;
      const count = isGroup ? Math.max(0, countBitsSet(readRecords[id]) - 1) : 1;
      const prev = pendingReadReceipts[id];
      pendingReadReceipts[id] = {
        readCount: Math.max(prev?.readCount || 0, count),
        isRead: true,
      };
    }
    return { messagesMap: { ...s.messagesMap, [convId]: nextList } };
  });
}

/** 消息落本地时消费缓存里的回执 */
function consumePendingReceipt(m: Message): Message {
  const pend = pendingReadReceipts[m.id];
  if (!pend) return m;
  delete pendingReadReceipts[m.id];
  return { ...m, isRead: pend.isRead, readCount: pend.readCount };
}

/** 把某条消息原位置为撤回态（撤回事件 / 本人撤回乐观更新共用） */
function applyRecall(convId: string, msgId: string | undefined, recalledBy?: string) {
  if (!msgId) return;
  useChatStore.setState(s => {
    const list = s.messagesMap[convId];
    if (!list) return {};
    let changed = false;
    const msgs = list.map(m => {
      if (m.id === msgId && !m.recalled) {
        changed = true;
        return { ...m, recalled: true, recalledBy };
      }
      return m;
    });
    if (!changed) return {};
    // 同步会话列表的最后一条预览（若被撤回的就是最新一条）
    const convs = s.conversations.map(c =>
      c.id === convId && msgs.length > 0 && msgs[msgs.length - 1].id === msgId
        ? { ...c, lastMessage: '撤回了一条消息' }
        : c,
    );
    return { messagesMap: { ...s.messagesMap, [convId]: msgs }, conversations: convs };
  });
}

/** 处理服务端推送的消息 */
function handlePush(chat: WsChatData, rawId: string | undefined, currentUserId: string) {
  // 先拦截已读回执：不落为一条新消息，只去更新既有消息的 isRead
  if (chat.msg?.mType === MsgType.ContentMakeRead) {
    handleReadReceipt(chat);
    return;
  }

  // 发送方回响：服务端写库后把消息带着真实 MsgId 回推给发送方，
  // 我们用它把前端之前插入的 local_<timestamp> 占位记录升级为真实 mongoID。
  // 忽略全零 ObjectID（异常数据，避免 key 冲突）。
  if (chat.contentType === ContentType.MsgAck && chat.sendId === currentUserId
      && rawId && !/^0+$/.test(rawId)) {
    reconcileLocalMessageId(chat.conversationId, chat.msg?.content || '', rawId);
    return;
  }
  if (chat.contentType === ContentType.MsgAck) {
    // 回响来了但 ID 异常，静默丢弃
    return;
  }

  // 撤回事件：把对应消息原位置为撤回态，并刷新会话最后一条预览
  if (chat.contentType === ContentType.Recall) {
    applyRecall(chat.conversationId, rawId, chat.recalledBy);
    return;
  }

  const convId = chat.conversationId;
  const baseMsg: Message = {
    // 服务端在 pusher 把 MongoDB MsgId 写入 WS Message.Id，优先用它
    id: rawId || `push_${Date.now()}`,
    senderId: chat.sendId,
    content: chat.msg?.content || '',
    timestamp: chat.sendTime ? parseTimestamp(chat.sendTime) : new Date(),
    type: backMsgTypeMap[chat.msg?.mType] || 'text',
    replyTo: quoteToReplyTo(chat.msg?.quote),
    atUsers: chat.msg?.atUsers,
    atAll: chat.msg?.atAll,
  };
  // 消费可能先到的乱序回执
  const msg = consumePendingReceipt(baseMsg);

  // 这条群消息是否 @了我（别人发的 + @所有人 或 @列表含我）
  const atMe = chat.sendId !== currentUserId && chat.chatType === ChatType.Group &&
    (!!chat.msg?.atAll || (chat.msg?.atUsers || []).includes(currentUserId));

  // 通话记录分类：对方发来的通话记录，按状态区分会话列表未读表现
  const fromPeer = chat.sendId !== currentUserId;
  const callStatus = msg.type === 'call' ? (() => {
    try { return (JSON.parse(msg.content) as { status?: string }).status; } catch { return undefined; }
  })() : undefined;
  // 未接来电（超时/被取消）：你没接到 → 计未读 + 像被@一样红标提示
  const missedCall = fromPeer && (callStatus === 'no_answer' || callStatus === 'canceled');
  // 已参与的通话（已接通/已拒接）：不该在会话列表当成未读新消息冒红点（结束后标记已读）
  const seenCall = fromPeer && msg.type === 'call' && (callStatus === 'completed' || callStatus === 'rejected');

  useChatStore.setState(s => {
    // 添加消息（若正在浏览历史窗口，则不追加到该窗口，避免新消息与旧上下文错误相邻；
    // 回到最新页时会从服务端重新拉取）
    const anchored = s.anchoredConvs[convId];
    const msgs = anchored ? (s.messagesMap[convId] || []) : [...(s.messagesMap[convId] || []), msg];

    // 更新或创建会话
    let convs = [...s.conversations];
    const idx = convs.findIndex(c => c.id === convId);
    if (idx >= 0) {
      convs[idx] = {
        ...convs[idx],
        lastMessage: mediaPreview(msg.type, msg.content),
        lastMessageTime: msg.timestamp,
        hasAtMe: convs[idx].hasAtMe || atMe,
        hasMissedCall: convs[idx].hasMissedCall || missedCall,
        unreadCount: convs[idx].unreadCount + (chat.sendId !== currentUserId ? 1 : 0),
      };
    } else {
      // 新会话
      const isGroup = chat.chatType === ChatType.Group;
      let name = convId;
      let avatar = '';
      if (!isGroup) {
        // 私聊：用对方资料
        const peerId = convId.split('_').find(p => p !== currentUserId) || '';
        name = peerId ? getProfileName(peerId) : convId;
        avatar = peerId ? getProfileAvatar(peerId) : '';
      }
      // 群聊先用 convId 占位，后续异步回填群名
      convs = [{
        id: convId,
        type: isGroup ? 'group' : 'private',
        name,
        avatar,
        lastMessage: mediaPreview(msg.type, msg.content),
        lastMessageTime: msg.timestamp,
        unreadCount: chat.sendId !== currentUserId ? 1 : 0,
        pinned: false,
        muted: false,
        hasAtMe: atMe,
        hasMissedCall: missedCall,
      }, ...convs];
    }

    // 实时被 @：把这条 @我消息追加进 atMeMap，让顶部"有人@我"横幅即时出现/累计。
    // 仅收真实 mongoID（pusher 已把 MsgId 写入 WS 顶层 Id），临时 id 无法跳转故跳过；去重。
    let atMeMap = s.atMeMap;
    if (atMe && msg.id && !msg.id.startsWith('push_') && !msg.id.startsWith('local_')) {
      const cur = atMeMap[convId] || [];
      if (!cur.includes(msg.id)) atMeMap = { ...atMeMap, [convId]: [...cur, msg.id] };
    }

    return { messagesMap: { ...s.messagesMap, [convId]: msgs }, conversations: convs, atMeMap };
  });

  // 已参与的通话（已接通/已拒接）：用户已在通话里，不该在会话列表显示未读红点。
  // 结束后标记该会话已读（同步后端，刷新后也不再有未读）。
  if (seenCall) {
    useChatStore.getState().clearUnread(convId);
  }

  // 收到别人的消息 → 播放提示音 + 振动（免打扰会话不提示）。
  // 通话记录（mType=call）不当普通新消息提示：用户刚通完话，未接另有红标/专属提示，避免重复打扰。
  if (chat.sendId !== currentUserId && msg.type !== 'call' && typeof window !== 'undefined') {
    const convMuted = useChatStore.getState().conversations.find(c => c.id === convId)?.muted;
    if (!convMuted) notifyNewMessage();
  }

  const token = useIMStore.getState().currentUser?.token;
  if (!token) return;

  const isGroup = chat.chatType === ChatType.Group;

  // 异步加载发送者资料（用于群聊消息显示发送者名称）
  if (!useChatStore.getState().userProfiles[chat.sendId]) {
    resolveUserProfiles(token, [chat.sendId]);
  }

  // 检查会话名称是否需要补全（名称是纯 ID 格式说明还没填充过）
  const convNow = useChatStore.getState().conversations.find(c => c.id === convId);
  const needsResolve = convNow && (convNow.name === convId || /^\d+$/.test(convNow.name) || convNow.name.includes('_'));
  if (needsResolve) {
    if (isGroup) {
      // 群聊：查群信息 + 加载群成员
      (async () => {
        try {
          const res = await fetch('/api/social/groups', { headers: { Authorization: `Bearer ${token}` } });
          const d = await res.json();
          if (d.success) {
            const list = Array.isArray(d.data) ? d.data : (d.data?.list || []);
            const group = list.find((g: any) => String(g.id) === convId);
            if (group) {
              useChatStore.setState(s => ({
                conversations: s.conversations.map(c =>
                  c.id === convId ? { ...c, name: group.name || c.name, avatar: group.icon || c.avatar } : c,
                ),
              }));
            }
          }
        } catch { /* silent */ }
        // 预加载群成员
        useChatStore.getState().fetchGroupMembers(token, convId);
      })();
    } else {
      // 私聊：查对方用户资料（优先好友备注）
      const peerId = convId.split('_').find(p => p !== currentUserId) || '';
      if (peerId) {
        (async () => {
          // 1. 先从本地好友缓存查（按 friend_uid 优先，避免好友关系行号与他人 userId 撞号）
          let friends = useIMStore.getState().friends;
          let friend = friends.find(f => f.friend_uid === peerId) || friends.find(f => f.id === peerId);

          // 2. 本地没有则从 API 加载好友列表
          if (!friend && friends.length === 0) {
            try {
              const res = await fetch('/api/social/friends', { headers: { Authorization: `Bearer ${token}` } });
              const d = await res.json();
              if (d.success && d.data?.list) {
                friend = d.data.list.find((f: any) => String(f.friend_uid) === peerId);
              }
            } catch { /* silent */ }
          }

          if (friend) {
            const name = (friend as any).remark || (friend as any).nickname || friend.name || peerId;
            const avatar = (friend as any).user_avatar_url || friend.avatar || '';
            useChatStore.setState(s => ({
              conversations: s.conversations.map(c =>
                c.id === convId ? { ...c, name, avatar: avatar || c.avatar } : c,
              ),
            }));
            return;
          }

          // 3. 非好友：查用户 API
          await resolveUserProfiles(token, [peerId]);
          const profile = useChatStore.getState().userProfiles[peerId];
          if (profile) {
            useChatStore.setState(s => ({
              conversations: s.conversations.map(c =>
                c.id === convId ? { ...c, name: profile.nickname, avatar: profile.avatar } : c,
              ),
            }));
          }
        })();
      }
    }
  }
}

/** 新消息提示音 + 振动 */
function notifyNewMessage() {
  const s = useSettingsStore.getState();
  if (!s.notifyEnabled) return;
  if (s.notifySound) playMessageSound();
  if (s.notifyVibrate) vibrate();
}
