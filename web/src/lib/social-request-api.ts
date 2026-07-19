export type FriendRequestUnread = { total: number; apply: number; result: number };
export type GroupRequestUnread = { total: number; apply: number; result: number; invite: number };
export type GroupRequestTab = 'received' | 'sent' | 'invitations';
export type GroupRequestStatus = 'pending' | 'accepted' | 'rejected' | 'invalidated' | 'expired';
export type GroupRequestStatusFilter = 'all' | GroupRequestStatus;

export type GroupRequest = {
  id: string;
  tab: 'received' | 'sent';
  applicantUid: string;
  applicantName: string;
  applicantAvatar: string;
  inviterUid: string;
  groupId: string;
  groupName: string;
  groupIcon: string;
  message: string;
  status: GroupRequestStatus;
  actionable: boolean;
  read: boolean;
  createdAt: Date;
  handledAt?: Date;
  source: number;
};

export type GroupInvitation = {
  id: string;
  tab: 'invitations';
  inviterUid: string;
  inviteeUid: string;
  groupId: string;
  groupName: string;
  groupIcon: string;
  message: string;
  status: GroupRequestStatus;
  actionable: boolean;
  read: boolean;
  createdAt: Date;
  handledAt?: Date;
  expiresAt?: Date;
};

export type GroupRequestItem = GroupRequest | GroupInvitation;
export type GroupInvitationJoinState = 'joined' | 'pending_approval' | 'approval_rejected' | 'rejected' | 'expired' | 'invalidated';
export type GroupInvitationHandleResult = {
  invitationId: string;
  status: GroupRequestStatus;
  joinState: GroupInvitationJoinState;
  groupRequestId?: string;
  groupId: string;
};
export type GroupInvitationAcceptEffect = 'membership_joined' | 'approval_pending' | 'terminal_invalidated' | 'terminal_expired' | 'no_membership';
export type GroupInvitationAcceptPlan = {
  effect: GroupInvitationAcceptEffect;
  invalidateRequests: true;
  refreshGroups: boolean;
  refreshConversations: boolean;
};

export function groupInvitationAcceptEffect(joinState: GroupInvitationJoinState): GroupInvitationAcceptEffect {
  if (joinState === 'joined') return 'membership_joined';
  if (joinState === 'pending_approval') return 'approval_pending';
  if (joinState === 'invalidated') return 'terminal_invalidated';
  if (joinState === 'expired') return 'terminal_expired';
  return 'no_membership';
}

export function groupInvitationAcceptPlan(joinState: GroupInvitationJoinState): GroupInvitationAcceptPlan {
  const effect = groupInvitationAcceptEffect(joinState);
  const joined = effect === 'membership_joined';
  return { effect, invalidateRequests: true, refreshGroups: joined, refreshConversations: joined };
}
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

const groupStatusCode: Record<GroupRequestStatus, string> = {
  pending: '0', accepted: '1', rejected: '2', invalidated: '3', expired: '4',
};

function exactID(value: unknown, kind: string): string {
  if (typeof value !== 'string' || !/^[1-9]\d*$/.test(value)) throw new Error(`Invalid ${kind} id`);
  return value;
}

function unixDate(value: unknown): Date {
  return new Date(Number(value || 0) * 1000);
}

function optionalUnixDate(value: unknown): Date | undefined {
  return Number(value || 0) > 0 ? unixDate(value) : undefined;
}

export function buildFriendRequestListURL(requestClass: FriendRequestClass, status: FriendRequestStatusFilter, page: number, size: number) {
  const params = new URLSearchParams({ class: requestClass === 'received' ? '1' : '0', page: String(page), size: String(size) });
  if (status !== 'all') params.set('status', friendStatusCode[status]);
  return `/api/social/friend/putIns?${params}`;
}

export function buildGroupRequestListURL(tab: GroupRequestTab, status: GroupRequestStatusFilter, page: number, size: number) {
  const path = tab === 'invitations' ? '/api/social/group/invitations' : '/api/social/group/putInsByUid';
  const params = new URLSearchParams();
  if (tab !== 'invitations') params.set('class', tab === 'received' ? '2' : '1');
  params.set('page', String(page));
  params.set('size', String(size));
  if (status !== 'all') params.set('status', groupStatusCode[status]);
  return `${path}?${params}`;
}

export function groupRequestTargetFromBizID(bizId?: string): { tab: GroupRequestTab; itemId: string } | undefined {
  const match = /^(group|group_invite):([1-9]\d*):(apply|accept|reject|invalidated|invite)$/.exec(bizId || '');
  if (!match) return undefined;
  return { tab: match[1] === 'group_invite' ? 'invitations' : match[3] === 'apply' ? 'received' : 'sent', itemId: match[2] };
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

function mapGroupStatus(value: unknown): GroupRequestStatus {
  return Number(value) === 1 ? 'accepted' : Number(value) === 2 ? 'rejected' : Number(value) === 3 ? 'invalidated' : Number(value) === 4 ? 'expired' : 'pending';
}

export function mapGroupRequest(item: Record<string, unknown>, tab: 'received' | 'sent'): GroupRequest {
  const user = item.user && typeof item.user === 'object' ? item.user as Record<string, unknown> : {};
  const group = item.group && typeof item.group === 'object' ? item.group as Record<string, unknown> : {};
  return {
    id: exactID(item.request_id, 'group request'), tab,
    applicantUid: String(item.applicant_uid || item.user_id || ''),
    applicantName: String(user.nickname || ''), applicantAvatar: String(user.avatar || ''),
    inviterUid: String(item.inviter_user_id || ''), groupId: String(item.group_id || ''),
    groupName: String(group.name || ''), groupIcon: String(group.icon || ''), message: String(item.req_msg || ''),
    status: mapGroupStatus(item.handle_result), actionable: item.actionable === true, read: Number(item.read_state) === 1,
    createdAt: unixDate(item.req_time), handledAt: optionalUnixDate(item.handle_time), source: Number(item.actual_join_source || item.join_source || 1),
  };
}

export function mapGroupInvitation(item: Record<string, unknown>): GroupInvitation {
  return {
    id: exactID(item.id, 'group invitation'), tab: 'invitations', inviterUid: String(item.inviter_uid || ''),
    inviteeUid: String(item.invitee_uid || ''), groupId: String(item.group_id || ''), groupName: '',
    groupIcon: '', message: String(item.message || ''), status: mapGroupStatus(item.status),
    actionable: item.actionable === true, read: Number(item.read_state) === 1, createdAt: unixDate(item.created_at),
    handledAt: optionalUnixDate(item.handled_at), expiresAt: optionalUnixDate(item.expires_at),
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

export async function listGroupRequests(token: string, tab: GroupRequestTab, status: GroupRequestStatusFilter, page: number, size: number) {
  const data = await socialRequest<{ list?: Record<string, unknown>[]; total?: number }>(buildGroupRequestListURL(tab, status, page, size), token);
  return {
    list: (data.list || []).map(item => tab === 'invitations' ? mapGroupInvitation(item) : mapGroupRequest(item, tab)),
    total: data.total ?? 0,
  };
}

export function handleGroupRequest(token: string, requestId: string, result: 1 | 2) {
  return socialRequest('/api/social/group/putIn', token, { method: 'PUT', body: JSON.stringify({ request_id: requestId, handle_result: result }) });
}

export async function handleGroupInvitation(token: string, invitationId: string, result: 1 | 2, handleMsg?: string): Promise<GroupInvitationHandleResult> {
  const reason = handleMsg?.trim();
  const data = await socialRequest<Record<string, unknown>>(`/api/social/group/invitations/${encodeURIComponent(invitationId)}`, token, { method: 'PUT', body: JSON.stringify({ result, ...(reason ? { handle_msg: reason } : {}) }) });
  const joinState = String(data.join_state || '') as GroupInvitationJoinState;
  if (!['joined', 'pending_approval', 'approval_rejected', 'rejected', 'expired', 'invalidated'].includes(joinState)) throw new Error('Invalid invitation join state');
  return {
    invitationId: exactID(data.invitation_id, 'group invitation'), status: mapGroupStatus(data.status), joinState,
    groupRequestId: typeof data.group_request_id === 'string' && data.group_request_id ? exactID(data.group_request_id, 'group request') : undefined,
    groupId: String(data.group_id || ''),
  };
}

export async function markGroupRequestsRead(token: string, requestIds: string[]): Promise<GroupRequestUnread> {
  if (requestIds.length === 0) throw new Error('requestIds must not be empty');
  const data = await socialRequest<{ count?: number; apply?: number; result?: number; invite?: number }>('/api/social/group/putIns/read', token, { method: 'PUT', body: JSON.stringify({ request_ids: requestIds }) });
  return { total: data.count ?? 0, apply: data.apply ?? 0, result: data.result ?? 0, invite: data.invite ?? 0 };
}

export async function markGroupInvitationsRead(token: string, invitationIds: string[]): Promise<GroupRequestUnread> {
  if (invitationIds.length === 0) throw new Error('invitationIds must not be empty');
  const data = await socialRequest<{ count?: number; apply?: number; result?: number; invite?: number }>('/api/social/group/invitations/read', token, { method: 'PUT', body: JSON.stringify({ invitation_ids: invitationIds }) });
  return { total: data.count ?? 0, apply: data.apply ?? 0, result: data.result ?? 0, invite: data.invite ?? 0 };
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
