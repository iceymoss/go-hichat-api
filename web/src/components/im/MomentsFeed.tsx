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
import { getAvatarColor } from '@/lib/utils';
import { useIMStore } from '@/lib/im-store';
import { useIsMobile } from '@/hooks/use-mobile';
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
  getUnread,
  markCommentsRead,
  markLikesRead,
  toggleLike as apiToggleLike,
  getBatchLikeSummary,
  mapBackendTrend,
  commentTreeToMap,
  batchSummaryToLikeUsersMap,
  unreadToNotifications,
} from '@/lib/trend-api';

/* ═══════════════════════════════════════
   Helpers
   ═══════════════════════════════════════ */

function fmtTime(date: Date): string {
  const now = Date.now();
  const diff = now - date.getTime();
  const m = Math.floor(diff / 60000);
  const h = Math.floor(diff / 3600000);
  const d = Math.floor(diff / 86400000);
  if (m < 1) return '刚刚';
  if (m < 60) return `${m}分钟前`;
  if (h < 24) return `${h}小时前`;
  if (d < 30) return `${new Date(date).getMonth() + 1}月${new Date(date).getDate()}日`;
  return `${new Date(date).getFullYear()}/${new Date(date).getMonth() + 1}/${new Date(date).getDate()}`;
}

function getUserName(userId: string, fallback?: string): string {
  if (userId === 'me') return fallback || useIMStore.getState().currentUser?.name || '我';
  if (fallback) return fallback;
  const c = useIMStore.getState().friends.find(ct => ct.id === userId);
  return c ? (c.remark || c.name) : userId;
}

function trendDisplayName(trend: Trend): string {
  return trend.userName || getUserName(trend.userId);
}

function avatarCircle(name: string, size: number, extra?: React.ReactNode) {
  return (
    <div className="relative shrink-0" style={{ width: size, height: size, borderRadius: '50%', backgroundColor: getAvatarColor(name), display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: size * 0.38, fontWeight: 600, color: '#FFF', overflow: 'hidden' }}>
      {name[0]}
      {extra}
    </div>
  );
}

const inputStyle = { width: '100%', borderRadius: '10px', border: '1px solid rgba(0,0,0,0.1)', padding: '10px 12px', fontSize: '14px', color: '#1C2733', outline: 'none', background: '#F5F7FA', boxSizing: 'border-box' as const };
const focusInput = (e: React.FocusEvent<HTMLInputElement | HTMLTextAreaElement>) => { e.target.style.borderColor = '#3390EC'; e.target.style.boxShadow = '0 0 0 3px rgba(51,144,236,0.15)'; };
const blurInput = (e: React.FocusEvent<HTMLInputElement | HTMLTextAreaElement>) => { e.target.style.borderColor = 'rgba(0,0,0,0.1)'; e.target.style.boxShadow = 'none'; };

const trendTypeLabels: Record<number, { label: string; icon: React.ReactNode }> = {
  1: { label: '文本', icon: <Type className="w-4 h-4" /> },
  2: { label: '图文', icon: <ImageIcon className="w-4 h-4" /> },
  3: { label: '文章', icon: <FileText className="w-4 h-4" /> },
  4: { label: '分享', icon: <Link2 className="w-4 h-4" /> },
  5: { label: '视频', icon: <Video className="w-4 h-4" /> },
};

const scopeLabels: Record<number, { label: string; icon: React.ReactNode }> = {
  1: { label: '仅自己', icon: <Lock className="w-3.5 h-3.5" /> },
  2: { label: '仅好友', icon: <Users className="w-3.5 h-3.5" /> },
  3: { label: '所有人', icon: <Globe className="w-3.5 h-3.5" /> },
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
          <button onClick={opts.onCancel} disabled={opts.loading} style={{ padding: '8px 20px', borderRadius: '8px', border: '1px solid rgba(0,0,0,0.1)', background: '#FFF', color: '#646A73', fontSize: '14px', fontWeight: 500, cursor: opts.loading ? 'not-allowed' : 'pointer' }}>取消</button>
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
  return (
    <div>
      <div className="flex items-start gap-2" style={{ padding: '4px 0' }}>
        <div style={{ fontSize: '12px', lineHeight: '1.6', flex: 1 }}>
          <span style={{ color: '#3390EC', fontWeight: 600, cursor: 'pointer' }}>{comment.replyer.name}</span>
          {comment.father !== 0 && comment.user && comment.user.id !== comment.replyer.id && (
            <span> 回复 <span style={{ color: '#3390EC', fontWeight: 500 }}>{comment.user.name}</span></span>
          )}
          <span style={{ color: '#1C2733' }}>：{comment.content}</span>
        </div>
        <div className="flex items-center gap-1 shrink-0" style={{ marginTop: 2 }}>
          {comment.replyer.id === 'me' && (
            <button onClick={() => onDelete(comment)} style={{ padding: '2px 6px', border: 'none', background: 'transparent', color: '#A2ACB5', fontSize: '11px', cursor: 'pointer', borderRadius: '4px' }} onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.color = '#E53935'; }} onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.color = '#A2ACB5'; }}>
              删除
            </button>
          )}
          <button onClick={() => onReply(comment)} style={{ padding: '2px 6px', border: 'none', background: 'transparent', color: '#A2ACB5', fontSize: '11px', cursor: 'pointer', borderRadius: '4px' }} onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.color = '#3390EC'; }} onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.color = '#A2ACB5'; }}>
            回复
          </button>
        </div>
      </div>
      {comment.children && comment.children.length > 0 && (
        <div style={{ marginLeft: 12, paddingLeft: 10, borderLeft: '2px solid rgba(51,144,236,0.15)' }}>
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
  trend, liked, likeCount, likeUsers, comments, expanded,
  replyTarget, commentText, selected,
  onToggleLike, onLikeCountClick, onExpandComments,
  onSetReplyTarget, onCommentTextChange, onSubmitComment, onDeleteComment,
  onOpenDetail, onAvatarClick, onManage,
}: TrendCardProps) {
  const [likeAnim, setLikeAnim] = useState(false);
  const [likeNamesExpanded, setLikeNamesExpanded] = useState(false);
  const userName = trendDisplayName(trend);
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
        return <span key={i} style={{ color: '#3390EC', cursor: 'pointer' }}>{atUser ? atUser.name : part}</span>;
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
        background: selected ? 'rgba(51,144,236,0.06)' : undefined,
        borderLeft: selected ? '3px solid #3390EC' : undefined,
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
              style={{ color: '#3390EC' }}
              onClick={onAvatarClick}
            >
              {userName}
            </span>
            {trend.isTop && (
              <span style={{ fontSize: '10px', fontWeight: 500, color: '#F5A623', backgroundColor: 'rgba(245,166,35,0.1)', borderRadius: '4px', padding: '1px 6px', display: 'inline-flex', alignItems: 'center', gap: 2 }}>
                <Pin className="w-3 h-3" />置顶
              </span>
            )}
            {!trend.openReply && (
              <span style={{ fontSize: '10px', fontWeight: 500, color: '#A2ACB5', backgroundColor: 'rgba(0,0,0,0.04)', borderRadius: '4px', padding: '1px 6px', display: 'inline-flex', alignItems: 'center', gap: 2 }}>
                <MessageSquareOff className="w-3 h-3" />评论已关闭
              </span>
            )}
            <span style={{ fontSize: '11px', color: '#A2ACB5', marginLeft: 'auto', whiteSpace: 'nowrap' }}>{fmtTime(trend.createTime)}</span>
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
              <div style={{ width: 36, height: 36, borderRadius: 8, background: 'rgba(51,144,236,0.1)', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                <ExternalLink className="w-4 h-4" style={{ color: '#3390EC' }} />
              </div>
              <div className="flex-1 min-w-0">
                <div style={{ fontSize: '12px', fontWeight: 600, color: '#1C2733', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>分享链接</div>
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
                  background: liked ? 'rgba(51,144,236,0.1)' : 'transparent',
                  cursor: 'pointer',
                  transform: likeAnim ? 'scale(1.2)' : 'scale(1)',
                  transition: 'transform 0.2s cubic-bezier(0.175, 0.885, 0.32, 1.275)',
                }}
              >
                <Heart className="w-3.5 h-3.5" style={{ color: liked ? '#3390EC' : '#A2ACB5', fill: liked ? '#3390EC' : 'none', transition: 'all 0.15s' }} />
                {likeCount > 0 && <span style={{ fontSize: '11px', color: liked ? '#3390EC' : '#A2ACB5', fontWeight: liked ? 600 : 400 }}>{likeCount}</span>}
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
                onClick={() => toast.success('链接已复制')}
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
                <Heart className="w-3 h-3 shrink-0 mt-0.5" style={{ color: '#3390EC', fill: '#3390EC' }} />
                <span style={{ color: '#708499' }}>
                  {(likeNamesExpanded ? likeUsers : likeUsers.slice(0, FEED_LIKE_COLLAPSE_LIMIT)).map((u, i) => {
                    const displayName = u.id === 'me' ? '我' : (useIMStore.getState().friends.find(ct => ct.id === u.id)?.remark || u.name);
                    return (
                      <span key={u.id || `like-${i}`}>
                        {i > 0 && <span style={{ margin: '0 2px' }}>、</span>}
                        <span
                          style={{ color: '#3390EC', cursor: 'pointer' }}
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
                        color: '#3390EC', background: 'none', border: 'none',
                        cursor: 'pointer', fontSize: '12px', padding: 0, fontWeight: 500,
                      }}
                    >
                      {likeNamesExpanded ? '收起' : `等${likeUsers.length}人`}
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
                  style={{ fontSize: '12px', color: '#3390EC', background: 'none', border: 'none', cursor: 'pointer', padding: '4px 0 0', fontWeight: 500 }}
                >
                  展开全部{trend.replyCount}条评论
                </button>
              )}
            </div>
          )}

          {/* Inline comment input */}
          {expanded && trend.openReply && (
            <div className="flex items-center gap-2 mt-2">
              {replyTarget && (
                <div className="flex items-center gap-1 text-xs w-full" style={{ color: '#708499', marginBottom: 4, padding: '0 4px' }}>
                  <span>回复</span>
                  <span style={{ color: '#3390EC', fontWeight: 500 }}>{replyTarget.replyer.name}</span>
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
                  placeholder={replyTarget ? `回复 ${replyTarget.replyer.name}...` : '写评论...'}
                  style={{ ...inputStyle, flex: 1, padding: '7px 12px', fontSize: '13px', borderRadius: '18px' }}
                  onFocus={focusInput}
                  onBlur={blurInput}
                />
                {commentText.trim() && (
                  <button
                    onClick={onSubmitComment}
                    style={{ width: 32, height: 32, borderRadius: '50%', border: 'none', background: '#3390EC', color: '#FFF', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}
                  >
                    <Send className="w-3.5 h-3.5" />
                  </button>
                )}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

/* ═══════════════════════════════════════
   Publish Modal
   ═══════════════════════════════════════ */

interface PublishModalProps {
  open: boolean;
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

function PublishModal({ open, onClose, onSubmit }: PublishModalProps) {
  const [type, setType] = useState<number>(1);
  const [content, setContent] = useState('');
  const [title, setTitle] = useState('');
  const [imageInput, setImageInput] = useState('');
  const [images, setImages] = useState<string[]>([]);
  const [shareUrl, setShareUrl] = useState('');
  const [location, setLocation] = useState('');
  const [scope, setScope] = useState<number>(3);

  if (!open) return null;

  const handleAddImage = () => {
    if (imageInput.trim()) {
      setImages(prev => [...prev, imageInput.trim()]);
      setImageInput('');
    }
  };

  const handleRemoveImage = (idx: number) => {
    setImages(prev => prev.filter((_, i) => i !== idx));
  };

  const handleSubmit = () => {
    if (type !== 4 && !content.trim()) {
      toast.error('请输入内容');
      return;
    }
    if (type === 3 && !title.trim()) {
      toast.error('请输入文章标题');
      return;
    }
    if (type === 4 && !shareUrl.trim()) {
      toast.error('请输入分享链接');
      return;
    }
    onSubmit({ type, content: content.trim(), title: title.trim(), images, shareUrl: shareUrl.trim(), location: location.trim(), scope });
    // Reset
    setContent('');
    setTitle('');
    setImageInput('');
    setImages([]);
    setShareUrl('');
    setLocation('');
    setScope(3);
    setType(1);
  };

  return (
    <div className="fixed inset-0" style={{ zIndex: 10001, display: 'flex', alignItems: 'center', justifyContent: 'center' }} onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}>
      <div className="absolute inset-0" style={{ background: 'rgba(0,0,0,0.5)' }} />
      <div className="relative" style={{ background: '#FFF', borderRadius: '16px', width: '92%', maxWidth: '480px', maxHeight: '85vh', display: 'flex', flexDirection: 'column', boxShadow: '0 8px 40px rgba(0,0,0,0.15)' }}>
        {/* Header */}
        <div className="flex items-center justify-between shrink-0" style={{ padding: '16px 20px', borderBottom: '1px solid rgba(0,0,0,0.06)' }}>
          <button onClick={onClose} style={{ width: 32, height: 32, borderRadius: '50%', border: 'none', background: 'transparent', color: '#708499', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <X className="w-5 h-5" />
          </button>
          <span style={{ fontSize: '16px', fontWeight: 600, color: '#1C2733' }}>发布动态</span>
          <button
            onClick={handleSubmit}
            style={{ padding: '6px 16px', borderRadius: '16px', border: 'none', background: '#3390EC', color: '#FFF', fontSize: '13px', fontWeight: 500, cursor: 'pointer' }}
          >
            发布
          </button>
        </div>

        {/* Body */}
        <div className="flex-1 overflow-y-auto" style={{ padding: '16px 20px' }}>
          {/* Type selector */}
          <div className="flex items-center gap-2 mb-4 flex-wrap">
            {[1, 2, 3, 4, 5].map(t => (
              <button
                key={t}
                onClick={() => setType(t)}
                style={{
                  display: 'flex', alignItems: 'center', gap: 4, padding: '5px 12px', borderRadius: '16px',
                  border: 'none', background: type === t ? 'rgba(51,144,236,0.1)' : 'rgba(0,0,0,0.04)',
                  color: type === t ? '#3390EC' : '#708499', fontSize: '12px', fontWeight: type === t ? 500 : 400, cursor: 'pointer',
                }}
              >
                {trendTypeLabels[t].icon}
                {trendTypeLabels[t].label}
              </button>
            ))}
          </div>

          {/* Title (article) */}
          {type === 3 && (
            <div style={{ marginBottom: 12 }}>
              <input value={title} onChange={(e) => setTitle(e.target.value)} placeholder="文章标题" style={{ ...inputStyle, fontWeight: 600, fontSize: '15px' }} onFocus={focusInput} onBlur={blurInput} />
            </div>
          )}

          {/* Content */}
          {type !== 4 && (
            <div style={{ marginBottom: 12 }}>
              <textarea
                value={content}
                onChange={(e) => setContent(e.target.value)}
                placeholder={type === 1 ? '分享你的想法...' : type === 3 ? '文章内容...' : '添加描述...'}
                rows={4}
                style={{ ...inputStyle, resize: 'vertical', minHeight: 80, lineHeight: '1.6' }}
                onFocus={focusInput}
                onBlur={blurInput}
              />
            </div>
          )}

          {/* Image URLs (type 2 or 5) */}
          {(type === 2 || type === 5) && (
            <div style={{ marginBottom: 12 }}>
              <div className="flex items-center gap-2 mb-2">
                <input value={imageInput} onChange={(e) => setImageInput(e.target.value)} placeholder="输入图片URL" style={{ ...inputStyle, flex: 1 }} onKeyDown={(e) => { if (e.key === 'Enter') handleAddImage(); }} onFocus={focusInput} onBlur={blurInput} />
                <button onClick={handleAddImage} style={{ padding: '8px 14px', borderRadius: '10px', border: 'none', background: '#3390EC', color: '#FFF', fontSize: '13px', cursor: 'pointer', whiteSpace: 'nowrap' }}>添加</button>
              </div>
              {images.length > 0 && (
                <div className="flex items-center gap-2 flex-wrap">
                  {images.map((img, idx) => (
                    <div key={idx} className="relative" style={{ width: 60, height: 60, borderRadius: 8, overflow: 'hidden' }}>
                      <img src={img} alt="" className="w-full h-full object-cover" />
                      <button onClick={() => handleRemoveImage(idx)} className="absolute" style={{ top: 2, right: 2, width: 18, height: 18, borderRadius: '50%', background: 'rgba(0,0,0,0.5)', border: 'none', color: '#FFF', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 10 }}>
                        <X className="w-3 h-3" />
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* Share link (type 4) */}
          {type === 4 && (
            <div style={{ marginBottom: 12 }}>
              <input value={shareUrl} onChange={(e) => setShareUrl(e.target.value)} placeholder="分享链接" style={inputStyle} onFocus={focusInput} onBlur={blurInput} />
              <textarea value={content} onChange={(e) => setContent(e.target.value)} placeholder="推荐理由（可选）" rows={3} style={{ ...inputStyle, resize: 'vertical', marginTop: 8, minHeight: 60 }} onFocus={focusInput} onBlur={blurInput} />
            </div>
          )}

          {/* Location */}
          <div style={{ marginBottom: 12 }}>
            <div className="flex items-center gap-2">
              <MapPin className="w-4 h-4" style={{ color: '#A2ACB5', flexShrink: 0 }} />
              <input value={location} onChange={(e) => setLocation(e.target.value)} placeholder="添加位置（可选）" style={{ ...inputStyle, flex: 1 }} onFocus={focusInput} onBlur={blurInput} />
            </div>
          </div>

          {/* Scope */}
          <div>
            <div style={{ fontSize: '13px', fontWeight: 500, color: '#1C2733', marginBottom: '8px' }}>可见范围</div>
            <div className="flex items-center gap-2">
              {[1, 2, 3].map(s => (
                <button
                  key={s}
                  onClick={() => setScope(s)}
                  style={{
                    display: 'flex', alignItems: 'center', gap: 4, padding: '6px 14px', borderRadius: '16px',
                    border: scope === s ? '1.5px solid #3390EC' : '1.5px solid rgba(0,0,0,0.1)',
                    background: scope === s ? 'rgba(51,144,236,0.06)' : '#FFF',
                    color: scope === s ? '#3390EC' : '#708499', fontSize: '12px', fontWeight: scope === s ? 500 : 400, cursor: 'pointer',
                  }}
                >
                  {scopeLabels[s].icon}{scopeLabels[s].label}
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
  if (!open) return null;
  return (
    <div className="fixed inset-0" style={{ zIndex: 10001, display: 'flex', alignItems: 'center', justifyContent: 'center' }} onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}>
      <div className="absolute inset-0" style={{ background: 'rgba(0,0,0,0.5)' }} />
      <div className="relative" style={{ background: '#FFF', borderRadius: '16px', width: '90%', maxWidth: '400px', maxHeight: '60vh', display: 'flex', flexDirection: 'column', boxShadow: '0 8px 40px rgba(0,0,0,0.15)' }}>
        <div className="flex items-center justify-between shrink-0" style={{ padding: '16px 20px', borderBottom: '1px solid rgba(0,0,0,0.06)' }}>
          <span style={{ fontSize: '16px', fontWeight: 600, color: '#1C2733' }}>点赞列表</span>
          <button onClick={onClose} style={{ width: 28, height: 28, borderRadius: '50%', border: 'none', background: 'transparent', color: '#A2ACB5', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <X className="w-4 h-4" />
          </button>
        </div>
        <div className="flex-1 overflow-y-auto" style={{ padding: '8px 12px' }}>
          {users.length === 0 ? (
            <div className="flex flex-col items-center justify-center" style={{ padding: '40px 0' }}>
              <Heart className="w-12 h-12" style={{ color: '#E8EDEF', marginBottom: 12 }} />
              <div style={{ fontSize: '13px', color: '#A2ACB5' }}>暂无点赞</div>
            </div>
          ) : (
            users.map(user => (
              <div key={user.id} className="flex items-center gap-3" style={{ padding: '8px 8px', borderRadius: 8, cursor: 'pointer' }} onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'rgba(0,0,0,0.02)'; }} onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}>
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
  if (!open || !trend) return null;
  const userName = trendDisplayName(trend);
  const userAvatar = trend.userAvatar || '';

  const renderContent = () => {
    const parts = trend.content.split(/(@\S+)/g);
    return parts.map((part, i) => {
      if (part.startsWith('@')) {
        const atUser = trend.atUsers.find(u => `@${u.name}` === part);
        return <span key={i} style={{ color: '#3390EC' }}>{atUser ? atUser.name : part}</span>;
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
          <span style={{ fontSize: '16px', fontWeight: 600, color: '#1C2733' }}>动态详情</span>
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
                <span style={{ fontSize: '15px', fontWeight: 600, color: '#3390EC', cursor: 'pointer' }} onClick={() => onAvatarClick(trend.userId)}>{userName}</span>
                {trend.isTop && (
                  <span style={{ fontSize: '10px', fontWeight: 500, color: '#F5A623', backgroundColor: 'rgba(245,166,35,0.1)', borderRadius: '4px', padding: '1px 6px' }}>置顶</span>
                )}
              </div>
              <span style={{ fontSize: '12px', color: '#A2ACB5' }}>{fmtTime(trend.createTime)}</span>
            </div>
          </div>

          {/* Badges */}
          <div className="flex items-center gap-2 mb-3 flex-wrap">
            {trend.scope !== 3 && (
              <span style={{ fontSize: '11px', color: '#708499', display: 'flex', alignItems: 'center', gap: 3 }}>
                {scopeLabels[trend.scope].icon}{scopeLabels[trend.scope].label}
              </span>
            )}
            {!trend.openReply && (
              <span style={{ fontSize: '11px', color: '#A2ACB5', display: 'flex', alignItems: 'center', gap: 3 }}>
                <MessageSquareOff className="w-3 h-3" />评论已关闭
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
                <div key={idx} className="overflow-hidden" style={{ borderRadius: 8, background: '#E8EDEF' }}>
                  <img src={img} alt="" className="w-full object-cover" style={{ display: 'block', aspectRatio: trend.resources.length === 1 ? '16/9' : '1' }} />
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
              <div style={{ width: 36, height: 36, borderRadius: 8, background: 'rgba(51,144,236,0.1)', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                <ExternalLink className="w-4 h-4" style={{ color: '#3390EC' }} />
              </div>
              <div className="flex-1 min-w-0">
                <div style={{ fontSize: '12px', fontWeight: 600, color: '#1C2733', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>分享链接</div>
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
            <button onClick={onToggleLike} className="flex items-center gap-1.5" style={{ padding: '6px 14px', borderRadius: '18px', border: 'none', background: liked ? 'rgba(51,144,236,0.1)' : 'transparent', cursor: 'pointer' }}>
              <Heart className="w-4 h-4" style={{ color: liked ? '#3390EC' : '#A2ACB5', fill: liked ? '#3390EC' : 'none' }} />
              <span style={{ fontSize: '13px', color: liked ? '#3390EC' : '#708499', fontWeight: liked ? 600 : 400 }}>{likeCount > 0 ? likeCount : '赞'}</span>
            </button>
            <button onClick={onLikeCountClick} style={{ padding: '6px 14px', borderRadius: '18px', border: 'none', background: 'transparent', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 4 }}>
              <ThumbsUp className="w-4 h-4" style={{ color: '#A2ACB5' }} />
              <span style={{ fontSize: '13px', color: '#708499' }}>赞过的人</span>
            </button>
          </div>

          {/* Comments */}
          <div>
            <div style={{ fontSize: '14px', fontWeight: 600, color: '#1C2733', marginBottom: 8 }}>
              评论 ({trend.replyCount})
            </div>
            {comments.length === 0 ? (
              <div className="flex flex-col items-center justify-center" style={{ padding: '32px 0' }}>
                <MessageCircle className="w-10 h-10" style={{ color: '#E8EDEF', marginBottom: 8 }} />
                <div style={{ fontSize: '13px', color: '#A2ACB5' }}>暂无评论</div>
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
                <span>回复</span>
                <span style={{ color: '#3390EC', fontWeight: 500 }}>{replyTarget.replyer.name}</span>
              </div>
            )}
            <input
              value={commentText}
              onChange={(e) => onCommentTextChange(e.target.value)}
              onKeyDown={(e) => { if (e.key === 'Enter' && commentText.trim()) onSubmitComment(); }}
              placeholder={replyTarget ? `回复 ${replyTarget.replyer.name}...` : '写评论...'}
              style={{ ...inputStyle, flex: 1, padding: '8px 14px', fontSize: '13px', borderRadius: '20px' }}
              onFocus={focusInput}
              onBlur={blurInput}
            />
            {commentText.trim() && (
              <button onClick={onSubmitComment} style={{ width: 34, height: 34, borderRadius: '50%', border: 'none', background: '#3390EC', color: '#FFF', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                <Send className="w-4 h-4" />
              </button>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

/* ═══════════════════════════════════════
   Main Component
   ═══════════════════════════════════════ */

type View = 'feed' | 'notifications' | 'userTrends';
type NotifTab = 'all' | 'reply' | 'like';

export default function MomentsFeed() {
  const isMobile = useIsMobile();
  const { selectedTrendId, setSelectedTrendId, currentUser: meAuth, trendVersions, bumpTrendVersion, openUserProfile, updateCurrentUser } = useIMStore();
  const meName = meAuth?.name || '';
  const meAvatar = meAuth?.avatar || '';
  const meSignature = meAuth?.introduction || '';
  const meCover = meAuth?.momentsCover || '';
  const meUserId = meAuth?.id || '';
  const token = meAuth?.token || '';
  const coverInputRef = useRef<HTMLInputElement>(null);

  const handleCoverUpload = useCallback(async (file: File) => {
    if (!token) return;
    try {
      const fd = new FormData();
      fd.append('file', file);
      const up = await fetch('/api/user/avatar', { method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: fd });
      const upData = await up.json();
      if (!upData.success || !upData.data?.url) { toast.error('封面上传失败'); return; }
      const url = upData.data.url as string;
      const save = await fetch('/api/user/update', { method: 'PUT', headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' }, body: JSON.stringify({ moments_cover: url }) });
      const saveData = await save.json();
      if (saveData.success) { updateCurrentUser({ momentsCover: url }); toast.success('朋友圈封面已更新'); }
      else toast.error(saveData.message || '保存封面失败');
    } catch { toast.error('封面上传失败'); }
  }, [token, updateCurrentUser]);

  // ── View state ──
  const [view, setView] = useState<View>('feed');
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
      toast.error('加载动态失败');
    } finally {
      setLoading(false);
    }
  }, [token, meUserId]);

  // ── Notifications loader ──
  const loadNotifications = useCallback(async () => {
    if (!token) return;
    try {
      const resp = await getUnread(token);
      if (resp.success && resp.data) {
        setNotifications(unreadToNotifications(resp.data, meUserId));
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
      if (a.isTop !== b.isTop) return a.isTop ? -1 : 1;
      return b.createTime.getTime() - a.createTime.getTime();
    });
    if (view === 'userTrends') {
      list = list.filter(t => t.userId === userTrendsUserId);
    }
    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase();
      list = list.filter(t => t.content.toLowerCase().includes(q) || trendDisplayName(t).toLowerCase().includes(q));
    }
    return list;
  }, [trends, view, userTrendsUserId, searchQuery]);

  const filteredNotifications = useMemo(() => {
    let list = [...notifications].sort((a, b) => b.createTime.getTime() - a.createTime.getTime());
    if (notifTab === 'reply') list = list.filter(n => n.type === 'reply');
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
      toast.error('请先登录');
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
      toast.error('点赞操作失败');
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
        toast.error(r.message || '评论失败');
        return;
      }
      await refreshCommentsFor(trendId);
      setTrends(ts => ts.map(t => t.id === trendId ? { ...t, replyCount: t.replyCount + 1 } : t));
      setExpandedTrendIds(prev => new Set(prev).add(trendId));
      seenVersionsRef.current[trendId] = (trendVersions[trendId] || 0) + 1;
      bumpTrendVersion(trendId);
      toast.success('评论已发布');
    }).catch(err => {
      console.error('createComment error', err);
      toast.error('评论失败');
    });
  }, [token, trends, meUserId, refreshCommentsFor, trendVersions, bumpTrendVersion]);

  const handleDeleteComment = useCallback((trendId: number, comment: TrendComment) => {
    if (!token) return;
    doConfirm('删除评论', '确定删除这条评论吗？', '删除', '#E53935', () => {
      apiDeleteComment(token, comment.id).then(async r => {
        if (!r.success) {
          toast.error(r.message || '删除失败');
          return;
        }
        await refreshCommentsFor(trendId);
        setTrends(ts => ts.map(t => t.id === trendId ? { ...t, replyCount: Math.max(0, t.replyCount - 1) } : t));
        seenVersionsRef.current[trendId] = (trendVersions[trendId] || 0) + 1;
        bumpTrendVersion(trendId);
        toast.success('评论已删除');
      }).catch(() => toast.error('删除失败'));
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
        toast.error(r.message || '发布失败');
        return;
      }
      setShowPublish(false);
      toast.success('动态已发布');
      loadFeed();
    }).catch(err => {
      console.error('publish error', err);
      toast.error('发布失败');
    });
  }, [token, loadFeed]);

  // ── Notification handlers ──
  const handleMarkAllRead = useCallback(() => {
    if (!token) return;
    const unreadReplies = notifications.filter(n => !n.read && n.type === 'reply').map(n => n.id);
    const unreadLikes = notifications.filter(n => !n.read && n.type === 'like').map(n => n.id);
    const ops: Promise<unknown>[] = [];
    if (unreadReplies.length) ops.push(markCommentsRead(token, unreadReplies));
    if (unreadLikes.length) ops.push(markLikesRead(token, unreadLikes));
    Promise.all(ops).then(() => {
      setNotifications(prev => prev.map(n => ({ ...n, read: true })));
      toast.success('已全部标记为已读');
    }).catch(() => toast.error('标记失败'));
  }, [token, notifications]);

  const handleNotifClick = useCallback((notif: MomentsNotification) => {
    if (!notif.read && token) {
      if (notif.type === 'reply') markCommentsRead(token, [notif.id]).catch(() => {});
      else markLikesRead(token, [notif.id]).catch(() => {});
    }
    setNotifications(prev => prev.map(n => n.id === notif.id ? { ...n, read: true } : n));
    if (isMobile) {
      setDetailTrendId(notif.trendId);
      setView('feed');
    } else {
      setSelectedTrendId(notif.trendId);
      setView('feed');
    }
  }, [isMobile, setSelectedTrendId, token]);

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
          ? (currentVal ? '已取消置顶' : '已置顶')
          : (currentVal ? '评论已关闭' : '评论已开启'),
      );
    }).catch(err => {
      console.error('updateTrend error', err);
      setTrends(prev => prev.map(t => t.id === trendId ? { ...t, [field]: currentVal } : t));
      toast.error('操作失败');
    });
  }, [token, trends, trendVersions, bumpTrendVersion]);

  const handleDeleteTrend = useCallback((trendId: number) => {
    if (!token) return;
    apiDeleteTrend(token, trendId).then(r => {
      if (!r.success) {
        toast.error(r.message || '删除失败');
        return;
      }
      // Close the detail panel if it was showing the deleted trend.
      if (selectedTrendId === trendId) setSelectedTrendId(null);
      bumpTrendVersion(trendId);
      setTrends(prev => prev.filter(t => t.id !== trendId));
      setCommentsMap(prev => { const next = { ...prev }; delete next[trendId]; return next; });
      setLikeUsersMap(prev => { const next = { ...prev }; delete next[trendId]; return next; });
      toast.success('动态已删除');
    }).catch(() => toast.error('删除失败'));
  }, [token, selectedTrendId, setSelectedTrendId, bumpTrendVersion]);

  // ── Navigation ──
  const handleAvatarClick = useCallback((userId: string, _userName?: string) => {
    const uid = userId === 'me' ? meUserId : userId;
    if (!uid) return;
    openUserProfile(uid);
  }, [meUserId, openUserProfile]);

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
    setView('feed');
    setNotifTab('all');
  }, []);

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
          placeholder="搜索动态"
          style={{ width: '100%', borderRadius: '20px', background: '#F0F2F5', border: 'none', padding: '7px 14px 7px 34px', fontSize: '13px', color: '#1C2733', outline: 'none' }}
        />
      </div>
      <div className="flex items-center gap-1 shrink-0 ml-2">
        {/* Notification bell */}
        <div className="relative">
          <button
            onClick={() => { setView('notifications'); setNotifTab('all'); }}
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
          style={{ width: 36, height: 36, borderRadius: '50%', border: 'none', background: '#3390EC', color: '#FFF', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}
        >
          <Plus className="w-5 h-5" style={{ color: '#FFF' }} />
        </button>
      </div>
    </div>
  );

  const renderCoverSection = () => (
    <div className="relative">
      <div
        className="relative overflow-hidden"
        style={{ height: 176 }}
      >
        {meCover ? (
          <img
            src={meCover}
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
                background: 'linear-gradient(135deg, rgba(51,144,236,0.7), rgba(111,177,252,0.5))',
              }}
            />
          </>
        )}
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
          更换封面
        </button>
        <div
          className="absolute inset-0"
          style={{
            background: 'linear-gradient(to top, rgba(23,33,43,0.45) 0%, transparent 60%)',
          }}
        />
      </div>
      <div className="absolute" style={{ bottom: -16, left: 12, cursor: 'pointer' }} onClick={() => meUserId && openUserProfile(meUserId)}>
        {meAvatar ? (
          <img
            src={meAvatar}
            alt={meName}
            style={{ width: 68, height: 68, borderRadius: '50%', objectFit: 'cover', boxShadow: '0 0 0 3px rgba(51,144,236,0.25), 0 2px 8px rgba(0,0,0,0.1)' }}
          />
        ) : (
          <div
            style={{ width: 68, height: 68, borderRadius: '50%', backgroundColor: getAvatarColor(meName), display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 26, fontWeight: 600, color: '#FFF', boxShadow: '0 0 0 3px rgba(51,144,236,0.25), 0 2px 8px rgba(0,0,0,0.1)' }}
          >
            {meName ? meName[0] : '?'}
          </div>
        )}
      </div>
    </div>
  );

  const renderUserInfo = () => (
    <div style={{ padding: '20px 12px 10px' }}>
      <h3 style={{ fontSize: '16px', fontWeight: 600, color: '#1C2733', cursor: 'pointer', display: 'inline-block' }} onClick={() => meUserId && openUserProfile(meUserId)}>{meName}</h3>
      <p style={{ fontSize: '12px', color: '#708499', marginTop: 2 }}>{meSignature}</p>
    </div>
  );

  const pillTab = (active: boolean, onClick: () => void, label: string, badge?: number) => (
    <button onClick={onClick} style={{ padding: '6px 18px', borderRadius: '17px', border: 'none', background: active ? '#3390EC' : 'transparent', color: active ? '#FFF' : '#646A73', fontSize: '13px', fontWeight: 500, cursor: 'pointer', transition: 'all 0.2s', display: 'flex', alignItems: 'center', gap: 4 }}>
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
          {renderCoverSection()}
          {renderUserInfo()}
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
                  {searchQuery ? '没有找到相关动态' : '暂无动态'}
                </div>
                <div style={{ fontSize: '13px', color: '#A2ACB5' }}>
                  {searchQuery ? '换个关键词试试' : '点击右上角 + 发布第一条动态'}
                </div>
              </div>
            )}
          </div>

          <div style={{ padding: '20px 0 12px', textAlign: 'center', fontSize: '12px', color: '#A2ACB5' }}>
            — 已经到底了 —
          </div>
        </div>

        {/* FAB button */}
        <button
          onClick={() => setShowPublish(true)}
          className="absolute"
          style={{ bottom: 24, right: 20, width: 52, height: 52, borderRadius: '50%', background: '#3390EC', color: '#FFF', border: 'none', boxShadow: '0 4px 16px rgba(51,144,236,0.4)', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', transition: 'transform 0.2s' }}
          onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.transform = 'scale(1.05)'; }}
          onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.transform = 'scale(1)'; }}
        >
          <Plus className="w-6 h-6" style={{ color: '#FFF' }} />
        </button>

        {/* Modals */}
        <PublishModal open={showPublish} onClose={() => setShowPublish(false)} onSubmit={handlePublish} />
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
                label: (() => { const t = trends.find(tr => tr.id === actionMenu.trendId); return t?.isTop ? '取消置顶' : '置顶'; })(),
                icon: (() => { const t = trends.find(tr => tr.id === actionMenu.trendId); return t?.isTop ? <PinOff className="w-3.5 h-3.5" /> : <Pin className="w-3.5 h-3.5" />; })(),
                onClick: () => handleToggleTrendField(actionMenu.trendId, 'isTop'),
              },
              {
                label: (() => { const t = trends.find(tr => tr.id === actionMenu.trendId); return t?.openReply ? '关闭评论' : '开启评论'; })(),
                icon: (() => { const t = trends.find(tr => tr.id === actionMenu.trendId); return t?.openReply ? <MessageSquareOff className="w-3.5 h-3.5" /> : <MessageCircle className="w-3.5 h-3.5" />; })(),
                onClick: () => handleToggleTrendField(actionMenu.trendId, 'openReply'),
              },
              {
                label: '删除动态',
                color: '#FF5252',
                icon: <Trash2 className="w-3.5 h-3.5" />,
                onClick: () => {
                  const tid = actionMenu.trendId;
                  doConfirm('删除动态', '确定删除这条动态吗？删除后不可恢复。', '删除', '#E53935', () => handleDeleteTrend(tid));
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
    const unreadReplyCount = notifications.filter(n => !n.read && n.type === 'reply').length;
    const unreadLikeCount = notifications.filter(n => !n.read && n.type === 'like').length;

    return (
      <div className="h-full flex flex-col" style={{ background: '#F5F7FA' }}>
        {/* Header */}
        <div className="flex items-center justify-between shrink-0" style={{ height: 56, background: '#FFF', borderBottom: '1px solid rgba(0,0,0,0.08)', paddingLeft: 4, paddingRight: 16 }}>
          <button onClick={handleBackFromNotifications} style={{ width: 40, height: 40, borderRadius: '50%', border: 'none', background: 'transparent', color: '#3390EC', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <ArrowLeft className="w-5 h-5" />
          </button>
          <span style={{ fontSize: '17px', fontWeight: 600, color: '#1C2733' }}>通知</span>
          <button onClick={handleMarkAllRead} style={{ fontSize: '13px', color: '#3390EC', background: 'none', border: 'none', cursor: 'pointer', fontWeight: 500 }}>全部已读</button>
        </div>

        {/* Tabs */}
        <div className="flex items-center shrink-0" style={{ padding: '12px 16px 8px', background: '#FFF', borderBottom: '1px solid rgba(0,0,0,0.05)' }}>
          <div className="flex items-center" style={{ borderRadius: '20px', background: 'rgba(0,0,0,0.04)', padding: '3px' }}>
            {pillTab(notifTab === 'all', () => setNotifTab('all'), '全部', unreadNotifCount)}
            {pillTab(notifTab === 'reply', () => setNotifTab('reply'), '评论', unreadReplyCount)}
            {pillTab(notifTab === 'like', () => setNotifTab('like'), '点赞', unreadLikeCount)}
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
                    {avatarCircle(notif.actor.name, 40)}
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2" style={{ marginBottom: 3 }}>
                        <span style={{ fontSize: '14px', fontWeight: 600, color: '#1C2733' }}>{notif.actor.name}</span>
                        {notif.type === 'reply' ? (
                          <span style={{ fontSize: '12px', color: '#708499' }}>评论了你的动态</span>
                        ) : (
                          <span style={{ fontSize: '12px', color: '#708499' }}>赞了你的动态</span>
                        )}
                        {!notif.read && (
                          <span style={{ width: 8, height: 8, borderRadius: '50%', background: '#E53935', flexShrink: 0, marginLeft: 'auto' }} />
                        )}
                      </div>
                      <div style={{ fontSize: '13px', color: '#646A73', lineHeight: '1.5', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                        {notif.type === 'reply' && notif.content ? (
                          <span>{notif.content}</span>
                        ) : (
                          <span>「{notif.trendContent}」</span>
                        )}
                      </div>
                      <div style={{ fontSize: '11px', color: '#A2ACB5', marginTop: 4 }}>{fmtTime(notif.createTime)}</div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center" style={{ padding: '60px 24px' }}>
              <Bell className="w-16 h-16" style={{ color: '#D1D5DB', marginBottom: 16 }} />
              <div style={{ fontSize: '15px', fontWeight: 500, color: '#646A73', marginBottom: 8 }}>
                {notifTab === 'all' ? '暂无通知' : notifTab === 'reply' ? '暂无评论通知' : '暂无点赞通知'}
              </div>
              <div style={{ fontSize: '13px', color: '#A2ACB5' }}>这里空空如也</div>
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
    return (
      <div className="h-full flex flex-col" style={{ background: '#F5F7FA' }}>
        {/* Header */}
        <div className="flex items-center justify-between shrink-0" style={{ height: 56, background: '#FFF', borderBottom: '1px solid rgba(0,0,0,0.08)', paddingLeft: 4, paddingRight: 16 }}>
          <button onClick={handleBackFromUserTrends} style={{ width: 40, height: 40, borderRadius: '50%', border: 'none', background: 'transparent', color: '#3390EC', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <ArrowLeft className="w-5 h-5" />
          </button>
          <span style={{ fontSize: '17px', fontWeight: 600, color: '#1C2733' }}>{userTrendsUserName}的动态</span>
          <div style={{ width: 40 }} />
        </div>

        {/* User info bar */}
        <div className="shrink-0" style={{ background: '#FFF', padding: '12px 16px', borderBottom: '1px solid rgba(0,0,0,0.06)' }}>
          <div className="flex items-center gap-3">
            {avatarCircle(userTrendsUserName, 44)}
            <div>
              <div style={{ fontSize: '15px', fontWeight: 600, color: '#1C2733' }}>{userTrendsUserName}</div>
              <div style={{ fontSize: '12px', color: '#A2ACB5', marginTop: 1 }}>{filteredTrends.length} 条动态</div>
            </div>
          </div>
        </div>

        {/* Trend list */}
        <div className="flex-1 overflow-y-auto im-scroll">
          {filteredTrends.length > 0 ? (
            <div style={{ background: '#FFF', borderRadius: '12px', margin: '8px', overflow: 'hidden' }}>
              {filteredTrends.map(trend => (
                <TrendCard
                  key={trend.id}
                  trend={trend}
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
                  onAvatarClick={() => {}}
                  onManage={(x, y) => handleManageTrend(trend.id, x, y)}
                />
              ))}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center" style={{ padding: '60px 24px' }}>
              <Inbox className="w-16 h-16" style={{ color: '#D1D5DB', marginBottom: 16 }} />
              <div style={{ fontSize: '15px', fontWeight: 500, color: '#646A73', marginBottom: 8 }}>暂无动态</div>
              <div style={{ fontSize: '13px', color: '#A2ACB5' }}>该用户还没有发布动态</div>
            </div>
          )}
        </div>

        {/* Modals */}
        <PublishModal open={showPublish} onClose={() => setShowPublish(false)} onSubmit={handlePublish} />
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
                label: (() => { const t = trends.find(tr => tr.id === actionMenu.trendId); return t?.isTop ? '取消置顶' : '置顶'; })(),
                icon: (() => { const t = trends.find(tr => tr.id === actionMenu.trendId); return t?.isTop ? <PinOff className="w-3.5 h-3.5" /> : <Pin className="w-3.5 h-3.5" />; })(),
                onClick: () => handleToggleTrendField(actionMenu.trendId, 'isTop'),
              },
              {
                label: '删除动态',
                color: '#FF5252',
                icon: <Trash2 className="w-3.5 h-3.5" />,
                onClick: () => {
                  const tid = actionMenu.trendId;
                  doConfirm('删除动态', '确定删除这条动态吗？', '删除', '#E53935', () => handleDeleteTrend(tid));
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
