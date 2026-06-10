'use client';

import React, { useState, useEffect } from 'react';
import { type Contact } from '@/lib/mock-data';
import { useIMStore } from '@/lib/im-store';
import { getUserTrends } from '@/lib/trend-api';
import { useT } from '@/hooks/use-i18n';
import { getAvatarColor } from '@/lib/utils';
import { Phone, Video, Send, UserPlus, Copy, MapPin, Briefcase } from 'lucide-react';
import { toast } from 'sonner';
import ProfileActionMenu from './ProfileActionMenu';
import SetRemarkDialog from './SetRemarkDialog';
import ConfirmDialog from './ConfirmDialog';

interface UserProfileCardProps {
  contact: Contact;
  /** If true, hide phone/region/moments, show "添加好友" button */
  isStranger?: boolean;
  isBlocked?: boolean;
  onSendMessage?: () => void;
  onVoiceCall?: () => void;
  onVideoCall?: () => void;
  onAddFriend?: () => void;
  onSettings?: () => void;
  onRecommend?: () => void;
  onSetPermissions?: () => void;
  onToggleBlock?: (blocked: boolean) => void;
  onReport?: () => void;
  onDeleteFriend?: () => void;
  /** Compact mode for small cards */
  compact?: boolean;
}

/** 陌生人资料：性别/地区/职业/标签（仅搜索加好友场景展示） */
function StrangerInfo({ contact }: { contact: Contact }) {
  const sexLabel = contact.gender === 'male' ? '男' : contact.gender === 'female' ? '女' : '';
  let tagList: string[] = [];
  if (contact.tags) {
    try {
      const parsed = JSON.parse(contact.tags);
      if (Array.isArray(parsed)) tagList = parsed.filter(Boolean).map(String);
    } catch {
      tagList = contact.tags.split(/[,，]/).map((s) => s.trim()).filter(Boolean);
    }
  }

  const rows: Array<[string, string]> = [];
  if (sexLabel) rows.push(['性别', sexLabel]);
  if (contact.region) rows.push(['地区', contact.region]);
  if (contact.occupation) rows.push(['职业', contact.occupation]);

  if (rows.length === 0 && tagList.length === 0) return null;

  return (
    <div style={{ marginBottom: 10 }}>
      {rows.map(([k, v]) => (
        <div key={k} className="flex" style={{ fontSize: 13, lineHeight: 1.9 }}>
          <span style={{ width: 44, color: '#A2ACB5', flexShrink: 0 }}>{k}</span>
          <span style={{ color: '#1C2733' }}>{v}</span>
        </div>
      ))}
      {tagList.length > 0 && (
        <div className="flex flex-wrap" style={{ gap: 6, marginTop: 6 }}>
          {tagList.map((t, i) => (
            <span key={i} style={{ fontSize: 12, color: '#3390EC', background: 'rgba(51,144,236,0.1)', borderRadius: 6, padding: '2px 8px' }}>{t}</span>
          ))}
        </div>
      )}
    </div>
  );
}

export default function UserProfileCard({
  contact,
  isStranger = false,
  isBlocked: isBlockedProp = false,
  onSendMessage,
  onVoiceCall,
  onVideoCall,
  onAddFriend,
  onSettings,
  onRecommend,
  onSetPermissions,
  onToggleBlock,
  onReport,
  onDeleteFriend,
  compact = false,
}: UserProfileCardProps) {
  const avatarColor = getAvatarColor(contact.name);
  const displayName = contact.remark || contact.name;
  const t = useT();
  const openUserTrends = useIMStore((s) => s.openUserTrends);
  const authToken = useIMStore((s) => s.currentUser?.token || '');

  // Latest moments thumbnails (WeChat-style): show up to 3 real images from this user's trends.
  const [moments, setMoments] = useState<string[]>([]);
  useEffect(() => {
    if (isStranger || !contact.id || !authToken) { setMoments([]); return; }
    let cancelled = false;
    (async () => {
      try {
        const r = await getUserTrends(authToken, contact.id);
        if (cancelled || !r.success || !r.data) return;
        const ordered = [...(r.data.top_list || []), ...(r.data.list || [])];
        const imgs: string[] = [];
        for (const tr of ordered) {
          for (const url of tr.resources || []) {
            if (url && !imgs.includes(url)) imgs.push(url);
            if (imgs.length >= 3) break;
          }
          if (imgs.length >= 3) break;
        }
        setMoments(imgs);
      } catch {
        if (!cancelled) setMoments([]);
      }
    })();
    return () => { cancelled = true; };
  }, [contact.id, authToken, isStranger]);

  // 解析标签（JSON 数组字符串或逗号分隔）
  const parsedTags: string[] = (() => {
    if (!contact.tags) return [];
    try {
      const p = JSON.parse(contact.tags);
      if (Array.isArray(p)) return p.filter(Boolean).map(String);
    } catch {
      return contact.tags.split(/[,，]/).map((s) => s.trim()).filter(Boolean);
    }
    return [];
  })();

  // ActionMenu state
  const [showMenu, setShowMenu] = useState(false);
  const [menuPos, setMenuPos] = useState({ top: 0, left: 0 });

  // Sub-dialog states
  const [localBlocked, setLocalBlocked] = useState(isBlockedProp);
  const [showSetRemark, setShowSetRemark] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [showBlockConfirm, setShowBlockConfirm] = useState(false);

  // Sync external isBlocked prop
  React.useEffect(() => {
    setLocalBlocked(isBlockedProp);
  }, [isBlockedProp]);

  const handleCopySignature = () => {
    if (contact.signature) {
      navigator.clipboard.writeText(contact.signature);
      toast.success('签名已复制');
    }
  };

  const handleCopyAccount = () => {
    if (contact.account) {
      navigator.clipboard.writeText(contact.account);
      toast.success('HiChat号已复制');
    }
  };

  const genderIcon = contact.gender === 'male' ? '🚹' : contact.gender === 'female' ? '🚺' : '';
  const genderText = contact.gender === 'male' ? '男' : contact.gender === 'female' ? '女' : '';

  // ── ActionMenu callbacks ──
  const handleSetRemark = () => {
    setShowSetRemark(true);
  };

  const handleRecommend = () => {
    if (onRecommend) {
      onRecommend();
    } else {
      toast('功能开发中');
    }
  };

  const handleSetPermissions = () => {
    if (onSetPermissions) {
      onSetPermissions();
    } else {
      toast('功能开发中');
    }
  };

  const handleToggleBlock = () => {
    if (localBlocked) {
      // Unblock directly
      setLocalBlocked(false);
      onToggleBlock?.(false);
      toast.success('已将 ' + contact.name + ' 移出黑名单');
    } else {
      // Show block confirmation
      setShowBlockConfirm(true);
    }
  };

  const handleReport = () => {
    if (onReport) {
      onReport();
    } else {
      toast('功能开发中');
    }
  };

  const handleDeleteFriend = () => {
    setShowDeleteConfirm(true);
  };

  const handleConfirmDelete = () => {
    onDeleteFriend?.();
    toast.success('已删除好友 ' + contact.name);
  };

  const handleConfirmBlock = () => {
    setLocalBlocked(true);
    onToggleBlock?.(true);
    toast.success('已将 ' + contact.name + ' 加入黑名单');
  };

  const handleAddFriend = () => {
    onAddFriend?.();
  };

  const token = useIMStore.getState().currentUser?.token;
  const invalidateFriends = useIMStore.getState().invalidateFriends;

  const handleSaveRemark = async (remark: string, tags: string[]) => {
    if (!token) return;
    try {
      // Save remark
      if (remark) {
        const r = await fetch('/api/social/friend/remark', {
          method: 'POST',
          headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
          body: JSON.stringify({ friend_uid: contact.id, remark }),
        }).then(res => res.json());
        if (!r.success) { toast.error(r.message || '备注更新失败'); return; }
      }
      // Save tags
      if (tags.length > 0) {
        const r = await fetch('/api/social/friend/tags', {
          method: 'POST',
          headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
          body: JSON.stringify({ friend_uid: contact.id, tags }),
        }).then(res => res.json());
        if (!r.success) { toast.error(r.message || '标签更新失败'); return; }
      }
      const tagStr = tags.length > 0 ? `，标签: ${tags.join('、')}` : '';
      toast.success(`备注已更新: ${remark}${tagStr}`);
      // 同步刷新好友列表
      invalidateFriends();
    } catch {
      toast.error('更新失败');
    }
  };

  if (compact) {
    return (
      <div
        style={{
          padding: '14px 16px',
          backgroundColor: '#FFFFFF',
          borderRadius: 16,
          boxShadow: '0 2px 12px rgba(0,0,0,0.08)',
        }}
      >
        {/* Header: avatar + name + online */}
        <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 10 }}>
          <div className="relative shrink-0">
            {contact.avatar ? (
              <img
                src={contact.avatar}
                alt={contact.name}
                style={{
                  width: 48,
                  height: 48,
                  borderRadius: '50%',
                  objectFit: 'cover',
                }}
                onError={(e) => {
                  (e.target as HTMLImageElement).style.display = 'none';
                  const next = (e.target as HTMLImageElement).nextElementSibling as HTMLElement;
                  if (next) next.style.display = 'flex';
                }}
              />
            ) : null}
            <div
              style={{
                width: 48,
                height: 48,
                borderRadius: '50%',
                background: avatarColor,
                display: contact.avatar ? 'none' : 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: '#FFFFFF',
                fontSize: 20,
                fontWeight: 600,
              }}
            >
              {contact.name[0]}
            </div>
            {contact.online && (
              <span
                style={{
                  position: 'absolute',
                  bottom: 0,
                  right: 0,
                  width: 12,
                  height: 12,
                  backgroundColor: '#4DCD5E',
                  borderRadius: '50%',
                  border: '2.5px solid #FFFFFF',
                }}
              />
            )}
          </div>
          <div style={{ flex: 1, minWidth: 0 }}>
            <div style={{ fontSize: 15, fontWeight: 600, color: '#1C2733', lineHeight: 1.3 }}>
              {displayName}
            </div>
            <div style={{ fontSize: 12, color: '#A2ACB5', lineHeight: 1.3, marginTop: 2 }}>
              {contact.remark ? `昵称: ${contact.nickname || contact.name}` : ''}
            </div>
          </div>
          <div style={{ fontSize: 12, color: contact.online ? '#4DCD5E' : '#A2ACB5' }}>
            {contact.online ? '在线' : '离线'}
          </div>
        </div>

        {/* Signature */}
        {contact.signature && (
          <div
            onClick={handleCopySignature}
            style={{
              padding: '8px 12px',
              marginBottom: 10,
              backgroundColor: '#F5F7FA',
              borderRadius: 8,
              cursor: 'pointer',
            }}
          >
            <div
              style={{
                fontSize: 13,
                color: '#646A73',
                lineHeight: 1.6,
                overflow: 'hidden',
                display: '-webkit-box',
                WebkitLineClamp: 2,
                WebkitBoxOrient: 'vertical',
              }}
            >
              {contact.signature}
            </div>
          </div>
        )}

        {/* Stranger info: 性别/地区/职业/标签 */}
        {isStranger && (
          <StrangerInfo contact={contact} />
        )}

        {/* Action button */}
        {!isStranger ? (
          onSendMessage && (
            <button
              onClick={onSendMessage}
              style={{
                width: '100%',
                height: 36,
                borderRadius: 8,
                background: '#3390EC',
                color: '#FFFFFF',
                fontSize: 13,
                fontWeight: 600,
                border: 'none',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: 6,
                transition: 'background 0.2s',
              }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLButtonElement).style.background = '#2A7BD6';
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLButtonElement).style.background = '#3390EC';
              }}
            >
              <Send size={14} />
              发消息
            </button>
          )
        ) : (
          onAddFriend && (
            <button
              onClick={onAddFriend}
              style={{
                width: '100%',
                height: 36,
                borderRadius: 8,
                background: '#3390EC',
                color: '#FFFFFF',
                fontSize: 13,
                fontWeight: 600,
                border: 'none',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: 6,
                transition: 'background 0.2s',
              }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLButtonElement).style.background = '#2A7BD6';
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLButtonElement).style.background = '#3390EC';
              }}
            >
              <UserPlus size={14} />
              添加好友
            </button>
          )
        )}
      </div>
    );
  }

  // ═══════════════════════════════════════════════
  // Full profile card — 按照 PRD 设计原型图布局
  // ═══════════════════════════════════════════════
  return (
    <div style={{ position: 'relative' }}>
      {/* Settings icon - top right — opens ActionMenu */}
      {(onSettings || onRecommend || onDeleteFriend) && (
        <button
          onClick={(e) => {
            const rect = e.currentTarget.getBoundingClientRect();
            let top = rect.bottom + 4;
            let left = rect.right - 200;
            if (left < 8) left = 8;
            if (left + 200 > window.innerWidth - 8) left = window.innerWidth - 208;
            if (top + 300 > window.innerHeight) top = window.innerHeight - 310;
            setMenuPos({ top, left });
            setShowMenu(true);
          }}
          style={{
            position: 'absolute',
            top: 0,
            right: 0,
            width: 32,
            height: 32,
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
            (e.currentTarget as HTMLButtonElement).style.background = 'rgba(0,0,0,0.04)';
            (e.currentTarget as HTMLButtonElement).style.color = '#646A73';
          }}
          onMouseLeave={(e) => {
            (e.currentTarget as HTMLButtonElement).style.background = 'transparent';
            (e.currentTarget as HTMLButtonElement).style.color = '#8F959E';
          }}
          title="更多设置"
        >
          <MoreDotsIcon size={18} />
        </button>
      )}

      {/* ProfileActionMenu dropdown */}
      <ProfileActionMenu
        open={showMenu}
        position={menuPos}
        onClose={() => setShowMenu(false)}
        isStranger={isStranger}
        isBlocked={localBlocked}
        contactName={contact.name}
        onSetRemark={handleSetRemark}
        onRecommend={handleRecommend}
        onSetPermissions={handleSetPermissions}
        onToggleBlock={handleToggleBlock}
        onReport={handleReport}
        onDeleteFriend={handleDeleteFriend}
        onAddFriend={handleAddFriend}
      />

      {/* SetRemarkDialog */}
      <SetRemarkDialog
        open={showSetRemark}
        onClose={() => setShowSetRemark(false)}
        currentRemark={contact.remark || contact.name}
        onSave={handleSaveRemark}
      />

      {/* Delete friend confirm */}
      <ConfirmDialog
        open={showDeleteConfirm}
        onClose={() => setShowDeleteConfirm(false)}
        title="删除好友"
        description={`确定删除好友「${contact.name}」吗？删除后将无法接收对方消息。`}
        confirmText="删除"
        cancelText="取消"
        confirmVariant="danger"
        onConfirm={handleConfirmDelete}
      />

      {/* Block confirm */}
      <ConfirmDialog
        open={showBlockConfirm}
        onClose={() => setShowBlockConfirm(false)}
        title="加入黑名单"
        description={`确定将「${contact.name}」加入黑名单吗？加入后你将不再收到对方的消息。`}
        confirmText="加入黑名单"
        cancelText="取消"
        confirmVariant="default"
        onConfirm={handleConfirmBlock}
      />

      {/* ── Header: 左头像 + 右信息 ── */}
      <div style={{ display: 'flex', alignItems: 'flex-start', marginBottom: 20 }}>
        {/* 头像: 100px, 圆角 12px, 阴影 */}
        <div
          className="relative shrink-0"
          style={{ width: 100, height: 100, flexShrink: 0 }}
        >
          {contact.avatar ? (
            <img
              src={contact.avatar}
              alt={contact.name}
              style={{
                width: 100,
                height: 100,
                borderRadius: 12,
                objectFit: 'cover',
                boxShadow: '0 2px 8px rgba(0,0,0,0.12)',
              }}
              onError={(e) => {
                // 图片加载失败时隐藏，显示备用方块
                (e.target as HTMLImageElement).style.display = 'none';
                const next = (e.target as HTMLImageElement).nextElementSibling as HTMLElement;
                if (next) next.style.display = 'flex';
              }}
            />
          ) : null}
          <div
            style={{
              width: 100,
              height: 100,
              borderRadius: 12,
              background: avatarColor,
              display: contact.avatar ? 'none' : 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: '#FFFFFF',
              fontSize: 40,
              fontWeight: 600,
              boxShadow: '0 2px 8px rgba(0,0,0,0.12)',
            }}
          >
            {contact.name[0]}
          </div>
          {contact.online && (
            <span
              style={{
                position: 'absolute',
                bottom: 2,
                right: 2,
                width: 14,
                height: 14,
                backgroundColor: '#4DCD5E',
                borderRadius: '50%',
                border: '3px solid #FFFFFF',
              }}
            />
          )}
        </div>

        {/* 右侧信息区 */}
        <div style={{ flex: 1, marginLeft: 20, display: 'flex', flexDirection: 'column', gap: 5, minWidth: 0 }}>
          {/* Row 1: 备注（没有备注就显示昵称）— 22px bold */}
          <div
            style={{
              fontSize: 22,
              fontWeight: 600,
              color: '#1F2329',
              lineHeight: 1.3,
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
            }}
          >
            {contact.remark || contact.name}
          </div>

          {/* Row 2: 用户账号 + 性别年龄 并排 */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 16, lineHeight: 1.8 }}>
            {contact.account && (
              <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                <span style={{ fontSize: 14, color: '#646A73' }}>
                  {contact.account}
                </span>
                <button
                  onClick={handleCopyAccount}
                  style={{
                    background: 'transparent',
                    border: 'none',
                    cursor: 'pointer',
                    color: '#A2ACB5',
                    padding: 2,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    transition: 'color 0.15s',
                  }}
                  onMouseEnter={(e) => {
                    (e.currentTarget as HTMLButtonElement).style.color = '#3390EC';
                  }}
                  onMouseLeave={(e) => {
                    (e.currentTarget as HTMLButtonElement).style.color = '#A2ACB5';
                  }}
                  title="复制HiChat号"
                >
                  <Copy size={13} />
                </button>
              </div>
            )}
            {(genderIcon || contact.age) && (
              <span style={{ fontSize: 14, color: '#646A73' }}>
                {genderIcon} {genderText}{contact.age ? ` / ${contact.age}岁` : ''}
              </span>
            )}
          </div>

          {/* Row 3: 用户昵称 + 手机号 并排 */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 16, lineHeight: 1.8 }}>
            <span style={{ fontSize: 14, color: '#646A73' }}>
              昵称：{contact.nickname || contact.name}
            </span>
            {contact.phone && (
              <span style={{ fontSize: 14, color: '#8F959E' }}>
                手机：{contact.phone}
              </span>
            )}
          </div>

          {/* Row 4: 地址 */}
          {contact.region && (
            <div style={{ display: 'flex', alignItems: 'center', gap: 4, lineHeight: 1.8 }}>
              <MapPin size={13} style={{ color: '#A2ACB5', flexShrink: 0 }} />
              <span style={{ fontSize: 14, color: '#8F959E' }}>{contact.region}</span>
            </div>
          )}

          {/* Row 5: 职业 */}
          {contact.occupation && (
            <div style={{ display: 'flex', alignItems: 'center', gap: 4, lineHeight: 1.8 }}>
              <Briefcase size={13} style={{ color: '#A2ACB5', flexShrink: 0 }} />
              <span style={{ fontSize: 14, color: '#8F959E' }}>{contact.occupation}</span>
            </div>
          )}

          {/* Row 6: 标签 */}
          {parsedTags.length > 0 && (
            <div className="flex flex-wrap" style={{ gap: 6, marginTop: 2 }}>
              {parsedTags.map((t, i) => (
                <span key={i} style={{ fontSize: 12, color: '#3390EC', background: 'rgba(51,144,236,0.1)', borderRadius: 6, padding: '2px 8px' }}>{t}</span>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* ── 个性签名 ── */}
      {contact.signature && (
        <div
          onClick={handleCopySignature}
          style={{
            marginBottom: 20,
            padding: 12,
            backgroundColor: '#F5F7FA',
            borderRadius: 8,
            cursor: 'pointer',
            transition: 'background 0.15s',
          }}
          onMouseEnter={(e) => {
            (e.currentTarget as HTMLDivElement).style.backgroundColor = '#EEF0F4';
          }}
          onMouseLeave={(e) => {
            (e.currentTarget as HTMLDivElement).style.backgroundColor = '#F5F7FA';
          }}
          title="点击复制签名"
        >
          <div style={{ fontSize: 14, color: '#646A73', lineHeight: 1.6 }}>
            {contact.signature}
          </div>
        </div>
      )}

      {/* ── 朋友圈 ── */}
      {!isStranger && (
        <div style={{ marginBottom: 20 }}>
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              marginBottom: moments.length > 0 ? 10 : 0,
            }}
          >
            <span style={{ fontSize: 16, fontWeight: 600, color: '#1F2329' }}>朋友圈</span>
            <button
              onClick={() => openUserTrends(contact.id, displayName)}
              style={{
                fontSize: 13,
                color: '#3390EC',
                background: 'transparent',
                border: 'none',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                gap: 2,
                transition: 'opacity 0.15s',
              }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLButtonElement).style.opacity = '0.7';
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLButtonElement).style.opacity = '1';
              }}
            >
              {t('moments.view')}
              <span style={{ fontSize: 14 }}>›</span>
            </button>
          </div>
          {moments.length > 0 && (
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 4 }}>
              {moments.slice(0, 3).map((img, idx) => (
                <img
                  key={idx}
                  src={img}
                  alt=""
                  style={{
                    width: '100%',
                    aspectRatio: '1',
                    borderRadius: 8,
                    objectFit: 'cover',
                  }}
                />
              ))}
            </div>
          )}
        </div>
      )}

      {/* ── 底部操作按钮 — 竖排 ── */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
        {!isStranger ? (
          <>
            {onSendMessage && (
              <button
                onClick={onSendMessage}
                style={{
                  width: '100%',
                  height: 40,
                  borderRadius: 8,
                  background: '#3390EC',
                  color: '#FFFFFF',
                  fontSize: 14,
                  fontWeight: 600,
                  border: 'none',
                  cursor: 'pointer',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: 6,
                  transition: 'background 0.2s',
                }}
                onMouseEnter={(e) => {
                  (e.currentTarget as HTMLButtonElement).style.background = '#2A7BD6';
                }}
                onMouseLeave={(e) => {
                  (e.currentTarget as HTMLButtonElement).style.background = '#3390EC';
                }}
              >
                <Send size={16} />
                发消息
              </button>
            )}
            {onVoiceCall && (
              <button
                onClick={onVoiceCall}
                style={{
                  width: '100%',
                  height: 40,
                  borderRadius: 8,
                  background: '#FFFFFF',
                  border: '1px solid #3390EC',
                  color: '#3390EC',
                  fontSize: 14,
                  fontWeight: 500,
                  cursor: 'pointer',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: 6,
                  transition: 'all 0.2s',
                }}
                onMouseEnter={(e) => {
                  (e.currentTarget as HTMLButtonElement).style.background = 'rgba(51,144,236,0.06)';
                }}
                onMouseLeave={(e) => {
                  (e.currentTarget as HTMLButtonElement).style.background = '#FFFFFF';
                }}
              >
                <Phone size={16} />
                语音通话
              </button>
            )}
            {onVideoCall && (
              <button
                onClick={onVideoCall}
                style={{
                  width: '100%',
                  height: 40,
                  borderRadius: 8,
                  background: '#FFFFFF',
                  border: '1px solid #3390EC',
                  color: '#3390EC',
                  fontSize: 14,
                  fontWeight: 500,
                  cursor: 'pointer',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: 6,
                  transition: 'all 0.2s',
                }}
                onMouseEnter={(e) => {
                  (e.currentTarget as HTMLButtonElement).style.background = 'rgba(51,144,236,0.06)';
                }}
                onMouseLeave={(e) => {
                  (e.currentTarget as HTMLButtonElement).style.background = '#FFFFFF';
                }}
              >
                <Video size={16} />
                视频通话
              </button>
            )}
          </>
        ) : (
          onAddFriend && (
            <button
              onClick={onAddFriend}
              style={{
                width: '100%',
                height: 40,
                borderRadius: 8,
                background: '#3390EC',
                color: '#FFFFFF',
                fontSize: 14,
                fontWeight: 600,
                border: 'none',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: 6,
                transition: 'background 0.2s',
              }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLButtonElement).style.background = '#2A7BD6';
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLButtonElement).style.background = '#3390EC';
              }}
            >
              <UserPlus size={16} />
              添加好友
            </button>
          )
        )}
      </div>
    </div>
  );
}

/* Three dots "..." icon */
function MoreDotsIcon({ size = 18 }: { size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor">
      <circle cx="5" cy="12" r="1.5" />
      <circle cx="12" cy="12" r="1.5" />
      <circle cx="19" cy="12" r="1.5" />
    </svg>
  );
}
