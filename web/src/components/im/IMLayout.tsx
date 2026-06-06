'use client';

import React from 'react';
import { useIMStore, TabType } from '@/lib/im-store';
import { type Contact } from '@/lib/mock-data';
import { useChatStore } from '@/lib/chat-store';
import { ChatListProvider, ChatListToolbar, ChatListContent } from './ChatList';
import ChatDetail from './ChatDetail';
import ContactList from './ContactList';
import ContactDetailPanel from './ContactDetailPanel';
import FloatingProfileCard from './FloatingProfileCard';
import FriendRequestList from './FriendRequestList';
import GroupList from './GroupList';
import MomentsFeed from './MomentsFeed';
import TrendDetailPanel from './TrendDetailPanel';
import ProfilePage from './ProfilePage';
import MyProfileEditPage from './MyProfileEditPage';
import FavoritesPage from './FavoritesPage';
import AlbumPage from './AlbumPage';
import EmojisPage from './EmojisPage';
import SettingsPage from './SettingsPage';
import { Avatar, AvatarImage, AvatarFallback } from '@/components/ui/avatar';
import { useIsMobile } from '@/hooks/use-mobile';
import { toast } from 'sonner';
import {
  MessageCircle,
  Users,
  Compass,
  User,
  Search,
} from 'lucide-react';
import { useT } from '@/hooks/use-i18n';

const navItems: { tab: TabType; icon: React.ReactNode }[] = [
  { tab: 'chats', icon: <MessageCircle className="w-5 h-5" /> },
  { tab: 'contacts', icon: <Users className="w-5 h-5" /> },
  { tab: 'moments', icon: <Compass className="w-5 h-5" /> },
  { tab: 'me', icon: <User className="w-5 h-5" /> },
];

export default function IMLayout() {
  const { activeTab, setActiveTab, showChatDetail, selectedContactId, showFriendRequests, showGroupPanel, selectedTrendId, meSubPage, setMeSubPage, currentUser, friends, viewedProfile, floatingProfile, floatingProfileIsStranger, closeUserCard, openUserProfile, setSelectedConversationId, setShowChatDetail } = useIMStore();
  const chatConversations = useChatStore(s => s.conversations);

  // Resolve selected contact for detail panel from store friends, falling back
  // to viewedProfile (set when opening self / a non-friend from moments).
  const selectedContact: Contact | null = React.useMemo(() => {
    if (!selectedContactId) return null;
    return friends.find(c => c.id === selectedContactId)
      || (viewedProfile && viewedProfile.id === selectedContactId ? viewedProfile : null);
  }, [selectedContactId, friends, viewedProfile]);
  const isMobile = useIsMobile();
  // 免打扰会话不计入侧边栏总未读气泡（与会话列表只显示小灰点保持一致）
  const totalUnread = chatConversations.reduce((sum, c) => sum + (c.muted ? 0 : c.unreadCount), 0);
  const t = useT();

  const tabLabels: Record<TabType, string> = {
    chats: t('tab.chats'),
    contacts: t('tab.contacts'),
    moments: t('tab.moments'),
    me: t('tab.me'),
  };

  const renderContent = () => {
    switch (activeTab) {
      case 'contacts':
        return <ContactList />;
      case 'moments':
        return <MomentsFeed />;
      case 'me':
        return <ProfilePage />;
      default:
        return null;
    }
  };

  const showContactDetail = activeTab === 'contacts' && selectedContact !== null && !showFriendRequests && !showGroupPanel;
  const showFriendRequestPanel = activeTab === 'contacts' && showFriendRequests;
  const showGroupPanelView = activeTab === 'contacts' && showGroupPanel;
  const showTrendDetail = activeTab === 'moments' && selectedTrendId !== null;

  /* ═══════════════════════════════════════
     Shared: Non-chat header buttons (search + pencil / three dots)
     ═══════════════════════════════════════ */
  const renderNonChatHeaderButtons = () => (
    <div className="flex items-center gap-1">
      <button
        className="flex items-center justify-center"
        style={{
          width: 36, height: 36, borderRadius: '50%',
          border: 'none', background: 'transparent',
          color: 'rgba(255,255,255,0.7)', cursor: 'pointer',
        }}
      >
        <Search className="w-5 h-5" />
      </button>
      {isMobile ? (
        <button
          className="flex items-center justify-center"
          style={{
            width: 36, height: 36, borderRadius: '50%',
            border: 'none', background: 'transparent',
            color: 'rgba(255,255,255,0.7)', cursor: 'pointer',
          }}
        >
          <svg className="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <circle cx="12" cy="5" r="1.5" /><circle cx="12" cy="12" r="1.5" /><circle cx="12" cy="19" r="1.5" />
          </svg>
        </button>
      ) : (
        <button
          className="flex items-center justify-center"
          style={{
            width: 36, height: 36, borderRadius: '50%',
            border: 'none', background: 'transparent',
            color: 'rgba(255,255,255,0.7)', cursor: 'pointer',
          }}
        >
          <svg className="w-[18px] h-[18px]" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M12 20h9" /><path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" />
          </svg>
        </button>
      )}
    </div>
  );

  /* ── Floating user profile card (opened from moments; overlays everything,
        closing returns to the exact trend underneath) ── */
  const floatingCard = floatingProfile ? (
    <FloatingProfileCard
      contact={floatingProfile}
      isStranger={floatingProfileIsStranger}
      zIndex={10020}
      onClose={closeUserCard}
      onSendMessage={async () => {
        if (!currentUser?.token || !currentUser?.id) return;
        try {
          const conv = await useChatStore.getState().getOrCreateConversation(currentUser.token, currentUser.id, floatingProfile.id);
          closeUserCard();
          setActiveTab('chats');
          setSelectedConversationId(conv.id);
          setShowChatDetail(true);
        } catch {
          toast.error('打开会话失败');
        }
      }}
      onAddFriend={floatingProfileIsStranger ? () => { toast('请在通讯录中搜索添加好友'); closeUserCard(); } : undefined}
      onViewProfile={() => { const id = floatingProfile.id; closeUserCard(); openUserProfile(id); }}
    />
  ) : null;

  /* ═══════════════════════════════════════
     MOBILE LAYOUT
     ═══════════════════════════════════════ */
  if (isMobile) {
    const headerHeight = 52;
    return (
      <div className="h-full flex flex-col" style={{ background: '#F5F7FA' }}>
        {activeTab === 'chats' ? (
          /* ── Mobile Chats Tab: provider wraps header + content ── */
          <ChatListProvider>
            <header
              className="flex items-center px-4 shrink-0"
              style={{
                height: headerHeight,
                background: '#2C3E50',
                border: 'none',
                gap: 12,
              }}
            >
              <h1 style={{ fontSize: 17, fontWeight: 600, color: '#FFFFFF', flexShrink: 0 }}>
                {tabLabels.chats}
              </h1>
              <ChatListToolbar />
            </header>
            <main
              className="flex-1 overflow-hidden"
              style={{ background: showChatDetail ? '#F0F2F5' : '#FFFFFF' }}
            >
              {showChatDetail ? <ChatDetail /> : <ChatListContent />}
            </main>
          </ChatListProvider>
        ) : (
          /* ── Mobile Non-Chat Tabs ── */
          <>
            <header
              className="flex items-center justify-between px-4 shrink-0"
              style={{
                height: headerHeight,
                background: '#2C3E50',
                border: 'none',
              }}
            >
              <h1 style={{ fontSize: 17, fontWeight: 600, color: '#FFFFFF' }}>
                {tabLabels[activeTab]}
              </h1>
              {renderNonChatHeaderButtons()}
            </header>
            <main
              className="flex-1 overflow-hidden"
              style={{ background: '#FFFFFF' }}
            >
              {showGroupPanelView ? <GroupList /> : showFriendRequestPanel ? <FriendRequestList /> : showContactDetail ? <ContactDetailPanel contact={selectedContact} /> : renderContent()}
            </main>
          </>
        )}

        {/* Bottom Tab Bar */}
        <nav
          className="flex items-stretch justify-around shrink-0"
          style={{
            background: '#FFFFFF',
            borderTop: '1px solid rgba(0,0,0,0.08)',
            paddingBottom: 'env(safe-area-inset-bottom)',
            borderLeft: 'none',
            borderRight: 'none',
            borderBottom: 'none',
          }}
        >
          {navItems.map(({ tab, icon }) => {
            const unread = tab === 'chats' ? totalUnread : 0;
            return (
              <button
                key={tab}
                onClick={() => setActiveTab(tab)}
                className={`im-bottom-tab flex-1 ${activeTab === tab ? 'active' : ''}`}
                style={{ border: 'none' }}
              >
                <span className="tab-icon">
                  {icon}
                  {unread > 0 && (
                    <span className="tab-badge">{unread > 99 ? '99+' : unread}</span>
                  )}
                </span>
                <span className="text-[10px]">{tabLabels[tab]}</span>
              </button>
            );
          })}
        </nav>
        {floatingCard}
      </div>
    );
  }

  /* ═══════════════════════════════════════
     DESKTOP LAYOUT — dark left panel + light chat
     ═══════════════════════════════════════ */
  const desktopHeaderHeight = 56;

  return (
    <div className="h-full flex overflow-hidden" style={{ background: '#F5F7FA' }}>
      {/* ── Icon Sidebar: TG dark ── */}
      <aside className="im-sidebar flex flex-col items-center py-3 gap-1 shrink-0" style={{ width: 56 }}>
        {/* User Avatar */}
        <div className="mb-3">
          <Avatar className="w-8 h-8 cursor-pointer" style={{ border: '2px solid rgba(255,255,255,0.15)' }}>
            <AvatarImage src={currentUser?.avatar} alt={currentUser?.name} />
            <AvatarFallback>{currentUser?.name}</AvatarFallback>
          </Avatar>
        </div>

        {/* Navigation Icons */}
        <div className="flex flex-col gap-1 flex-1">
          {navItems.map(({ tab, icon }) => {
            const unread = tab === 'chats' ? totalUnread : 0;
            return (
              <button
                key={tab}
                onClick={() => setActiveTab(tab)}
                className={`im-sidebar-icon ${activeTab === tab ? 'active' : ''}`}
                title={tabLabels[tab]}
                style={{ width: 38, height: 38, borderRadius: 12 }}
              >
                {icon}
                {unread > 0 && (
                  <span className="unread-dot">{unread > 99 ? '99+' : unread}</span>
                )}
              </button>
            );
          })}
        </div>

        {/* Settings */}
        <div className="flex flex-col gap-1 mt-auto">
          <button className="im-sidebar-icon" title="设置" style={{ width: 38, height: 38, borderRadius: 12 }}>
            <svg className="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8">
              <path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z" />
              <circle cx="12" cy="12" r="3" />
            </svg>
          </button>
        </div>
      </aside>

      {/* ── Left Panel: dark bg for chat list, white for others ── */}
      <div
        className="w-[420px] flex flex-col shrink-0"
        style={{
          background: activeTab === 'chats' ? '#2C3E50' : '#FFFFFF',
          borderRight: (showChatDetail && activeTab === 'chats') || showContactDetail || showFriendRequestPanel || showGroupPanelView || showTrendDetail || (activeTab === 'me' && meSubPage)
            ? '1px solid rgba(0,0,0,0.08)'
            : 'none',
        }}
      >
        {activeTab === 'chats' ? (
          /* ── Desktop Chats Tab: provider wraps header + content ── */
          <ChatListProvider>
            <header
              className="flex items-center px-4 shrink-0"
              style={{
                height: desktopHeaderHeight,
                background: '#2C3E50',
                borderBottom: '1px solid rgba(255,255,255,0.08)',
                borderLeft: 'none',
                borderRight: 'none',
                borderTop: 'none',
                gap: 12,
              }}
            >
              <h1 style={{
                fontSize: 17,
                fontWeight: 600,
                color: '#FFFFFF',
                letterSpacing: '-0.01em',
                flexShrink: 0,
              }}>
                {tabLabels.chats}
              </h1>
              <ChatListToolbar />
            </header>
            <main className="flex-1 overflow-hidden">
              <ChatListContent />
            </main>
          </ChatListProvider>
        ) : (
          /* ── Desktop Non-Chat Tabs ── */
          <>
            <header
              className="flex items-center justify-between px-4 shrink-0"
              style={{
                height: desktopHeaderHeight,
                background: '#FFFFFF',
                borderBottom: '1px solid rgba(0,0,0,0.08)',
                borderLeft: 'none',
                borderRight: 'none',
                borderTop: 'none',
              }}
            >
              <h1 style={{
                fontSize: 17,
                fontWeight: 600,
                color: '#1C2733',
                letterSpacing: '-0.01em',
              }}>
                {tabLabels[activeTab]}
              </h1>
              <div className="flex items-center gap-1">
                <button
                  className="flex items-center justify-center"
                  style={{
                    width: 36, height: 36, borderRadius: '50%',
                    border: 'none', background: 'transparent',
                    color: '#708499',
                    cursor: 'pointer',
                  }}
                >
                  <Search className="w-[18px] h-[18px]" />
                </button>
                <button
                  className="flex items-center justify-center"
                  style={{
                    width: 36, height: 36, borderRadius: '50%',
                    border: 'none', background: 'transparent',
                    color: '#708499',
                    cursor: 'pointer',
                  }}
                >
                  <svg className="w-[18px] h-[18px]" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M12 20h9" /><path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" />
                  </svg>
                </button>
              </div>
            </header>
            <main className="flex-1 overflow-hidden">
              {renderContent()}
            </main>
          </>
        )}
      </div>

      {/* ── Right Panel: Chat detail or Contact detail (desktop) ── */}
      {(showChatDetail && activeTab === 'chats') && (
        <div className="flex-1 min-w-0 animate-fade-in">
          <ChatDetail />
        </div>
      )}
      {showContactDetail && (
        <div className="flex-1 min-w-0 animate-fade-in">
          <ContactDetailPanel contact={selectedContact} />
        </div>
      )}
      {showFriendRequestPanel && (
        <div className="flex-1 min-w-0 animate-fade-in">
          <FriendRequestList />
        </div>
      )}
      {showGroupPanelView && (
        <div className="flex-1 min-w-0 animate-fade-in">
          <GroupList />
        </div>
      )}
      {showTrendDetail && (
        <div className="flex-1 min-w-0 animate-fade-in">
          <TrendDetailPanel />
        </div>
      )}
      {activeTab === 'me' && meSubPage === 'profile' && (
        <div className="flex-1 min-w-0 animate-fade-in">
          <MyProfileEditPage onBack={() => setMeSubPage(null)} />
        </div>
      )}
      {activeTab === 'me' && meSubPage === 'favorites' && (
        <div className="flex-1 min-w-0 animate-fade-in">
          <FavoritesPage onBack={() => setMeSubPage(null)} />
        </div>
      )}
      {activeTab === 'me' && meSubPage === 'album' && (
        <div className="flex-1 min-w-0 animate-fade-in">
          <AlbumPage onBack={() => setMeSubPage(null)} />
        </div>
      )}
      {activeTab === 'me' && meSubPage === 'emojis' && (
        <div className="flex-1 min-w-0 animate-fade-in">
          <EmojisPage onBack={() => setMeSubPage(null)} />
        </div>
      )}
      {activeTab === 'me' && meSubPage === 'settings' && (
        <div className="flex-1 min-w-0 animate-fade-in">
          <SettingsPage onBack={() => setMeSubPage(null)} />
        </div>
      )}
      {floatingCard}
    </div>
  );
}
