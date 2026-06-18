'use client';

import React, { useEffect, useMemo, useRef, useState } from 'react';
import { Copy, Forward, Reply, RotateCcw, Trash2, Star } from 'lucide-react';
import type { Message } from '@/lib/mock-data';
import { useT } from '@/hooks/use-i18n';

interface MessageContextMenuProps {
  message: Message;
  senderName: string;
  isOwn: boolean;
  /** 是否展示「撤回」入口：本人消息，或群管理员/群主撤回他人消息。未传时回退到 isOwn */
  canRecall?: boolean;
  position: { x: number; y: number };
  onClose: () => void;
  onCopy: (msg: Message) => void;
  onForward: (msg: Message) => void;
  onReply: (msg: Message, senderName: string) => void;
  onRecall: (msgId: string) => void;
  onDelete: (msgId: string) => void;
  /** 把图片/表情消息收藏到「我的表情」（仅图片/表情类消息显示） */
  onSaveSticker?: (msg: Message) => void;
}

const MENU_ITEM_SIZE = 42;
const MENU_GAP = 4;
const MENU_PADDING = 6;
const MENU_Y_OFFSET = 10; // gap between menu and bubble

export default function MessageContextMenu({
  message,
  senderName,
  isOwn,
  canRecall,
  position,
  onClose,
  onCopy,
  onForward,
  onReply,
  onRecall,
  onDelete,
  onSaveSticker,
}: MessageContextMenuProps) {
  const [visible, setVisible] = useState(false);
  const t = useT();

  const canSaveSticker = (message.type === 'image' || message.type === 'memes') && !!onSaveSticker;
  // 撤回入口：优先用显式 canRecall（含管理员撤回他人消息），未传时回退到本人消息
  const showRecall = canRecall ?? isOwn;

  // Build menu items based on ownership
  const items = useMemo(() => [
    { key: 'copy', icon: <Copy size={20} />, label: t('msgmenu.copy'), color: '#FFFFFF', onClick: () => onCopy(message) },
    { key: 'forward', icon: <Forward size={20} />, label: t('msgmenu.forward'), color: '#FFFFFF', onClick: () => onForward(message) },
    { key: 'reply', icon: <Reply size={20} />, label: t('msgmenu.reply'), color: '#FFFFFF', onClick: () => onReply(message, senderName) },
    ...(canSaveSticker
      ? [{ key: 'sticker', icon: <Star size={20} />, label: t('msgmenu.sticker'), color: '#FFFFFF', onClick: () => onSaveSticker!(message) }]
      : []),
    ...(showRecall
      ? [{ key: 'recall', icon: <RotateCcw size={20} />, label: t('msgmenu.recall'), color: '#FFFFFF', onClick: () => onRecall(message.id) }]
      : []),
    { key: 'delete', icon: <Trash2 size={20} />, label: t('msgmenu.delete'), color: '#E53935', onClick: () => onDelete(message.id) },
  ], [t, showRecall, canSaveSticker, message, onCopy, onForward, onReply, onRecall, onDelete, onSaveSticker, senderName]);

  // Calculate total menu width
  const totalItemsWidth = items.length * MENU_ITEM_SIZE + (items.length - 1) * MENU_GAP;
  const menuWidth = totalItemsWidth + MENU_PADDING * 2;

  // Compute menu position (pure calculation, no state)
  const menuPosition = useMemo(() => {
    if (typeof window === 'undefined') return { left: 0, top: 0 };

    const menuHeight = MENU_ITEM_SIZE + MENU_PADDING * 2;
    const vw = window.innerWidth;

    let left = position.x - menuWidth / 2;
    let top = position.y - menuHeight - MENU_Y_OFFSET;

    // Clamp horizontal
    if (left < 8) left = 8;
    if (left + menuWidth > vw - 8) left = vw - menuWidth - 8;

    // If near top, show below
    if (top < 8) {
      top = position.y + MENU_Y_OFFSET;
    }

    return { left, top };
  }, [position, menuWidth]);

  // Entrance animation trigger
  useEffect(() => {
    const id = requestAnimationFrame(() => setVisible(true));
    return () => cancelAnimationFrame(id);
  }, []);

  // Escape key handler
  useEffect(() => {
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', handleKey);
    return () => window.removeEventListener('keydown', handleKey);
  }, [onClose]);

  return (
    <div
      className="msg-context-menu-backdrop"
      onClick={(e) => {
        e.preventDefault();
        onClose();
      }}
      onContextMenu={(e) => {
        e.preventDefault();
        onClose();
      }}
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 9998,
      }}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        style={{
          position: 'fixed',
          left: menuPosition.left,
          top: menuPosition.top,
          zIndex: 9999,
          display: 'inline-flex',
          alignItems: 'center',
          gap: MENU_GAP,
          padding: `${MENU_PADDING}px ${MENU_PADDING}px`,
          background: '#2C3E50',
          borderRadius: 12,
          boxShadow: '0 4px 24px rgba(0,0,0,0.3)',
          pointerEvents: 'auto',
          transform: visible ? 'scale(1)' : 'scale(0.85)',
          opacity: visible ? 1 : 0,
          transition: 'transform 0.15s cubic-bezier(0.2, 0, 0, 1.4), opacity 0.12s ease',
        }}
      >
        {items.map((item) => (
          <button
            key={item.key}
            onClick={(e) => {
              e.stopPropagation();
              item.onClick();
            }}
            title={item.label}
            style={{
              width: MENU_ITEM_SIZE,
              height: MENU_ITEM_SIZE,
              borderRadius: '50%',
              border: 'none',
              background: 'transparent',
              color: item.color,
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              transition: 'background 0.15s',
              flexShrink: 0,
            }}
            onMouseEnter={(e) => {
              (e.currentTarget as HTMLButtonElement).style.background = 'rgba(255,255,255,0.1)';
            }}
            onMouseLeave={(e) => {
              (e.currentTarget as HTMLButtonElement).style.background = 'transparent';
            }}
          >
            {item.icon}
          </button>
        ))}
      </div>
    </div>
  );
}
