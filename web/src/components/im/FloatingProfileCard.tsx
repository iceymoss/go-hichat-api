'use client';

import React, { useEffect, useCallback } from 'react';
import { type Contact } from '@/lib/mock-data';
import { toast } from 'sonner';
import UserProfileCard from './UserProfileCard';
import { X, User } from 'lucide-react';

interface FloatingProfileCardProps {
  contact: Contact;
  isStranger?: boolean;
  isBlocked?: boolean;
  zIndex?: number;
  onClose: () => void;
  /** Optional: jump to the full contact detail page (contacts tab). Renders a footer entry. */
  onViewProfile?: () => void;
  onSendMessage?: () => void;
  onVoiceCall?: () => void;
  onVideoCall?: () => void;
  onAddFriend?: () => void;
  onRecommend?: () => void;
  onSetPermissions?: () => void;
  onToggleBlock?: (blocked: boolean) => void;
  onReport?: () => void;
  onDeleteFriend?: () => void;
}

export default function FloatingProfileCard({
  contact,
  isStranger = false,
  isBlocked = false,
  zIndex = 9998,
  onClose,
  onViewProfile,
  onSendMessage,
  onVoiceCall,
  onVideoCall,
  onAddFriend,
  onRecommend,
  onSetPermissions,
  onToggleBlock,
  onReport,
  onDeleteFriend,
}: FloatingProfileCardProps) {
  // Close on Escape
  useEffect(() => {
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', handleKey);
    return () => window.removeEventListener('keydown', handleKey);
  }, [onClose]);

  const handleSendMessage = useCallback(() => {
    onSendMessage?.();
    onClose();
  }, [onSendMessage, onClose]);

  const handleVoiceCall = useCallback(() => {
    onVoiceCall?.();
    onClose();
  }, [onVoiceCall, onClose]);

  const handleVideoCall = useCallback(() => {
    onVideoCall?.();
    onClose();
  }, [onVideoCall, onClose]);

  const handleRecommend = useCallback(() => {
    toast('功能开发中');
  }, []);

  const handleSetPermissions = useCallback(() => {
    toast('功能开发中');
  }, []);

  const handleReport = useCallback(() => {
    toast('功能开发中');
  }, []);

  const handleDeleteFriend = useCallback(() => {
    onDeleteFriend?.();
    onClose();
  }, [onDeleteFriend, onClose]);

  const handleToggleBlock = useCallback((blocked: boolean) => {
    onToggleBlock?.(blocked);
  }, [onToggleBlock]);

  return (
    <div
      className="animate-fade-in"
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: zIndex,
        background: 'rgba(0,0,0,0.4)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: 20,
      }}
      onClick={onClose}
    >
      <div
        style={{
          width: 480,
          maxWidth: 'calc(100vw - 40px)',
          backgroundColor: '#FFFFFF',
          borderRadius: 16,
          boxShadow: '0 4px 24px rgba(0,0,0,0.15)',
          position: 'relative',
          padding: '28px',
          animation: 'fade-in 0.15s ease-out',
        }}
        onClick={(e) => e.stopPropagation()}
      >
        {/* Close button - top right */}
        <button
          onClick={onClose}
          style={{
            position: 'absolute',
            top: 12,
            right: 12,
            width: 28,
            height: 28,
            borderRadius: '50%',
            border: 'none',
            background: 'transparent',
            color: '#8F959E',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            transition: 'all 0.15s',
            zIndex: 1,
          }}
          onMouseEnter={(e) => {
            (e.currentTarget as HTMLButtonElement).style.background = 'rgba(0,0,0,0.06)';
            (e.currentTarget as HTMLButtonElement).style.color = '#646A73';
          }}
          onMouseLeave={(e) => {
            (e.currentTarget as HTMLButtonElement).style.background = 'transparent';
            (e.currentTarget as HTMLButtonElement).style.color = '#8F959E';
          }}
        >
          <X size={18} />
        </button>

        <UserProfileCard
          contact={contact}
          isStranger={isStranger}
          isBlocked={isBlocked}
          onSendMessage={handleSendMessage}
          onVoiceCall={handleVoiceCall}
          onVideoCall={handleVideoCall}
          onAddFriend={onAddFriend}
          onRecommend={onRecommend || handleRecommend}
          onSetPermissions={onSetPermissions || handleSetPermissions}
          onToggleBlock={handleToggleBlock}
          onReport={onReport || handleReport}
          onDeleteFriend={onDeleteFriend || handleDeleteFriend}
        />

        {onViewProfile && (
          <button
            onClick={onViewProfile}
            style={{
              marginTop: 16,
              width: '100%',
              height: 40,
              borderRadius: 8,
              background: 'transparent',
              color: '#3390EC',
              fontSize: 13,
              fontWeight: 600,
              border: '1px solid rgba(51,144,236,0.4)',
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              gap: 6,
              transition: 'background 0.2s',
            }}
            onMouseEnter={(e) => { (e.currentTarget as HTMLButtonElement).style.background = 'rgba(51,144,236,0.06)'; }}
            onMouseLeave={(e) => { (e.currentTarget as HTMLButtonElement).style.background = 'transparent'; }}
          >
            <User size={14} />
            查看完整资料
          </button>
        )}
      </div>
    </div>
  );
}
