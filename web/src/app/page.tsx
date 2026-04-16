'use client';

import { useState, useEffect } from 'react';
import { useIMStore } from '@/lib/im-store';
import { useChatStore } from '@/lib/chat-store';
import IMLayout from '@/components/im/IMLayout';
import AuthPage from '@/components/auth/AuthPage';
import SettingsProvider from '@/components/SettingsProvider';

async function fetchFriendsForStore(token: string) {
  try {
    const resp = await fetch('/api/social/friends', {
      headers: { Authorization: `Bearer ${token}` },
    });
    const json = await resp.json();
    const list = json?.list || json?.data?.list || [];
    const mapped = list.map((f: any) => ({
      id: String(f.id || f.friend_uid || ''),
      friend_uid: String(f.friend_uid || ''),
      name: f.nickname || f.remark || String(f.friend_uid || ''),
      remark: f.remark || '',
      avatar: f.avatar || '',
      pinyin: '', letter: '',
    }));
    useIMStore.getState().setFriends(mapped);
  } catch { /* ignore */ }
}

export default function Home() {
  const { isAuthenticated, currentUser } = useIMStore();
  const initWs = useChatStore(s => s.initWs);
  const destroyWs = useChatStore(s => s.destroyWs);
  const fetchConversations = useChatStore(s => s.fetchConversations);
  const [hydrated, setHydrated] = useState(false);

  const fetchFriendsAndConversations = async (token: string) => {
    await fetchFriendsForStore(token);
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
  }, [isAuthenticated, currentUser?.token, currentUser?.id]);

  if (!hydrated) {
    return null;
  }

  if (!isAuthenticated) {
    return <AuthPage />;
  }

  return (
    <SettingsProvider>
      <main className="h-dvh overflow-hidden">
        <IMLayout />
      </main>
    </SettingsProvider>
  );
}
