/**
 * Chat Store — 管理会话列表、消息、WebSocket 生命周期
 *
 * 替代 mock-data 中的 conversations / conversationMessagesMap，
 * 通过 IM REST API 获取数据 + WebSocket 实时推送。
 */

import { create } from 'zustand';
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
  type ChatLogItem,
  type ConversationItem,
  type BackendUser,
} from './api-client';
import { useIMStore } from './im-store';
import { playMessageSound, vibrate } from './notification';
import { useSettingsStore } from './settings-store';
import type { Message, Conversation } from './mock-data';

// ========== 消息类型映射 ==========

const backMsgTypeMap: Record<number, Message['type']> = {
  [MsgType.Text]: 'text',
  [MsgType.File]: 'text',     // 暂时作为 text 展示
  [MsgType.Voice]: 'voice',
  [MsgType.Image]: 'image',
  [MsgType.Memes]: 'text',
};

const frontMsgTypeMap: Record<string, number> = {
  text: MsgType.Text,
  image: MsgType.Image,
  voice: MsgType.Voice,
};

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
  role_level: number; // 1=member, 2=admin, 3=owner
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
  fetchConversations: (token: string) => Promise<void>;
  fetchMessages: (token: string, conversationId: string, oldestMsgId?: string) => Promise<void>;
  sendMessage: (token: string, userId: string, conversationId: string, content: string, msgType?: string) => void;
  resendMessage: (token: string, userId: string, conversationId: string, msgId: string) => void;
  markRead: (userId: string, conversationId: string, msgIds: string[]) => void;
  getOrCreateConversation: (token: string, userId: string, targetId: string) => Promise<Conversation>;
  deleteConversation: (token: string, conversationId: string) => void;
  clearUnread: (conversationId: string) => void;
  fetchGroupMembers: (token: string, groupId: string) => Promise<void>;
}

export const useChatStore = create<ChatState>()((set, get) => ({
  wsState: 'disconnected',
  groupMembers: {},
  conversations: [],
  messagesMap: {},
  userProfiles: {},
  loadingConversations: false,
  loadingMessages: {},

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
      onStateChange: (state) => set({ wsState: state }),
      onError: (err) => console.error('[ChatStore] ws error:', err),
    });

    // 服务端推送消息 — push.go NewMessage 不设 method，所以 method 为 ""
    ws.on('', (data, raw) => {
      const chat = data as WsChatData | null;
      if (!chat?.conversationId) return;
      handlePush(chat, raw?.id, userId);
    });

    ws.on('chat.ping', () => { /* pong */ });

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
    set({ loadingConversations: true });
    try {
      const resp = await getConversations(token);
      const map = resp?.conversationList || {};
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
        for (const gc of groupConvs) {
          get().fetchGroupMembers(token, gc.id);
        }
      }
    } catch (e) {
      console.error('[ChatStore] fetch conversations error:', e);
    } finally {
      set({ loadingConversations: false });
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
        return { messagesMap: { ...s.messagesMap, [conversationId]: merged } };
      });
    } catch (e) {
      console.error('[ChatStore] fetch messages error:', e);
    } finally {
      set(s => ({ loadingMessages: { ...s.loadingMessages, [conversationId]: false } }));
    }
  },

  // ==================== 发送消息 ====================

  sendMessage: (token, userId, conversationId, content, msgType = 'text') => {
    const conv = get().conversations.find(c => c.id === conversationId);
    const chatType = conv?.type === 'group' ? ChatType.Group : ChatType.Single;

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
    };

    set(s => {
      const msgs = [...(s.messagesMap[conversationId] || []), localMsg];
      const convs = s.conversations.map(c =>
        c.id === conversationId
          ? { ...c, lastMessage: content, lastMessageTime: new Date() }
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
        readRecords: {},
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

    ws.send('chat.user', wsData)
      .then(() => updateMsgStatus('sent'))
      .catch(err => {
        console.error('[ChatStore] send failed:', err);
        updateMsgStatus('failed');
      });
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

  // ==================== 清除未读 ====================

  clearUnread: (conversationId) => {
    const conv = get().conversations.find(c => c.id === conversationId);
    const unreadCount = conv?.unreadCount || 0;

    // 1. 清除前端未读计数
    set(s => ({
      conversations: s.conversations.map(c =>
        c.id === conversationId ? { ...c, unreadCount: 0 } : c,
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
}));

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
  return {
    id: raw.conversationId,
    type: raw.chatType === ChatType.Group ? 'group' : 'private',
    name: (raw as any).targetName || raw.conversationId,
    avatar: (raw as any).targetAvatar || '',
    lastMessage: msg?.msgContent || '',
    lastMessageTime: msg?.sendTime ? parseTimestamp(msg.sendTime) : new Date(),
    unreadCount: calcUnread(raw.seq, raw.read),
    pinned: false,
    muted: false,
    members: (raw as any).memberCount || undefined,
  };
}

function mapChatLog(log: ChatLogItem): Message {
  const base: Message = {
    id: log.id,
    senderId: log.sendId,
    content: log.msgContent,
    timestamp: log.sendTime ? parseTimestamp(log.sendTime) : new Date(),
    type: backMsgTypeMap[log.msgType] || 'text',
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

  const convId = chat.conversationId;
  const baseMsg: Message = {
    // 服务端在 pusher 把 MongoDB MsgId 写入 WS Message.Id，优先用它
    id: rawId || `push_${Date.now()}`,
    senderId: chat.sendId,
    content: chat.msg?.content || '',
    timestamp: chat.sendTime ? parseTimestamp(chat.sendTime) : new Date(),
    type: backMsgTypeMap[chat.msg?.mType] || 'text',
  };
  // 消费可能先到的乱序回执
  const msg = consumePendingReceipt(baseMsg);

  useChatStore.setState(s => {
    // 添加消息
    const msgs = [...(s.messagesMap[convId] || []), msg];

    // 更新或创建会话
    let convs = [...s.conversations];
    const idx = convs.findIndex(c => c.id === convId);
    if (idx >= 0) {
      convs[idx] = {
        ...convs[idx],
        lastMessage: msg.content,
        lastMessageTime: msg.timestamp,
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
        lastMessage: msg.content,
        lastMessageTime: msg.timestamp,
        unreadCount: chat.sendId !== currentUserId ? 1 : 0,
        pinned: false,
        muted: false,
      }, ...convs];
    }

    return { messagesMap: { ...s.messagesMap, [convId]: msgs }, conversations: convs };
  });

  // 收到别人的消息 → 播放提示音 + 振动
  if (chat.sendId !== currentUserId && typeof window !== 'undefined') {
    notifyNewMessage();
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
          // 1. 先从本地好友缓存查
          let friends = useIMStore.getState().friends;
          let friend = friends.find(f => f.id === peerId || f.friend_uid === peerId);

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
