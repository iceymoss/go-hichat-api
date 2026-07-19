import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { createLatestRequest } from './latest-request';
import { getFriendRequestUnread, getGroupRequestUnread, type FriendRequestUnread, type GroupRequestUnread } from './social-request-api';
import { getNotificationUnreadCount } from './api-client';
import { type Contact } from '@/lib/types';

const friendUnreadRequest = createLatestRequest();
const groupUnreadRequest = createLatestRequest();
const friendsRequest = createLatestRequest();
const notificationUnreadRequest = createLatestRequest();

export type TabType = 'chats' | 'contacts' | 'moments' | 'me';
export type AuthView = 'login' | 'register' | 'forgot-password';
export type LoginMethod = 'id-password' | 'smart-code';

export interface AuthUser {
  id: string;
  phone: string;
  email: string | null;
  name: string;
  avatar: string;
  token: string;
  expire: number;
  sex: number;
  introduction: string;
  region: string;
  occupation: string;
  tags: string;
  momentsCover?: string;
}

interface IMState {
  // Auth state
  isAuthenticated: boolean;
  currentUser: AuthUser | null;
  authView: AuthView;
  loginMethod: LoginMethod;
  setAuthView: (view: AuthView) => void;
  setLoginMethod: (method: LoginMethod) => void;
  login: (user: AuthUser) => void;
  logout: () => void;
  updateCurrentUser: (partial: Partial<AuthUser>) => void;

  // IM state
  activeTab: TabType;
  setActiveTab: (tab: TabType) => void;
  selectedConversationId: string | null;
  setSelectedConversationId: (id: string | null) => void;
  selectedContactId: string | null;
  setSelectedContactId: (id: string | null) => void;
  // Profile of the user currently being viewed in the contact detail panel.
  // Lets us open detail for self / non-friends (not in the friends list).
  viewedProfile: Contact | null;
  openUserProfile: (uid: string) => void;
  // Floating profile card overlay (does NOT navigate away — used from moments so
  // closing the card returns to the exact trend underneath).
  floatingProfile: Contact | null;
  floatingProfileIsStranger: boolean;
  showUserCard: (uid: string) => void;
  closeUserCard: () => void;
  // One-shot request to open a specific user's trends (朋友圈) in the moments tab.
  // MomentsFeed consumes it to switch into the userTrends view, then clears it.
  userTrendsTarget: { id: string; name: string } | null;
  openUserTrends: (uid: string, name?: string) => void;
  clearUserTrendsTarget: () => void;
  searchQuery: string;
  setSearchQuery: (query: string) => void;
  showChatDetail: boolean;
  setShowChatDetail: (show: boolean) => void;

  // Friend request panel
  showFriendRequests: boolean;
  setShowFriendRequests: (show: boolean) => void;
  friendRequestUnreadCount: number;
  setFriendRequestUnreadCount: (count: number) => void;
  friendRequestUnread: FriendRequestUnread;
  friendRequestsVersion: number;
  invalidateFriendRequests: () => void;
  refreshFriendRequestUnread: () => Promise<void>;

  // Moments (动态) message-center unread + realtime notify version
  momentsUnreadCount: number;
  setMomentsUnreadCount: (count: number) => void;
  trendNotifyVersion: number;
  bumpTrendNotifyVersion: () => void;

  // 公共通知（好友/群申请等）：收到 ws notify 帧后 bump，供通知中心刷新历史列表
  notificationVersion: number;
  bumpNotificationVersion: () => void;

  // 公共通知未读总数：登录拉一次(持久化) + ws 实时 +1，驱动「联系人」tab 气泡 + 铃铛角标（同一真相）
  notificationUnreadCount: number;
  setNotificationUnreadCount: (count: number) => void;
  refreshNotificationUnread: () => Promise<void>;

  // 通知点击/气泡跳转意图：把用户带到对应入口的具体子 tab（received=我收到 / sent=我发起）
  friendReqNavTab: 'received' | 'sent' | null;
  groupAppNavTab: 'received' | 'sent' | null;
  clearFriendReqNavTab: () => void;
  clearGroupAppNavTab: () => void;
  // 从群聊深链到通讯录群详情：携带 groupId，GroupList 读取后自动打开该群详情
  groupDetailNavId: string | null;
  openGroupDetail: (groupId: string) => void;
  clearGroupDetailNav: () => void;
  // 按 notifyType 跳到来源（好友→新的朋友；群→群申请），并定位子 tab。通知中心点击 + toast 共用。
  navigateToNotificationSource: (notifyType: string) => void;

  // Group panel
  showGroupPanel: boolean;
  setShowGroupPanel: (show: boolean) => void;
  groupAppUnreadCount: number;
  setGroupAppUnreadCount: (count: number) => void;
  groupRequestUnread: GroupRequestUnread;
  groupRequestsVersion: number;
  invalidateGroupRequests: () => void;
  groupsVersion: number;
  invalidateGroups: () => void;
  refreshGroupRequestUnread: () => Promise<void>;
  refreshSocialRequestUnread: () => Promise<void>;

  // Trend detail panel
  selectedTrendId: number | null;
  setSelectedTrendId: (id: number | null) => void;

  // Per-trend version counter: any panel that mutates a trend bumps its version;
  // other panels subscribe to the map and refetch when the version changes for
  // an id they care about.
  trendVersions: Record<number, number>;
  bumpTrendVersion: (trendId: number) => void;

  // Me tab sub-page (shown in right panel)
  meSubPage: 'profile' | 'favorites' | 'album' | 'emojis' | null;
  setMeSubPage: (page: 'profile' | 'favorites' | 'album' | 'emojis' | null) => void;

  // System panel opened from the sidebar settings gear (independent of activeTab)
  systemPanel: 'settings' | 'about' | null;
  setSystemPanel: (panel: 'settings' | 'about' | null) => void;

  // Friends list (fetched from API)
  friends: Contact[];
  setFriends: (friends: Contact[]) => void;
  friendsVersion: number;
  invalidateFriends: () => void;
  refreshFriends: () => Promise<void>;
}

// ── Shared profile resolvers (used by openUserProfile + showUserCard) ──
function meToContact(me: AuthUser): Contact {
  return { id: me.id, name: me.name, nickname: me.name, avatar: me.avatar, pinyin: '', letter: '', gender: me.sex === 1 ? 'male' : me.sex === 2 ? 'female' : undefined, phone: me.phone, email: me.email || undefined, region: me.region, signature: me.introduction, introduction: me.introduction, occupation: me.occupation };
}
function searchUserToContact(u: { id: string | number; nickname?: string; avatar?: string; sex?: number; region?: string; introduction?: string; occupation?: string }, fallbackId: string): Contact {
  return { id: String(u.id), name: u.nickname || fallbackId, nickname: u.nickname, avatar: u.avatar || '', pinyin: '', letter: '', gender: u.sex === 1 ? 'male' : u.sex === 2 ? 'female' : undefined, region: u.region, signature: u.introduction, introduction: u.introduction, occupation: u.occupation };
}

export const useIMStore = create<IMState>()(persist((set) => ({
  // Auth
  isAuthenticated: false,
  currentUser: null,
  authView: 'login',
  loginMethod: 'id-password',
  setAuthView: (view) => set({ authView: view }),
  setLoginMethod: (method) => set({ loginMethod: method }),
  login: (user) => {
    friendUnreadRequest.invalidate();
    groupUnreadRequest.invalidate();
    friendsRequest.invalidate();
    notificationUnreadRequest.invalidate();
    if (typeof window !== 'undefined') window.dispatchEvent(new Event('hichat:auth-reset'));
    set({ isAuthenticated: true, currentUser: user, authView: 'login' });
    // Load settings from backend after login
    import('./settings-store').then(({ useSettingsStore }) => {
      useSettingsStore.getState().loadFromBackend(user.token);
    });
  },
  logout: () => {
    friendUnreadRequest.invalidate();
    groupUnreadRequest.invalidate();
    friendsRequest.invalidate();
    notificationUnreadRequest.invalidate();
    if (typeof window !== 'undefined') window.dispatchEvent(new Event('hichat:auth-reset'));
    const token = useIMStore.getState().currentUser?.token;
    if (token) {
      fetch('/api/auth/logout', {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}` },
      }).catch(() => {});
    }
    set({ isAuthenticated: false, currentUser: null, activeTab: 'chats', selectedConversationId: null, showChatDetail: false, friendRequestUnreadCount: 0, groupAppUnreadCount: 0, notificationUnreadCount: 0, friendRequestUnread: { total: 0, apply: 0, result: 0 }, groupRequestUnread: { total: 0, apply: 0, result: 0, invite: 0 }, friends: [] });
  },
  updateCurrentUser: (partial) => set((state) => ({
    currentUser: state.currentUser ? { ...state.currentUser, ...partial } : null,
  })),

  // IM
  activeTab: 'chats',
  setActiveTab: (tab) => set({ activeTab: tab, selectedConversationId: null, selectedContactId: null, showChatDetail: false, showFriendRequests: false, showGroupPanel: false, selectedTrendId: null, meSubPage: null, systemPanel: null, groupDetailNavId: null }),
  selectedConversationId: null,
  setSelectedConversationId: (id) => set({ selectedConversationId: id, showChatDetail: true }),
  selectedContactId: null,
  setSelectedContactId: (id) => set({ selectedContactId: id }),
  viewedProfile: null,
  openUserProfile: (uid) => {
    if (!uid) return;
    const st = useIMStore.getState();
    set({ activeTab: 'contacts', selectedContactId: uid, showChatDetail: false, showFriendRequests: false, showGroupPanel: false, selectedTrendId: null, meSubPage: null });

    const friend = st.friends.find(f => f.id === uid);
    if (friend) { set({ viewedProfile: friend }); return; }

    const me = st.currentUser;
    if (me && uid === me.id) {
      set({ viewedProfile: meToContact(me) });
      return;
    }

    // Non-friend: fetch the basic profile by id.
    set({ viewedProfile: { id: uid, name: uid, avatar: '', pinyin: '', letter: '' } });
    if (me?.token) {
      fetch(`/api/user/search?ids=${encodeURIComponent(uid)}`, { headers: { Authorization: `Bearer ${me.token}` } })
        .then(r => r.json())
        .then(j => {
          const u = (j?.data?.users || j?.users || [])[0];
          if (u && useIMStore.getState().selectedContactId === uid) {
            set({ viewedProfile: searchUserToContact(u, uid) });
          }
        })
        .catch(() => {});
    }
  },
  floatingProfile: null,
  floatingProfileIsStranger: false,
  showUserCard: (uidRaw) => {
    if (!uidRaw) return;
    const st = useIMStore.getState();
    const me = st.currentUser;
    // Comment/like ids may be the literal 'me' — normalize to the real id.
    const uid = (uidRaw === 'me' && me) ? me.id : uidRaw;

    const friend = st.friends.find(f => f.id === uid);
    if (friend) { set({ floatingProfile: friend, floatingProfileIsStranger: false }); return; }

    if (me && uid === me.id) {
      set({ floatingProfile: meToContact(me), floatingProfileIsStranger: false });
      return;
    }

    // Non-friend: show a placeholder immediately, then refine via search.
    set({ floatingProfile: { id: uid, name: uid, avatar: '', pinyin: '', letter: '' }, floatingProfileIsStranger: true });
    if (me?.token) {
      fetch(`/api/user/search?ids=${encodeURIComponent(uid)}`, { headers: { Authorization: `Bearer ${me.token}` } })
        .then(r => r.json())
        .then(j => {
          const u = (j?.data?.users || j?.users || [])[0];
          if (u && useIMStore.getState().floatingProfile?.id === uid) {
            set({ floatingProfile: searchUserToContact(u, uid) });
          }
        })
        .catch(() => {});
    }
  },
  closeUserCard: () => set({ floatingProfile: null }),
  userTrendsTarget: null,
  openUserTrends: (uidRaw, name = '') => {
    if (!uidRaw) return;
    const me = useIMStore.getState().currentUser;
    // Comment/like ids may be the literal 'me' — normalize to the real id.
    const uid = (uidRaw === 'me' && me) ? me.id : uidRaw;
    set({
      activeTab: 'moments',
      userTrendsTarget: { id: uid, name },
      floatingProfile: null,
      selectedConversationId: null,
      selectedContactId: null,
      showChatDetail: false,
      showFriendRequests: false,
      showGroupPanel: false,
      selectedTrendId: null,
      meSubPage: null,
    });
  },
  clearUserTrendsTarget: () => set({ userTrendsTarget: null }),
  searchQuery: '',
  setSearchQuery: (query) => set({ searchQuery: query }),
  showChatDetail: false,
  setShowChatDetail: (show) => set({ showChatDetail: show }),

  // Friend request panel
  showFriendRequests: false,
  setShowFriendRequests: (show) => set({ showFriendRequests: show, selectedContactId: null, showGroupPanel: false }),
  friendRequestUnreadCount: 0,
  setFriendRequestUnreadCount: (count) => set({ friendRequestUnreadCount: count }),
  friendRequestUnread: { total: 0, apply: 0, result: 0 },
  friendRequestsVersion: 0,
  invalidateFriendRequests: () => set((state) => ({ friendRequestsVersion: state.friendRequestsVersion + 1 })),
  refreshFriendRequestUnread: async () => {
    const token = useIMStore.getState().currentUser?.token;
    if (!token) return;
    const isLatest = friendUnreadRequest.begin();
    try {
      const unread = await getFriendRequestUnread(token);
      if (!isLatest() || useIMStore.getState().currentUser?.token !== token) return;
      set({ friendRequestUnread: unread, friendRequestUnreadCount: unread.total });
    } catch { /* preserve the last successful value */ }
  },

  momentsUnreadCount: 0,
  setMomentsUnreadCount: (count) => set({ momentsUnreadCount: count }),
  trendNotifyVersion: 0,
  bumpTrendNotifyVersion: () => set((s) => ({ trendNotifyVersion: s.trendNotifyVersion + 1 })),

  notificationVersion: 0,
  bumpNotificationVersion: () => set((s) => ({ notificationVersion: s.notificationVersion + 1 })),

  notificationUnreadCount: 0,
  setNotificationUnreadCount: (count) => set({ notificationUnreadCount: Math.max(0, count) }),
  refreshNotificationUnread: async () => {
    const token = useIMStore.getState().currentUser?.token;
    if (!token) return;
    const isLatest = notificationUnreadRequest.begin();
    try {
      const body = await getNotificationUnreadCount(token);
      if (!isLatest() || useIMStore.getState().currentUser?.token !== token) return;
      set({ notificationUnreadCount: Math.max(0, body.count ?? 0) });
    } catch { /* preserve the last successful value */ }
  },

  friendReqNavTab: null,
  groupAppNavTab: null,
  clearFriendReqNavTab: () => set({ friendReqNavTab: null }),
  clearGroupAppNavTab: () => set({ groupAppNavTab: null }),
  groupDetailNavId: null,
  openGroupDetail: (groupId) => set({
    activeTab: 'contacts', showGroupPanel: true, showFriendRequests: false,
    selectedContactId: null, showChatDetail: false, selectedTrendId: null, meSubPage: null,
    groupDetailNavId: groupId,
  }),
  clearGroupDetailNav: () => set({ groupDetailNavId: null }),
  navigateToNotificationSource: (notifyType) => {
    // 公共导航字段，避免被 setActiveTab/setShowXxx 互相重置：一次 set 落定目标视图 + 子 tab 意图
    const base = { selectedContactId: null, showChatDetail: false, selectedTrendId: null, meSubPage: null };
    if (notifyType.startsWith('friend.')) {
      // friend.apply=被申请人(我收到)；friend.accept/reject=申请人看结果(我发起)
      const tab = notifyType === 'friend.apply' ? 'received' : 'sent';
      set({ ...base, activeTab: 'contacts', showGroupPanel: false, showFriendRequests: true, friendReqNavTab: tab });
    } else if (notifyType.startsWith('group.')) {
      // group.apply=群主/管理员(我收到)；group.accept/reject=申请人看结果(我发起)
      const tab = notifyType === 'group.apply' ? 'received' : 'sent';
      set({ ...base, activeTab: 'contacts', showFriendRequests: false, showGroupPanel: true, groupAppNavTab: tab });
    }
  },

  // Group panel
  showGroupPanel: false,
  setShowGroupPanel: (show) => set({ showGroupPanel: show, selectedContactId: null, showFriendRequests: false }),
  groupAppUnreadCount: 0,
  setGroupAppUnreadCount: (count) => set({ groupAppUnreadCount: count }),
  groupRequestUnread: { total: 0, apply: 0, result: 0, invite: 0 },
  groupRequestsVersion: 0,
  invalidateGroupRequests: () => set((state) => ({ groupRequestsVersion: state.groupRequestsVersion + 1 })),
  groupsVersion: 0,
  invalidateGroups: () => set((state) => ({ groupsVersion: state.groupsVersion + 1 })),
  refreshGroupRequestUnread: async () => {
    const token = useIMStore.getState().currentUser?.token;
    if (!token) return;
    const isLatest = groupUnreadRequest.begin();
    try {
      const unread = await getGroupRequestUnread(token);
      if (!isLatest() || useIMStore.getState().currentUser?.token !== token) return;
      set({ groupRequestUnread: unread, groupAppUnreadCount: unread.total });
    } catch { /* preserve the last successful value */ }
  },
  refreshSocialRequestUnread: async () => {
    const state = useIMStore.getState();
    await Promise.all([state.refreshFriendRequestUnread(), state.refreshGroupRequestUnread()]);
  },

  // Trend detail panel
  selectedTrendId: null,
  setSelectedTrendId: (id) => set({ selectedTrendId: id }),
  trendVersions: {},
  bumpTrendVersion: (trendId) => set((state) => ({
    trendVersions: { ...state.trendVersions, [trendId]: (state.trendVersions[trendId] || 0) + 1 },
  })),

  // Me tab sub-page
  meSubPage: null,
  setMeSubPage: (page) => set({ meSubPage: page }),

  // System panel (settings / about) opened from the sidebar gear
  systemPanel: null,
  setSystemPanel: (panel) => set({ systemPanel: panel }),

  // Friends list
  friends: [],
  setFriends: (friends) => set({ friends }),
  friendsVersion: 0,
  invalidateFriends: () => set((state) => ({ friendsVersion: state.friendsVersion + 1 })),
  refreshFriends: async () => {
    const token = useIMStore.getState().currentUser?.token;
    if (!token) return;
    const isLatest = friendsRequest.begin();
    try {
      const [friendsResponse, onlineResponse] = await Promise.all([
        fetch('/api/social/friends', { headers: { Authorization: `Bearer ${token}` } }),
        fetch('/api/social/friends/online', { headers: { Authorization: `Bearer ${token}` } }),
      ]);
      const friendsBody = await friendsResponse.json();
      const onlineBody = await onlineResponse.json();
      if (!isLatest() || useIMStore.getState().currentUser?.token !== token) return;
      if (!friendsResponse.ok || !friendsBody?.success) return;
      const onlineMap: Record<string, boolean> = onlineBody?.data?.onLineList || {};
      const list = friendsBody?.data?.list || friendsBody?.list || [];
      const mapped: Contact[] = list.map((friend: any) => {
        const id = String(friend.friend_uid || friend.id || '');
        const name = friend.remark || friend.nickname || id;
        return { id, friend_uid: String(friend.friend_uid || ''), name, remark: friend.remark || '', nickname: friend.nickname || '', avatar: friend.avatar || '', pinyin: name, letter: /^[a-z]/i.test(name[0] || '') ? name[0].toUpperCase() : '#', online: !!onlineMap[id], gender: friend.sex === 1 ? 'male' : friend.sex === 2 ? 'female' : undefined, phone: friend.phone || undefined, email: friend.email || undefined, region: friend.region || undefined, signature: friend.introduction || undefined, introduction: friend.introduction || undefined, occupation: friend.occupation || undefined, tags: friend.tags || undefined, account: id, blacklisted: !!friend.blacklisted, moments_permission: friend.moments_permission, notify_enabled: friend.notify_enabled, pinned: !!friend.pinned, muted: !!friend.muted, friend_tags: friend.friend_tags || [] };
      });
      set({ friends: mapped });
    } catch { /* preserve the last successful value */ }
  },
}), {
  name: 'hichat-auth',
  partialize: (state) => ({
    isAuthenticated: state.isAuthenticated,
    currentUser: state.currentUser,
  }),
}));

/**
 * 按 userId 解析展示名：好友备注优先，其次用传入的后端昵称兜底。
 * 用于动态评论 / 点赞 / 消息中心等显示他人名字的地方，统一"备注优先"。
 */
export function friendDisplayName(userId: string, fallback: string): string {
  if (!userId) return fallback;
  const f = useIMStore.getState().friends.find(c => c.id === userId);
  return f?.remark || fallback;
}
