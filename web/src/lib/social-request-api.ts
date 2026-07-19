export type FriendRequestUnread = { total: number; apply: number; result: number };
export type GroupRequestUnread = { total: number; apply: number; result: number; invite: number };
export type FriendRequestStatus = 'pending' | 'accepted' | 'rejected' | 'ignored';
export type FriendRequestClass = 'received' | 'sent';
export type FriendRequestStatusFilter = 'all' | FriendRequestStatus;

export type FriendRequest = {
  id: string;
  class: FriendRequestClass;
  peerUid: string;
  nickname: string;
  avatar: string;
  sex: 'male' | 'female' | 'unknown';
  region: string;
  occupation: string;
  introduction: string;
  tags: string[];
  reqMsg: string;
  handleMsg: string;
  status: FriendRequestStatus;
  readState: boolean;
  actionable: boolean;
  reqTime: Date;
  handledAt?: Date;
  hiChatId?: string;
  email?: string;
  phone?: string;
};

const friendStatusCode: Record<FriendRequestStatus, string> = {
  pending: '0', accepted: '1', rejected: '2', ignored: '3',
};

export function buildFriendRequestListURL(requestClass: FriendRequestClass, status: FriendRequestStatusFilter, page: number, size: number) {
  const params = new URLSearchParams({ class: requestClass === 'received' ? '1' : '0', page: String(page), size: String(size) });
  if (status !== 'all') params.set('status', friendStatusCode[status]);
  return `/api/social/friend/putIns?${params}`;
}

export function friendRequestIDFromBizID(bizId?: string) {
  const match = /^friend:([1-9]\d*):(apply|accept|reject)$/.exec(bizId || '');
  return match?.[1];
}

function mapHandleResult(value: number): FriendRequestStatus {
  return value === 1 ? 'accepted' : value === 2 ? 'rejected' : value === 3 ? 'ignored' : 'pending';
}

export function mapFriendRequest(item: Record<string, unknown>, requestClass: FriendRequestClass): FriendRequest {
  let tags: string[] = [];
  if (Array.isArray(item.tags)) tags = item.tags.filter((tag): tag is string => typeof tag === 'string');
  else if (typeof item.tags === 'string' && item.tags) {
    try { const parsed = JSON.parse(item.tags); tags = Array.isArray(parsed) ? parsed.filter((tag): tag is string => typeof tag === 'string') : []; }
    catch { tags = item.tags.split(',').filter(Boolean); }
  }
  const requestId = typeof item.request_id === 'string' ? item.request_id : '';
  if (!/^[1-9]\d*$/.test(requestId)) throw new Error('Invalid friend request id');
  const peerUid = String(item.peer_uid || (requestClass === 'received' ? item.user_id : item.req_uid) || '');
  const status = typeof item.status === 'string' && ['pending', 'accepted', 'rejected', 'ignored'].includes(item.status)
    ? item.status as FriendRequestStatus : mapHandleResult(Number(item.handle_result || 0));
  const handledAt = Number(item.handled_at || 0);
  return {
    id: requestId, class: requestClass, peerUid, nickname: String(item.nickname || ''), avatar: String(item.avatar || ''),
    sex: item.sex === 1 ? 'male' : item.sex === 2 ? 'female' : 'unknown', region: String(item.region || ''),
    occupation: String(item.occupation || ''), introduction: String(item.introduction || ''), tags,
    reqMsg: String(item.req_msg || ''), handleMsg: String(item.handle_msg || ''), status,
    readState: item.read_state === 1, actionable: item.actionable === true,
    reqTime: new Date(Number(item.req_time || 0) * 1000), handledAt: handledAt > 0 ? new Date(handledAt * 1000) : undefined,
    hiChatId: peerUid || undefined, email: typeof item.email === 'string' ? item.email : undefined,
    phone: typeof item.phone === 'string' ? item.phone : undefined,
  };
}

async function socialRequest<T>(url: string, token: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, { ...init, headers: { Authorization: `Bearer ${token}`, ...(init?.body ? { 'Content-Type': 'application/json' } : {}), ...init?.headers } });
  const body = await response.json();
  if (!response.ok || !body.success) throw new Error(body.message || 'Request failed');
  return body.data as T;
}

export async function listFriendRequests(token: string, requestClass: FriendRequestClass, status: FriendRequestStatusFilter, page: number, size: number) {
  const data = await socialRequest<{ list?: Record<string, unknown>[]; total?: number }>(buildFriendRequestListURL(requestClass, status, page, size), token);
  return { list: (data.list || []).map(item => mapFriendRequest(item, requestClass)), total: data.total ?? 0 };
}

export function handleFriendRequest(token: string, requestId: string, handleResult: 1 | 2, handleMsg?: string) {
  return socialRequest('/api/social/friend/putIn', token, { method: 'PUT', body: JSON.stringify({ request_id: requestId, handle_result: handleResult, ...(handleMsg ? { handle_msg: handleMsg } : {}) }) });
}

export function deleteFriendRequest(token: string, requestId: string) {
  return socialRequest('/api/social/friend/putIn/delete', token, { method: 'POST', body: JSON.stringify({ request_id: requestId }) });
}

export async function markFriendRequestsRead(token: string, requestIds: string[]): Promise<FriendRequestUnread> {
  if (requestIds.length === 0) throw new Error('requestIds must not be empty');
  const data = await socialRequest<{ count?: number; apply?: number; result?: number }>('/api/social/friend/putIn/read', token, { method: 'PUT', body: JSON.stringify({ request_ids: requestIds }) });
  return { total: data.count ?? 0, apply: data.apply ?? 0, result: data.result ?? 0 };
}

async function getCount<T>(url: string, token: string): Promise<T> {
  const response = await fetch(url, { headers: { Authorization: `Bearer ${token}` } });
  const body = await response.json();
  if (!response.ok || !body.success) throw new Error(body.message || 'Request failed');
  return body.data as T;
}

export async function getFriendRequestUnread(token: string): Promise<FriendRequestUnread> {
  const data = await getCount<{ count?: number; apply?: number; result?: number }>('/api/social/friend/putIn/messageCount', token);
  return { total: data.count ?? 0, apply: data.apply ?? 0, result: data.result ?? 0 };
}

export async function getGroupRequestUnread(token: string): Promise<GroupRequestUnread> {
  const data = await getCount<{ count?: number; apply?: number; result?: number; invite?: number }>('/api/social/group/putIn/messageCount', token);
  return { total: data.count ?? 0, apply: data.apply ?? 0, result: data.result ?? 0, invite: data.invite ?? 0 };
}
