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
    set(s => ({
      conversations: s.conversations.map(c =>
        c.id === conversationId ? { ...c, unreadCount: 0 } : c,
      ),
    }));
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
  const s = seq || 0;
  const r = read || 0;
  const diff = s - r;
  // seq 应该是消息序列号（小数字），不是时间戳（大数字）
  // 如果差值超过 10000，说明数据异常，返回 0
  if (diff <= 0 || diff > 10000) return 0;
  return diff;
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
  return {
    id: log.id,
    senderId: log.sendId,
    content: log.msgContent,
    timestamp: log.sendTime ? parseTimestamp(log.sendTime) : new Date(),
    type: backMsgTypeMap[log.msgType] || 'text',
  };
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

/** 处理服务端推送的消息 */
function handlePush(chat: WsChatData, rawId: string | undefined, currentUserId: string) {
  const convId = chat.conversationId;
  const msg: Message = {
    id: rawId || `push_${Date.now()}`,
    senderId: chat.sendId,
    content: chat.msg?.content || '',
    timestamp: chat.sendTime ? parseTimestamp(chat.sendTime) : new Date(),
    type: backMsgTypeMap[chat.msg?.mType] || 'text',
  };

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
