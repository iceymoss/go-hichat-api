'use client';

import React, { useState, useCallback, useEffect } from 'react';
import { ChevronDown, ChevronUp } from 'lucide-react';
import {
  Heart, MessageCircle, MapPin, Send, Pin, MessageSquareOff,
  Play, ArrowLeft, ExternalLink, ThumbsUp, Users, Lock, Globe,
  X,
} from 'lucide-react';
import { toast } from 'sonner';
import { getAvatarColor } from '@/lib/utils';
import { useIMStore } from '@/lib/im-store';
import {
  type Trend,
  type TrendComment,
} from '@/lib/mock-data';
import {
  getTrendDetail,
  getCommentTree,
  getLikedUsers,
  toggleLike as apiToggleLike,
  createComment,
  deleteComment as apiDeleteComment,
  mapBackendTrend,
  commentTreeToMap,
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

const scopeLabels: Record<number, { label: string; icon: React.ReactNode }> = {
  1: { label: '仅自己', icon: <Lock className="w-3 h-3" /> },
  2: { label: '仅好友', icon: <Users className="w-3 h-3" /> },
  3: { label: '所有人', icon: <Globe className="w-3 h-3" /> },
};

/* ═══════════════════════════════════════
   CommentItem (recursive)
   ═══════════════════════════════════════ */

interface CommentItemProps {
  comment: TrendComment;
  onReply: (comment: TrendComment) => void;
  onDelete: (comment: TrendComment) => void;
}

function CommentItem({ comment, onReply, onDelete }: CommentItemProps) {
  const [showActions, setShowActions] = useState(false);
  const isOwn = comment.replyer.id === 'me';
  const userName = comment.replyer.name;
  const userAvatar = comment.replyer.avatar;
  const replyToName = comment.user?.name;

  return (
    <div
      className="relative"
      style={{ padding: '8px 0', borderBottom: '1px solid rgba(0,0,0,0.04)' }}
      onMouseEnter={() => setShowActions(true)}
      onMouseLeave={() => setShowActions(false)}
    >
      <div className="flex gap-2.5">
        <div
          className="overflow-hidden"
          style={{
            width: 32, height: 32, borderRadius: '50%', flexShrink: 0,
            backgroundColor: getAvatarColor(userName),
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            fontSize: 13, fontWeight: 600, color: '#FFF',
          }}
        >
          {userAvatar ? (
            <img src={userAvatar} alt="" className="w-full h-full object-cover" />
          ) : (
            userName[0]
          )}
        </div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <span style={{ fontSize: '13px', fontWeight: 600, color: '#3390EC' }}>{userName}</span>
            {replyToName && (
              <span style={{ fontSize: '12px', color: '#A2ACB5' }}>
                回复 <span style={{ color: '#3390EC' }}>{replyToName}</span>
              </span>
            )}
            <span style={{ fontSize: '11px', color: '#A2ACB5', marginLeft: 'auto', flexShrink: 0 }}>{fmtTime(comment.createTime)}</span>
          </div>
          <p style={{ fontSize: '13px', color: '#1C2733', marginTop: 3, lineHeight: '1.5' }}>{comment.content}</p>
          {/* Actions */}
          {showActions && (
            <div className="flex items-center gap-3" style={{ marginTop: 4 }}>
              <button
                onClick={() => onReply(comment)}
                style={{ fontSize: '12px', color: '#3390EC', background: 'none', border: 'none', cursor: 'pointer' }}
              >
                回复
              </button>
              {isOwn && (
                <button
                  onClick={() => onDelete(comment)}
                  style={{ fontSize: '12px', color: '#E53935', background: 'none', border: 'none', cursor: 'pointer' }}
                >
                  删除
                </button>
              )}
            </div>
          )}
        </div>
      </div>
      {/* Children */}
      {comment.children && comment.children.length > 0 && (
        <div style={{ marginLeft: 44, marginTop: 2 }}>
          {comment.children.map(child => (
            <CommentItem key={child.id} comment={child} onReply={onReply} onDelete={onDelete} />
          ))}
        </div>
      )}
    </div>
  );
}

/* ═══════════════════════════════════════
   Like Avatar Grid (detail view)
   ═══════════════════════════════════════ */

const LIKE_COLLAPSE_LIMIT = 70;

function LikeAvatarItem({ user }: { user: { id: string; name: string; avatar: string } }) {
  const name = user.name;
  const [showTip, setShowTip] = useState(false);
  const displayName = user.id === 'me' ? '我' : (name || user.id);
  return (
    <div
      className="relative overflow-hidden"
      style={{
        width: 32, height: 32, borderRadius: '50%', flexShrink: 0,
        backgroundColor: getAvatarColor(name),
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        fontSize: 12, fontWeight: 600, color: '#FFF', cursor: 'pointer',
        border: '2px solid #FFF', boxShadow: '0 0 0 1px rgba(0,0,0,0.06)',
        transition: 'transform 0.15s',
      }}
      onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.transform = 'scale(1.15)'; setShowTip(true); }}
      onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.transform = 'scale(1)'; setShowTip(false); }}
    >
      {user.avatar ? (
        <img src={user.avatar} alt="" className="w-full h-full object-cover" />
      ) : (
        name[0]
      )}
      {showTip && (
        <div
          className="absolute"
          style={{
            bottom: '100%', left: '50%', transform: 'translateX(-50%)',
            marginBottom: 6, padding: '3px 8px', borderRadius: 6,
            background: 'rgba(0,0,0,0.75)', color: '#FFF', fontSize: 11,
            whiteSpace: 'nowrap', pointerEvents: 'none', zIndex: 10,
          }}
        >
          {displayName}
        </div>
      )}
    </div>
  );
}

/* ═══════════════════════════════════════
   Main Panel Component
   ═══════════════════════════════════════ */

export default function TrendDetailPanel() {
  const { selectedTrendId, setSelectedTrendId, currentUser: meAuth, trendVersions, bumpTrendVersion } = useIMStore();
  const trendVersion = selectedTrendId != null ? (trendVersions[selectedTrendId] || 0) : 0;
  const token = meAuth?.token || '';
  const meUserId = meAuth?.id || '';
  const meName = meAuth?.name || '';
  const meAvatar = meAuth?.avatar || '';

  const [trend, setTrend] = useState<Trend | null>(null);
  const [comments, setComments] = useState<TrendComment[]>([]);
  const [likeUsers, setLikeUsers] = useState<{ id: string; name: string; avatar: string }[]>([]);
  const [liked, setLiked] = useState(false);
  const [commentText, setCommentText] = useState('');
  const [replyTarget, setReplyTarget] = useState<TrendComment | null>(null);
  const [likeAnim, setLikeAnim] = useState(false);
  const [likeExpanded, setLikeExpanded] = useState(false);

  const likeCount = trend?.agreeCount || 0;

  // Load trend + comments + like users when the selected id changes.
  useEffect(() => {
    if (!selectedTrendId || !token) {
      setTrend(null);
      setComments([]);
      setLikeUsers([]);
      setLiked(false);
      return;
    }
    let cancelled = false;
    (async () => {
      const [detailRes, treeRes, likedUsersRes] = await Promise.all([
        getTrendDetail(token, selectedTrendId),
        getCommentTree(token, [selectedTrendId]),
        getLikedUsers(token, selectedTrendId, 0, 100),
      ]);
      if (cancelled) return;
      if (detailRes.success && detailRes.data?.trend) {
        setTrend(mapBackendTrend(detailRes.data.trend, meUserId));
      } else {
        setTrend(null);
      }
      if (treeRes.success && treeRes.data) {
        const map = commentTreeToMap(treeRes.data, meUserId);
        setComments(map[selectedTrendId] || []);
      } else {
        setComments([]);
      }
      if (likedUsersRes.success && likedUsersRes.data) {
        const users = (likedUsersRes.data.users || []).map(u => ({
          id: u.id === meUserId ? 'me' : u.id,
          name: u.nickname || '',
          avatar: u.avatar || '',
        }));
        setLikeUsers(users);
        setLiked(users.some(u => u.id === 'me'));
      } else {
        setLikeUsers([]);
        setLiked(false);
      }
    })().catch(err => console.error('TrendDetailPanel load error', err));
    return () => { cancelled = true; };
  }, [selectedTrendId, token, meUserId, trendVersion]);

  // Handlers
  const handleClose = useCallback(() => {
    setSelectedTrendId(null);
  }, [setSelectedTrendId]);

  const refreshComments = useCallback(async () => {
    if (!selectedTrendId || !token) return;
    const r = await getCommentTree(token, [selectedTrendId]);
    if (r.success && r.data) {
      const map = commentTreeToMap(r.data, meUserId);
      setComments(map[selectedTrendId] || []);
    }
  }, [selectedTrendId, token, meUserId]);

  const handleToggleLike = useCallback(() => {
    if (!selectedTrendId || !trend || !token) return;
    setLikeAnim(true);
    setTimeout(() => setLikeAnim(false), 300);
    const wasLiked = liked;
    // Optimistic
    setLiked(!wasLiked);
    setLikeUsers(prev => wasLiked
      ? prev.filter(u => u.id !== 'me')
      : [{ id: 'me', name: meName, avatar: meAvatar }, ...prev]);
    setTrend(t => t ? { ...t, agreeCount: wasLiked ? Math.max(0, t.agreeCount - 1) : t.agreeCount + 1 } : t);

    const authorIdStr = trend.userId === 'me' ? meUserId : trend.userId;
    const authorId = Number(authorIdStr) || 0;
    // like_type: 1 点赞 / 0 取消 (后端按 LikeType > 0 判加减)
    apiToggleLike(token, selectedTrendId, authorId, wasLiked ? 0 : 1).then(() => {
      bumpTrendVersion(selectedTrendId);
    }).catch(err => {
      console.error('toggleLike error', err);
      setLiked(wasLiked);
      setLikeUsers(prev => wasLiked
        ? [{ id: 'me', name: meName, avatar: meAvatar }, ...prev]
        : prev.filter(u => u.id !== 'me'));
      setTrend(t => t ? { ...t, agreeCount: wasLiked ? t.agreeCount + 1 : Math.max(0, t.agreeCount - 1) } : t);
      toast.error('点赞操作失败');
    });
  }, [selectedTrendId, trend, token, liked, meUserId, meName, meAvatar, bumpTrendVersion]);

  const handleSubmitComment = useCallback(() => {
    if (!selectedTrendId || !commentText.trim() || !token || !trend) return;
    let replyUserIdStr = '';
    if (replyTarget) {
      replyUserIdStr = replyTarget.replyer.id === 'me' ? meUserId : replyTarget.replyer.id;
    } else {
      replyUserIdStr = trend.userId === 'me' ? meUserId : trend.userId;
    }
    const payload = {
      trend_id: selectedTrendId,
      father: replyTarget ? replyTarget.id : undefined,
      user_id: Number(replyUserIdStr) || 0,
      content: commentText.trim(),
    };
    setCommentText('');
    setReplyTarget(null);
    createComment(token, payload).then(async r => {
      if (!r.success) {
        toast.error(r.message || '评论失败');
        return;
      }
      await refreshComments();
      setTrend(t => t ? { ...t, replyCount: t.replyCount + 1 } : t);
      bumpTrendVersion(selectedTrendId);
      toast.success('评论已发布');
    }).catch(err => {
      console.error('createComment error', err);
      toast.error('评论失败');
    });
  }, [selectedTrendId, commentText, replyTarget, token, trend, meUserId, refreshComments, bumpTrendVersion]);

  const handleDeleteComment = useCallback((comment: TrendComment) => {
    if (!selectedTrendId || !token) return;
    apiDeleteComment(token, comment.id).then(async r => {
      if (!r.success) {
        toast.error(r.message || '删除失败');
        return;
      }
      await refreshComments();
      setTrend(t => t ? { ...t, replyCount: Math.max(0, t.replyCount - 1) } : t);
      bumpTrendVersion(selectedTrendId);
      toast.success('评论已删除');
    }).catch(() => toast.error('删除失败'));
  }, [selectedTrendId, token, refreshComments, bumpTrendVersion]);

  const imageGridClass = (count: number) => {
    if (count === 1) return 'grid-cols-1';
    if (count === 2 || count === 4) return 'grid-cols-2';
    return 'grid-cols-3';
  };

  const renderContent = () => {
    if (!trend) return null;
    const parts = trend.content.split(/(@\S+)/g);
    return parts.map((part, i) => {
      if (part.startsWith('@')) {
        const atUser = trend.atUsers.find(u => `@${u.name}` === part);
        return <span key={i} style={{ color: '#3390EC', cursor: 'pointer' }}>{atUser ? atUser.name : part}</span>;
      }
      return <span key={i}>{part}</span>;
    });
  };

  if (!trend) return null;
  const userName = trend.userName || getUserName(trend.userId);
  const userAvatar = trend.userAvatar || '';

  return (
    <div className="h-full flex flex-col" style={{ background: '#F0F2F5' }}>
      {/* Header */}
      <header
        className="flex items-center justify-between shrink-0"
        style={{
          height: 56,
          background: '#FFFFFF',
          borderBottom: '1px solid rgba(0,0,0,0.08)',
          paddingLeft: 4,
          paddingRight: 16,
        }}
      >
        <button
          onClick={handleClose}
          style={{
            width: 40, height: 40, borderRadius: '50%',
            border: 'none', background: 'transparent',
            color: '#3390EC', cursor: 'pointer',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
          }}
        >
          <ArrowLeft className="w-5 h-5" />
        </button>
        <span style={{ fontSize: '17px', fontWeight: 600, color: '#1C2733' }}>动态详情</span>
        <div style={{ width: 40 }} />
      </header>

      {/* Scrollable body */}
      <div className="flex-1 overflow-y-auto im-scroll" style={{ padding: '16px 20px' }}>
        <div style={{ maxWidth: 680, margin: '0 auto' }}>
          {/* ── Section 1: Trend Content ── */}
          <div style={{ background: '#FFF', borderRadius: 12, padding: '20px', marginBottom: 12, boxShadow: '0 1px 3px rgba(0,0,0,0.04)' }}>
            {/* User header */}
            <div className="flex items-center gap-3 mb-4">
              <div
                className="overflow-hidden"
                style={{
                  width: 44, height: 44, borderRadius: '50%', flexShrink: 0,
                  backgroundColor: getAvatarColor(userName),
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                  fontSize: 18, fontWeight: 600, color: '#FFF',
                }}
              >
                {userAvatar ? (
                  <img src={userAvatar} alt="" className="w-full h-full object-cover" />
                ) : (
                  userName[0]
                )}
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 flex-wrap">
                  <span style={{ fontSize: '15px', fontWeight: 600, color: '#3390EC' }}>{userName}</span>
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
                </div>
                <div className="flex items-center gap-2 mt-0.5">
                  <span style={{ fontSize: '12px', color: '#A2ACB5' }}>{fmtTime(trend.createTime)}</span>
                  {trend.scope !== 3 && (
                    <span style={{ fontSize: '11px', color: '#708499', display: 'flex', alignItems: 'center', gap: 3 }}>
                      {scopeLabels[trend.scope].icon}{scopeLabels[trend.scope].label}
                    </span>
                  )}
                </div>
              </div>
            </div>

            {/* Title (article type) */}
            {trend.type === 3 && trend.title && (
              <h3 style={{ fontSize: '17px', fontWeight: 600, color: '#1C2733', marginBottom: 8, lineHeight: '1.4' }}>{trend.title}</h3>
            )}

            {/* Text content */}
            <p style={{ fontSize: '14px', lineHeight: '1.7', color: '#1C2733', marginBottom: trend.positionName ? 8 : 0 }}>
              {renderContent()}
            </p>

            {/* Location */}
            {trend.positionName && (
              <div className="flex items-center gap-1.5" style={{ fontSize: '12px', color: '#A2ACB5', marginBottom: 12 }}>
                <MapPin className="w-3.5 h-3.5" />
                {trend.positionName}
              </div>
            )}

            {/* Media */}
            {trend.type === 2 && trend.resources.length > 0 && (
              <div className={`grid ${imageGridClass(trend.resources.length)} gap-2`} style={{ marginBottom: 12 }}>
                {trend.resources.map((img, idx) => (
                  <div key={idx} className="overflow-hidden" style={{ borderRadius: 8, background: '#E8EDEF' }}>
                    <img
                      src={img}
                      alt=""
                      className="w-full object-cover hover:scale-105 transition-transform duration-200"
                      style={{ display: 'block', aspectRatio: trend.resources.length === 1 ? '16/9' : '1' }}
                    />
                  </div>
                ))}
              </div>
            )}

            {trend.type === 3 && trend.coverUrl && (
              <div style={{ borderRadius: 10, overflow: 'hidden', marginBottom: 12 }}>
                <img src={trend.coverUrl} alt="" className="w-full object-cover" style={{ display: 'block' }} />
              </div>
            )}

            {trend.type === 4 && (
              <div
                className="flex items-center gap-3"
                style={{
                  padding: '10px 14px', borderRadius: 10, marginBottom: 12,
                  background: 'rgba(0,0,0,0.03)', border: '1px solid rgba(0,0,0,0.06)',
                  cursor: 'pointer',
                }}
                onClick={() => window.open(trend.shareUrl, '_blank')}
              >
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
              <div className="relative" style={{ borderRadius: 10, overflow: 'hidden', marginBottom: 12 }}>
                <img src={trend.resources[0]} alt="" className="w-full object-cover" style={{ display: 'block', aspectRatio: '16/9' }} />
                <div className="absolute inset-0 flex items-center justify-center" style={{ background: 'rgba(0,0,0,0.2)' }}>
                  <div style={{ width: 52, height: 52, borderRadius: '50%', background: 'rgba(255,255,255,0.9)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                    <Play className="w-6 h-6" style={{ color: '#1C2733', marginLeft: 2 }} />
                  </div>
                </div>
              </div>
            )}

            {/* Action bar */}
            <div className="flex items-center gap-3" style={{ paddingTop: 12, borderTop: '1px solid rgba(0,0,0,0.06)' }}>
              <button
                onClick={handleToggleLike}
                className="flex items-center gap-1.5 transition-all duration-200"
                style={{
                  padding: '6px 16px', borderRadius: '18px', border: 'none',
                  background: liked ? 'rgba(51,144,236,0.1)' : 'transparent',
                  cursor: 'pointer',
                  transform: likeAnim ? 'scale(1.15)' : 'scale(1)',
                  transition: 'transform 0.2s cubic-bezier(0.175, 0.885, 0.32, 1.275)',
                }}
              >
                <Heart className="w-4 h-4" style={{ color: liked ? '#3390EC' : '#A2ACB5', fill: liked ? '#3390EC' : 'none' }} />
                <span style={{ fontSize: '13px', color: liked ? '#3390EC' : '#708499', fontWeight: liked ? 600 : 400 }}>
                  {likeCount > 0 ? likeCount : '赞'}
                </span>
              </button>
              <div className="flex items-center gap-1.5" style={{ padding: '6px 16px', borderRadius: '18px', color: '#708499' }}>
                <MessageCircle className="w-4 h-4" />
                <span style={{ fontSize: '13px' }}>{trend.replyCount || '评论'}</span>
              </div>
              <button
                style={{ padding: '6px 16px', borderRadius: '18px', border: 'none', background: 'transparent', cursor: 'pointer', color: '#708499' }}
                onClick={() => toast.success('链接已复制')}
              >
                <Send className="w-4 h-4" />
              </button>
            </div>
          </div>

          {/* ── Section 2: Like Users Avatar Grid ── */}
          <div style={{ background: '#FFF', borderRadius: 12, marginBottom: 12, boxShadow: '0 1px 3px rgba(0,0,0,0.04)', overflow: 'hidden' }}>
            <div
              className="flex items-center justify-between"
              style={{ padding: '14px 20px', borderBottom: likeUsers.length > 0 ? '1px solid rgba(0,0,0,0.06)' : 'none' }}
            >
              <div className="flex items-center gap-2">
                <ThumbsUp className="w-4 h-4" style={{ color: '#3390EC' }} />
                <span style={{ fontSize: '14px', fontWeight: 600, color: '#1C2733' }}>
                  点赞好友
                </span>
              </div>
              <span style={{ fontSize: '12px', color: '#A2ACB5' }}>{likeUsers.length} 人</span>
            </div>
            {likeUsers.length > 0 ? (
              <div style={{ padding: '16px 20px' }}>
                <div className="flex flex-wrap" style={{ gap: 6 }}>
                  {(likeExpanded ? likeUsers : likeUsers.slice(0, LIKE_COLLAPSE_LIMIT)).map(user => (
                    <LikeAvatarItem key={user.id} user={user} />
                  ))}
                </div>
                {likeUsers.length > LIKE_COLLAPSE_LIMIT && (
                  <button
                    onClick={() => setLikeExpanded(prev => !prev)}
                    className="flex items-center gap-1 mt-3"
                    style={{
                      fontSize: '12px', color: '#3390EC', background: 'none',
                      border: 'none', cursor: 'pointer', padding: '4px 0',
                      fontWeight: 500,
                    }}
                  >
                    {likeExpanded ? (
                      <>
                        <ChevronUp className="w-3.5 h-3.5" />
                        收起
                      </>
                    ) : (
                      <>
                        <ChevronDown className="w-3.5 h-3.5" />
                        展开全部 {likeUsers.length} 人
                      </>
                    )}
                  </button>
                )}
              </div>
            ) : (
              <div className="flex flex-col items-center justify-center" style={{ padding: '32px 0' }}>
                <Heart className="w-10 h-10" style={{ color: '#E8EDEF', marginBottom: 8 }} />
                <div style={{ fontSize: '13px', color: '#A2ACB5' }}>暂无人点赞</div>
              </div>
            )}
          </div>

          {/* ── Section 3: Comments ── */}
          <div style={{ background: '#FFF', borderRadius: 12, boxShadow: '0 1px 3px rgba(0,0,0,0.04)', overflow: 'hidden' }}>
            <div
              className="flex items-center justify-between"
              style={{ padding: '14px 20px', borderBottom: '1px solid rgba(0,0,0,0.06)' }}
            >
              <div className="flex items-center gap-2">
                <MessageCircle className="w-4 h-4" style={{ color: '#3390EC' }} />
                <span style={{ fontSize: '14px', fontWeight: 600, color: '#1C2733' }}>
                  动态评论区
                </span>
              </div>
              <span style={{ fontSize: '12px', color: '#A2ACB5' }}>{trend.replyCount} 条</span>
            </div>
            {comments.length > 0 ? (
              <div style={{ padding: '4px 20px 8px' }}>
                {comments.map(comment => (
                  <CommentItem
                    key={comment.id}
                    comment={comment}
                    onReply={(c) => setReplyTarget(c)}
                    onDelete={handleDeleteComment}
                  />
                ))}
              </div>
            ) : (
              <div className="flex flex-col items-center justify-center" style={{ padding: '32px 0' }}>
                <MessageCircle className="w-10 h-10" style={{ color: '#E8EDEF', marginBottom: 8 }} />
                <div style={{ fontSize: '13px', color: '#A2ACB5' }}>暂无评论</div>
              </div>
            )}
          </div>

          {/* Bottom spacing */}
          <div style={{ height: 16 }} />
        </div>
      </div>

      {/* Comment Input Bar */}
      {trend.openReply && (
        <div
          className="shrink-0 flex items-center gap-2"
          style={{
            padding: '12px 20px',
            borderTop: '1px solid rgba(0,0,0,0.06)',
            background: '#FFFFFF',
          }}
        >
          {replyTarget && (
            <div className="flex items-center gap-1 shrink-0" style={{ fontSize: '11px', color: '#708499', maxWidth: 120, overflow: 'hidden', whiteSpace: 'nowrap' }}>
              <span>回复</span>
              <span style={{ color: '#3390EC', fontWeight: 500 }}>{replyTarget.replyer.name}</span>
              <button
                onClick={() => setReplyTarget(null)}
                style={{ marginLeft: 4, background: 'none', border: 'none', cursor: 'pointer', color: '#A2ACB5', padding: 0, display: 'flex' }}
              >
                <X className="w-3 h-3" />
              </button>
            </div>
          )}
          <input
            value={commentText}
            onChange={(e) => setCommentText(e.target.value)}
            onKeyDown={(e) => { if (e.key === 'Enter' && commentText.trim()) handleSubmitComment(); }}
            placeholder={replyTarget ? `回复 ${replyTarget.replyer.name}...` : '写评论...'}
            style={{
              flex: 1, borderRadius: '20px',
              border: '1px solid rgba(0,0,0,0.1)',
              padding: '8px 14px', fontSize: '13px', color: '#1C2733',
              outline: 'none', background: '#F5F7FA',
              transition: 'border-color 0.2s, box-shadow 0.2s',
            }}
            onFocus={(e) => { e.target.style.borderColor = '#3390EC'; e.target.style.boxShadow = '0 0 0 3px rgba(51,144,236,0.15)'; }}
            onBlur={(e) => { e.target.style.borderColor = 'rgba(0,0,0,0.1)'; e.target.style.boxShadow = 'none'; }}
          />
          {commentText.trim() && (
            <button
              onClick={handleSubmitComment}
              style={{
                width: 34, height: 34, borderRadius: '50%',
                border: 'none', background: '#3390EC', color: '#FFF',
                cursor: 'pointer', display: 'flex', alignItems: 'center',
                justifyContent: 'center', flexShrink: 0,
              }}
            >
              <Send className="w-4 h-4" />
            </button>
          )}
        </div>
      )}
    </div>
  );
}
