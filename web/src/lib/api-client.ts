// Backend API client for Go services

const BACKEND_BASE = process.env.BACKEND_API_URL || 'http://127.0.0.1:8887';
const SOCIAL_BASE = process.env.SOCIAL_API_URL || 'http://127.0.0.1:8888';
const TREND_BASE = process.env.TREND_API_URL || 'http://127.0.0.1:8891';
const IM_BASE = process.env.IM_API_URL || 'http://127.0.0.1:8890';

export interface BackendResp<T = unknown> {
  code: number;
  msg: string;
  data: T;
}

export interface BackendUser {
  id: string;
  mobile: string;
  nickname: string;
  sex: number;
  avatar: string;
  lastLogin: string;
  introduction: string;
  email: string;
  region: string;
  occupation: string;
  tags: string;
  moments_cover: string;
}

export async function backendFetch<T = unknown>(
  path: string,
  options: RequestInit = {},
): Promise<BackendResp<T>> {
  const url = `${BACKEND_BASE}${path}`;
  const res = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
  });
  return res.json() as Promise<BackendResp<T>>;
}

export async function backendGet<T = unknown>(
  path: string,
  token?: string,
): Promise<BackendResp<T>> {
  return backendFetch<T>(path, {
    method: 'GET',
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });
}

export async function backendPost<T = unknown>(
  path: string,
  body: unknown,
  token?: string,
): Promise<BackendResp<T>> {
  return backendFetch<T>(path, {
    method: 'POST',
    body: JSON.stringify(body),
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });
}

export async function backendPut<T = unknown>(
  path: string,
  body: unknown,
  token?: string,
): Promise<BackendResp<T>> {
  return backendFetch<T>(path, {
    method: 'PUT',
    body: JSON.stringify(body),
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });
}

// Social service clients (port 8888)
export async function socialGet<T = unknown>(
  path: string,
  token?: string,
): Promise<BackendResp<T>> {
  const url = `${SOCIAL_BASE}${path}`;
  const res = await fetch(url, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
  });
  return res.json() as Promise<BackendResp<T>>;
}

export async function socialPost<T = unknown>(
  path: string,
  body: unknown,
  token?: string,
): Promise<BackendResp<T>> {
  const url = `${SOCIAL_BASE}${path}`;
  const res = await fetch(url, {
    method: 'POST',
    body: JSON.stringify(body),
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
  });
  return res.json() as Promise<BackendResp<T>>;
}

export async function socialPut<T = unknown>(
  path: string,
  body: unknown,
  token?: string,
): Promise<BackendResp<T>> {
  const url = `${SOCIAL_BASE}${path}`;
  const res = await fetch(url, {
    method: 'PUT',
    body: JSON.stringify(body),
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
  });
  return res.json() as Promise<BackendResp<T>>;
}

// Trend service client (port 8891)
export async function trendGet<T = unknown>(
  path: string,
  token?: string,
): Promise<BackendResp<T>> {
  const url = `${TREND_BASE}${path}`;
  const res = await fetch(url, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
  });
  return res.json() as Promise<BackendResp<T>>;
}

// ========== IM service clients (port 8890) ==========

// IM API 直接返回数据对象（不是 {code, data} 包装），用独立的 fetch 函数
async function imFetch<T = unknown>(path: string, options: RequestInit = {}): Promise<T> {
  const url = `${IM_BASE}${path}`;
  const res = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
  });
  if (!res.ok) throw new Error(`IM API error: ${res.status} ${await res.text()}`);
  return res.json() as Promise<T>;
}

export function imGet<T = unknown>(path: string, token?: string): Promise<T> {
  return imFetch<T>(path, {
    method: 'GET',
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });
}

export function imPost<T = unknown>(path: string, body: unknown, token?: string): Promise<T> {
  return imFetch<T>(path, {
    method: 'POST',
    body: JSON.stringify(body),
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });
}

export function imPut<T = unknown>(path: string, body: unknown, token?: string): Promise<T> {
  return imFetch<T>(path, {
    method: 'PUT',
    body: JSON.stringify(body),
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });
}

// ========== IM typed API helpers ==========

export interface ChatLogItem {
  id: string;
  conversationId: string;
  sendId: string;
  recvId: string;
  msgType: number;
  msgContent: string;
  chatType: number;
  sendTime: number;
  /** base64(bitmap)；私聊 "AQ==" 表示对方已读，空串或缺失视为未读；群聊为成员 bitmap */
  readRecords?: string;
  /** 引用/回复消息（JSON 字符串：{"id","name","preview"}） */
  quote?: string;
  /** 消息状态：0=正常 1=已撤回（已撤回时 msgContent/quote 被后端置空） */
  status?: number;
  /** 撤回操作者 uid */
  recalledBy?: string;
  /** 被 @ 的成员 uid 列表（群聊） */
  atUsers?: string[];
  /** 是否 @所有人（群聊） */
  atAll?: boolean;
}

export interface ConversationItem {
  conversationId: string;
  chatType: number;
  isShow: boolean;
  isTop?: boolean;
  isMute?: boolean;
  /** 是否有未读的 @我（群聊） */
  hasAtMe?: boolean;
  seq: number;
  read: number;
  message?: ChatLogItem;
}

/** 获取聊天记录 — 返回 {list: ChatLogItem[]}。direction: older(默认)/newer/around */
export function getChatLog(token: string, conversationId: string, msgId = '', count = 30, direction = '') {
  const params = new URLSearchParams({ conversationId, count: String(count) });
  if (msgId) params.set('msgId', msgId);
  if (direction) params.set('direction', direction);
  return imGet<{ list: ChatLogItem[] }>(`/v1/im/chatlog?${params}`, token);
}

export interface ReadRecordUser {
  id: string;
  nickname: string;
  avatar: string;
  /** 已读时间戳（unix nano），未读时为 0 或不存在 */
  readAt?: number;
}

/** 获取消息已读记录（群聊查看谁读/谁未读） */
export function getChatLogReadRecords(token: string, msgId: string) {
  return imGet<{ reads: ReadRecordUser[]; unReads: ReadRecordUser[] }>(`/v1/im/chatlog/readRecords?msgId=${msgId}`, token);
}

/** 获取会话中 @我 且未读的消息列表（按 sendTime 升序），用于"有人@我"快速跳转 */
export function getAtMeMessages(token: string, conversationId: string, count = 0) {
  const params = new URLSearchParams({ conversationId });
  if (count > 0) params.set('count', String(count));
  return imGet<{ list: ChatLogItem[] }>(`/v1/im/chatlog/atme?${params}`, token);
}

// ---------------- 公共通知（好友/群申请等实时通知） ----------------

export interface NotificationItem {
  id: string;
  notifyType: string; // friend.apply / friend.accept / group.apply ...
  bizId?: string;
  actorId?: string;
  groupId?: string;
  title?: string;
  content?: string;
  payload?: string;
  isRead: number; // 0 未读 1 已读
  createTime: number;
}

/** 拉取当前用户的通知列表（公共通知通道） */
export function listNotifications(token: string, unreadOnly = false, offset = 0, limit = 20) {
  const params = new URLSearchParams({ offset: String(offset), limit: String(limit) });
  if (unreadOnly) params.set('unreadOnly', 'true');
  return imGet<{ list: NotificationItem[] }>(`/v1/im/notifications?${params}`, token);
}

/** 当前用户未读通知数 */
export function getNotificationUnreadCount(token: string) {
  return imGet<{ count: number }>(`/v1/im/notifications/unreadCount`, token);
}

export type NotificationReadResult = { affected: number; unreadCount: number };
export type NotificationBusinessTarget = { notify_type: string; biz_id: string };

async function markNotificationsRead(token: string, body: unknown): Promise<NotificationReadResult> {
  const result = await imPost<{ affected: number; unread_count: number }>(`/v1/im/notifications/read`, body, token);
  return { affected: result.affected, unreadCount: result.unread_count };
}

/** 标记明确的通知 ID 已读。 */
export function markNotificationIdsRead(token: string, ids: string[]) {
  if (ids.length === 0) throw new Error('notification ids must not be empty');
  return markNotificationsRead(token, { ids });
}

/** 标记当前用户的全部通知已读。 */
export function markAllNotificationsRead(token: string) {
  return markNotificationsRead(token, {});
}

/** 按本次已提交的业务目标精确同步公共通知已读状态。 */
export async function markBusinessNotificationsRead(token: string, targets: NotificationBusinessTarget[]) {
  if (targets.length === 0) throw new Error('notification targets must not be empty');
  let affected = 0;
  let unreadCount = 0;
  for (let offset = 0; offset < targets.length; offset += 100) {
    const result = await markNotificationsRead(token, { targets: targets.slice(offset, offset + 100) });
    affected += result.affected;
    unreadCount = result.unreadCount;
  }
  return { affected, unreadCount };
}

/** 获取会话列表 — 返回 {conversationList: Record<string, ConversationItem>} */
export function getConversations(token: string) {
  return imGet<{ conversationList: Record<string, ConversationItem> }>('/v1/im/conversation', token);
}

/** 建立会话 */
export function setupConversation(token: string, sendId: string, recvId: string, chatType: number) {
  return imPost('/v1/im/setup/conversation', { sendId, recvId, chatType }, token);
}

export interface ImUploadResp {
  url: string;
  name: string;
  size: number;
  /** image / video / voice / file */
  fileType: string;
}

/** 上传富媒体文件到 im 服务（直连，CORS 已开），返回访问 URL 与类型 */
export async function imUpload(token: string, file: File | Blob, filename?: string): Promise<ImUploadResp> {
  const fd = new FormData();
  fd.append('file', file, filename || (file as File).name);
  const res = await fetch(`${IM_BASE}/v1/im/upload`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: fd,
  });
  if (!res.ok) throw new Error(`IM upload error: ${res.status} ${await res.text()}`);
  return res.json() as Promise<ImUploadResp>;
}

/** 更新会话 */
export function updateConversations(token: string, conversationList: Record<string, Partial<ConversationItem>>) {
  return imPut('/v1/im/conversation', { conversationList }, token);
}

/** 设置会话置顶 / 免打扰（全量覆盖两个标记） */
export function setConversationSettings(token: string, conversationId: string, isTop: boolean, isMute: boolean) {
  return imPut('/v1/im/conversation/settings', { conversationId, isTop, isMute }, token);
}

/** 撤回消息（本人限时 / 群管理员不限时，校验在后端） */
export function recallMsg(token: string, conversationId: string, msgId: string, chatType: number) {
  return imPost('/v1/im/chatlog/recall', { conversationId, msgId, chatType }, token);
}
