'use client';

import React, { useState, useMemo, useCallback, useEffect, useRef } from 'react';
import {
  Search,
  Heart,
  MessageCircle,
  MapPin,
  Plus,
  Bell,
  X,
  Send,
  Pin,
  PinOff,
  MessageSquareOff,
  Trash2,
  Play,
  ArrowLeft,
  ExternalLink,
  Clock,
  ThumbsUp,
  EyeOff,
  Image as ImageIcon,
  FileText,
  Link2,
  Video,
  Type,
  Users,
  Lock,
  Globe,
  User,
  Loader2,
  Inbox,
  ImagePlus,
} from 'lucide-react';
import { toast } from 'sonner';
import ImageViewer from './ImageViewer';
import { getAvatarColor } from '@/lib/utils';
import { useIMStore, friendDisplayName } from '@/lib/im-store';
import { useIsMobile } from '@/hooks/use-mobile';
import { useT } from '@/hooks/use-i18n';
import {
  type Trend,
  type TrendComment,
  type MomentsNotification,
} from '@/lib/mock-data';
import {
  createTrend,
  deleteTrend as apiDeleteTrend,
  updateTrend as apiUpdateTrend,
  getLatestTrends,
  getUserTrends,
  getTrendDetail,
  getLikedUsers,
  getCommentTree,
  createComment,
  deleteComment as apiDeleteComment,
  getTrendMessages,
  markTrendMessagesRead,
  trendMessagesToNotifications,
  toggleLike as apiToggleLike,
  getBatchLikeSummary,
  getTrendPublishConfig,
  uploadTrendMedia,
  getTrendDraft,
  saveTrendDraft,
  deleteTrendDraft,
  type TrendPublishConfig,
  mapBackendTrend,
  commentTreeToMap,
  batchSummaryToLikeUsersMap,
} from '@/lib/trend-api';

/* ═══════════════════════════════════════
   Helpers
   ═══════════════════════════════════════ */

function fmtTime(date: Date, t: (k: string) => string): string {
  const now = Date.now();
  const diff = now - date.getTime();
  const m = Math.floor(diff / 60000);
  const h = Math.floor(diff / 3600000);
  const d = Math.floor(diff / 86400000);
  if (m < 1) return t('group.time.justNow');
  if (m < 60) return t('group.time.minutesAgo').replace('{m}', String(m));
  if (h < 24) return t('group.time.hoursAgo').replace('{h}', String(h));
  if (d < 30) return t('group.time.monthDay').replace('{month}', String(new Date(date).getMonth() + 1)).replace('{day}', String(new Date(date).getDate()));
  return `${new Date(date).getFullYear()}/${new Date(date).getMonth() + 1}/${new Date(date).getDate()}`;
}

function getUserName(userId: string, fallback?: string, t?: (k: string) => string): string {
  if (userId === 'me') return fallback || useIMStore.getState().currentUser?.name || (t ? t('trend.me') : '我');
  // 好友备注优先，其次后端昵称兜底
  const c = useIMStore.getState().friends.find(ct => ct.id === userId);
  if (c?.remark) return c.remark;
  return fallback || c?.name || userId;
}

function trendDisplayName(trend: Trend, t?: (k: string) => string): string {
  return getUserName(trend.userId, trend.userName, t);
}

function avatarCircle(name: string, size: number, extra?: React.ReactNode, avatar?: string) {
  return (
    <div className="relative shrink-0" style={{ width: size, height: size, borderRadius: '50%', backgroundColor: getAvatarColor(name), display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: size * 0.38, fontWeight: 600, color: '#FFF', overflow: 'hidden' }}>
      {avatar ? (
        <img src={avatar} alt="" className="w-full h-full object-cover" />
      ) : (
        name[0]
      )}
      {extra}
    </div>
  );
}

const inputStyle = { width: '100%', borderRadius: '10px', border: '1px solid rgba(0,0,0,0.1)', padding: '10px 12px', fontSize: '14px', color: '#1C2733', outline: 'none', background: '#F5F7FA', boxSizing: 'border-box' as const };
const focusInput = (e: React.FocusEvent<HTMLInputElement | HTMLTextAreaElement>) => { e.target.style.borderColor = '#1BB45B'; e.target.style.boxShadow = '0 0 0 3px rgba(27,180,91,0.15)'; };
const blurInput = (e: React.FocusEvent<HTMLInputElement | HTMLTextAreaElement>) => { e.target.style.borderColor = 'rgba(0,0,0,0.1)'; e.target.style.boxShadow = 'none'; };

const trendTypeLabels: Record<number, { labelKey: string; icon: React.ReactNode }> = {
  1: { labelKey: 'trend.type.text', icon: <Type className="w-4 h-4" /> },
  2: { labelKey: 'trend.type.imageText', icon: <ImageIcon className="w-4 h-4" /> },
  3: { labelKey: 'trend.type.article', icon: <FileText className="w-4 h-4" /> },
  4: { labelKey: 'trend.type.share', icon: <Link2 className="w-4 h-4" /> },
  5: { labelKey: 'trend.type.video', icon: <Video className="w-4 h-4" /> },
};

const scopeLabels: Record<number, { labelKey: string; icon: React.ReactNode }> = {
  1: { labelKey: 'trend.vis.private', icon: <Lock className="w-3.5 h-3.5" /> },
  2: { labelKey: 'trend.vis.friends', icon: <Users className="w-3.5 h-3.5" /> },
  3: { labelKey: 'trend.vis.public', icon: <Globe className="w-3.5 h-3.5" /> },
};

/* ═══════════════════════════════════════
   ConfirmModal
   ═══════════════════════════════════════ */

interface ConfirmOpts {
  title: string;
  description: string;
  confirmLabel: string;
  confirmColor: string;
  loading?: boolean;
  onConfirm: () => void;
  onCancel: () => void;
  onClose: () => void;
}

function ConfirmModal({ open, opts }: { open: boolean; opts: ConfirmOpts | null }) {
  const t = useT();
  if (!open || !opts) return null;
  return (
    <div className="fixed inset-0" style={{ zIndex: 10001, display: 'flex', alignItems: 'center', justifyContent: 'center' }} onClick={(e) => { if (e.target === e.currentTarget) opts.onClose(); }}>
      <div className="absolute inset-0" style={{ background: 'rgba(0,0,0,0.5)' }} />
      <div className="relative" style={{ background: '#FFF', borderRadius: '16px', padding: '24px', width: '90%', maxWidth: '400px', boxShadow: '0 8px 40px rgba(0,0,0,0.15)' }}>
        <button onClick={opts.onClose} className="absolute" style={{ top: 16, right: 16, width: 28, height: 28, borderRadius: '50%', border: 'none', background: 'transparent', color: '#A2ACB5', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <X className="w-4 h-4" />
        </button>
        <h3 style={{ fontSize: '16px', fontWeight: 600, color: '#1C2733', marginBottom: '8px', paddingRight: 32 }}>{opts.title}</h3>
        <p style={{ fontSize: '14px', color: '#646A73', marginBottom: '20px', lineHeight: '1.5' }}>{opts.description}</p>
        <div className="flex items-center justify-end gap-3">
          <button onClick={opts.onCancel} disabled={opts.loading} style={{ padding: '8px 20px', borderRadius: '8px', border: '1px solid rgba(0,0,0,0.1)', background: '#FFF', color: '#646A73', fontSize: '14px', fontWeight: 500, cursor: opts.loading ? 'not-allowed' : 'pointer' }}>{t('common.cancel')}</button>
          <button onClick={opts.onConfirm} disabled={opts.loading} style={{ padding: '8px 20px', borderRadius: '8px', border: 'none', background: opts.confirmColor, color: '#FFF', fontSize: '14px', fontWeight: 500, cursor: opts.loading ? 'not-allowed' : 'pointer', display: 'flex', alignItems: 'center', gap: 6, opacity: opts.loading ? 0.7 : 1 }}>
            {opts.loading && <Loader2 className="w-4 h-4" style={{ animation: 'spin 1s linear infinite' }} />}
            {opts.confirmLabel}
          </button>
        </div>
      </div>
    </div>
  );
}

/* ═══════════════════════════════════════
   Action Menu
   ═══════════════════════════════════════ */

interface ActionMenuItem {
  label: string;
  color?: string;
  icon?: React.ReactNode;
  onClick: () => void;
}

function ActionMenu({ x, y, items, onClose }: { x: number; y: number; items: ActionMenuItem[]; onClose: () => void }) {
  const ref = useRef<HTMLDivElement>(null);
  useEffect(() => {
    const handler = (e: MouseEvent) => { if (ref.current && !ref.current.contains(e.target as Node)) onClose(); };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [onClose]);

  return (
    <div ref={ref} className="fixed" style={{ zIndex: 10002, left: Math.min(x, window.innerWidth - 170), top: Math.min(y, window.innerHeight - items.length * 40 - 10) }}>
      <div style={{ background: '#2C3E50', borderRadius: '12px', padding: '4px', boxShadow: '0 4px 24px rgba(0,0,0,0.3)', minWidth: '150px' }}>
        {items.map((item, i) => (
          <button key={i} onClick={() => { item.onClick(); onClose(); }} style={{ display: 'flex', alignItems: 'center', gap: 8, width: '100%', padding: '9px 14px', border: 'none', background: 'transparent', color: item.color || '#FFF', fontSize: '13px', cursor: 'pointer', borderRadius: '8px', transition: 'background 0.15s' }} onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'rgba(255,255,255,0.1)'; }} onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}>
            {item.icon}{item.label}
          </button>
        ))}
      </div>
    </div>
  );
}

/* ═══════════════════════════════════════
   Skeleton Card
   ═══════════════════════════════════════ */

function SkeletonCard() {
  return (
    <div style={{ padding: '14px 12px', borderBottom: '1px solid rgba(0,0,0,0.06)' }}>
      <div className="flex gap-3">
        <div style={{ width: 40, height: 40, borderRadius: '50%', background: '#E8EDEF', flexShrink: 0 }} />
        <div className="flex-1">
          <div style={{ width: 80, height: 14, borderRadius: 4, background: '#E8EDEF', marginBottom: 8 }} />
          <div style={{ width: '100%', height: 12, borderRadius: 4, background: '#E8EDEF', marginBottom: 6 }} />
          <div style={{ width: '70%', height: 12, borderRadius: 4, background: '#E8EDEF' }} />
        </div>
      </div>
    </div>
  );
}

/* ═══════════════════════════════════════
   CommentItem (recursive)
   ═══════════════════════════════════════ */

interface CommentItemProps {
  comment: TrendComment;
  onReply: (comment: TrendComment) => void;
  onDelete: (comment: TrendComment) => void;
  depth?: number;
}

function CommentItem({ comment, onReply, onDelete, depth = 0 }: CommentItemProps) {
  const showUserCard = useIMStore(s => s.showUserCard);
  const t = useT();
  return (
    <div>
      <div className="flex items-start gap-2" style={{ padding: '4px 0' }}>
        <div style={{ fontSize: '12px', lineHeight: '1.6', flex: 1 }}>
          <span style={{ color: '#576b95', fontWeight: 600, cursor: 'pointer' }} onClick={(e) => { e.stopPropagation(); showUserCard(comment.replyer.id); }}>{friendDisplayName(comment.replyer.id, comment.replyer.name)}</span>
          {comment.father !== 0 && comment.user && comment.user.id !== comment.replyer.id && (
            <span> {t('trend.replyConnector')} <span style={{ color: '#576b95', fontWeight: 500, cursor: 'pointer' }} onClick={(e) => { e.stopPropagation(); showUserCard(comment.user.id); }}>{friendDisplayName(comment.user.id, comment.user.name)}</span></span>
          )}
          <span style={{ color: '#1C2733' }}>：{comment.content}</span>
        </div>
        <div className="flex items-center gap-1 shrink-0" style={{ marginTop: 2 }}>
          {comment.replyer.id === 'me' && (
            <button onClick={() => onDelete(comment)} style={{ padding: '2px 6px', border: 'none', background: 'transparent', color: '#A2ACB5', fontSize: '11px', cursor: 'pointer', borderRadius: '4px' }} onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.color = '#E53935'; }} onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.color = '#A2ACB5'; }}>
              {t('trend.delete')}
            </button>
          )}
          <button onClick={() => onReply(comment)} style={{ padding: '2px 6px', border: 'none', background: 'transparent', color: '#A2ACB5', fontSize: '11px', cursor: 'pointer', borderRadius: '4px' }} onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.color = '#1BB45B'; }} onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.color = '#A2ACB5'; }}>
            {t('trend.reply')}
          </button>
        </div>
      </div>
      {comment.children && comment.children.length > 0 && (
        <div style={{ marginLeft: 12, paddingLeft: 10, borderLeft: '2px solid rgba(27,180,91,0.15)' }}>
          {comment.children.map(child => (
            <CommentItem key={child.id} comment={child} onReply={onReply} onDelete={onDelete} depth={depth + 1} />
          ))}
        </div>
      )}
    </div>
  );
}

/* ═══════════════════════════════════════
   TrendCard
   ═══════════════════════════════════════ */

const FEED_LIKE_COLLAPSE_LIMIT = 70;

interface TrendCardProps {
  trend: Trend;
  showTopBadge?: boolean;
  liked: boolean;
  likeCount: number;
  likeUsers: { id: string; name: string; avatar: string }[];
  comments: TrendComment[];
  expanded: boolean;
  replyTarget: TrendComment | null;
  commentText: string;
  selected?: boolean;
  onToggleLike: () => void;
  onLikeCountClick: () => void;
  onExpandComments: () => void;
  onSetReplyTarget: (c: TrendComment | null) => void;
  onCommentTextChange: (v: string) => void;
  onSubmitComment: () => void;
  onDeleteComment: (c: TrendComment) => void;
  onOpenDetail: () => void;
  onAvatarClick: () => void;
  onManage: (x: number, y: number) => void;
}

function TrendCard({
  trend, showTopBadge = true, liked, likeCount, likeUsers, comments, expanded,
  replyTarget, commentText, selected,
  onToggleLike, onLikeCountClick, onExpandComments,
  onSetReplyTarget, onCommentTextChange, onSubmitComment, onDeleteComment,
  onOpenDetail, onAvatarClick, onManage,
}: TrendCardProps) {
  const [likeAnim, setLikeAnim] = useState(false);
  const [likeNamesExpanded, setLikeNamesExpanded] = useState(false);
  const [viewerIndex, setViewerIndex] = useState(-1);
  const showUserCard = useIMStore(s => s.showUserCard);
  const t = useT();
  const userName = trendDisplayName(trend, t);
  const userAvatar = trend.userAvatar || '';
  const totalComments = trend.replyCount + (expanded ? comments.reduce((acc, c) => acc + 1 + (c.children?.length || 0), 0) - comments.reduce((acc, c) => acc, 0) : 0);
  const visibleComments = expanded ? comments : comments.slice(0, 2);
  const hiddenCount = comments.length > 2 && !expanded ? comments.length - 2 : 0;

  const handleLike = () => {
    setLikeAnim(true);
    onToggleLike();
    setTimeout(() => setLikeAnim(false), 300);
  };

  const imageGridClass = (count: number) => {
    if (count === 1) return 'grid-cols-1';
    if (count === 2 || count === 4) return 'grid-cols-2';
    return 'grid-cols-3';
  };

  const renderContent = () => {
    const parts = trend.content.split(/(@\S+)/g);
    return parts.map((part, i) => {
      if (part.startsWith('@')) {
        const atUser = trend.atUsers.find(u => `@${u.name}` === part);
        return <span key={i} style={{ color: '#576b95', cursor: 'pointer' }} onClick={(e) => { if (atUser) { e.stopPropagation(); showUserCard(atUser.id); } }}>{atUser ? atUser.name : part}</span>;
      }
      return <span key={i}>{part}</span>;
    });
  };

  return (
    <div
      className="cursor-pointer"
      style={{
        padding: '14px 12px',
        borderBottom: '1px solid rgba(0,0,0,0.06)',
        background: selected ? 'rgba(27,180,91,0.06)' : undefined,
        borderLeft: selected ? '3px solid #1BB45B' : undefined,
        paddingLeft: selected ? 9 : 12,
        transition: 'background 0.15s',
      }}
      onClick={(e) => {
        const target = e.target as HTMLElement;
        if (target.closest('button') || target.closest('input') || target.closest('a')) return;
        onOpenDetail();
      }}
    >
      <div className="flex gap-3">
        {/* Avatar */}
        <div
          className="shrink-0 mt-0.5 cursor-pointer overflow-hidden"
          style={{ width: 40, height: 40, borderRadius: '50%', backgroundColor: getAvatarColor(userName), display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 16, fontWeight: 600, color: '#FFF' }}
          onClick={onAvatarClick}
        >
          {userAvatar ? (
            <img src={userAvatar} alt="" className="w-full h-full object-cover" />
          ) : (
            userName[0]
          )}
        </div>

        {/* Content */}
        <div className="flex-1 min-w-0">
          {/* Header row */}
          <div className="flex items-center gap-2 flex-wrap">
            <span
              className="text-sm font-semibold cursor-pointer"
              style={{ color: '#576b95' }}
              onClick={onAvatarClick}
            >
              {userName}
            </span>
            {showTopBadge && trend.isTop && (
              <span style={{ fontSize: '10px', fontWeight: 500, color: '#F5A623', backgroundColor: 'rgba(245,166,35,0.1)', borderRadius: '4px', padding: '1px 6px', display: 'inline-flex', alignItems: 'center', gap: 2 }}>
                <Pin className="w-3 h-3" />{t('trend.top')}
              </span>
            )}
            {!trend.openReply && (
              <span style={{ fontSize: '10px', fontWeight: 500, color: '#A2ACB5', backgroundColor: 'rgba(0,0,0,0.04)', borderRadius: '4px', padding: '1px 6px', display: 'inline-flex', alignItems: 'center', gap: 2 }}>
                <MessageSquareOff className="w-3 h-3" />{t('trend.commentClosed')}
              </span>
            )}
            <span style={{ fontSize: '11px', color: '#A2ACB5', marginLeft: 'auto', whiteSpace: 'nowrap' }}>{fmtTime(trend.createTime, t)}</span>
          </div>

          {/* Text content */}
          <p className="text-sm mt-1.5 leading-relaxed" style={{ color: '#1C2733' }}>
            {renderContent()}
          </p>

          {/* Title (article type) */}
          {trend.type === 3 && trend.title && (
            <h4 className="text-sm font-semibold mt-2" style={{ color: '#1C2733' }}>{trend.title}</h4>
          )}

          {/* Media */}
          {trend.type === 2 && trend.resources.length > 0 && (
            <div className={`grid ${imageGridClass(trend.resources.length)} gap-1.5 mt-2.5`} style={{ maxWidth: 280 }}>
              {trend.resources.map((img, idx) => (
                <div
                  key={idx}
                  className="overflow-hidden cursor-pointer"
                  style={{ borderRadius: 6, background: '#E8EDEF', aspectSquare: trend.resources.length > 1 ? 'auto' : undefined, maxHeight: trend.resources.length === 1 ? 200 : undefined }}
                  onClick={(e) => { e.stopPropagation(); setViewerIndex(idx); }}
                >
                  <img src={img} alt="" className="w-full h-full object-cover hover:scale-105 transition-transform duration-200" style={{ display: 'block', ...(trend.resources.length === 1 ? { height: 200 } : { aspectRatio: '1' }) }} />
                </div>
              ))}
            </div>
          )}

          {trend.type === 3 && trend.coverUrl && (
            <div className="mt-2.5" style={{ borderRadius: 10, overflow: 'hidden', cursor: 'pointer', maxWidth: 300 }} onClick={onOpenDetail}>
              <img src={trend.coverUrl} alt="" className="w-full object-cover hover:scale-105 transition-transform duration-200" style={{ display: 'block', height: 150 }} />
            </div>
          )}

          {trend.type === 4 && (
            <div
              className="mt-2.5 flex items-center gap-3 cursor-pointer"
              style={{ padding: '10px 12px', borderRadius: 10, background: 'rgba(0,0,0,0.03)', border: '1px solid rgba(0,0,0,0.06)', maxWidth: 300 }}
              onClick={onOpenDetail}
            >
              <div style={{ width: 36, height: 36, borderRadius: 8, background: 'rgba(27,180,91,0.1)', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                <ExternalLink className="w-4 h-4" style={{ color: '#1BB45B' }} />
              </div>
              <div className="flex-1 min-w-0">
                <div style={{ fontSize: '12px', fontWeight: 600, color: '#1C2733', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{t('trend.shareLink')}</div>
                <div style={{ fontSize: '11px', color: '#A2ACB5', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', marginTop: 2 }}>
                  {trend.shareUrl}
                </div>
              </div>
            </div>
          )}

          {trend.type === 5 && trend.resources.length > 0 && (
            <div className="mt-2.5 relative cursor-pointer" style={{ borderRadius: 10, overflow: 'hidden', maxWidth: 280 }} onClick={onOpenDetail}>
              <img src={trend.resources[0]} alt="" className="w-full object-cover" style={{ display: 'block', aspectRatio: '16/9' }} />
              <div className="absolute inset-0 flex items-center justify-center" style={{ background: 'rgba(0,0,0,0.2)' }}>
                <div style={{ width: 48, height: 48, borderRadius: '50%', background: 'rgba(255,255,255,0.9)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                  <Play className="w-5 h-5" style={{ color: '#1C2733', marginLeft: 2 }} />
                </div>
              </div>
            </div>
          )}

          {/* Location */}
          {trend.positionName && (
            <div className="flex items-center gap-1 mt-1.5">
              <MapPin className="w-3 h-3" style={{ color: '#A2ACB5' }} />
              <span style={{ fontSize: '11px', color: '#A2ACB5' }}>{trend.positionName}</span>
            </div>
          )}

          {/* Action bar */}
          <div className="flex items-center justify-between mt-2.5">
            <div className="flex items-center gap-3">
              {/* Like */}
              <button
                onClick={handleLike}
                className="flex items-center gap-1 transition-all duration-200"
                style={{
                  padding: '4px 10px',
                  borderRadius: '16px',
                  border: 'none',
                  background: liked ? 'rgba(27,180,91,0.1)' : 'transparent',
                  cursor: 'pointer',
                  transform: likeAnim ? 'scale(1.2)' : 'scale(1)',
                  transition: 'transform 0.2s cubic-bezier(0.175, 0.885, 0.32, 1.275)',
                }}
              >
                <Heart className="w-3.5 h-3.5" style={{ color: liked ? '#FA5151' : '#A2ACB5', fill: liked ? '#FA5151' : 'none', transition: 'all 0.15s' }} />
                {likeCount > 0 && <span style={{ fontSize: '11px', color: liked ? '#FA5151' : '#A2ACB5', fontWeight: liked ? 600 : 400 }}>{likeCount}</span>}
              </button>

              {/* Comment */}
              <button
                onClick={onOpenDetail}
                className="flex items-center gap-1"
                style={{ padding: '4px 10px', borderRadius: '16px', border: 'none', background: 'transparent', cursor: 'pointer' }}
              >
                <MessageCircle className="w-3.5 h-3.5" style={{ color: '#A2ACB5' }} />
                {trend.replyCount > 0 && <span style={{ fontSize: '11px', color: '#A2ACB5' }}>{trend.replyCount}</span>}
              </button>

              {/* Share */}
              <button
                style={{ padding: '4px 10px', borderRadius: '16px', border: 'none', background: 'transparent', cursor: 'pointer' }}
                onClick={() => toast.success(t('trend.linkCopied'))}
              >
                <Send className="w-3.5 h-3.5" style={{ color: '#A2ACB5' }} />
              </button>
            </div>

            {/* Manage (own posts) */}
            {trend.userId === 'me' && (
              <button
                onClick={(e) => onManage(e.clientX, e.clientY)}
                style={{ padding: '4px', borderRadius: '50%', border: 'none', background: 'transparent', cursor: 'pointer' }}
              >
                <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
                  <circle cx="8" cy="3" r="1.5" fill="#A2ACB5" />
                  <circle cx="8" cy="8" r="1.5" fill="#A2ACB5" />
                  <circle cx="8" cy="13" r="1.5" fill="#A2ACB5" />
                </svg>
              </button>
            )}
          </div>

          {/* Like names preview */}
          {likeCount > 0 && likeUsers.length > 0 && (
            <div
              className="mt-2"
              style={{ fontSize: '12px', lineHeight: '1.8' }}
              onClick={(e) => e.stopPropagation()}
            >
              <div className="flex items-start gap-1.5">
                <Heart className="w-3 h-3 shrink-0 mt-0.5" style={{ color: '#FA5151', fill: '#FA5151' }} />
                <span style={{ color: '#708499' }}>
                  {(likeNamesExpanded ? likeUsers : likeUsers.slice(0, FEED_LIKE_COLLAPSE_LIMIT)).map((u, i) => {
                    const displayName = u.id === 'me' ? t('trend.me') : friendDisplayName(u.id, u.name);
                    return (
                      <span key={u.id || `like-${i}`}>
                        {i > 0 && <span style={{ margin: '0 2px' }}>、</span>}
                        <span
                          style={{ color: '#576b95', cursor: 'pointer' }}
                          onClick={(e) => { e.stopPropagation(); showUserCard(u.id); }}
                          onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.textDecoration = 'underline'; }}
                          onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.textDecoration = 'none'; }}
                        >
                          {displayName}
                        </span>
                      </span>
                    );
                  })}
                  {likeUsers.length > FEED_LIKE_COLLAPSE_LIMIT && !likeNamesExpanded && (
                    <span style={{ margin: '0 2px' }}>、</span>
                  )}
                  {likeUsers.length > FEED_LIKE_COLLAPSE_LIMIT && (
                    <button
                      onClick={(e) => { e.stopPropagation(); setLikeNamesExpanded(prev => !prev); }}
                      style={{
                        color: '#1BB45B', background: 'none', border: 'none',
                        cursor: 'pointer', fontSize: '12px', padding: 0, fontWeight: 500,
                      }}
                    >
                      {likeNamesExpanded ? t('trend.collapse') : t('trend.andNPeople').replace('{count}', String(likeUsers.length))}
                    </button>
                  )}
                </span>
              </div>
            </div>
          )}

          {/* Comments section */}
          {visibleComments.length > 0 && (
            <div className="mt-2" style={{ background: 'rgba(0,0,0,0.03)', borderRadius: 10, padding: '8px 12px' }}>
              {visibleComments.map(comment => (
                <CommentItem key={comment.id} comment={comment} onReply={onSetReplyTarget} onDelete={onDeleteComment} />
              ))}
              {hiddenCount > 0 && (
                <button
                  onClick={onExpandComments}
                  style={{ fontSize: '12px', color: '#1BB45B', background: 'none', border: 'none', cursor: 'pointer', padding: '4px 0 0', fontWeight: 500 }}
                >
                  {t('trend.expandComments').replace('{count}', String(trend.replyCount))}
                </button>
              )}
            </div>
          )}

          {/* Inline comment input */}
          {expanded && trend.openReply && (
            <div className="flex items-center gap-2 mt-2">
              {replyTarget && (
                <div className="flex items-center gap-1 text-xs w-full" style={{ color: '#708499', marginBottom: 4, padding: '0 4px' }}>
                  <span>{t('trend.reply')}</span>
                  <span style={{ color: '#576b95', fontWeight: 500 }}>{friendDisplayName(replyTarget.replyer.id, replyTarget.replyer.name)}</span>
                  <button onClick={() => onSetReplyTarget(null)} style={{ marginLeft: 'auto', background: 'none', border: 'none', cursor: 'pointer', color: '#A2ACB5', padding: 0 }}>
                    <X className="w-3 h-3" />
                  </button>
                </div>
              )}
              <div className="flex items-center gap-2 flex-1">
                <input
                  value={commentText}
                  onChange={(e) => onCommentTextChange(e.target.value)}
                  onKeyDown={(e) => { if (e.key === 'Enter' && commentText.trim()) onSubmitComment(); }}
                  placeholder={replyTarget ? t('trend.replyPlaceholder').replace('{name}', friendDisplayName(replyTarget.replyer.id, replyTarget.replyer.name)) : t('trend.writeComment')}
                  style={{ ...inputStyle, flex: 1, padding: '7px 12px', fontSize: '13px', borderRadius: '18px' }}
                  onFocus={focusInput}
                  onBlur={blurInput}
                />
                {commentText.trim() && (
                  <button
                    onClick={onSubmitComment}
                    style={{ width: 32, height: 32, borderRadius: '50%', border: 'none', background: '#1BB45B', color: '#FFF', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}
                  >
                    <Send className="w-3.5 h-3.5" />
                  </button>
                )}
              </div>
            </div>
          )}
        </div>
      </div>

      <ImageViewer images={trend.resources} index={viewerIndex} onClose={() => setViewerIndex(-1)} onIndexChange={setViewerIndex} />
    </div>
  );
}

/* ═══════════════════════════════════════
   Publish Modal
   ═══════════════════════════════════════ */

interface PublishModalProps {
  open: boolean;
  token: string | null;
  onClose: () => void;
  onSubmit: (data: {
    type: number;
    content: string;
    title: string;
    images: string[];
    shareUrl: string;
    location: string;
    scope: number;
  }) => void;
}

async function compressImageFile(file: File): Promise<File> {
  if (!file.type.startsWith('image/') || file.type === 'image/gif') return file;
  const imageUrl = URL.createObjectURL(file);
  try {
    const img = await new Promise<HTMLImageElement>((resolve, reject) => {
      const image = new window.Image();
      image.onload = () => resolve(image);
      image.onerror = reject;
      image.src = imageUrl;
    });
    const maxSide = 1920;
    const scale = Math.min(1, maxSide / Math.max(img.width, img.height));
    const width = Math.max(1, Math.round(img.width * scale));
    const height = Math.max(1, Math.round(img.height * scale));
    const canvas = document.createElement('canvas');
    canvas.width = width;
    canvas.height = height;
    const ctx = canvas.getContext('2d');
    if (!ctx) return file;
    ctx.drawImage(img, 0, 0, width, height);
    const blob = await new Promise<Blob | null>(resolve => canvas.toBlob(resolve, 'image/webp', 0.82));
    if (!blob || blob.size >= file.size) return file;
    const name = file.name.replace(/\.[^.]+$/, '.webp');
    return new File([blob], name, { type: 'image/webp', lastModified: Date.now() });
  } finally {
    URL.revokeObjectURL(imageUrl);
  }
}

function PublishModal({ open, token, onClose, onSubmit }: PublishModalProps) {
  const t = useT();
  const [type, setType] = useState<number>(1);
  const [content, setContent] = useState('');
  const [title, setTitle] = useState('');
  const [images, setImages] = useState<string[]>([]);
  const [uploading, setUploading] = useState(false);
  const [config, setConfig] = useState<TrendPublishConfig | null>(null);
  const [draftId, setDraftId] = useState<number | undefined>();
  const [dragMediaIndex, setDragMediaIndex] = useState<number | null>(null);
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const [shareUrl, setShareUrl] = useState('');
  const [location, setLocation] = useState('');
  const [scope, setScope] = useState<number>(3);

  useEffect(() => {
    if (!open || !token) return;
    getTrendPublishConfig(token).then(r => {
      if (r.success && r.data) setConfig(r.data);
    }).catch(() => undefined);
  }, [open, token]);

  useEffect(() => {
    if (!open || !token) return;
    getTrendDraft(token).then(r => {
      const draft = r.data?.draft;
      if (!r.success || !draft || content || images.length > 0) return;
      setDraftId(draft.id);
      setType(draft.type || 1);
      setContent(draft.content || '');
      setTitle(draft.title || '');
      setImages(draft.resources || []);
      setShareUrl(draft.share_url || '');
      setLocation(draft.position_name || '');
      setScope(draft.scope || 3);
    }).catch(() => undefined);
  }, [open, token]);

  useEffect(() => {
    if (!open || !token) return;
    const hasDraft = content.trim() || title.trim() || images.length > 0 || shareUrl.trim() || location.trim();
    if (!hasDraft) return;
    const timer = window.setTimeout(() => {
      saveTrendDraft(token, {
        id: draftId,
        type,
        content,
        title,
        resources: images,
        share_url: shareUrl,
        position_name: location,
        scope,
        open_reply: 1,
      }).then(r => {
        if (r.success && r.data?.draft?.id) setDraftId(r.data.draft.id);
      }).catch(() => undefined);
    }, 900);
    return () => window.clearTimeout(timer);
  }, [open, token, draftId, type, content, title, images, shareUrl, location, scope]);

  if (!open) return null;

  const maxMediaCount = type === 5 ? (config?.max_video_count || 1) : (config?.max_image_count || 9);

  const handleFiles = async (files: FileList | File[]) => {
    if (!token) {
      toast.error(t('trend.needLogin'));
      return;
    }
    const nextFiles = Array.from(files);
    if (nextFiles.length === 0) return;
    const allowed = Math.max(0, maxMediaCount - images.length);
    if (allowed <= 0) {
      toast.error(t('trend.maxMedia').replace('{count}', String(maxMediaCount)));
      return;
    }
    const picked = nextFiles.slice(0, allowed);
    if (picked.length < nextFiles.length) toast.warning(t('trend.maxMediaIgnored').replace('{count}', String(maxMediaCount)));

    setUploading(true);
    try {
      const urls: string[] = [];
      for (const file of picked) {
        const uploadFile = config?.image_compression_enabled && file.type.startsWith('image/')
          ? await compressImageFile(file).catch(() => file)
          : file;
        const r = await uploadTrendMedia(token, uploadFile);
        if (!r.success || !r.data?.url) throw new Error(r.message || t('trend.uploadFail'));
        urls.push(r.data.url);
      }
      setImages(prev => [...prev, ...urls]);
      toast.success(t('trend.mediaUploaded'));
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('trend.uploadRetry'));
    } finally {
      setUploading(false);
    }
  };

  const handleRemoveImage = (idx: number) => {
    setImages(prev => prev.filter((_, i) => i !== idx));
  };

  const moveImage = (from: number, to: number) => {
    setImages(prev => {
      const next = [...prev];
      if (from < 0 || to < 0 || from >= next.length || to >= next.length || from === to) return prev;
      const [item] = next.splice(from, 1);
      next.splice(to, 0, item);
      return next;
    });
  };

  const requestClose = () => {
    if (uploading) return;
    const hasDraft = content.trim() || title.trim() || images.length > 0 || shareUrl.trim() || location.trim();
    if (hasDraft && !window.confirm(t('trend.draftConfirm'))) return;
    onClose();
  };

  const handleSubmit = () => {
    if (uploading) {
      toast.error(t('trend.mediaUploading'));
      return;
    }
    if (type !== 4 && !content.trim() && images.length === 0) {
      toast.error(type === 1 ? t('trend.needContent') : t('trend.needContentOrMedia'));
      return;
    }
    if (type === 3 && !title.trim()) {
      toast.error(t('trend.needArticleTitle'));
      return;
    }
    if (type === 4 && !shareUrl.trim()) {
      toast.error(t('trend.needShareUrl'));
      return;
    }
    onSubmit({ type, content: content.trim(), title: title.trim(), images, shareUrl: shareUrl.trim(), location: location.trim(), scope });
    if (token && draftId) deleteTrendDraft(token, draftId).catch(() => undefined);
    // Reset
    setDraftId(undefined);
    setContent('');
    setTitle('');
    setImages([]);
    setShareUrl('');
    setLocation('');
    setScope(3);
    setType(1);
  };

  return (
    <div className="fixed inset-0" style={{ zIndex: 10001, display: 'flex', alignItems: 'center', justifyContent: 'center' }} onClick={(e) => { if (e.target === e.currentTarget) requestClose(); }}>
      <div className="absolute inset-0" style={{ background: 'rgba(0,0,0,0.5)' }} />
      <div className="relative" style={{ background: '#FFF', borderRadius: '16px', width: '92%', maxWidth: '480px', maxHeight: '85vh', display: 'flex', flexDirection: 'column', boxShadow: '0 8px 40px rgba(0,0,0,0.15)' }}>
        {/* Header */}
        <div className="flex items-center justify-between shrink-0" style={{ padding: '16px 20px', borderBottom: '1px solid rgba(0,0,0,0.06)' }}>
          <button onClick={requestClose} style={{ width: 32, height: 32, borderRadius: '50%', border: 'none', background: 'transparent', color: '#708499', cursor: uploading ? 'not-allowed' : 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <X className="w-5 h-5" />
          </button>
          <span style={{ fontSize: '16px', fontWeight: 600, color: '#1C2733' }}>{t('trend.publish')}</span>
          <button
            onClick={handleSubmit}
            disabled={uploading}
            style={{ padding: '6px 16px', borderRadius: '16px', border: 'none', background: uploading ? '#A2ACB5' : '#1BB45B', color: '#FFF', fontSize: '13px', fontWeight: 500, cursor: uploading ? 'not-allowed' : 'pointer' }}
          >
            {uploading ? t('trend.uploading') : t('trend.publishBtn')}
          </button>
        </div>

        {/* Body */}
        <div className="flex-1 overflow-y-auto" style={{ padding: '16px 20px' }}>
          {/* Type selector */}
          <div className="flex items-center gap-2 mb-4 flex-wrap">
            {[1, 2, 3, 4, 5].map(ty => (
              <button
                key={ty}
                onClick={() => setType(ty)}
                style={{
                  display: 'flex', alignItems: 'center', gap: 4, padding: '5px 12px', borderRadius: '16px',
                  border: 'none', background: type === ty ? 'rgba(27,180,91,0.1)' : 'rgba(0,0,0,0.04)',
                  color: type === ty ? '#1BB45B' : '#708499', fontSize: '12px', fontWeight: type === ty ? 500 : 400, cursor: 'pointer',
                }}
              >
                {trendTypeLabels[ty].icon}
                {t(trendTypeLabels[ty].labelKey)}
              </button>
            ))}
          </div>

          {/* Title (article) */}
          {type === 3 && (
            <div style={{ marginBottom: 12 }}>
              <input value={title} onChange={(e) => setTitle(e.target.value)} placeholder={t('trend.articleTitle')} style={{ ...inputStyle, fontWeight: 600, fontSize: '15px' }} onFocus={focusInput} onBlur={blurInput} />
            </div>
          )}

          {/* Content */}
          {type !== 4 && (
            <div style={{ marginBottom: 12 }}>
              <textarea
                value={content}
                onChange={(e) => setContent(e.target.value)}
                placeholder={type === 1 ? t('trend.placeholderText') : type === 3 ? t('trend.placeholderArticle') : t('trend.placeholderDesc')}
                rows={4}
                style={{ ...inputStyle, resize: 'vertical', minHeight: 80, lineHeight: '1.6' }}
                onFocus={focusInput}
                onBlur={blurInput}
              />
            </div>
          )}

          {/* Media upload (type 2 or 5) */}
          {(type === 2 || type === 5) && (
            <div style={{ marginBottom: 12 }}>
              {images.length < maxMediaCount ? (
                <>
                  <input
                    ref={fileInputRef}
                    type="file"
                    accept={type === 5 ? 'video/*' : 'image/*'}
                    multiple={type !== 5}
                    className="hidden"
                    onChange={(e) => { if (e.target.files) handleFiles(e.target.files); e.currentTarget.value = ''; }}
                  />
                  <div
                    className="flex flex-col items-center justify-center gap-2"
                    style={{ border: '1.5px dashed rgba(27,180,91,0.35)', borderRadius: 12, padding: '18px 12px', background: 'rgba(27,180,91,0.04)', cursor: 'pointer', marginBottom: 10 }}
                    onClick={() => fileInputRef.current?.click()}
                    onDragOver={(e) => { e.preventDefault(); }}
                    onDrop={(e) => { e.preventDefault(); handleFiles(e.dataTransfer.files); }}
                  >
                    <ImagePlus className="w-7 h-7" style={{ color: '#1BB45B' }} />
                    <div style={{ fontSize: '13px', fontWeight: 600, color: '#1C2733' }}>{type === 5 ? t('trend.pickVideo') : t('trend.pickImage')}</div>
                    <div style={{ fontSize: '12px', color: '#708499' }}>{type === 5 ? t('trend.videoLimit') : t('trend.imageLimit').replace('{count}', String(maxMediaCount)).replace('{size}', String(config?.max_image_size_mb || 50))}</div>
                    {uploading && <Loader2 className="w-4 h-4 animate-spin" style={{ color: '#1BB45B' }} />}
                  </div>
                </>
              ) : (
                <div style={{ borderRadius: 10, padding: '10px 12px', background: 'rgba(245,166,35,0.08)', color: '#A06400', fontSize: 12, marginBottom: 10 }}>
                  {t('trend.maxReached').replace('{count}', String(maxMediaCount))}
                </div>
              )}
              {images.length > 0 && (
                <div className="grid grid-cols-3 gap-2">
                  {images.map((img, idx) => (
                    <div
                      key={img}
                      draggable
                      className="relative"
                      style={{ aspectRatio: '1', borderRadius: 8, overflow: 'hidden', background: '#E8EDEF', cursor: 'grab', opacity: dragMediaIndex === idx ? 0.55 : 1 }}
                      onDragStart={() => setDragMediaIndex(idx)}
                      onDragOver={(e) => e.preventDefault()}
                      onDrop={(e) => { e.preventDefault(); if (dragMediaIndex !== null) moveImage(dragMediaIndex, idx); setDragMediaIndex(null); }}
                      onDragEnd={() => setDragMediaIndex(null)}
                    >
                      {type === 5 ? (
                        <div className="w-full h-full flex items-center justify-center" style={{ color: '#708499' }}><Video className="w-7 h-7" /></div>
                      ) : (
                        <img src={img} alt="" className="w-full h-full object-cover" />
                      )}
                      <button onClick={() => handleRemoveImage(idx)} className="absolute" style={{ top: 5, right: 5, width: 24, height: 24, borderRadius: '50%', background: '#E53935', border: '2px solid #FFF', color: '#FFF', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', boxShadow: '0 2px 8px rgba(0,0,0,0.35)' }}>
                        <X className="w-4 h-4" strokeWidth={3} />
                      </button>
                      <div className="absolute" style={{ left: 4, bottom: 4, borderRadius: 999, background: 'rgba(0,0,0,0.45)', color: '#FFF', fontSize: 10, padding: '1px 6px' }}>{t('trend.dragSort')}</div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* Share link (type 4) */}
          {type === 4 && (
            <div style={{ marginBottom: 12 }}>
              <input value={shareUrl} onChange={(e) => setShareUrl(e.target.value)} placeholder={t('trend.shareUrlPlaceholder')} style={inputStyle} onFocus={focusInput} onBlur={blurInput} />
              <textarea value={content} onChange={(e) => setContent(e.target.value)} placeholder={t('trend.shareReasonPlaceholder')} rows={3} style={{ ...inputStyle, resize: 'vertical', marginTop: 8, minHeight: 60 }} onFocus={focusInput} onBlur={blurInput} />
            </div>
          )}

          {/* Location */}
          <div style={{ marginBottom: 12 }}>
            <div className="flex items-center gap-2">
              <MapPin className="w-4 h-4" style={{ color: '#A2ACB5', flexShrink: 0 }} />
              <input value={location} onChange={(e) => setLocation(e.target.value)} placeholder={t('trend.locationPlaceholder')} style={{ ...inputStyle, flex: 1 }} onFocus={focusInput} onBlur={blurInput} />
            </div>
          </div>

          {/* Scope */}
          <div>
            <div style={{ fontSize: '13px', fontWeight: 500, color: '#1C2733', marginBottom: '8px' }}>{t('trend.visRange')}</div>
            <div className="flex items-center gap-2">
              {[1, 2, 3].map(s => (
                <button
                  key={s}
                  onClick={() => setScope(s)}
                  style={{
                    display: 'flex', alignItems: 'center', gap: 4, padding: '6px 14px', borderRadius: '16px',
                    border: scope === s ? '1.5px solid #1BB45B' : '1.5px solid rgba(0,0,0,0.1)',
                    background: scope === s ? 'rgba(27,180,91,0.06)' : '#FFF',
                    color: scope === s ? '#1BB45B' : '#708499', fontSize: '12px', fontWeight: scope === s ? 500 : 400, cursor: 'pointer',
                  }}
                >
                  {scopeLabels[s].icon}{t(scopeLabels[s].labelKey)}
                </button>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

/* ═══════════════════════════════════════
   Like List Modal
   ═══════════════════════════════════════ */

function LikeListModal({ open, users, onClose }: { open: boolean; users: { id: string; name: string; avatar: string }[]; onClose: () => void }) {
  const showUserCard = useIMStore(s => s.showUserCard);
  const t = useT();
  if (!open) return null;
  return (
    <div className="fixed inset-0" style={{ zIndex: 10001, display: 'flex', alignItems: 'center', justifyContent: 'center' }} onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}>
      <div className="absolute inset-0" style={{ background: 'rgba(0,0,0,0.5)' }} />
      <div className="relative" style={{ background: '#FFF', borderRadius: '16px', width: '90%', maxWidth: '400px', maxHeight: '60vh', display: 'flex', flexDirection: 'column', boxShadow: '0 8px 40px rgba(0,0,0,0.15)' }}>
        <div className="flex items-center justify-between shrink-0" style={{ padding: '16px 20px', borderBottom: '1px solid rgba(0,0,0,0.06)' }}>
          <span style={{ fontSize: '16px', fontWeight: 600, color: '#1C2733' }}>{t('trend.likeList')}</span>
          <button onClick={onClose} style={{ width: 28, height: 28, borderRadius: '50%', border: 'none', background: 'transparent', color: '#A2ACB5', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <X className="w-4 h-4" />
          </button>
        </div>
        <div className="flex-1 overflow-y-auto" style={{ padding: '8px 12px' }}>
          {users.length === 0 ? (
            <div className="flex flex-col items-center justify-center" style={{ padding: '40px 0' }}>
              <Heart className="w-12 h-12" style={{ color: '#E8EDEF', marginBottom: 12 }} />
              <div style={{ fontSize: '13px', color: '#A2ACB5' }}>{t('trend.noLikes')}</div>
            </div>
          ) : (
            users.map(user => (
              <div key={user.id} className="flex items-center gap-3" style={{ padding: '8px 8px', borderRadius: 8, cursor: 'pointer' }} onClick={() => showUserCard(user.id)} onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'rgba(0,0,0,0.02)'; }} onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}>
                {avatarCircle(user.name, 36)}
                <span style={{ fontSize: '14px', fontWeight: 500, color: '#1C2733' }}>{user.name}</span>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
}

/* ═══════════════════════════════════════
   Trend Detail Modal
   ═══════════════════════════════════════ */

interface TrendDetailModalProps {
  open: boolean;
  trend: Trend | null;
  liked: boolean;
  likeCount: number;
  likeUsers: { id: string; name: string; avatar: string }[];
  comments: TrendComment[];
  commentText: string;
  replyTarget: TrendComment | null;
  onToggleLike: () => void;
  onLikeCountClick: () => void;
  onSubmitComment: () => void;
  onCommentTextChange: (v: string) => void;
  onSetReplyTarget: (c: TrendComment | null) => void;
  onDeleteComment: (c: TrendComment) => void;
  onAvatarClick: (userId: string) => void;
  onClose: () => void;
}

function TrendDetailModal({
  open, trend, liked, likeCount, likeUsers, comments, commentText, replyTarget,
  onToggleLike, onLikeCountClick, onSubmitComment, onCommentTextChange,
  onSetReplyTarget, onDeleteComment, onAvatarClick, onClose,
}: TrendDetailModalProps) {
  const [viewerIndex, setViewerIndex] = useState(-1);
  const t = useT();
  if (!open || !trend) return null;
  const userName = trendDisplayName(trend, t);
  const userAvatar = trend.userAvatar || '';

  const renderContent = () => {
    const parts = trend.content.split(/(@\S+)/g);
    return parts.map((part, i) => {
      if (part.startsWith('@')) {
        const atUser = trend.atUsers.find(u => `@${u.name}` === part);
        return <span key={i} style={{ color: '#576b95' }}>{atUser ? atUser.name : part}</span>;
      }
      return <span key={i}>{part}</span>;
    });
  };

  const imageGridClass = (count: number) => {
    if (count === 1) return 'grid-cols-1';
    if (count === 2 || count === 4) return 'grid-cols-2';
    return 'grid-cols-3';
  };

  return (
    <div className="fixed inset-0" style={{ zIndex: 10001, display: 'flex', alignItems: 'center', justifyContent: 'center' }} onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}>
      <div className="absolute inset-0" style={{ background: 'rgba(0,0,0,0.5)' }} />
      <div className="relative" style={{ background: '#FFF', borderRadius: '16px', width: '94%', maxWidth: '520px', maxHeight: '85vh', display: 'flex', flexDirection: 'column', boxShadow: '0 8px 40px rgba(0,0,0,0.15)' }}>
        {/* Header */}
        <div className="flex items-center justify-between shrink-0" style={{ padding: '16px 20px', borderBottom: '1px solid rgba(0,0,0,0.06)' }}>
          <span style={{ fontSize: '16px', fontWeight: 600, color: '#1C2733' }}>{t('trend.detail')}</span>
          <button onClick={onClose} style={{ width: 28, height: 28, borderRadius: '50%', border: 'none', background: 'transparent', color: '#A2ACB5', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Body */}
        <div className="flex-1 overflow-y-auto" style={{ padding: '16px 20px' }}>
          {/* User header */}
          <div className="flex items-center gap-3 mb-4">
            <div
              className="cursor-pointer shrink-0 overflow-hidden"
              style={{ width: 44, height: 44, borderRadius: '50%', backgroundColor: getAvatarColor(userName), display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 18, fontWeight: 600, color: '#FFF' }}
              onClick={() => onAvatarClick(trend.userId)}
            >
              {userAvatar ? (
                <img src={userAvatar} alt="" className="w-full h-full object-cover" />
              ) : (
                userName[0]
              )}
            </div>
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2">
                <span style={{ fontSize: '15px', fontWeight: 600, color: '#1BB45B', cursor: 'pointer' }} onClick={() => onAvatarClick(trend.userId)}>{userName}</span>
                {trend.isTop && (
                  <span style={{ fontSize: '10px', fontWeight: 500, color: '#F5A623', backgroundColor: 'rgba(245,166,35,0.1)', borderRadius: '4px', padding: '1px 6px' }}>{t('trend.top')}</span>
                )}
              </div>
              <span style={{ fontSize: '12px', color: '#A2ACB5' }}>{fmtTime(trend.createTime, t)}</span>
            </div>
          </div>

          {/* Badges */}
          <div className="flex items-center gap-2 mb-3 flex-wrap">
            {trend.scope !== 3 && (
              <span style={{ fontSize: '11px', color: '#708499', display: 'flex', alignItems: 'center', gap: 3 }}>
                {scopeLabels[trend.scope].icon}{t(scopeLabels[trend.scope].labelKey)}
              </span>
            )}
            {!trend.openReply && (
              <span style={{ fontSize: '11px', color: '#A2ACB5', display: 'flex', alignItems: 'center', gap: 3 }}>
                <MessageSquareOff className="w-3 h-3" />{t('trend.commentClosed')}
              </span>
            )}
          </div>

          {/* Content */}
          {trend.type === 3 && trend.title && (
            <h3 style={{ fontSize: '16px', fontWeight: 600, color: '#1C2733', marginBottom: 8 }}>{trend.title}</h3>
          )}
          <p style={{ fontSize: '14px', lineHeight: '1.7', color: '#1C2733', marginBottom: 12 }}>{renderContent()}</p>

          {/* Location */}
          {trend.positionName && (
            <div className="flex items-center gap-1.5 mb-4" style={{ fontSize: '12px', color: '#A2ACB5' }}>
              <MapPin className="w-3.5 h-3.5" />
              {trend.positionName}
            </div>
          )}

          {/* Media */}
          {trend.type === 2 && trend.resources.length > 0 && (
            <div className={`grid ${imageGridClass(trend.resources.length)} gap-2 mb-4`} style={{ maxWidth: '100%' }}>
              {trend.resources.map((img, idx) => (
                <div
                  key={idx}
                  className="overflow-hidden cursor-pointer"
                  style={{ borderRadius: 8, background: '#E8EDEF' }}
                  onClick={() => setViewerIndex(idx)}
                >
                  <img src={img} alt="" className="w-full object-cover hover:scale-105 transition-transform duration-200" style={{ display: 'block', aspectRatio: trend.resources.length === 1 ? '16/9' : '1' }} />
                </div>
              ))}
            </div>
          )}

          {trend.type === 3 && trend.coverUrl && (
            <div className="mb-4" style={{ borderRadius: 10, overflow: 'hidden' }}>
              <img src={trend.coverUrl} alt="" className="w-full object-cover" style={{ display: 'block' }} />
            </div>
          )}

          {trend.type === 4 && (
            <div className="flex items-center gap-3 mb-4" style={{ padding: '10px 14px', borderRadius: 10, background: 'rgba(0,0,0,0.03)', border: '1px solid rgba(0,0,0,0.06)' }}>
              <div style={{ width: 36, height: 36, borderRadius: 8, background: 'rgba(27,180,91,0.1)', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                <ExternalLink className="w-4 h-4" style={{ color: '#1BB45B' }} />
              </div>
              <div className="flex-1 min-w-0">
                <div style={{ fontSize: '12px', fontWeight: 600, color: '#1C2733', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{t('trend.shareLink')}</div>
                <div style={{ fontSize: '11px', color: '#A2ACB5', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', marginTop: 2 }}>{trend.shareUrl}</div>
              </div>
            </div>
          )}

          {trend.type === 5 && trend.resources.length > 0 && (
            <div className="relative mb-4" style={{ borderRadius: 10, overflow: 'hidden' }}>
              <img src={trend.resources[0]} alt="" className="w-full object-cover" style={{ display: 'block', aspectRatio: '16/9' }} />
              <div className="absolute inset-0 flex items-center justify-center" style={{ background: 'rgba(0,0,0,0.2)' }}>
                <div style={{ width: 52, height: 52, borderRadius: '50%', background: 'rgba(255,255,255,0.9)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                  <Play className="w-6 h-6" style={{ color: '#1C2733', marginLeft: 2 }} />
                </div>
              </div>
            </div>
          )}

          {/* Actions */}
          <div className="flex items-center gap-4 mb-4" style={{ paddingBottom: 12, borderBottom: '1px solid rgba(0,0,0,0.06)' }}>
            <button onClick={onToggleLike} className="flex items-center gap-1.5" style={{ padding: '6px 14px', borderRadius: '18px', border: 'none', background: liked ? 'rgba(27,180,91,0.1)' : 'transparent', cursor: 'pointer' }}>
              <Heart className="w-4 h-4" style={{ color: liked ? '#FA5151' : '#A2ACB5', fill: liked ? '#FA5151' : 'none' }} />
              <span style={{ fontSize: '13px', color: liked ? '#FA5151' : '#708499', fontWeight: liked ? 600 : 400 }}>{likeCount > 0 ? likeCount : t('trend.like')}</span>
            </button>
            <button onClick={onLikeCountClick} style={{ padding: '6px 14px', borderRadius: '18px', border: 'none', background: 'transparent', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 4 }}>
              <ThumbsUp className="w-4 h-4" style={{ color: '#A2ACB5' }} />
              <span style={{ fontSize: '13px', color: '#708499' }}>{t('trend.likedPeople')}</span>
            </button>
          </div>

          {/* Comments */}
          <div>
            <div style={{ fontSize: '14px', fontWeight: 600, color: '#1C2733', marginBottom: 8 }}>
              {t('trend.commentCount').replace('{count}', String(trend.replyCount))}
            </div>
            {comments.length === 0 ? (
              <div className="flex flex-col items-center justify-center" style={{ padding: '32px 0' }}>
                <MessageCircle className="w-10 h-10" style={{ color: '#E8EDEF', marginBottom: 8 }} />
                <div style={{ fontSize: '13px', color: '#A2ACB5' }}>{t('trend.noComments')}</div>
              </div>
            ) : (
              <div style={{ background: 'rgba(0,0,0,0.03)', borderRadius: 10, padding: '10px 14px' }}>
                {comments.map(comment => (
                  <CommentItem key={comment.id} comment={comment} onReply={onSetReplyTarget} onDelete={onDeleteComment} />
                ))}
              </div>
            )}
          </div>
        </div>

        {/* Comment input bar */}
        {trend.openReply && (
          <div className="shrink-0 flex items-center gap-2" style={{ padding: '12px 20px', borderTop: '1px solid rgba(0,0,0,0.06)', background: '#FFF', borderBottomLeftRadius: 16, borderBottomRightRadius: 16 }}>
            {replyTarget && (
              <div className="flex items-center gap-1 shrink-0" style={{ fontSize: '11px', color: '#708499', maxWidth: 100, overflow: 'hidden' }}>
                <span>{t('trend.reply')}</span>
                <span style={{ color: '#576b95', fontWeight: 500 }}>{friendDisplayName(replyTarget.replyer.id, replyTarget.replyer.name)}</span>
              </div>
            )}
            <input
              value={commentText}
              onChange={(e) => onCommentTextChange(e.target.value)}
              onKeyDown={(e) => { if (e.key === 'Enter' && commentText.trim()) onSubmitComment(); }}
              placeholder={replyTarget ? t('trend.replyPlaceholder').replace('{name}', friendDisplayName(replyTarget.replyer.id, replyTarget.replyer.name)) : t('trend.writeComment')}
              style={{ ...inputStyle, flex: 1, padding: '8px 14px', fontSize: '13px', borderRadius: '20px' }}
              onFocus={focusInput}
              onBlur={blurInput}
            />
            {commentText.trim() && (
              <button onClick={onSubmitComment} style={{ width: 34, height: 34, borderRadius: '50%', border: 'none', background: '#1BB45B', color: '#FFF', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                <Send className="w-4 h-4" />
              </button>
            )}
          </div>
        )}
      </div>

      <ImageViewer images={trend.resources} index={viewerIndex} onClose={() => setViewerIndex(-1)} onIndexChange={setViewerIndex} />
    </div>
  );
}

/* ═══════════════════════════════════════
   Main Component
   ═══════════════════════════════════════ */

type View = 'feed' | 'notifications' | 'userTrends';
type NotifTab = 'all' | 'reply' | 'like';

// 动态消息通知的动作文案
function notifActionText(type: MomentsNotification['type'], t: (k: string) => string): string {
  switch (type) {
    case 'like': return t('trend.notify.act.like');
    case 'comment': return t('trend.notify.act.comment');
    case 'reply': return t('trend.notify.act.reply');
    case 'at_trend': return t('trend.notify.act.atTrend');
    case 'at_comment': return t('trend.notify.act.atComment');
    default: return '';
  }
}

export default function MomentsFeed() {
  const isMobile = useIsMobile();
  const t = useT();
  const { selectedTrendId, setSelectedTrendId, currentUser: meAuth, trendVersions, bumpTrendVersion, showUserCard, updateCurrentUser, setMomentsUnreadCount, trendNotifyVersion } = useIMStore();
  const meName = meAuth?.name || '';
  const meAvatar = meAuth?.avatar || '';
  const meSignature = meAuth?.introduction || '';
  const meCover = meAuth?.momentsCover || '';
  const meUserId = meAuth?.id || '';
  const token = meAuth?.token || '';
  const userTrendsTarget = useIMStore(s => s.userTrendsTarget);
  const clearUserTrendsTarget = useIMStore(s => s.clearUserTrendsTarget);
  const coverInputRef = useRef<HTMLInputElement>(null);

  const handleCoverUpload = useCallback(async (file: File) => {
    if (!token) return;
    try {
      const fd = new FormData();
      fd.append('file', file);
      const up = await fetch('/api/user/avatar', { method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: fd });
      const upData = await up.json();
      if (!upData.success || !upData.data?.url) { toast.error(t('trend.coverUploadFail')); return; }
      const url = upData.data.url as string;
      const save = await fetch('/api/user/update', { method: 'PUT', headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' }, body: JSON.stringify({ moments_cover: url }) });
      const saveData = await save.json();
      if (saveData.success) { updateCurrentUser({ momentsCover: url }); toast.success(t('trend.coverUpdated')); }
      else toast.error(saveData.message || t('trend.coverSaveFail'));
    } catch { toast.error(t('trend.coverUploadFail')); }
  }, [token, updateCurrentUser]);

  // ── View state ──
  const [view, setView] = useState<View>('feed');
  // Where the notifications view returns to (feed vs. my own userTrends).
  const [notifBackView, setNotifBackView] = useState<View>('feed');
  const [userTrendsUserId, setUserTrendsUserId] = useState<string>('');
  const [notifTab, setNotifTab] = useState<NotifTab>('all');
  const [loading, setLoading] = useState(true);

  // ── Data state ──
  const [trends, setTrends] = useState<Trend[]>([]);
  const [likedTrends, setLikedTrends] = useState<Set<number>>(new Set());
  const [commentsMap, setCommentsMap] = useState<Record<number, TrendComment[]>>({});
  const [likeUsersMap, setLikeUsersMap] = useState<Record<number, { id: string; name: string; avatar: string }[]>>({});
  const [notifications, setNotifications] = useState<MomentsNotification[]>([]);
  const [searchQuery, setSearchQuery] = useState('');

  // ── UI state ──
  const [showPublish, setShowPublish] = useState(false);
  const [detailTrendId, setDetailTrendId] = useState<number | null>(null);
  const [likeListTrendId, setLikeListTrendId] = useState<number | null>(null);
  const [expandedTrendIds, setExpandedTrendIds] = useState<Set<number>>(new Set());
  const [replyTargets, setReplyTargets] = useState<Record<number, TrendComment | null>>({});
  const [commentTexts, setCommentTexts] = useState<Record<number, string>>({});
  const [detailCommentText, setDetailCommentText] = useState('');
  const [detailReplyTarget, setDetailReplyTarget] = useState<TrendComment | null>(null);
  const [confirmOpts, setConfirmOpts] = useState<ConfirmOpts | null>(null);
  const [actionMenu, setActionMenu] = useState<{ x: number; y: number; trendId: number } | null>(null);

  // Name lookup for user trends header (filled when user clicks an avatar).
  const [userTrendsUserName, setUserTrendsUserName] = useState<string>('');
  // Viewed user's profile for the userTrends cover/header (parity with the public feed).
  const [userTrendsProfile, setUserTrendsProfile] = useState<{ avatar: string; name: string; signature: string; cover: string }>({ avatar: '', name: '', signature: '', cover: '' });

  // ── Feed loader: pulls latest trends + comments + like users ──
  const loadFeed = useCallback(async () => {
    if (!token) return;
    setLoading(true);
    try {
      const resp = await getLatestTrends(token, { count: 20 });
      if (!resp.success || !resp.data) {
        setTrends([]);
        return;
      }
      const mapped = (resp.data.list || []).map(t => mapBackendTrend(t, meUserId));
      setTrends(mapped);

      const ids = mapped.map(t => t.id);
      if (ids.length > 0) {
        const [treeRes, likesRes] = await Promise.all([
          getCommentTree(token, ids),
          getBatchLikeSummary(token, ids),
        ]);
        if (treeRes.success && treeRes.data) {
          setCommentsMap(commentTreeToMap(treeRes.data, meUserId));
        }
        if (likesRes.success && likesRes.data) {
          const map = batchSummaryToLikeUsersMap(likesRes.data, meUserId);
          setLikeUsersMap(map);
          // Derive likedTrends from whether 'me' is in the like users.
          const liked = new Set<number>();
          Object.entries(map).forEach(([tid, users]) => {
            if (users.some(u => u.id === 'me')) liked.add(Number(tid));
          });
          setLikedTrends(liked);
        }
      } else {
        setCommentsMap({});
        setLikeUsersMap({});
        setLikedTrends(new Set());
      }
    } catch (err) {
      console.error('loadFeed error', err);
      toast.error(t('trend.loadFail'));
    } finally {
      setLoading(false);
    }
  }, [token, meUserId]);

  // ── Notifications loader（统一消息中心：赞/评论/回复/@） ──
  const loadNotifications = useCallback(async () => {
    if (!token) return;
    try {
      const resp = await getTrendMessages(token, 0, 30);
      if (resp.success && resp.data) {
        setNotifications(trendMessagesToNotifications(resp.data.list || [], meUserId));
      }
    } catch (err) {
      console.error('loadNotifications error', err);
    }
  }, [token, meUserId]);

  useEffect(() => {
    if (!token) {
      setLoading(false);
      return;
    }
    loadFeed();
    loadNotifications();
  }, [token, loadFeed, loadNotifications]);

  // ws 实时通知到达（trendNotifyVersion 变化）：刷新消息列表
  useEffect(() => {
    if (!token || trendNotifyVersion === 0) return;
    loadNotifications();
  }, [trendNotifyVersion, token, loadNotifications]);

  // 进入消息中心：保留未读标记可见，不自动倒计时已读。
  // 由用户自己点单条（handleNotifClick）/「全部已读」按钮标记；
  // 离开消息列表（切走视图 / 卸载）时再自动「全部已读」+ 清零全局红点。
  useEffect(() => {
    if (view !== 'notifications' || !token) return;
    return () => {
      markTrendMessagesRead(token).catch(() => { /* silent */ });
      setMomentsUnreadCount(0);
      setNotifications(prev => prev.map(n => ({ ...n, read: true })));
    };
  }, [view, token, setMomentsUnreadCount]);

  // Re-sync a single trend (meta + comments + like users) into local state.
  // Called when an external panel (e.g. TrendDetailPanel) bumps the trend version.
  const refreshTrend = useCallback(async (trendId: number) => {
    if (!token) return;
    try {
      const [detailRes, treeRes, likedUsersRes] = await Promise.all([
        getTrendDetail(token, trendId),
        getCommentTree(token, [trendId]),
        getLikedUsers(token, trendId, 0, 100),
      ]);
      if (detailRes.success && detailRes.data?.trend) {
        const mapped = mapBackendTrend(detailRes.data.trend, meUserId);
        setTrends(prev => prev.map(t => t.id === trendId ? mapped : t));
      }
      if (treeRes.success && treeRes.data) {
        const map = commentTreeToMap(treeRes.data, meUserId);
        setCommentsMap(prev => ({ ...prev, [trendId]: map[trendId] || [] }));
      }
      if (likedUsersRes.success && likedUsersRes.data) {
        const users = (likedUsersRes.data.users || []).map(u => ({
          id: u.id === meUserId ? 'me' : u.id,
          name: u.nickname || '',
          avatar: u.avatar || '',
        }));
        setLikeUsersMap(prev => ({ ...prev, [trendId]: users }));
        setLikedTrends(prev => {
          const next = new Set(prev);
          if (users.some(u => u.id === 'me')) next.add(trendId); else next.delete(trendId);
          return next;
        });
      }
    } catch (err) {
      console.error('refreshTrend error', err);
    }
  }, [token, meUserId]);

  // Watch per-trend version bumps and refetch only ids this component has
  // not yet applied. Using a ref avoids re-fetching from our own mutations
  // more than once — we update the ref right after any refresh.
  const seenVersionsRef = useRef<Record<number, number>>({});
  useEffect(() => {
    if (!token || trends.length === 0) return;
    const seen = seenVersionsRef.current;
    const ids = trends.map(t => t.id);
    ids.forEach(id => {
      const v = trendVersions[id] || 0;
      if (v > (seen[id] || 0)) {
        seen[id] = v;
        refreshTrend(id);
      }
    });
  }, [token, trendVersions, trends, refreshTrend]);

  // ── Computed ──
  const unreadNotifCount = useMemo(() => notifications.filter(n => !n.read).length, [notifications]);

  const filteredTrends = useMemo(() => {
    let list = [...trends].sort((a, b) => {
      if (view === 'userTrends' && a.isTop !== b.isTop) return a.isTop ? -1 : 1;
      return b.createTime.getTime() - a.createTime.getTime();
    });
    if (view === 'userTrends') {
      // Own trends carry the 'me' sentinel as userId; normalize before matching so
      // "我的朋友圈" (userTrendsUserId = my real id) still includes them.
      const norm = (uid: string) => (uid === 'me' ? meUserId : uid);
      list = list.filter(t => norm(t.userId) === norm(userTrendsUserId));
    }
    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase();
      list = list.filter(t => t.content.toLowerCase().includes(q) || trendDisplayName(t).toLowerCase().includes(q));
    }
    return list;
  }, [trends, view, userTrendsUserId, searchQuery, meUserId]);

  const filteredNotifications = useMemo(() => {
    let list = [...notifications].sort((a, b) => b.createTime.getTime() - a.createTime.getTime());
    // 「评论」tab 聚合所有非点赞类型（评论/回复/@）
    if (notifTab === 'reply') list = list.filter(n => n.type !== 'like');
    if (notifTab === 'like') list = list.filter(n => n.type === 'like');
    return list;
  }, [notifications, notifTab]);

  const detailTrend = useMemo(() => trends.find(t => t.id === detailTrendId) || null, [trends, detailTrendId]);
  const detailComments = useMemo(() => detailTrendId ? (commentsMap[detailTrendId] || []) : [], [commentsMap, detailTrendId]);
  const likeListUsers = useMemo(() => likeListTrendId ? (likeUsersMap[likeListTrendId] || []) : [], [likeUsersMap, likeListTrendId]);

  // ── Confirm helper ──
  const doConfirm = useCallback((title: string, desc: string, confirmLabel: string, confirmColor: string, action: () => void) => {
    setConfirmOpts({
      title, description: desc, confirmLabel, confirmColor, loading: false,
      onConfirm: () => {
        setConfirmOpts(prev => prev ? { ...prev, loading: true } : prev);
        setTimeout(() => { action(); setConfirmOpts(null); }, 400);
      },
      onCancel: () => setConfirmOpts(null),
      onClose: () => setConfirmOpts(null),
    });
  }, []);

  // Resolve the real backend user id for a trend's poster. When the poster is
  // the current user, our local 'me' sentinel has been mapped, so fall back to meUserId.
  const trendAuthorBackendId = useCallback((trendId: number): string => {
    const t = trends.find(tr => tr.id === trendId);
    if (!t) return '';
    return t.userId === 'me' ? meUserId : t.userId;
  }, [trends, meUserId]);

  // ── Like handlers ──
  const handleToggleLike = useCallback((trendId: number) => {
    if (!token) {
      toast.error(t('trend.needLogin'));
      return;
    }
    // Optimistic update
    const wasLiked = likedTrends.has(trendId);
    setLikedTrends(prev => {
      const next = new Set(prev);
      if (wasLiked) next.delete(trendId); else next.add(trendId);
      return next;
    });
    setLikeUsersMap(lum => {
      if (wasLiked) {
        return { ...lum, [trendId]: (lum[trendId] || []).filter(u => u.id !== 'me') };
      }
      return { ...lum, [trendId]: [{ id: 'me', name: meName, avatar: meAvatar }, ...(lum[trendId] || [])] };
    });
    setTrends(ts => ts.map(t => t.id === trendId ? {
      ...t, agreeCount: wasLiked ? Math.max(0, t.agreeCount - 1) : t.agreeCount + 1,
    } : t));

    const authorIdStr = trendAuthorBackendId(trendId);
    const authorId = Number(authorIdStr) || 0;
    // like_type: 1 点赞 / 0 取消 (后端按 LikeType > 0 判加减)
    apiToggleLike(token, trendId, authorId, wasLiked ? 0 : 1).then(r => {
      if (!r.success) throw new Error(r.message || 'toggle failed');
      // Bump after our own seen-version is already at the current value, so no extra refetch.
      seenVersionsRef.current[trendId] = (trendVersions[trendId] || 0) + 1;
      bumpTrendVersion(trendId);
    }).catch(err => {
      console.error('toggleLike error', err);
      // Rollback optimistic update
      setLikedTrends(prev => {
        const next = new Set(prev);
        if (wasLiked) next.add(trendId); else next.delete(trendId);
        return next;
      });
      setLikeUsersMap(lum => {
        if (wasLiked) {
          return { ...lum, [trendId]: [{ id: 'me', name: meName, avatar: meAvatar }, ...(lum[trendId] || [])] };
        }
        return { ...lum, [trendId]: (lum[trendId] || []).filter(u => u.id !== 'me') };
      });
      setTrends(ts => ts.map(t => t.id === trendId ? {
        ...t, agreeCount: wasLiked ? t.agreeCount + 1 : Math.max(0, t.agreeCount - 1),
      } : t));
      toast.error(t('trend.likeFail'));
    });
  }, [token, likedTrends, meName, meAvatar, trendAuthorBackendId, trendVersions, bumpTrendVersion]);

  // Refetch the comment tree for a single trend, merging it into the map.
  const refreshCommentsFor = useCallback(async (trendId: number) => {
    if (!token) return;
    try {
      const r = await getCommentTree(token, [trendId]);
      if (r.success && r.data) {
        const mapped = commentTreeToMap(r.data, meUserId);
        setCommentsMap(prev => ({ ...prev, [trendId]: mapped[trendId] || [] }));
      }
    } catch (err) {
      console.error('refreshCommentsFor error', err);
    }
  }, [token, meUserId]);

  // ── Comment handlers ──
  const handleSubmitComment = useCallback((trendId: number, text: string, replyTo: TrendComment | null, isDetail?: boolean) => {
    if (!text.trim() || !token) return;
    const trend = trends.find(t => t.id === trendId);
    // user_id in the backend payload = the user being replied to (or trend author for top-level).
    let replyUserIdStr = '';
    if (replyTo) {
      replyUserIdStr = replyTo.replyer.id === 'me' ? meUserId : replyTo.replyer.id;
    } else if (trend) {
      replyUserIdStr = trend.userId === 'me' ? meUserId : trend.userId;
    }

    const payload = {
      trend_id: trendId,
      father: replyTo ? replyTo.id : undefined,
      user_id: Number(replyUserIdStr) || 0,
      content: text.trim(),
    };

    // Clear input immediately for snappy UX.
    if (isDetail) {
      setDetailCommentText('');
      setDetailReplyTarget(null);
    } else {
      setCommentTexts(prev => ({ ...prev, [trendId]: '' }));
      setReplyTargets(prev => ({ ...prev, [trendId]: null }));
    }

    createComment(token, payload).then(async r => {
      if (!r.success) {
        toast.error(r.message || t('trend.commentFail'));
        return;
      }
      await refreshCommentsFor(trendId);
      setTrends(ts => ts.map(t => t.id === trendId ? { ...t, replyCount: t.replyCount + 1 } : t));
      setExpandedTrendIds(prev => new Set(prev).add(trendId));
      seenVersionsRef.current[trendId] = (trendVersions[trendId] || 0) + 1;
      bumpTrendVersion(trendId);
      toast.success(t('trend.commentPublished'));
    }).catch(err => {
      console.error('createComment error', err);
      toast.error(t('trend.commentFail'));
    });
  }, [token, trends, meUserId, refreshCommentsFor, trendVersions, bumpTrendVersion]);

  const handleDeleteComment = useCallback((trendId: number, comment: TrendComment) => {
    if (!token) return;
    doConfirm(t('trend.deleteCommentTitle'), t('trend.deleteCommentDesc'), t('trend.delete'), '#E53935', () => {
      apiDeleteComment(token, comment.id).then(async r => {
        if (!r.success) {
          toast.error(r.message || t('trend.deleteFail'));
          return;
        }
        await refreshCommentsFor(trendId);
        setTrends(ts => ts.map(t => t.id === trendId ? { ...t, replyCount: Math.max(0, t.replyCount - 1) } : t));
        seenVersionsRef.current[trendId] = (trendVersions[trendId] || 0) + 1;
        bumpTrendVersion(trendId);
        toast.success(t('trend.commentDeleted'));
      }).catch(() => toast.error(t('trend.deleteFail')));
    });
  }, [token, doConfirm, refreshCommentsFor, trendVersions, bumpTrendVersion]);

  // ── Publish handler ──
  const handlePublish = useCallback((data: { type: number; content: string; title: string; images: string[]; shareUrl: string; location: string; scope: number }) => {
    if (!token) return;
    createTrend(token, {
      type: data.type,
      content: data.content,
      scope: data.scope,
      resources: data.type === 5
        ? (data.images.length > 0 ? [data.images[0]] : [])
        : (data.type === 2 ? data.images : undefined),
      position_name: data.location || undefined,
      title: data.type === 3 ? data.title : undefined,
      cover_url: data.type === 3 && data.images.length > 0 ? data.images[0] : undefined,
      share_url: data.type === 4 ? data.shareUrl : undefined,
      open_reply: true,
    }).then(r => {
      if (!r.success) {
        toast.error(r.message || t('trend.publishFail'));
        return;
      }
      setShowPublish(false);
      toast.success(t('trend.published'));
      loadFeed();
    }).catch(err => {
      console.error('publish error', err);
      toast.error(t('trend.publishFail'));
    });
  }, [token, loadFeed]);

  // ── Notification handlers ──
  const handleMarkAllRead = useCallback(() => {
    if (!token) return;
    markTrendMessagesRead(token).then(() => {
      setNotifications(prev => prev.map(n => ({ ...n, read: true })));
      setMomentsUnreadCount(0);
      toast.success(t('trend.notify.allReadDone'));
    }).catch(() => toast.error(t('trend.notify.markFailed')));
  }, [token, setMomentsUnreadCount, t]);

  const handleNotifClick = useCallback((notif: MomentsNotification) => {
    setNotifications(prev => prev.map(n => n.id === notif.id ? { ...n, read: true } : n));
    if (isMobile) {
      setDetailTrendId(notif.trendId);
      setView('feed');
    } else {
      setSelectedTrendId(notif.trendId);
      setView('feed');
    }
  }, [isMobile, setSelectedTrendId]);

  // ── Trend management ──
  // The ActionMenu items are constructed inline in the render; this handler just opens the menu.
  const handleManageTrend = useCallback((trendId: number, x: number, y: number) => {
    const trend = trends.find(t => t.id === trendId);
    if (!trend || trend.userId !== 'me') return;
    setActionMenu({ x, y, trendId });
  }, [trends]);

  // API-backed toggle of pin/reply; optimistic with rollback.
  const handleToggleTrendField = useCallback((trendId: number, field: 'isTop' | 'openReply') => {
    if (!token) return;
    const trend = trends.find(t => t.id === trendId);
    if (!trend) return;
    const currentVal = trend[field];
    // Optimistic
    setTrends(prev => prev.map(t => t.id === trendId ? { ...t, [field]: !currentVal } : t));
    const body: { trend_id: number; is_top?: number; open_reply?: number } = { trend_id: trendId };
    if (field === 'isTop') body.is_top = currentVal ? 0 : 1;
    else body.open_reply = currentVal ? 0 : 1;
    apiUpdateTrend(token, body).then(r => {
      if (!r.success) throw new Error(r.message);
      seenVersionsRef.current[trendId] = (trendVersions[trendId] || 0) + 1;
      bumpTrendVersion(trendId);
      toast.success(
        field === 'isTop'
          ? (currentVal ? t('trend.unpinned') : t('trend.pinnedToast'))
          : (currentVal ? t('trend.commentClosed') : t('trend.commentOpened')),
      );
    }).catch(err => {
      console.error('updateTrend error', err);
      setTrends(prev => prev.map(t => t.id === trendId ? { ...t, [field]: currentVal } : t));
      toast.error(t('trend.opFail'));
    });
  }, [token, trends, trendVersions, bumpTrendVersion]);

  const handleDeleteTrend = useCallback((trendId: number) => {
    if (!token) return;
    apiDeleteTrend(token, trendId).then(r => {
      if (!r.success) {
        toast.error(r.message || t('trend.deleteFail'));
        return;
      }
      // Close the detail panel if it was showing the deleted trend.
      if (selectedTrendId === trendId) setSelectedTrendId(null);
      bumpTrendVersion(trendId);
      setTrends(prev => prev.filter(t => t.id !== trendId));
      setCommentsMap(prev => { const next = { ...prev }; delete next[trendId]; return next; });
      setLikeUsersMap(prev => { const next = { ...prev }; delete next[trendId]; return next; });
      toast.success(t('trend.trendDeleted'));
    }).catch(() => toast.error(t('trend.deleteFail')));
  }, [token, selectedTrendId, setSelectedTrendId, bumpTrendVersion]);

  // ── Navigation ──
  const handleAvatarClick = useCallback((userId: string, _userName?: string) => {
    if (!userId) return;
    showUserCard(userId);
  }, [showUserCard]);

  // Consume a cross-component request to open a user's 朋友圈 (from profile cards / 我的-朋友圈).
  useEffect(() => {
    if (!userTrendsTarget) return;
    setView('userTrends');
    setUserTrendsUserId(userTrendsTarget.id);
    setUserTrendsUserName(userTrendsTarget.name || '');
    setSearchQuery('');
    clearUserTrendsTarget();
  }, [userTrendsTarget, clearUserTrendsTarget]);

  // Load the viewed user's profile (cover/avatar/name/signature) for the userTrends header.
  useEffect(() => {
    if (view !== 'userTrends' || !userTrendsUserId || !token) return;
    const isSelf = !!meUserId && userTrendsUserId === meUserId;
    if (isSelf) {
      setUserTrendsProfile({ avatar: meAvatar, name: meName, signature: meSignature, cover: meCover });
      return;
    }
    setUserTrendsProfile({ avatar: '', name: userTrendsUserName, signature: '', cover: '' });
    (async () => {
      try {
        const r = await fetch(`/api/user/search?ids=${encodeURIComponent(userTrendsUserId)}`, { headers: { Authorization: `Bearer ${token}` } });
        const j = await r.json();
        const u = (j?.data?.users || [])[0];
        if (u) {
          setUserTrendsProfile({
            avatar: u.avatar || '',
            name: friendDisplayName(userTrendsUserId, u.nickname || userTrendsUserName),
            signature: u.introduction || '',
            cover: u.moments_cover || '',
          });
        }
      } catch (err) {
        console.error('load userTrends profile error', err);
      }
    })();
  }, [view, userTrendsUserId, token, meUserId, meAvatar, meName, meSignature, meCover, userTrendsUserName]);

  // Load user-trends when entering that view.
  useEffect(() => {
    if (view !== 'userTrends' || !userTrendsUserId || !token) return;
    (async () => {
      try {
        const r = await getUserTrends(token, userTrendsUserId);
        if (r.success && r.data) {
          const list = (r.data.list || []).map(t => mapBackendTrend(t, meUserId));
          const topList = (r.data.top_list || []).map(t => mapBackendTrend(t, meUserId));
          const combined = [...topList, ...list];
          // Merge user trends into the trends state (replace if id exists, else add).
          setTrends(prev => {
            const byId = new Map(prev.map(t => [t.id, t]));
            combined.forEach(t => byId.set(t.id, t));
            return Array.from(byId.values());
          });
          if (!userTrendsUserName && combined.length > 0 && combined[0].userName) {
            setUserTrendsUserName(combined[0].userName);
          }
        }
      } catch (err) {
        console.error('getUserTrends error', err);
      }
    })();
  }, [view, userTrendsUserId, token, meUserId, userTrendsUserName]);

  const handleBackFromUserTrends = useCallback(() => {
    setView('feed');
    setUserTrendsUserId('');
    setSearchQuery('');
  }, []);

  const handleBackFromNotifications = useCallback(() => {
    setView(notifBackView);
    setNotifTab('all');
  }, [notifBackView]);

  // ── Detail navigation helpers ──
  const handleOpenDetail = useCallback((trendId: number) => {
    if (isMobile) {
      // Mobile: open modal
      setDetailTrendId(trendId);
      setDetailCommentText('');
      setDetailReplyTarget(null);
    } else {
      // Desktop: use store to open right panel
      setSelectedTrendId(trendId);
    }
  }, [isMobile, setSelectedTrendId]);

  // ── Render ──
  const renderFeedHeader = () => (
    <div className="flex items-center justify-between shrink-0" style={{ padding: '8px 12px', background: '#FFF', borderBottom: '1px solid rgba(0,0,0,0.06)' }}>
      <div className="relative flex-1" style={{ maxWidth: 280 }}>
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4" style={{ color: '#A2ACB5' }} />
        <input
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          placeholder={t('trend.searchPlaceholder')}
          style={{ width: '100%', borderRadius: '20px', background: '#F0F2F5', border: 'none', padding: '7px 14px 7px 34px', fontSize: '13px', color: '#1C2733', outline: 'none' }}
        />
      </div>
      <div className="flex items-center gap-1 shrink-0 ml-2">
        {/* Notification bell */}
        <div className="relative">
          <button
            onClick={() => { setNotifBackView('feed'); setView('notifications'); setNotifTab('all'); }}
            style={{ width: 36, height: 36, borderRadius: '50%', border: 'none', background: 'transparent', color: '#708499', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}
          >
            <Bell className="w-5 h-5" />
          </button>
          {unreadNotifCount > 0 && (
            <span className="absolute flex items-center justify-center" style={{ top: 4, right: 2, minWidth: 16, height: 16, borderRadius: 8, background: '#E53935', color: '#FFF', fontSize: '10px', fontWeight: 700, padding: '0 4px' }}>
              {unreadNotifCount > 99 ? '99+' : unreadNotifCount}
            </span>
          )}
        </div>
        {/* Publish */}
        <button
          onClick={() => setShowPublish(true)}
          style={{ width: 36, height: 36, borderRadius: '50%', border: 'none', background: '#1BB45B', color: '#FFF', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}
        >
          <Plus className="w-5 h-5" style={{ color: '#FFF' }} />
        </button>
      </div>
    </div>
  );

  const renderCoverSection = (opts: { cover: string; avatar: string; name: string; uid: string; editable: boolean }) => (
    <div className="relative">
      <div
        className="relative overflow-hidden"
        style={{ height: 176 }}
      >
        {opts.cover ? (
          <img
            src={opts.cover}
            alt="cover"
            className="w-full h-full object-cover"
          />
        ) : (
          <>
            <img
              src="https://picsum.photos/seed/tg-cover-feed/800/400"
              alt="cover"
              className="w-full h-full object-cover"
              style={{ opacity: 0.3, mixBlendMode: 'overlay' }}
            />
            <div
              className="absolute inset-0"
              style={{
                background: 'linear-gradient(135deg, rgba(27,180,91,0.7), rgba(95,214,143,0.5))',
              }}
            />
          </>
        )}
        {opts.editable && (
          <>
            <input
              ref={coverInputRef}
              type="file"
              accept="image/*"
              style={{ display: 'none' }}
              onChange={(e) => { const f = e.target.files?.[0]; if (f) handleCoverUpload(f); e.target.value = ''; }}
            />
            <button
              onClick={() => coverInputRef.current?.click()}
              style={{ position: 'absolute', top: 10, right: 10, zIndex: 2, border: 'none', borderRadius: 14, padding: '4px 10px', background: 'rgba(0,0,0,0.35)', color: '#FFF', fontSize: 12, cursor: 'pointer' }}
            >
              {t('trend.changeCover')}
            </button>
          </>
        )}
        <div
          className="absolute inset-0"
          style={{
            background: 'linear-gradient(to top, rgba(23,33,43,0.45) 0%, transparent 60%)',
          }}
        />
      </div>
      <div className="absolute" style={{ bottom: -16, left: 12, cursor: 'pointer' }} onClick={() => opts.uid && showUserCard(opts.uid)}>
        {opts.avatar ? (
          <img
            src={opts.avatar}
            alt={opts.name}
            style={{ width: 68, height: 68, borderRadius: '50%', objectFit: 'cover', boxShadow: '0 0 0 3px rgba(27,180,91,0.25), 0 2px 8px rgba(0,0,0,0.1)' }}
          />
        ) : (
          <div
            style={{ width: 68, height: 68, borderRadius: '50%', backgroundColor: getAvatarColor(opts.name), display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 26, fontWeight: 600, color: '#FFF', boxShadow: '0 0 0 3px rgba(27,180,91,0.25), 0 2px 8px rgba(0,0,0,0.1)' }}
          >
            {opts.name ? opts.name[0] : '?'}
          </div>
        )}
      </div>
    </div>
  );

  const renderUserInfo = (opts: { name: string; signature: string; uid: string }) => (
    <div style={{ padding: '20px 12px 10px' }}>
      <h3 style={{ fontSize: '16px', fontWeight: 600, color: '#1C2733', cursor: 'pointer', display: 'inline-block' }} onClick={() => opts.uid && showUserCard(opts.uid)}>{opts.name}</h3>
      <p style={{ fontSize: '12px', color: '#708499', marginTop: 2 }}>{opts.signature}</p>
    </div>
  );

  const pillTab = (active: boolean, onClick: () => void, label: string, badge?: number) => (
    <button onClick={onClick} style={{ padding: '6px 18px', borderRadius: '17px', border: 'none', background: active ? '#1BB45B' : 'transparent', color: active ? '#FFF' : '#646A73', fontSize: '13px', fontWeight: 500, cursor: 'pointer', transition: 'all 0.2s', display: 'flex', alignItems: 'center', gap: 4 }}>
      {label}
      {badge !== undefined && badge > 0 && !active && (
        <span className="inline-flex items-center justify-center" style={{ width: 16, height: 16, borderRadius: 8, background: '#E53935', color: '#FFF', fontSize: '10px', fontWeight: 700 }}>{badge > 9 ? '9+' : badge}</span>
      )}
    </button>
  );

  // ═══════════════════════════════════════
  // VIEW: Feed
  // ═══════════════════════════════════════
  if (view === 'feed') {
    return (
      <div className="h-full flex flex-col relative" style={{ background: '#F5F7FA' }}>
        {renderFeedHeader()}

        <div className="flex-1 overflow-y-auto im-scroll">
          {renderCoverSection({ cover: meCover, avatar: meAvatar, name: meName, uid: meUserId, editable: true })}
          {renderUserInfo({ name: meName, signature: meSignature, uid: meUserId })}
          <div style={{ height: 1, background: 'rgba(0,0,0,0.06)' }} />

          {/* Trend cards */}
          <div style={{ background: '#FFF', borderRadius: '12px 12px 0 0', margin: '0 0 0' }}>
            {loading ? (
              <>
                <SkeletonCard />
                <SkeletonCard />
                <SkeletonCard />
              </>
            ) : filteredTrends.length > 0 ? (
              filteredTrends.map(trend => (
                <TrendCard
                  key={trend.id}
                  trend={trend}
                  showTopBadge={false}
                  liked={likedTrends.has(trend.id)}
                  likeCount={trend.agreeCount}
                  likeUsers={likeUsersMap[trend.id] || []}
                  comments={commentsMap[trend.id] || []}
                  expanded={expandedTrendIds.has(trend.id)}
                  replyTarget={replyTargets[trend.id] || null}
                  commentText={commentTexts[trend.id] || ''}
                  selected={!isMobile && selectedTrendId === trend.id}
                  onToggleLike={() => handleToggleLike(trend.id)}
                  onLikeCountClick={() => setLikeListTrendId(trend.id)}
                  onExpandComments={() => setExpandedTrendIds(prev => new Set(prev).add(trend.id))}
                  onSetReplyTarget={(c) => setReplyTargets(prev => ({ ...prev, [trend.id]: c }))}
                  onCommentTextChange={(v) => setCommentTexts(prev => ({ ...prev, [trend.id]: v }))}
                  onSubmitComment={() => handleSubmitComment(trend.id, commentTexts[trend.id] || '', replyTargets[trend.id] || null)}
                  onDeleteComment={(c) => handleDeleteComment(trend.id, c)}
                  onOpenDetail={() => handleOpenDetail(trend.id)}
                  onAvatarClick={() => handleAvatarClick(trend.userId, trendDisplayName(trend))}
                  onManage={(x, y) => handleManageTrend(trend.id, x, y)}
                />
              ))
            ) : (
              <div className="flex flex-col items-center justify-center" style={{ padding: '60px 24px' }}>
                <Inbox className="w-16 h-16" style={{ color: '#D1D5DB', marginBottom: 16 }} />
                <div style={{ fontSize: '15px', fontWeight: 500, color: '#646A73', marginBottom: 8 }}>
                  {searchQuery ? t('trend.noMatch') : t('trend.empty')}
                </div>
                <div style={{ fontSize: '13px', color: '#A2ACB5' }}>
                  {searchQuery ? t('trend.noMatchHint') : t('trend.emptyHint')}
                </div>
              </div>
            )}
          </div>

          <div style={{ padding: '20px 0 12px', textAlign: 'center', fontSize: '12px', color: '#A2ACB5' }}>
            {t('trend.reachedEnd')}
          </div>
        </div>

        {/* FAB button */}
        <button
          onClick={() => setShowPublish(true)}
          className="absolute"
          style={{ bottom: 24, right: 20, width: 52, height: 52, borderRadius: '50%', background: '#1BB45B', color: '#FFF', border: 'none', boxShadow: '0 4px 16px rgba(27,180,91,0.4)', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', transition: 'transform 0.2s' }}
          onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.transform = 'scale(1.05)'; }}
          onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.transform = 'scale(1)'; }}
        >
          <Plus className="w-6 h-6" style={{ color: '#FFF' }} />
        </button>

        {/* Modals */}
        <PublishModal open={showPublish} token={token} onClose={() => setShowPublish(false)} onSubmit={handlePublish} />
        <TrendDetailModal
          open={!!detailTrendId}
          trend={detailTrend}
          liked={detailTrendId ? likedTrends.has(detailTrendId) : false}
          likeCount={detailTrend?.agreeCount || 0}
          likeUsers={detailTrendId ? (likeUsersMap[detailTrendId] || []) : []}
          comments={detailComments}
          commentText={detailCommentText}
          replyTarget={detailReplyTarget}
          onToggleLike={() => detailTrendId && handleToggleLike(detailTrendId)}
          onLikeCountClick={() => detailTrendId && setLikeListTrendId(detailTrendId)}
          onSubmitComment={() => detailTrendId && handleSubmitComment(detailTrendId, detailCommentText, detailReplyTarget, true)}
          onCommentTextChange={setDetailCommentText}
          onSetReplyTarget={setDetailReplyTarget}
          onDeleteComment={(c) => detailTrendId && handleDeleteComment(detailTrendId, c)}
          onAvatarClick={handleAvatarClick}
          onClose={() => { setDetailTrendId(null); setDetailCommentText(''); setDetailReplyTarget(null); }}
        />
        <LikeListModal open={!!likeListTrendId} users={likeListUsers} onClose={() => setLikeListTrendId(null)} />
        <ConfirmModal open={!!confirmOpts} opts={confirmOpts} />

        {/* Action menu */}
        {actionMenu && (
          <ActionMenu
            x={actionMenu.x}
            y={actionMenu.y}
            items={[
              {
                label: (() => { const tr = trends.find(tr => tr.id === actionMenu.trendId); return tr?.isTop ? t('trend.unpin') : t('trend.pin'); })(),
                icon: (() => { const tr = trends.find(tr => tr.id === actionMenu.trendId); return tr?.isTop ? <PinOff className="w-3.5 h-3.5" /> : <Pin className="w-3.5 h-3.5" />; })(),
                onClick: () => handleToggleTrendField(actionMenu.trendId, 'isTop'),
              },
              {
                label: (() => { const tr = trends.find(tr => tr.id === actionMenu.trendId); return tr?.openReply ? t('trend.closeComment') : t('trend.openComment'); })(),
                icon: (() => { const tr = trends.find(tr => tr.id === actionMenu.trendId); return tr?.openReply ? <MessageSquareOff className="w-3.5 h-3.5" /> : <MessageCircle className="w-3.5 h-3.5" />; })(),
                onClick: () => handleToggleTrendField(actionMenu.trendId, 'openReply'),
              },
              {
                label: t('trend.deleteTrend'),
                color: '#FF5252',
                icon: <Trash2 className="w-3.5 h-3.5" />,
                onClick: () => {
                  const tid = actionMenu.trendId;
                  doConfirm(t('trend.deleteTrend'), t('trend.deleteTrendDesc'), t('trend.delete'), '#E53935', () => handleDeleteTrend(tid));
                },
              },
            ]}
            onClose={() => setActionMenu(null)}
          />
        )}
      </div>
    );
  }

  // ═══════════════════════════════════════
  // VIEW: Notifications
  // ═══════════════════════════════════════
  if (view === 'notifications') {
    const unreadReplyCount = notifications.filter(n => !n.read && n.type !== 'like').length;
    const unreadLikeCount = notifications.filter(n => !n.read && n.type === 'like').length;

    return (
      <div className="h-full flex flex-col" style={{ background: '#F5F7FA' }}>
        {/* Header */}
        <div className="flex items-center justify-between shrink-0" style={{ height: 56, background: '#FFF', borderBottom: '1px solid rgba(0,0,0,0.08)', paddingLeft: 4, paddingRight: 16 }}>
          <button onClick={handleBackFromNotifications} style={{ width: 40, height: 40, borderRadius: '50%', border: 'none', background: 'transparent', color: '#1BB45B', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <ArrowLeft className="w-5 h-5" />
          </button>
          <span style={{ fontSize: '17px', fontWeight: 600, color: '#1C2733' }}>{t('trend.notify.title')}</span>
          <button onClick={handleMarkAllRead} style={{ fontSize: '13px', color: '#1BB45B', background: 'none', border: 'none', cursor: 'pointer', fontWeight: 500 }}>{t('trend.notify.markAllRead')}</button>
        </div>

        {/* Tabs */}
        <div className="flex items-center shrink-0" style={{ padding: '12px 16px 8px', background: '#FFF', borderBottom: '1px solid rgba(0,0,0,0.05)' }}>
          <div className="flex items-center" style={{ borderRadius: '20px', background: 'rgba(0,0,0,0.04)', padding: '3px' }}>
            {pillTab(notifTab === 'all', () => setNotifTab('all'), t('trend.notify.tab.all'), unreadNotifCount)}
            {pillTab(notifTab === 'reply', () => setNotifTab('reply'), t('trend.notify.tab.comment'), unreadReplyCount)}
            {pillTab(notifTab === 'like', () => setNotifTab('like'), t('trend.notify.tab.like'), unreadLikeCount)}
          </div>
        </div>

        {/* Notification list */}
        <div className="flex-1 overflow-y-auto im-scroll">
          {filteredNotifications.length > 0 ? (
            <div style={{ background: '#FFF', borderRadius: '12px', margin: '8px', overflow: 'hidden' }}>
              {filteredNotifications.map(notif => (
                <div
                  key={notif.id}
                  className="relative"
                  style={{
                    padding: '12px 16px',
                    borderBottom: '1px solid rgba(0,0,0,0.05)',
                    background: notif.read ? '#FFF' : 'rgba(245,166,35,0.03)',
                    borderLeft: notif.read ? 'none' : '3px solid #F5A623',
                    cursor: 'pointer',
                    transition: 'background 0.15s',
                  }}
                  onClick={() => handleNotifClick(notif)}
                  onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'rgba(0,0,0,0.02)'; }}
                  onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = notif.read ? '#FFF' : 'rgba(245,166,35,0.03)'; }}
                >
                  <div className="flex gap-3">
                    {avatarCircle(friendDisplayName(notif.actor.id, notif.actor.name), 40, undefined, notif.actor.avatar)}
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2" style={{ marginBottom: 3 }}>
                        <span style={{ fontSize: '14px', fontWeight: 600, color: '#1C2733' }}>{friendDisplayName(notif.actor.id, notif.actor.name)}</span>
                        <span style={{ fontSize: '12px', color: '#708499' }}>{notifActionText(notif.type, t)}</span>
                        {!notif.read && (
                          <span style={{ width: 8, height: 8, borderRadius: '50%', background: '#E53935', flexShrink: 0, marginLeft: 'auto' }} />
                        )}
                      </div>
                      {notif.content ? (
                        <div style={{ fontSize: '13px', color: '#646A73', lineHeight: '1.5', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                          <span>{notif.content}</span>
                        </div>
                      ) : null}
                      <div style={{ fontSize: '11px', color: '#A2ACB5', marginTop: 4 }}>{fmtTime(notif.createTime, t)}</div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center" style={{ padding: '60px 24px' }}>
              <Bell className="w-16 h-16" style={{ color: '#D1D5DB', marginBottom: 16 }} />
              <div style={{ fontSize: '15px', fontWeight: 500, color: '#646A73', marginBottom: 8 }}>
                {notifTab === 'all' ? t('trend.notify.empty.all') : notifTab === 'reply' ? t('trend.notify.empty.comment') : t('trend.notify.empty.like')}
              </div>
              <div style={{ fontSize: '13px', color: '#A2ACB5' }}>{t('trend.notify.emptyHint')}</div>
            </div>
          )}
        </div>

        {/* Modals (reuse from feed) */}
        <TrendDetailModal
          open={!!detailTrendId}
          trend={detailTrend}
          liked={detailTrendId ? likedTrends.has(detailTrendId) : false}
          likeCount={detailTrend?.agreeCount || 0}
          likeUsers={detailTrendId ? (likeUsersMap[detailTrendId] || []) : []}
          comments={detailComments}
          commentText={detailCommentText}
          replyTarget={detailReplyTarget}
          onToggleLike={() => detailTrendId && handleToggleLike(detailTrendId)}
          onLikeCountClick={() => detailTrendId && setLikeListTrendId(detailTrendId)}
          onSubmitComment={() => detailTrendId && handleSubmitComment(detailTrendId, detailCommentText, detailReplyTarget, true)}
          onCommentTextChange={setDetailCommentText}
          onSetReplyTarget={setDetailReplyTarget}
          onDeleteComment={(c) => detailTrendId && handleDeleteComment(detailTrendId, c)}
          onAvatarClick={handleAvatarClick}
          onClose={() => { setDetailTrendId(null); setDetailCommentText(''); setDetailReplyTarget(null); }}
        />
        <LikeListModal open={!!likeListTrendId} users={likeListUsers} onClose={() => setLikeListTrendId(null)} />
        <ConfirmModal open={!!confirmOpts} opts={confirmOpts} />
      </div>
    );
  }

  // ═══════════════════════════════════════
  // VIEW: User Trends
  // ═══════════════════════════════════════
  if (view === 'userTrends') {
    const isSelfTrends = !!meUserId && userTrendsUserId === meUserId;
    const headerName = userTrendsProfile.name || userTrendsUserName;
    return (
      <div className="h-full flex flex-col relative" style={{ background: '#F5F7FA' }}>
        {/* Slim top bar: back + title (+ publish on my own circle) */}
        <div className="flex items-center justify-between shrink-0" style={{ height: 56, background: '#FFF', borderBottom: '1px solid rgba(0,0,0,0.08)', paddingLeft: 4, paddingRight: 16 }}>
          <button onClick={handleBackFromUserTrends} style={{ width: 40, height: 40, borderRadius: '50%', border: 'none', background: 'transparent', color: '#1BB45B', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <ArrowLeft className="w-5 h-5" />
          </button>
          <span style={{ fontSize: '17px', fontWeight: 600, color: '#1C2733' }}>{isSelfTrends ? t('moments.mine') : t('trend.othersTrends').replace('{name}', headerName)}</span>
          {/* Message queue (互动消息) + publish — only on my own 朋友圈 */}
          {isSelfTrends ? (
            <div className="flex items-center gap-1 shrink-0">
              <div className="relative">
                <button
                  onClick={() => { setNotifBackView('userTrends'); setView('notifications'); setNotifTab('all'); }}
                  style={{ width: 36, height: 36, borderRadius: '50%', border: 'none', background: 'transparent', color: '#708499', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}
                >
                  <Bell className="w-5 h-5" />
                </button>
                {unreadNotifCount > 0 && (
                  <span className="absolute flex items-center justify-center" style={{ top: 4, right: 2, minWidth: 16, height: 16, borderRadius: 8, background: '#E53935', color: '#FFF', fontSize: '10px', fontWeight: 700, padding: '0 4px' }}>
                    {unreadNotifCount > 99 ? '99+' : unreadNotifCount}
                  </span>
                )}
              </div>
              <button
                onClick={() => setShowPublish(true)}
                style={{ width: 36, height: 36, borderRadius: '50%', border: 'none', background: '#1BB45B', color: '#FFF', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}
              >
                <Plus className="w-5 h-5" style={{ color: '#FFF' }} />
              </button>
            </div>
          ) : (
            <div style={{ width: 40 }} />
          )}
        </div>

        {/* Same layout as the public feed: cover + user info + trend cards */}
        <div className="flex-1 overflow-y-auto im-scroll">
          {renderCoverSection({ cover: userTrendsProfile.cover, avatar: userTrendsProfile.avatar, name: headerName, uid: userTrendsUserId, editable: isSelfTrends })}
          {renderUserInfo({ name: headerName, signature: userTrendsProfile.signature, uid: userTrendsUserId })}
          <div style={{ height: 1, background: 'rgba(0,0,0,0.06)' }} />

          <div style={{ background: '#FFF', borderRadius: '12px 12px 0 0', margin: '0 0 0' }}>
            {filteredTrends.length > 0 ? (
              filteredTrends.map(trend => (
                <TrendCard
                  key={trend.id}
                  trend={trend}
                  showTopBadge={false}
                  liked={likedTrends.has(trend.id)}
                  likeCount={trend.agreeCount}
                  likeUsers={likeUsersMap[trend.id] || []}
                  comments={commentsMap[trend.id] || []}
                  expanded={expandedTrendIds.has(trend.id)}
                  replyTarget={replyTargets[trend.id] || null}
                  commentText={commentTexts[trend.id] || ''}
                  selected={!isMobile && selectedTrendId === trend.id}
                  onToggleLike={() => handleToggleLike(trend.id)}
                  onLikeCountClick={() => setLikeListTrendId(trend.id)}
                  onExpandComments={() => setExpandedTrendIds(prev => new Set(prev).add(trend.id))}
                  onSetReplyTarget={(c) => setReplyTargets(prev => ({ ...prev, [trend.id]: c }))}
                  onCommentTextChange={(v) => setCommentTexts(prev => ({ ...prev, [trend.id]: v }))}
                  onSubmitComment={() => handleSubmitComment(trend.id, commentTexts[trend.id] || '', replyTargets[trend.id] || null)}
                  onDeleteComment={(c) => handleDeleteComment(trend.id, c)}
                  onOpenDetail={() => handleOpenDetail(trend.id)}
                  onAvatarClick={() => handleAvatarClick(trend.userId, trendDisplayName(trend))}
                  onManage={(x, y) => handleManageTrend(trend.id, x, y)}
                />
              ))
            ) : (
              <div className="flex flex-col items-center justify-center" style={{ padding: '60px 24px' }}>
                <Inbox className="w-16 h-16" style={{ color: '#D1D5DB', marginBottom: 16 }} />
                <div style={{ fontSize: '15px', fontWeight: 500, color: '#646A73', marginBottom: 8 }}>{t('trend.empty')}</div>
                <div style={{ fontSize: '13px', color: '#A2ACB5' }}>{isSelfTrends ? t('trend.emptyHint') : t('trend.userEmptyHint')}</div>
              </div>
            )}
          </div>

          <div style={{ padding: '20px 0 12px', textAlign: 'center', fontSize: '12px', color: '#A2ACB5' }}>
            {t('trend.reachedEnd')}
          </div>
        </div>

        {/* Modals */}
        <PublishModal open={showPublish} token={token} onClose={() => setShowPublish(false)} onSubmit={handlePublish} />
        <TrendDetailModal
          open={!!detailTrendId}
          trend={detailTrend}
          liked={detailTrendId ? likedTrends.has(detailTrendId) : false}
          likeCount={detailTrend?.agreeCount || 0}
          likeUsers={detailTrendId ? (likeUsersMap[detailTrendId] || []) : []}
          comments={detailComments}
          commentText={detailCommentText}
          replyTarget={detailReplyTarget}
          onToggleLike={() => detailTrendId && handleToggleLike(detailTrendId)}
          onLikeCountClick={() => detailTrendId && setLikeListTrendId(detailTrendId)}
          onSubmitComment={() => detailTrendId && handleSubmitComment(detailTrendId, detailCommentText, detailReplyTarget, true)}
          onCommentTextChange={setDetailCommentText}
          onSetReplyTarget={setDetailReplyTarget}
          onDeleteComment={(c) => detailTrendId && handleDeleteComment(detailTrendId, c)}
          onAvatarClick={handleAvatarClick}
          onClose={() => { setDetailTrendId(null); setDetailCommentText(''); setDetailReplyTarget(null); }}
        />
        <LikeListModal open={!!likeListTrendId} users={likeListUsers} onClose={() => setLikeListTrendId(null)} />
        <ConfirmModal open={!!confirmOpts} opts={confirmOpts} />

        {actionMenu && (
          <ActionMenu
            x={actionMenu.x}
            y={actionMenu.y}
            items={[
              {
                label: (() => { const tr = trends.find(tr => tr.id === actionMenu.trendId); return tr?.isTop ? t('trend.unpin') : t('trend.pin'); })(),
                icon: (() => { const tr = trends.find(tr => tr.id === actionMenu.trendId); return tr?.isTop ? <PinOff className="w-3.5 h-3.5" /> : <Pin className="w-3.5 h-3.5" />; })(),
                onClick: () => handleToggleTrendField(actionMenu.trendId, 'isTop'),
              },
              {
                label: t('trend.deleteTrend'),
                color: '#FF5252',
                icon: <Trash2 className="w-3.5 h-3.5" />,
                onClick: () => {
                  const tid = actionMenu.trendId;
                  doConfirm(t('trend.deleteTrend'), t('trend.deleteTrendDescShort'), t('trend.delete'), '#E53935', () => handleDeleteTrend(tid));
                },
              },
            ]}
            onClose={() => setActionMenu(null)}
          />
        )}
      </div>
    );
  }

  // Fallback (should never reach)
  return null;
}
