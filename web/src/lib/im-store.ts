import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { type Contact } from '@/lib/mock-data';

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
  searchQuery: string;
  setSearchQuery: (query: string) => void;
  showChatDetail: boolean;
  setShowChatDetail: (show: boolean) => void;

  // Friend request panel
  showFriendRequests: boolean;
  setShowFriendRequests: (show: boolean) => void;
  friendRequestUnreadCount: number;
  setFriendRequestUnreadCount: (count: number) => void;

  // Group panel
  showGroupPanel: boolean;
  setShowGroupPanel: (show: boolean) => void;
  groupAppUnreadCount: number;
  setGroupAppUnreadCount: (count: number) => void;

  // Trend detail panel
  selectedTrendId: number | null;
  setSelectedTrendId: (id: number | null) => void;

  // Per-trend version counter: any panel that mutates a trend bumps its version;
  // other panels subscribe to the map and refetch when the version changes for
  // an id they care about.
  trendVersions: Record<number, number>;
  bumpTrendVersion: (trendId: number) => void;

  // Me tab sub-page (shown in right panel)
  meSubPage: 'profile' | 'favorites' | 'album' | 'emojis' | 'settings' | null;
  setMeSubPage: (page: 'profile' | 'favorites' | 'album' | 'emojis' | 'settings' | null) => void;

  // Friends list (fetched from API)
  friends: Contact[];
  setFriends: (friends: Contact[]) => void;
  friendsVersion: number;
  invalidateFriends: () => void;
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
    set({ isAuthenticated: true, currentUser: user, authView: 'login' });
    // Load settings from backend after login
    import('./settings-store').then(({ useSettingsStore }) => {
      useSettingsStore.getState().loadFromBackend(user.token);
    });
  },
  logout: () => {
    const token = useIMStore.getState().currentUser?.token;
    if (token) {
      fetch('/api/auth/logout', {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}` },
      }).catch(() => {});
    }
    set({ isAuthenticated: false, currentUser: null, activeTab: 'chats', selectedConversationId: null, showChatDetail: false });
  },
  updateCurrentUser: (partial) => set((state) => ({
    currentUser: state.currentUser ? { ...state.currentUser, ...partial } : null,
  })),

  // IM
  activeTab: 'chats',
  setActiveTab: (tab) => set({ activeTab: tab, selectedConversationId: null, selectedContactId: null, showChatDetail: false, showFriendRequests: false, showGroupPanel: false, selectedTrendId: null, meSubPage: null }),
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
      set({ viewedProfile: { id: me.id, name: me.name, nickname: me.name, avatar: me.avatar, pinyin: '', letter: '', gender: me.sex === 1 ? 'male' : me.sex === 2 ? 'female' : undefined, phone: me.phone, email: me.email || undefined, region: me.region, signature: me.introduction, introduction: me.introduction, occupation: me.occupation } });
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
            set({ viewedProfile: { id: String(u.id), name: u.nickname || uid, nickname: u.nickname, avatar: u.avatar || '', pinyin: '', letter: '', gender: u.sex === 1 ? 'male' : u.sex === 2 ? 'female' : undefined, region: u.region, signature: u.introduction, introduction: u.introduction, occupation: u.occupation } });
          }
        })
        .catch(() => {});
    }
  },
  searchQuery: '',
  setSearchQuery: (query) => set({ searchQuery: query }),
  showChatDetail: false,
  setShowChatDetail: (show) => set({ showChatDetail: show }),

  // Friend request panel
  showFriendRequests: false,
  setShowFriendRequests: (show) => set({ showFriendRequests: show, selectedContactId: null, showGroupPanel: false }),
  friendRequestUnreadCount: 0,
  setFriendRequestUnreadCount: (count) => set({ friendRequestUnreadCount: count }),

  // Group panel
  showGroupPanel: false,
  setShowGroupPanel: (show) => set({ showGroupPanel: show, selectedContactId: null, showFriendRequests: false }),
  groupAppUnreadCount: 0,
  setGroupAppUnreadCount: (count) => set({ groupAppUnreadCount: count }),

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

  // Friends list
  friends: [],
  setFriends: (friends) => set({ friends }),
  friendsVersion: 0,
  invalidateFriends: () => set((state) => ({ friendsVersion: state.friendsVersion + 1 })),
}), {
  name: 'hichat-auth',
  partialize: (state) => ({
    isAuthenticated: state.isAuthenticated,
    currentUser: state.currentUser,
  }),
}));
