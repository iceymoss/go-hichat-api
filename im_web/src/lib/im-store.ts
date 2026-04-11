import { create } from 'zustand';

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

  // Me tab sub-page (shown in right panel)
  meSubPage: 'profile' | 'favorites' | 'album' | null;
  setMeSubPage: (page: 'profile' | 'favorites' | 'album' | null) => void;
}

export const useIMStore = create<IMState>((set) => ({
  // Auth
  isAuthenticated: false,
  currentUser: null,
  authView: 'login',
  loginMethod: 'id-password',
  setAuthView: (view) => set({ authView: view }),
  setLoginMethod: (method) => set({ loginMethod: method }),
  login: (user) => set({ isAuthenticated: true, currentUser: user, authView: 'login' }),
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

  // Me tab sub-page
  meSubPage: null,
  setMeSubPage: (page) => set({ meSubPage: page }),
}));
