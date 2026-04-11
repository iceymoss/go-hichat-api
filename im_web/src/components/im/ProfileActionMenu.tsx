'use client';

import React, { useEffect } from 'react';
import {
  Pencil,
  Share2,
  ShieldCheck,
  Ban,
  Flag,
  Trash2,
  UserPlus,
} from 'lucide-react';
import type { LucideIcon } from 'lucide-react';

interface Position {
  top: number;
  left: number;
}

interface ProfileActionMenuProps {
  open: boolean;
  onClose: () => void;
  position: Position;
  isStranger?: boolean;
  isBlocked?: boolean;
  contactName: string;
  onSetRemark: () => void;
  onRecommend: () => void;
  onSetPermissions: () => void;
  onToggleBlock: () => void;
  onReport: () => void;
  onDeleteFriend: () => void;
  onAddFriend: () => void;
}

interface MenuItem {
  key: string;
  label: string;
  icon: LucideIcon;
  danger?: boolean;
  highlighted?: boolean;
  action: () => void;
}

interface MenuGroup {
  items: MenuItem[];
  divider?: boolean;
}

function buildMenuItems(props: ProfileActionMenuProps): MenuGroup[] {
  const {
    isStranger = false,
    isBlocked = false,
    onSetRemark,
    onRecommend,
    onSetPermissions,
    onToggleBlock,
    onReport,
    onDeleteFriend,
    onAddFriend,
  } = props;

  if (isStranger) {
    return [
      {
        items: [
          {
            key: 'add-friend',
            label: '添加好友',
            icon: UserPlus,
            highlighted: true,
            action: onAddFriend,
          },
        ],
      },
      {
        items: [
          {
            key: 'block',
            label: '加入黑名单',
            icon: Ban,
            action: onToggleBlock,
          },
          {
            key: 'report',
            label: '投诉',
            icon: Flag,
            action: onReport,
          },
        ],
      },
    ];
  }

  // Friend (blocked or not)
  return [
    {
      items: [
        {
          key: 'set-remark',
          label: '设置备注和标签',
          icon: Pencil,
          action: onSetRemark,
        },
        {
          key: 'recommend',
          label: '把他推荐给朋友',
          icon: Share2,
          action: onRecommend,
        },
      ],
    },
    {
      items: [
        {
          key: 'permissions',
          label: '朋友权限',
          icon: ShieldCheck,
          action: onSetPermissions,
        },
        {
          key: 'block',
          label: isBlocked ? '移出黑名单' : '加入黑名单',
          icon: Ban,
          action: onToggleBlock,
        },
      ],
    },
    {
      divider: true,
      items: [
        {
          key: 'report',
          label: '投诉',
          icon: Flag,
          action: onReport,
        },
        {
          key: 'delete-friend',
          label: '删除好友',
          icon: Trash2,
          danger: true,
          action: onDeleteFriend,
        },
      ],
    },
  ];
}

export default function ProfileActionMenu(props: ProfileActionMenuProps) {
  const { open, onClose, position } = props;

  // Close on Escape
  useEffect(() => {
    if (!open) return;
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', handleKey);
    return () => window.removeEventListener('keydown', handleKey);
  }, [open, onClose]);

  if (!open) return null;

  const groups = buildMenuItems(props);

  return (
    <>
      {/* Invisible overlay to catch clicks outside */}
      <div
        style={{
          position: 'fixed',
          inset: 0,
          zIndex: 9998,
        }}
        onClick={onClose}
      />

      {/* Menu dropdown */}
      <div
        className="animate-fade-in"
        style={{
          position: 'fixed',
          top: position.top,
          left: position.left,
          width: 200,
          zIndex: 9999,
          backgroundColor: '#FFFFFF',
          borderRadius: 10,
          boxShadow: '0 4px 20px rgba(0,0,0,0.12)',
          padding: '4px 0',
          overflow: 'hidden',
        }}
      >
        {groups.map((group, gIdx) => (
          <React.Fragment key={gIdx}>
            {group.divider && gIdx > 0 && (
              <div
                style={{
                  height: 1,
                  backgroundColor: 'rgba(0,0,0,0.06)',
                  margin: '4px 0',
                }}
              />
            )}
            {group.items.map((item) => {
              const Icon = item.icon;
              return (
                <button
                  key={item.key}
                  onClick={() => {
                    item.action();
                    onClose();
                  }}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 10,
                    width: '100%',
                    height: 44,
                    padding: '0 16px',
                    border: 'none',
                    background: item.highlighted
                      ? 'rgba(51,144,236,0.06)'
                      : 'transparent',
                    color: item.danger
                      ? '#FF5252'
                      : item.highlighted
                        ? '#3390EC'
                        : '#1C2733',
                    fontSize: 14,
                    fontWeight: item.highlighted ? 600 : 400,
                    cursor: 'pointer',
                    transition: 'background 0.12s',
                    textAlign: 'left',
                  }}
                  onMouseEnter={(e) => {
                    const el = e.currentTarget as HTMLButtonElement;
                    el.style.background = item.danger
                      ? '#FFF0F0'
                      : item.highlighted
                        ? 'rgba(51,144,236,0.1)'
                        : '#F5F7FA';
                  }}
                  onMouseLeave={(e) => {
                    const el = e.currentTarget as HTMLButtonElement;
                    el.style.background = item.highlighted
                      ? 'rgba(51,144,236,0.06)'
                      : 'transparent';
                  }}
                >
                  <Icon size={17} style={{ flexShrink: 0 }} />
                  <span>{item.label}</span>
                </button>
              );
            })}
          </React.Fragment>
        ))}
      </div>
    </>
  );
}
