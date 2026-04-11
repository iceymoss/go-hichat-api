'use client';

import React, { useEffect, useCallback } from 'react';
import { type Contact, conversations } from '@/lib/mock-data';
import { useIMStore } from '@/lib/im-store';
import { toast } from 'sonner';
import UserProfileCard from './UserProfileCard';
import { ArrowLeft } from 'lucide-react';

interface ContactDetailPanelProps {
  contact: Contact;
}

export default function ContactDetailPanel({ contact }: ContactDetailPanelProps) {
  const { setActiveTab, setSelectedConversationId, setShowChatDetail, setSelectedContactId } = useIMStore();

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

  const handleSendMessage = () => {
    const conv = conversations.find(c => c.name === contact.name);
    if (conv) {
      setSelectedContactId(null);
      setActiveTab('chats');
      setSelectedConversationId(conv.id);
      setShowChatDetail(true);
    } else {
      toast.info(`暂无与 ${contact.name} 的会话`);
    }
  };

  const handleVoiceCall = () => {
    toast('功能开发中');
  };

  const handleVideoCall = () => {
    toast('功能开发中');
  };

  const handleRecommend = useCallback(() => {
    toast('功能开发中');
  }, []);

  const handleSetPermissions = useCallback(() => {
    toast('功能开发中');
  }, []);

  const handleToggleBlock = useCallback((blocked: boolean) => {
    toast(blocked ? '已加入黑名单' : '已移出黑名单');
  }, []);

  const handleReport = useCallback(() => {
    toast('功能开发中');
  }, []);

  const handleDeleteFriend = useCallback(() => {
    toast('功能开发中');
  }, []);

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
            color: '#3390EC',
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
