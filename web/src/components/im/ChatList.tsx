'use client';

import React, { useState, useEffect, useMemo, useCallback, createContext, useContext } from 'react';
import {
  contacts,
  conversationMessagesMap,
  formatTime,
  type Conversation,
  type Contact,
  type Message,
} from '@/lib/mock-data';
import { useIMStore } from '@/lib/im-store';
import { useChatStore } from '@/lib/chat-store';
import { useIsMobile } from '@/hooks/use-mobile';
import {
  Search,
  Pencil,
  X,
  UserPlus,
  CheckCheck,
  Pin,
  Trash2,
  BellOff,
} from 'lucide-react';
import { getAvatarColor } from '@/lib/utils';
import { useT } from '@/hooks/use-i18n';
import AddFriendPanel from './AddFriendPanel';
import NotificationCenter from './NotificationCenter';
import { toast } from 'sonner';

/* ═══════════════════════════════════════
   Types
   ═══════════════════════════════════════ */

interface SearchResultMessage {
  conversationId: string;
  conversationName: string;
  message: Message;
}

/* ═══════════════════════════════════════
   ChatList Context — shared state between toolbar and content
   ═══════════════════════════════════════ */

interface ChatListContextType {
  localSearch: string;
  setLocalSearch: React.Dispatch<React.SetStateAction<string>>;
  editMode: boolean;
  setEditMode: React.Dispatch<React.SetStateAction<boolean>>;
  selectedIds: Set<string>;
  setSelectedIds: React.Dispatch<React.SetStateAction<Set<string>>>;
  localConversations: Conversation[];
  setLocalConversations: React.Dispatch<React.SetStateAction<Conversation[]>>;
  deleteConfirm: boolean;
  setDeleteConfirm: React.Dispatch<React.SetStateAction<boolean>>;
}

const ChatListContext = createContext<ChatListContextType | null>(null);

export function useChatListContext() {
  const ctx = useContext(ChatListContext);
  if (!ctx) throw new Error('useChatListContext must be used within ChatListProvider');
  return ctx;
}

/* ═══════════════════════════════════════
   ChatList Provider — manages all shared state
   ═══════════════════════════════════════ */

export function ChatListProvider({ children }: { children: React.ReactNode }) {
  const [localSearch, setLocalSearch] = useState('');
  const [editMode, setEditMode] = useState(false);
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const chatStoreConvs = useChatStore(s => s.conversations);
  const [localConversations, setLocalConversations] = useState<Conversation[]>(() => chatStoreConvs);
  const [deleteConfirm, setDeleteConfirm] = useState(false);

  const syncFromStore = useCallback(() => {
    if (chatStoreConvs.length > 0) {
      setLocalConversations(chatStoreConvs);
    }
  }, [chatStoreConvs]);

  useEffect(() => {
    syncFromStore();
  }, [syncFromStore]);

  return (
    <ChatListContext.Provider value={{
      localSearch, setLocalSearch,
      editMode, setEditMode,
      selectedIds, setSelectedIds,
      localConversations, setLocalConversations,
      deleteConfirm, setDeleteConfirm,
    }}>
      {children}
    </ChatListContext.Provider>
  );
}

/* ═══════════════════════════════════════
   Keyword Highlight Helper
   ═══════════════════════════════════════ */

function HighlightText({ text, query }: { text: string; query: string }) {
  if (!query) return <>{text}</>;
  const q = query.toLowerCase();
  const idx = text.toLowerCase().indexOf(q);
  if (idx === -1) return <>{text}</>;
  return (
    <>
      {text.slice(0, idx)}
      <span style={{ color: '#F59E0B', fontWeight: 600 }}>{text.slice(idx, idx + query.length)}</span>
      {text.slice(idx + query.length)}
    </>
  );
}

/* ═══════════════════════════════════════
   Custom Confirm Dialog (TG dark overlay)
   ═══════════════════════════════════════ */

function ConfirmDialog({
  message,
  onConfirm,
  onCancel,
}: {
  message: string;
  onConfirm: () => void;
  onCancel: () => void;
}) {
  const t = useT();
  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 9999,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'rgba(0,0,0,0.6)',
      }}
      onClick={onCancel}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        style={{
          background: '#2C3E50',
          borderRadius: 14,
          padding: '24px 20px 16px',
          width: 320,
          maxWidth: '90vw',
          boxShadow: '0 8px 32px rgba(0,0,0,0.4)',
        }}
      >
        <p
          style={{
            color: '#FFFFFF',
            fontSize: 15,
            lineHeight: 1.6,
            marginBottom: 20,
            textAlign: 'center',
          }}
        >
          {message}
        </p>
        <div style={{ display: 'flex', gap: 10 }}>
          <button
            onClick={onCancel}
            style={{
              flex: 1,
              padding: '10px 0',
              borderRadius: 10,
              border: '1px solid rgba(255,255,255,0.15)',
              background: 'transparent',
              color: 'rgba(255,255,255,0.7)',
              fontSize: 14,
              fontWeight: 500,
              cursor: 'pointer',
              transition: 'background 0.15s',
            }}
            onMouseEnter={(e) => {
              (e.currentTarget as HTMLElement).style.background = 'rgba(255,255,255,0.06)';
            }}
            onMouseLeave={(e) => {
              (e.currentTarget as HTMLElement).style.background = 'transparent';
            }}
          >
            {t('common.cancel')}
          </button>
          <button
            onClick={onConfirm}
            style={{
              flex: 1,
              padding: '10px 0',
              borderRadius: 10,
              border: 'none',
              background: '#E53935',
              color: '#FFFFFF',
              fontSize: 14,
              fontWeight: 500,
              cursor: 'pointer',
              transition: 'background 0.15s',
            }}
            onMouseEnter={(e) => {
              (e.currentTarget as HTMLElement).style.background = '#C62828';
            }}
            onMouseLeave={(e) => {
              (e.currentTarget as HTMLElement).style.background = '#E53935';
            }}
          >
            {t('chatlist.confirmDelete')}
          </button>
        </div>
      </div>
    </div>
  );
}

/* ═══════════════════════════════════════
   Conversation Item (with optional checkbox)
   ═══════════════════════════════════════ */

function ConversationItem({
  conversation,
  isActive,
  isMobile,
  onClick,
  editMode,
  editSelected,
  onToggleSelect,
  onDelete,
  onTogglePin,
  onToggleMute,
}: {
  conversation: Conversation;
  isActive: boolean;
  isMobile: boolean;
  onClick: () => void;
  editMode?: boolean;
  editSelected?: boolean;
  onToggleSelect?: () => void;
  onDelete?: () => void;
  onTogglePin?: () => void;
  onToggleMute?: () => void;
}) {
  const avatarSize = 48;
  const showCheckbox = editMode;
  const [ctxMenu, setCtxMenu] = useState<{ x: number; y: number } | null>(null);
  const t = useT();

  // 关闭右键菜单
  useEffect(() => {
    if (!ctxMenu) return;
    const close = () => setCtxMenu(null);
    window.addEventListener('click', close);
    return () => window.removeEventListener('click', close);
  }, [ctxMenu]);

  return (
    <div
      className={`im-conversation-item ${isActive ? 'active' : ''}`}
      onClick={editMode ? onToggleSelect : onClick}
      onContextMenu={(e) => {
        if (editMode) return;
        e.preventDefault();
        setCtxMenu({ x: e.clientX, y: e.clientY });
      }}
      style={{
        transition: 'background-color 0.15s',
        position: 'relative',
      }}
    >
      {/* 右键菜单 */}
      {ctxMenu && (
        <div
          style={{
            position: 'fixed', left: ctxMenu.x, top: ctxMenu.y, zIndex: 9999,
            background: '#2C3E50', borderRadius: 10, padding: '4px 0',
            boxShadow: '0 4px 20px rgba(0,0,0,0.4)', minWidth: 140,
            border: '1px solid rgba(255,255,255,0.1)',
          }}
          onClick={(e) => e.stopPropagation()}
        >
          {[
            { label: conversation.pinned ? t('chatlist.unpin') : t('chatlist.pin'), action: onTogglePin },
            { label: conversation.muted ? t('chatlist.unmute') : t('chatlist.mute'), action: onToggleMute },
            { label: t('chatlist.deleteConv'), action: onDelete, danger: true },
          ].map(({ label, action, danger }) => (
            <button
              key={label}
              onClick={() => { setCtxMenu(null); action?.(); }}
              style={{
                display: 'block', width: '100%', padding: '8px 16px', border: 'none',
                background: 'transparent', color: danger ? '#E53935' : '#FFFFFF',
                fontSize: 13, textAlign: 'left', cursor: 'pointer',
              }}
              onMouseEnter={(e) => { (e.target as HTMLElement).style.background = 'rgba(255,255,255,0.08)'; }}
              onMouseLeave={(e) => { (e.target as HTMLElement).style.background = 'transparent'; }}
            >
              {label}
            </button>
          ))}
        </div>
      )}
      {/* Checkbox (edit mode) */}
      {showCheckbox && (
        <div
          style={{
            width: 22,
            height: 22,
            borderRadius: '50%',
            border: editSelected ? 'none' : '2px solid rgba(255,255,255,0.35)',
            background: editSelected ? '#1BB45B' : 'transparent',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            flexShrink: 0,
            transition: 'all 0.15s',
          }}
        >
          {editSelected && (
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
              <polyline points="20 6 9 17 4 12" />
            </svg>
          )}
        </div>
      )}

      {/* Avatar */}
      <div className="relative shrink-0">
        <div
          style={{
            width: avatarSize,
            height: avatarSize,
            borderRadius: '50%',
            background: conversation.avatar ? 'transparent' : (isActive && !editMode ? 'rgba(255,255,255,0.2)' : getAvatarColor(conversation.name)),
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: '#FFFFFF',
            fontSize: 18,
            fontWeight: 600,
            flexShrink: 0,
            overflow: 'hidden',
          }}
        >
          {conversation.avatar
            ? <img src={conversation.avatar} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
            : (conversation.name?.[0] || '?')}
        </div>
        {/* Online indicator */}
        {conversation.online && conversation.type === 'private' && (
          <span
            className="absolute rounded-full"
            style={{
              bottom: 0,
              right: 0,
              width: 12,
              height: 12,
              background: '#4DCD5E',
              border: `2.5px solid ${isActive && !editMode ? '#1BB45B' : '#2C3E50'}`,
            }}
          />
        )}
      </div>

      {/* Content */}
      <div className="flex-1 min-w-0" style={{ display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
        {/* Top: name + time */}
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 3 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 6, minWidth: 0, flex: 1 }}>
            {conversation.pinned && (
              <svg
                className="conv-pin shrink-0"
                style={{ width: 13, height: 13, color: isActive && !editMode ? undefined : 'rgba(255,255,255,0.35)' }}
                viewBox="0 0 24 24"
                fill="currentColor"
              >
                <path d="M16 12V4h1V2H7v2h1v8l-2 2v2h5.2v6h1.6v-6H18v-2l-2-2z" />
              </svg>
            )}
            <span
              className="conv-name truncate"
              style={{
                fontSize: 14,
                fontWeight: 600,
                color: isActive && !editMode ? undefined : '#FFFFFF',
                lineHeight: '20px',
              }}
            >
              {conversation.name}
            </span>
            {conversation.muted && (
              <svg
                className="conv-mute shrink-0"
                style={{ width: 14, height: 14, color: isActive && !editMode ? undefined : 'rgba(255,255,255,0.35)' }}
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
              >
                <path d="M11 5 6 9H2v6h4l5 4V5z" />
                <path d="M15.54 8.46a5 5 0 0 1 0 7.07" />
                <path d="M19.07 4.93a10 10 0 0 1 0 14.14" />
              </svg>
            )}
          </div>
          <span
            className="conv-time shrink-0"
            style={{
              fontSize: 12,
              color: isActive && !editMode ? undefined : 'rgba(255,255,255,0.4)',
              marginLeft: 8,
              lineHeight: '16px',
              whiteSpace: 'nowrap',
            }}
          >
            {formatTime(conversation.lastMessageTime)}
          </span>
        </div>

        {/* Bottom: last message + unread */}
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <span
            className="conv-message truncate"
            style={{
              fontSize: 13,
              color: isActive && !editMode ? undefined : 'rgba(255,255,255,0.5)',
              lineHeight: '18px',
              paddingRight: 8,
              display: 'block',
              maxWidth: '100%',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
            }}
          >
            {conversation.hasAtMe && (
              <span style={{ color: '#FA5151', fontWeight: 600, marginRight: 4 }}>{t('chatlist.atMe')}</span>
            )}
            {conversation.lastMessage}
          </span>
          {conversation.unreadCount > 0 && (
            conversation.muted ? (
              /* 免打扰：仅显示小灰点，不显示数字气泡 */
              <span
                className="conv-unread-dot shrink-0"
                style={{
                  width: 8,
                  height: 8,
                  borderRadius: 4,
                  background: 'rgba(255,255,255,0.4)',
                }}
              />
            ) : (
              <span
                className="conv-unread shrink-0"
                style={{
                  minWidth: 20,
                  height: 20,
                  padding: '0 6px',
                  borderRadius: 10,
                  background: isActive && !editMode ? undefined : '#1BB45B',
                  color: '#FFFFFF',
                  fontSize: 11,
                  fontWeight: 700,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  lineHeight: 1,
                  whiteSpace: 'nowrap',
                }}
              >
                {conversation.unreadCount > 99 ? '99+' : conversation.unreadCount}
              </span>
            )
          )}
        </div>
      </div>
    </div>
  );
}

/* ═══════════════════════════════════════
   Search Result Item (conversation match)
   ═══════════════════════════════════════ */

function SearchConversationItem({
  conv,
  query,
  onClick,
}: {
  conv: Conversation;
  query: string;
  onClick: () => void;
}) {
  return (
    <div className="im-conversation-item" onClick={onClick}>
      <div className="relative shrink-0">
        <div
          style={{
            width: 44,
            height: 44,
            borderRadius: '50%',
            background: conv.avatar ? 'transparent' : getAvatarColor(conv.name),
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: '#FFFFFF',
            fontSize: 16,
            fontWeight: 600,
            flexShrink: 0,
            overflow: 'hidden',
          }}
        >
          {conv.avatar
            ? <img src={conv.avatar} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
            : (conv.name?.[0] || '?')}
        </div>
      </div>
      <div className="flex-1 min-w-0">
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <span style={{ fontSize: 14, fontWeight: 600, color: '#FFFFFF' }}>
            <HighlightText text={conv.name} query={query} />
          </span>
          <span style={{ fontSize: 11, color: 'rgba(255,255,255,0.35)', marginLeft: 8, whiteSpace: 'nowrap' }}>
            {formatTime(conv.lastMessageTime)}
          </span>
        </div>
        <span
          style={{
            fontSize: 13,
            color: 'rgba(255,255,255,0.45)',
            display: 'block',
            maxWidth: '100%',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap',
          }}
        >
          <HighlightText text={conv.lastMessage} query={query} />
        </span>
      </div>
    </div>
  );
}

/* ═══════════════════════════════════════
   Search Result Item (contact match)
   ═══════════════════════════════════════ */

function SearchContactItem({
  contact,
  query,
  onClick,
}: {
  contact: Contact;
  query: string;
  onClick: () => void;
}) {
  return (
    <div className="im-conversation-item" onClick={onClick}>
      <div className="relative shrink-0">
        <div
          style={{
            width: 44,
            height: 44,
            borderRadius: '50%',
            background: getAvatarColor(contact.name),
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: '#FFFFFF',
            fontSize: 16,
            fontWeight: 600,
            flexShrink: 0,
            overflow: 'hidden',
          }}
        >
          {contact.name[0]}
        </div>
        {contact.online && (
          <span
            className="absolute rounded-full"
            style={{
              bottom: 0,
              right: 0,
              width: 11,
              height: 11,
              background: '#4DCD5E',
              border: '2px solid #2C3E50',
            }}
          />
        )}
      </div>
      <div className="flex-1 min-w-0">
        <span style={{ fontSize: 14, fontWeight: 600, color: '#FFFFFF' }}>
          <HighlightText text={contact.name} query={query} />
        </span>
      </div>
    </div>
  );
}

/* ═══════════════════════════════════════
   Search Result Item (message match)
   ═══════════════════════════════════════ */

function SearchMessageItem({
  result,
  query,
  onClick,
}: {
  result: SearchResultMessage;
  query: string;
  onClick: () => void;
}) {
  return (
    <div className="im-conversation-item" onClick={onClick}>
      <div className="relative shrink-0">
        <div
          style={{
            width: 44,
            height: 44,
            borderRadius: '50%',
            background: getAvatarColor(result.conversationName),
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: '#FFFFFF',
            fontSize: 16,
            fontWeight: 600,
            flexShrink: 0,
            overflow: 'hidden',
          }}
        >
          {result.conversationName[0]}
        </div>
      </div>
      <div className="flex-1 min-w-0">
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <span style={{ fontSize: 13, fontWeight: 600, color: '#FFFFFF' }}>
            {result.conversationName}
          </span>
          <span style={{ fontSize: 11, color: 'rgba(255,255,255,0.3)', marginLeft: 8, whiteSpace: 'nowrap' }}>
            {formatTime(result.message.timestamp)}
          </span>
        </div>
        <span
          style={{
            fontSize: 13,
            color: 'rgba(255,255,255,0.45)',
            display: 'block',
            maxWidth: '100%',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap',
          }}
        >
          <HighlightText text={result.message.content} query={query} />
        </span>
      </div>
    </div>
  );
}

/* ═══════════════════════════════════════
   Search Section Header
   ═══════════════════════════════════════ */

function SearchSectionHeader({ title, count }: { title: string; count: number }) {
  return (
    <div
      style={{
        padding: '12px 14px 4px',
        fontSize: 12,
        fontWeight: 600,
        color: 'rgba(255,255,255,0.35)',
        letterSpacing: '0.02em',
      }}
    >
      {title} ({count})
    </div>
  );
}

/* ═══════════════════════════════════════
   Floating Action Bar (batch edit mode)
   ═══════════════════════════════════════ */

function FloatingActionBar({
  selectedCount,
  onMarkRead,
  onTogglePin,
  onDelete,
  onToggleMute,
}: {
  selectedCount: number;
  onMarkRead: () => void;
  onTogglePin: () => void;
  onDelete: () => void;
  onToggleMute: () => void;
}) {
  const t = useT();
  return (
    <div
      style={{
        position: 'absolute',
        bottom: 0,
        left: 0,
        right: 0,
        background: '#2C3E50',
        borderTop: '1px solid rgba(255,255,255,0.08)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-around',
        padding: '12px 8px',
        zIndex: 20,
        boxShadow: '0 -2px 12px rgba(0,0,0,0.3)',
      }}
    >
      <ActionBarButton
        icon={<CheckCheck size={20} />}
        label={t('chatlist.markAllRead')}
        onClick={onMarkRead}
      />
      <ActionBarButton
        icon={<Pin size={20} />}
        label={t('chatlist.pin')}
        onClick={onTogglePin}
      />
      <ActionBarButton
        icon={<Trash2 size={20} />}
        label={t('chatlist.delete')}
        danger
        onClick={onDelete}
      />
      <ActionBarButton
        icon={<BellOff size={20} />}
        label={t('chatlist.mute')}
        onClick={onToggleMute}
      />
    </div>
  );
}

function ActionBarButton({
  icon,
  label,
  onClick,
  danger,
}: {
  icon: React.ReactNode;
  label: string;
  onClick: () => void;
  danger?: boolean;
}) {
  const color = danger ? '#E53935' : '#FFFFFF';

  return (
    <button
      onClick={onClick}
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: 4,
        background: 'transparent',
        border: 'none',
        cursor: 'pointer',
        padding: '4px 8px',
        borderRadius: 8,
        transition: 'background 0.15s',
        color,
      }}
      onMouseEnter={(e) => {
        (e.currentTarget as HTMLElement).style.background = 'rgba(255,255,255,0.06)';
      }}
      onMouseLeave={(e) => {
        (e.currentTarget as HTMLElement).style.background = 'transparent';
      }}
    >
      {icon}
      <span style={{ fontSize: 11, fontWeight: 500, whiteSpace: 'nowrap' }}>{label}</span>
    </button>
  );
}

/* ═══════════════════════════════════════
   ChatListToolbar — rendered in IMLayout header
   (pencil + search input + userplus)
   ═══════════════════════════════════════ */

export function ChatListToolbar() {
  const { localSearch, setLocalSearch, editMode, setEditMode, setSelectedIds } = useChatListContext();
  const [showAddFriend, setShowAddFriend] = useState(false);
  const t = useT();

  const enterEditMode = useCallback(() => {
    setEditMode(true);
    setSelectedIds(new Set());
  }, [setEditMode, setSelectedIds]);

  const exitEditMode = useCallback(() => {
    setEditMode(false);
    setSelectedIds(new Set());
  }, [setEditMode, setSelectedIds]);

  const buttonStyle: React.CSSProperties = {
    width: 36,
    height: 36,
    borderRadius: 10,
    border: 'none',
    background: 'transparent',
    color: 'rgba(255,255,255,0.7)',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    cursor: 'pointer',
    flexShrink: 0,
    transition: 'background 0.15s',
  };

  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 8, flex: 1, minWidth: 0 }}>
      {/* Left: Pen / X icon */}
      {editMode ? (
        <button
          onClick={exitEditMode}
          style={buttonStyle}
          onMouseEnter={(e) => {
            (e.currentTarget as HTMLElement).style.background = 'rgba(255,255,255,0.08)';
          }}
          onMouseLeave={(e) => {
            (e.currentTarget as HTMLElement).style.background = 'transparent';
          }}
        >
          <X size={20} />
        </button>
      ) : (
        <button
          onClick={enterEditMode}
          style={buttonStyle}
          onMouseEnter={(e) => {
            (e.currentTarget as HTMLElement).style.background = 'rgba(255,255,255,0.08)';
          }}
          onMouseLeave={(e) => {
            (e.currentTarget as HTMLElement).style.background = 'transparent';
          }}
        >
          <Pencil size={18} />
        </button>
      )}

      {/* Center: Search input / batch mode indicator */}
      <div style={{ flex: 1, position: 'relative' }}>
        {editMode ? (
          <div
            style={{
              width: '100%',
              height: 36,
              borderRadius: 20,
              background: 'rgba(255,255,255,0.06)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: 'rgba(255,255,255,0.5)',
              fontSize: 13,
              fontWeight: 500,
            }}
          >
            {t('chatlist.selectConv')}
          </div>
        ) : (
          <>
            <Search
              className="pointer-events-none"
              style={{
                position: 'absolute',
                left: 10,
                top: '50%',
                transform: 'translateY(-50%)',
                width: 16,
                height: 16,
                color: 'rgba(255,255,255,0.35)',
                pointerEvents: 'none',
              }}
            />
            <input
              value={localSearch}
              onChange={(e) => setLocalSearch(e.target.value)}
              placeholder={t('chatlist.searchPlaceholder')}
              className="outline-none"
              style={{
                width: '100%',
                height: 36,
                paddingLeft: 34,
                paddingRight: 12,
                borderRadius: 20,
                background: '#34495E',
                border: 'none',
                fontSize: 13,
                color: '#FFFFFF',
                boxShadow: 'none',
              }}
            />
          </>
        )}
      </div>

      {/* Right: Plus icon (hidden in batch mode) */}
      {!editMode && (
        <>
          <NotificationCenter />
          <button
            style={buttonStyle}
            onClick={() => setShowAddFriend(true)}
            onMouseEnter={(e) => {
              (e.currentTarget as HTMLElement).style.background = 'rgba(255,255,255,0.08)';
            }}
            onMouseLeave={(e) => {
              (e.currentTarget as HTMLElement).style.background = 'transparent';
            }}
          >
            <UserPlus size={20} />
          </button>
        </>
      )}

      <AddFriendPanel open={showAddFriend} onClose={() => setShowAddFriend(false)} />
    </div>
  );
}

/* ═══════════════════════════════════════
   ChatListContent — the list area (no toolbar)
   ═══════════════════════════════════════ */

export function ChatListContent() {
  const {
    localSearch,
    editMode,
    setEditMode,
    selectedIds,
    setSelectedIds,
    localConversations,
    setLocalConversations,
    deleteConfirm,
    setDeleteConfirm,
  } = useChatListContext();

  const { setSelectedConversationId, selectedConversationId } = useIMStore();
  const isMobile = useIsMobile();
  const t = useT();

  const searchQuery = localSearch.trim();

  // ── Search Results ──
  const { matchedConversations, matchedContacts, matchedMessages } = useMemo(() => {
    if (!searchQuery) {
      return { matchedConversations: [], matchedContacts: [], matchedMessages: [] };
    }
    const q = searchQuery.toLowerCase();

    // Match conversations by name or last message
    const mc = localConversations.filter(
      (c) => c.name.toLowerCase().includes(q) || c.lastMessage.toLowerCase().includes(q)
    );

    // Match contacts by name or pinyin
    const mcontacts = contacts.filter(
      (c) => c.name.toLowerCase().includes(q) || c.pinyin.toLowerCase().includes(q)
    );

    // Match messages from conversationMessagesMap
    const mm: SearchResultMessage[] = [];
    for (const [convId, msgs] of Object.entries(conversationMessagesMap)) {
      for (const msg of msgs) {
        if (msg.type === 'text' && msg.content.toLowerCase().includes(q)) {
          const conv = localConversations.find((c) => c.id === convId);
          if (conv) {
            mm.push({
              conversationId: convId,
              conversationName: conv.name,
              message: msg,
            });
          }
        }
      }
    }

    return { matchedConversations: mc, matchedContacts: mcontacts, matchedMessages: mm };
  }, [searchQuery, localConversations]);

  const hasSearchResults = matchedConversations.length > 0 || matchedContacts.length > 0 || matchedMessages.length > 0;

  // ── Sorted conversations (normal mode) ──
  const sortedConversations = useMemo(() => {
    return [...localConversations].sort((a, b) => {
      if (a.pinned && !b.pinned) return -1;
      if (!a.pinned && b.pinned) return 1;
      return b.lastMessageTime.getTime() - a.lastMessageTime.getTime();
    });
  }, [localConversations]);

  // ── Edit mode handlers ──
  const toggleSelect = useCallback((id: string) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  }, [setSelectedIds]);

  const handleMarkRead = useCallback(() => {
    // 走 store.clearUnread：① 更新 store conversations → 触发同步 effect 清空列表气泡，
    // 并让侧边栏“会话”tab 的总未读数（IMLayout 从 store 重算）一并扣减；② 同步后端已读数。
    const { clearUnread } = useChatStore.getState();
    selectedIds.forEach((id) => clearUnread(id));
    // 乐观更新本地副本，保证即时反馈（store 同步 effect 随后也会覆盖一致）
    setLocalConversations((prev) =>
      prev.map((c) => (selectedIds.has(c.id) ? { ...c, unreadCount: 0 } : c))
    );
    toast.success(t('chatlist.markedRead'));
  }, [selectedIds, setLocalConversations]);

  const handleTogglePin = useCallback(() => {
    const token = useIMStore.getState().currentUser?.token;
    const allSelectedPinned = Array.from(selectedIds).every(
      (id) => localConversations.find((conv) => conv.id === id)?.pinned
    );
    const nextPinned = !allSelectedPinned;
    setLocalConversations((prev) =>
      prev.map((c) => (selectedIds.has(c.id) ? { ...c, pinned: nextPinned } : c))
    );
    if (token) {
      selectedIds.forEach(id =>
        useChatStore.getState().setConversationSettings(token, id, { pinned: nextPinned })
      );
    }
    toast.success(selectedIds.size > 0 ? t('chatlist.pinUpdated') : '');
  }, [selectedIds, localConversations, setLocalConversations]);

  const handleDelete = useCallback(() => {
    setDeleteConfirm(true);
  }, [setDeleteConfirm]);

  const confirmDelete = useCallback(() => {
    const token = useIMStore.getState().currentUser?.token;
    if (token) {
      selectedIds.forEach(id => useChatStore.getState().deleteConversation(token, id));
    }
    setLocalConversations((prev) => prev.filter((c) => !selectedIds.has(c.id)));
    setSelectedIds(new Set());
    setDeleteConfirm(false);
    setEditMode(false);
    toast.success(t('chatlist.convDeleted'));
  }, [selectedIds, setLocalConversations, setSelectedIds, setDeleteConfirm, setEditMode]);

  const handleToggleMute = useCallback(() => {
    const token = useIMStore.getState().currentUser?.token;
    const allSelectedMuted = Array.from(selectedIds).every(
      (id) => localConversations.find((conv) => conv.id === id)?.muted
    );
    const nextMuted = !allSelectedMuted;
    setLocalConversations((prev) =>
      prev.map((c) => (selectedIds.has(c.id) ? { ...c, muted: nextMuted } : c))
    );
    if (token) {
      selectedIds.forEach(id =>
        useChatStore.getState().setConversationSettings(token, id, { muted: nextMuted })
      );
    }
    toast.success(t('chatlist.muteUpdated'));
  }, [selectedIds, localConversations, setLocalConversations]);

  // ── Handle clicking a contact → open their conversation ──
  const handleContactClick = useCallback(
    (contact: Contact) => {
      // 按 id 匹配 (conversationId 包含对方 userId)
      const conv = localConversations.find(
        (c) => c.type === 'private' && c.id.split('_').includes(contact.id)
      );
      if (conv) {
        setSelectedConversationId(conv.id);
      }
    },
    [localConversations, setSelectedConversationId]
  );

  const isSearchMode = searchQuery.length > 0;
  const isBatchMode = editMode;

  return (
    <div className="h-full flex flex-col" style={{ position: 'relative' }}>
      {/* ═══ Content Area ═══ */}
      <div className="flex-1 overflow-y-auto im-scroll" style={{ paddingBottom: isBatchMode && selectedIds.size > 0 ? 68 : 0 }}>
        {/* ── Search Mode: Categorized Results ── */}
        {isSearchMode ? (
          <div>
            {hasSearchResults ? (
              <>
                {matchedConversations.length > 0 && (
                  <>
                    <SearchSectionHeader title={t('chatlist.section.conv')} count={matchedConversations.length} />
                    {matchedConversations.map((conv) => (
                      <SearchConversationItem
                        key={conv.id}
                        conv={conv}
                        query={searchQuery}
                        onClick={() => setSelectedConversationId(conv.id)}
                      />
                    ))}
                  </>
                )}
                {matchedContacts.length > 0 && (
                  <>
                    <SearchSectionHeader title={t('chatlist.section.contact')} count={matchedContacts.length} />
                    {matchedContacts.map((contact) => (
                      <SearchContactItem
                        key={contact.id}
                        contact={contact}
                        query={searchQuery}
                        onClick={() => handleContactClick(contact)}
                      />
                    ))}
                  </>
                )}
                {matchedMessages.length > 0 && (
                  <>
                    <SearchSectionHeader title={t('chatlist.section.message')} count={matchedMessages.length} />
                    {matchedMessages.slice(0, 20).map((result, idx) => (
                      <SearchMessageItem
                        key={`${result.message.id}-${idx}`}
                        result={result}
                        query={searchQuery}
                        onClick={() => setSelectedConversationId(result.conversationId)}
                      />
                    ))}
                  </>
                )}
              </>
            ) : (
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  height: 120,
                  color: 'rgba(255,255,255,0.35)',
                  fontSize: 13,
                }}
              >
                {t('chatlist.noResults')}
              </div>
            )}
          </div>
        ) : (
          /* ── Normal / Batch Mode: Conversation List ── */
          <>
            {sortedConversations.map((conv) => (
              <ConversationItem
                key={conv.id}
                conversation={conv}
                isActive={conv.id === selectedConversationId && !isBatchMode}
                isMobile={isMobile}
                onClick={() => setSelectedConversationId(conv.id)}
                editMode={isBatchMode}
                editSelected={selectedIds.has(conv.id)}
                onToggleSelect={() => toggleSelect(conv.id)}
                onDelete={() => {
                  const token = useIMStore.getState().currentUser?.token;
                  if (token) useChatStore.getState().deleteConversation(token, conv.id);
                  setLocalConversations(prev => prev.filter(c => c.id !== conv.id));
                  if (selectedConversationId === conv.id) setSelectedConversationId(null);
                  toast.success(t('chatlist.convDeleted'));
                }}
                onTogglePin={() => {
                  const token = useIMStore.getState().currentUser?.token;
                  setLocalConversations(prev => prev.map(c =>
                    c.id === conv.id ? { ...c, pinned: !c.pinned } : c
                  ));
                  if (token) useChatStore.getState().setConversationSettings(token, conv.id, { pinned: !conv.pinned });
                }}
                onToggleMute={() => {
                  const token = useIMStore.getState().currentUser?.token;
                  setLocalConversations(prev => prev.map(c =>
                    c.id === conv.id ? { ...c, muted: !c.muted } : c
                  ));
                  if (token) useChatStore.getState().setConversationSettings(token, conv.id, { muted: !conv.muted });
                }}
              />
            ))}

            {sortedConversations.length === 0 && (
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  height: 120,
                  color: 'rgba(255,255,255,0.35)',
                  fontSize: 13,
                }}
              >
                {t('chatlist.empty')}
              </div>
            )}
          </>
        )}
      </div>

      {/* ═══ Floating Action Bar (batch mode, when items selected) ═══ */}
      {isBatchMode && selectedIds.size > 0 && (
        <FloatingActionBar
          selectedCount={selectedIds.size}
          onMarkRead={handleMarkRead}
          onTogglePin={handleTogglePin}
          onDelete={handleDelete}
          onToggleMute={handleToggleMute}
        />
      )}

      {/* ═══ Delete Confirm Dialog ═══ */}
      {deleteConfirm && (
        <ConfirmDialog
          message={t('chatlist.deleteConfirm').replace('{count}', String(selectedIds.size))}
          onConfirm={confirmDelete}
          onCancel={() => setDeleteConfirm(false)}
        />
      )}
    </div>
  );
}
