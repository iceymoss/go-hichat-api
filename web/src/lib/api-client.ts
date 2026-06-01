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
}

export interface ConversationItem {
  conversationId: string;
  chatType: number;
  isShow: boolean;
  isTop?: boolean;
  isMute?: boolean;
  seq: number;
  read: number;
  message?: ChatLogItem;
}

/** 获取聊天记录 — 返回 {list: ChatLogItem[]} */
export function getChatLog(token: string, conversationId: string, msgId = '', count = 30) {
  const params = new URLSearchParams({ conversationId, count: String(count) });
  if (msgId) params.set('msgId', msgId);
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

/** 获取会话列表 — 返回 {conversationList: Record<string, ConversationItem>} */
export function getConversations(token: string) {
  return imGet<{ conversationList: Record<string, ConversationItem> }>('/v1/im/conversation', token);
}

/** 建立会话 */
export function setupConversation(token: string, sendId: string, recvId: string, chatType: number) {
  return imPost('/v1/im/setup/conversation', { sendId, recvId, chatType }, token);
}

/** 更新会话 */
export function updateConversations(token: string, conversationList: Record<string, Partial<ConversationItem>>) {
  return imPut('/v1/im/conversation', { conversationList }, token);
}

/** 设置会话置顶 / 免打扰（全量覆盖两个标记） */
export function setConversationSettings(token: string, conversationId: string, isTop: boolean, isMute: boolean) {
  return imPut('/v1/im/conversation/settings', { conversationId, isTop, isMute }, token);
}
