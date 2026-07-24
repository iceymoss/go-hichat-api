'use client';

import { useState, useEffect } from 'react';
import { useIMStore } from '@/lib/im-store';
import { useChatStore } from '@/lib/chat-store';
import IMLayout from '@/components/im/IMLayout';
import AuthPage from '@/components/auth/AuthPage';
import SettingsProvider from '@/components/SettingsProvider';

// Refresh the persisted currentUser with the latest profile so views that
// read name/avatar/introduction/momentsCover don't show stale or empty data.
async function refreshCurrentUser(token: string) {
  try {
    const resp = await fetch('/api/user/detail', {
      headers: { Authorization: `Bearer ${token}` },
    });
    const json = await resp.json();
    if (json?.success && json.data && useIMStore.getState().currentUser?.token === token) {
      useIMStore.getState().updateCurrentUser(json.data);
    }
  } catch { /* ignore */ }
}

export default function Home() {
  const { isAuthenticated, currentUser } = useIMStore();
  const initWs = useChatStore(s => s.initWs);
  const destroyWs = useChatStore(s => s.destroyWs);
  const fetchConversations = useChatStore(s => s.fetchConversations);
  const refreshFriends = useIMStore(s => s.refreshFriends);
  const [hydrated, setHydrated] = useState(false);

  const fetchFriendsAndConversations = async (token: string) => {
    await refreshCurrentUser(token);
    await refreshFriends();
    await fetchConversations(token);
  };

  // Wait for zustand persist to rehydrate from localStorage before rendering
  useEffect(() => {
    setHydrated(true);
  }, []);

  // 登录后立即初始化 WebSocket + 拉取好友列表和会话列表
  useEffect(() => {
    if (isAuthenticated && currentUser?.token && currentUser?.id) {
      initWs(currentUser.token, currentUser.id);
      // 先加载好友列表，再加载会话列表（会话名称依赖好友数据）
      fetchFriendsAndConversations(currentUser.token);
    }
    return () => { destroyWs(); };
  }, [isAuthenticated, currentUser?.token, currentUser?.id, initWs, destroyWs, fetchConversations, refreshFriends]);

  if (!hydrated) {
    return null;
  }

  if (!isAuthenticated) {
    return <AuthPage />;
  }

  return (
    <SettingsProvider>
      {/* h-full so the app follows the zoom-compensated height of .hc-app
          (SettingsProvider). Using h-dvh here would ignore that compensation and
          leave a blank strip at the bottom when the font-size zoom is not 1. */}
      <main className="h-full overflow-hidden">
        <IMLayout />
      </main>
    </SettingsProvider>
  );
}
