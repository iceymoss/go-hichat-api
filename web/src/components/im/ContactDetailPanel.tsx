'use client';

import React, { useEffect, useCallback } from 'react';
import { type Contact } from '@/lib/mock-data';
import { useIMStore } from '@/lib/im-store';
import { useChatStore } from '@/lib/chat-store';
import { toast } from 'sonner';
import UserProfileCard from './UserProfileCard';
import { ArrowLeft } from 'lucide-react';

interface ContactDetailPanelProps {
  contact: Contact;
}

export default function ContactDetailPanel({ contact }: ContactDetailPanelProps) {
  const { setActiveTab, setSelectedConversationId, setShowChatDetail, setSelectedContactId, currentUser, invalidateFriends } = useIMStore();
  const token = currentUser?.token;

  const handleClose = useCallback(() => {
    setSelectedContactId(null);
  }, [setSelectedContactId]);

  // Close on Escape
  useEffect(() => {
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') handleClose();
    };
    window.addEventListener('keydown', handleKey);
    return () => window.removeEventListener('keydown', handleKey);
  }, [handleClose]);

  const handleSendMessage = async () => {
    if (!currentUser?.token || !currentUser?.id) return;
    try {
      const conv = await useChatStore.getState().getOrCreateConversation(
        currentUser.token, currentUser.id, contact.id
      );
      setSelectedContactId(null);
      setActiveTab('chats');
      setSelectedConversationId(conv.id);
      setShowChatDetail(true);
    } catch {
      toast.error('打开会话失败');
    }
  };

  const handleVoiceCall = () => {
    toast('功能开发中');
  };

  const handleVideoCall = () => {
    toast('功能开发中');
  };

  const handleRecommend = useCallback(() => {
    if (!token) return;
    fetch('/api/social/friend/share', {
      method: 'POST',
      headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ friend_uid: contact.id }),
    })
      .then(r => r.json())
      .then(d => {
        if (d.success) toast.success('分享成功');
        else toast.error(d.message || '分享失败');
      })
      .catch(() => toast.error('分享失败'));
  }, [token, contact.id]);

  const handleSetPermissions = useCallback(() => {
    toast('功能开发中');
  }, []);

  const handleToggleBlock = useCallback((blocked: boolean) => {
    if (!token) return;
    fetch('/api/social/friend/block', {
      method: 'POST',
      headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ friend_uid: contact.id, block: blocked }),
    })
      .then(r => r.json())
      .then(d => {
        if (d.success) { toast.success(blocked ? '已加入黑名单' : '已移出黑名单'); invalidateFriends(); }
        else toast.error(d.message || '操作失败');
      })
      .catch(() => toast.error('操作失败'));
  }, [token, contact.id, invalidateFriends]);

  const handleReport = useCallback(() => {
    if (!token) return;
    fetch('/api/social/friend/report', {
      method: 'POST',
      headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ friend_uid: contact.id, reason: '' }),
    })
      .then(r => r.json())
      .then(d => {
        if (d.success) toast.success('举报已提交');
        else toast.error(d.message || '举报失败');
      })
      .catch(() => toast.error('举报失败'));
  }, [token, contact.id]);

  const handleDeleteFriend = useCallback(() => {
    if (!token) return;
    fetch('/api/social/friend/delete', {
      method: 'POST',
      headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ friend_uid: contact.id }),
    })
      .then(r => r.json())
      .then(d => {
        if (d.success) {
          toast.success('已删除好友');
          setSelectedContactId(null);
          invalidateFriends();
        } else {
          toast.error(d.message || '删除失败');
        }
      })
      .catch(() => toast.error('删除失败'));
  }, [token, contact.id, setSelectedContactId, invalidateFriends]);

  return (
    <div
      className="h-full flex flex-col"
      style={{ backgroundColor: '#F5F7FA' }}
    >
      {/* ── Panel Header ── */}
      <header
        className="flex items-center shrink-0"
        style={{
          height: 56,
          paddingLeft: 12,
          paddingRight: 16,
          background: '#FFFFFF',
          borderBottom: '1px solid rgba(0,0,0,0.08)',
        }}
      >
        <button
          onClick={handleClose}
          className="shrink-0 flex items-center justify-center"
          style={{
            width: 36,
            height: 36,
            borderRadius: 10,
            border: 'none',
            background: 'transparent',
            color: '#1BB45B',
            cursor: 'pointer',
          }}
        >
          <ArrowLeft style={{ width: 20, height: 20 }} />
        </button>
        <h2
          style={{
            fontSize: 15,
            fontWeight: 600,
            color: '#1C2733',
            marginLeft: 8,
          }}
        >
          联系人详情
        </h2>
      </header>

      {/* ── Content Area: max-width 600px, centered, padding 40px ── */}
      <div
        className="flex-1 overflow-y-auto im-scroll"
      >
        <div
          style={{
            maxWidth: 600,
            margin: '0 auto',
            padding: '40px 20px',
          }}
        >
          <UserProfileCard
            contact={contact}
            isStranger={false}
            isBlocked={contact.blacklisted}
            onSendMessage={handleSendMessage}
            onVoiceCall={handleVoiceCall}
            onVideoCall={handleVideoCall}
            onRecommend={handleRecommend}
            onSetPermissions={handleSetPermissions}
            onToggleBlock={handleToggleBlock}
            onReport={handleReport}
            onDeleteFriend={handleDeleteFriend}
          />
        </div>
      </div>
    </div>
  );
}
