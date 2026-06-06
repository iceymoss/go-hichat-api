// Trend service client (calls the Next.js /api/trend proxy).
// All functions return a { success, data, message } envelope.

import type { Trend, TrendComment, MomentsNotification, TrendType, TrendScope } from './mock-data';

/* ──────────────────────────────────────────────────────────────
   Backend shapes (mirrors apps/trend/api/internal/types/types.go)
   ────────────────────────────────────────────────────────────── */

export interface BackendTrendUser {
  id: string;
  nickname: string;
  sex: number;
  avatar: string;
}

export interface BackendTrend {
  id: number;
  user_id: BackendTrendUser; // NOTE: backend uses `user_id` field name for the user object
  type: number;
  content: string;
  scope: number;
  create_time: number; // unix seconds
  reply_count: number;
  agree_count: number;
  position_name: string;
  position_point?: number[];
  title: string;
  at_user?: BackendTrendUser[];
  resources?: string[];
  cover_url: string;
  share_url: string;
  open_reply: number; // 0/1
  device_id?: string;
  ip?: string;
  is_top: number; // 0/1
}

export interface BackendDiscuss {
  id: number;
  trend_id: number;
  root_id?: number;
  father?: number;
  replyer?: BackendTrendUser;
  user_id?: BackendTrendUser; // backend field `user_id` = the user being replied to
  at_user_ids?: BackendTrendUser[];
  level?: number;
  content: string;
  agree_count?: number;
  discuss_count?: number;
  state?: number;
  read?: boolean;
  create_time: number;
  update_time?: number;
  children?: BackendDiscuss[];
}

export interface BackendLike {
  id: number;
  user: BackendTrendUser;
  trend_id: number;
  like_time: number;
}

export interface ApiEnvelope<T> {
  success: boolean;
  data?: T;
  message?: string;
}

export interface TrendPublishConfig {
  max_image_count: number;
  max_image_size_mb: number;
  allowed_image_types: string[];
  max_video_count: number;
  max_video_size_mb: number;
  max_video_duration_sec: number;
  allowed_video_types: string[];
  image_compression_enabled: boolean;
  video_compression_enabled: boolean;
  media_review_enabled: boolean;
}

export interface UploadedTrendMedia {
  url: string;
  name: string;
  size: number;
  file_type: 'image' | 'video';
  content_type: string;
  width: number;
  height: number;
  duration: number;
  cover_url: string;
}

export interface TrendDraftPayload {
  id?: number;
  type: number;
  content?: string;
  title?: string;
  resources?: string[];
  cover_url?: string;
  share_url?: string;
  position_name?: string;
  longitude?: number;
  latitude?: number;
  scope: number;
  open_reply: number;
}

/* ──────────────────────────────────────────────────────────────
   HTTP helpers
   ────────────────────────────────────────────────────────────── */

function buildQuery(params: Record<string, unknown>): string {
  const sp = new URLSearchParams();
  Object.entries(params).forEach(([k, v]) => {
    if (v === undefined || v === null) return;
    if (Array.isArray(v)) {
      v.forEach(item => sp.append(k, String(item)));
    } else {
      sp.set(k, String(v));
    }
  });
  const s = sp.toString();
  return s ? `?${s}` : '';
}

async function send<T>(
  method: 'GET' | 'POST' | 'PUT' | 'DELETE',
  path: string,
  token: string,
  body?: unknown,
): Promise<ApiEnvelope<T>> {
  const res = await fetch(`/api/trend/${path}`, {
    method,
    headers: {
      Authorization: `Bearer ${token}`,
      ...(body ? { 'Content-Type': 'application/json' } : {}),
    },
    body: body ? JSON.stringify(body) : undefined,
  });
  try {
    return (await res.json()) as ApiEnvelope<T>;
  } catch {
    return { success: false, message: '解析响应失败' };
  }
}

/* ──────────────────────────────────────────────────────────────
   Mappers: backend → frontend domain types
   ────────────────────────────────────────────────────────────── */

function localizeUserId(id: string | undefined, currentUserId: string): string {
  if (!id) return '';
  return id === currentUserId ? 'me' : id;
}

function mapBackendUserLite(u: BackendTrendUser | undefined, currentUserId: string): { id: string; name: string; avatar: string } {
  if (!u) return { id: '', name: '', avatar: '' };
  return {
    id: localizeUserId(u.id, currentUserId),
    name: u.nickname || '',
    avatar: u.avatar || '',
  };
}

export function mapBackendTrend(t: BackendTrend, currentUserId: string): Trend {
  const user = t.user_id;
  return {
    id: t.id,
    userId: localizeUserId(user?.id, currentUserId),
    type: (t.type as TrendType) || 1,
    content: t.content || '',
    scope: (t.scope as TrendScope) || 3,
    createTime: new Date((t.create_time || 0) * 1000),
    replyCount: t.reply_count || 0,
    agreeCount: t.agree_count || 0,
    positionName: t.position_name || '',
    title: t.title || '',
    atUsers: (t.at_user || []).map(u => mapBackendUserLite(u, currentUserId)),
    resources: t.resources || [],
    coverUrl: t.cover_url || '',
    shareUrl: t.share_url || '',
    openReply: (t.open_reply ?? 1) !== 0,
    isTop: (t.is_top ?? 0) !== 0,
    userName: user?.nickname || '',
    userAvatar: user?.avatar || '',
  };
}

export function mapBackendDiscuss(d: BackendDiscuss, currentUserId: string): TrendComment {
  return {
    id: d.id,
    trendId: d.trend_id,
    rootId: d.root_id || 0,
    father: d.father || 0,
    replyer: mapBackendUserLite(d.replyer, currentUserId),
    user: mapBackendUserLite(d.user_id, currentUserId),
    level: d.level || 1,
    content: d.content || '',
    agreeCount: d.agree_count || 0,
    children: (d.children || []).map(c => mapBackendDiscuss(c, currentUserId)),
    createTime: new Date((d.create_time || 0) * 1000),
  };
}

/* ──────────────────────────────────────────────────────────────
   3.1 Trend management
   ────────────────────────────────────────────────────────────── */

export interface CreateTrendPayload {
  type: number;
  content?: string;
  scope: number;
  resources?: string[];
  position_name?: string;
  longitude?: number;
  latitude?: number;
  title?: string;
  at_user_ids?: number[];
  open_reply?: boolean;
  cover_url?: string;
  share_url?: string;
  device?: string;
  ip?: string;
}

export function createTrend(token: string, body: CreateTrendPayload) {
  return send<{ trend_id: number; code: number }>('POST', 'trend', token, body);
}

export function deleteTrend(token: string, trendId: number) {
  return send<{ success: boolean }>('POST', 'trend/delete', token, { trend_id: trendId });
}

export interface UpdateTrendPayload {
  trend_id: number;
  is_top?: number;     // 0/1
  open_reply?: number; // 0/1
  scope?: number;      // 1 visible, 2 invisible
}

export function updateTrend(token: string, body: UpdateTrendPayload) {
  return send<Record<string, unknown>>('PUT', 'trend/update', token, body);
}

export function getTrendDetail(token: string, trendId: number) {
  return send<{ trend: BackendTrend }>('GET', `trend/detail${buildQuery({ trend_id: trendId })}`, token);
}

export function getTrendPublishConfig(token: string) {
  return send<TrendPublishConfig>('GET', 'trend/config', token);
}

export async function uploadTrendMedia(token: string, file: File): Promise<ApiEnvelope<UploadedTrendMedia>> {
  const form = new FormData();
  form.set('file', file);
  const res = await fetch('/api/trend/trend/media/upload', {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: form,
  });
  try {
    return (await res.json()) as ApiEnvelope<UploadedTrendMedia>;
  } catch {
    return { success: false, message: '解析响应失败' };
  }
}

export function saveTrendDraft(token: string, draft: TrendDraftPayload) {
  return send<{ draft: TrendDraftPayload & { id: number; create_time: number; update_time: number } }>(
    'POST',
    'trend/draft',
    token,
    { draft },
  );
}

export function getTrendDraft(token: string, draftId?: number) {
  return send<{ draft: (TrendDraftPayload & { id: number; create_time: number; update_time: number }) | null }>(
    'GET',
    `trend/draft${buildQuery({ draft_id: draftId })}`,
    token,
  );
}

export function deleteTrendDraft(token: string, draftId: number) {
  return send<{ success: boolean }>('DELETE', 'trend/draft', token, { draft_id: draftId });
}

export interface LatestTrendsResp {
  list: BackendTrend[];
  last_trend_id: number;
  has_more: boolean;
}

export function getLatestTrends(token: string, opts: { last_trend_id?: number; count?: number } = {}) {
  return send<LatestTrendsResp>('GET', `trends/latest${buildQuery(opts)}`, token);
}

export interface UserTrendsResp {
  list: BackendTrend[];
  top_list: BackendTrend[];
  last_id: number;
  last_time: number;
}

export function getUserTrends(token: string, targetUserId: string, lastId = 0) {
  // Backend field is target_user_id (int). Numeric ids are strings on the frontend, but we pass as-is.
  return send<UserTrendsResp>(
    'GET',
    `user/trends${buildQuery({ target_user_id: targetUserId, last_id: lastId })}`,
    token,
  );
}

/* ──────────────────────────────────────────────────────────────
   3.2 Comments
   ────────────────────────────────────────────────────────────── */

export interface CreateCommentPayload {
  trend_id: number;
  father?: number;
  user_id: number | string;
  content: string;
  at_users?: number[];
}

export function createComment(token: string, body: CreateCommentPayload) {
  return send<Record<string, unknown>>('POST', 'trend/comment', token, body);
}

export interface CommentTreeResp {
  discuss: {
    trend_discuss: Record<string, { discusses: BackendDiscuss[] }>;
  };
  last_id: number;
  last_time: number;
}

export function getCommentTree(token: string, trendIds: number[], opts: { last_id?: number; last_time?: number } = {}) {
  // backend uses repeated query param trend_id=...
  return send<CommentTreeResp>('GET', `trend/comment/tree${buildQuery({ trend_id: trendIds, ...opts })}`, token);
}

export function getRootComments(token: string, trendId: number, opts: { last_id?: number; last_time?: number } = {}) {
  return send<{ list: BackendDiscuss[]; last_id: number; last_time: number; total: number }>(
    'GET',
    `trend/comment/root${buildQuery({ trend_id: trendId, ...opts })}`,
    token,
  );
}

export function getChildComments(token: string, father: number, opts: { last_id?: number; last_time?: number } = {}) {
  return send<{ list: BackendDiscuss[]; last_id: number; last_time: number; total: number }>(
    'GET',
    `trend/comment/children${buildQuery({ father, ...opts })}`,
    token,
  );
}

export function deleteComment(token: string, id: number) {
  return send<{ success: boolean }>('DELETE', 'trend/comment', token, { id });
}

export interface UnreadResp {
  replies: BackendDiscuss[];
  likes: BackendLike[];
  like_last_id: number;
  discuss_last_id: number;
  last_time: number;
}

export function getUnread(
  token: string,
  opts: { like_last_id?: number; discuss_last_id?: number; last_time?: number } = {},
) {
  return send<UnreadResp>('GET', `trend/unread${buildQuery(opts)}`, token);
}

export function markCommentsRead(token: string, disIds: number[]) {
  return send<Record<string, unknown>>('PUT', 'trend/comment/mark-read', token, { dis_ids: disIds });
}

/* ──────────────────────────────────────────────────────────────
   3.3 Likes
   ────────────────────────────────────────────────────────────── */

export function toggleLike(token: string, trendId: number, authorId: number, likeType = 1) {
  return send<Record<string, unknown>>('POST', 'trend/like', token, {
    trend_id: trendId,
    author_id: authorId,
    like_type: likeType,
  });
}

export function getLikeSummary(token: string, userId: string, trendId: string) {
  return send<{ summary_json: string }>(
    'GET',
    `trend/like/summary${buildQuery({ user_id: userId, trend_id: trendId })}`,
    token,
  );
}

export function getBatchLikeSummary(token: string, trendIds: number[]) {
  return send<{ trend_likes: Record<string, BackendTrendUser[]> }>(
    'GET',
    `trend/like/batch-summary${buildQuery({ trend_id: trendIds })}`,
    token,
  );
}

export function getLikedUsers(token: string, trendId: number, cursor = 0, limit = 50) {
  return send<{ users: BackendTrendUser[]; next_cursor: number; has_more: boolean; total: number }>(
    'GET',
    `trend/like/users${buildQuery({ trend_id: trendId, cursor, limit })}`,
    token,
  );
}

export function getUnreadLikes(token: string, userId: string, lastId = 0) {
  return send<{ likes_json: string; total_unread: number; last_id: number }>(
    'GET',
    `trend/like/unread${buildQuery({ user_id: userId, last_id: lastId })}`,
    token,
  );
}

export function markLikesRead(token: string, likeIds: number[]) {
  return send<Record<string, unknown>>('PUT', 'trend/like/mark-read', token, { like_ids: likeIds });
}

/* ──────────────────────────────────────────────────────────────
   Notification helpers (compose unread replies + likes)
   ────────────────────────────────────────────────────────────── */

export function unreadToNotifications(data: UnreadResp, currentUserId: string): MomentsNotification[] {
  const list: MomentsNotification[] = [];
  (data.replies || []).forEach(r => {
    list.push({
      id: r.id,
      type: 'reply',
      trendId: r.trend_id,
      trendContent: '',
      actor: mapBackendUserLite(r.replyer, currentUserId),
      content: r.content,
      read: !!r.read,
      createTime: new Date((r.create_time || 0) * 1000),
    });
  });
  (data.likes || []).forEach(l => {
    list.push({
      id: l.id,
      type: 'like',
      trendId: l.trend_id,
      trendContent: '',
      actor: mapBackendUserLite(l.user, currentUserId),
      read: false,
      createTime: new Date((l.like_time || 0) * 1000),
    });
  });
  return list;
}

/**
 * Flatten a CommentTreeResp into the frontend-friendly map<trendId, TrendComment[]>.
 */
export function commentTreeToMap(
  tree: CommentTreeResp,
  currentUserId: string,
): Record<number, TrendComment[]> {
  const out: Record<number, TrendComment[]> = {};
  const map = tree?.discuss?.trend_discuss || {};
  Object.entries(map).forEach(([tid, bucket]) => {
    out[Number(tid)] = (bucket?.discusses || []).map(d => mapBackendDiscuss(d, currentUserId));
  });
  return out;
}

export function batchSummaryToLikeUsersMap(
  payload: { trend_likes: Record<string, BackendTrendUser[]> } | undefined,
  currentUserId: string,
): Record<number, { id: string; name: string; avatar: string }[]> {
  const out: Record<number, { id: string; name: string; avatar: string }[]> = {};
  const map = payload?.trend_likes || {};
  Object.entries(map).forEach(([tid, users]) => {
    out[Number(tid)] = (users || []).map(u => mapBackendUserLite(u, currentUserId));
  });
  return out;
}
