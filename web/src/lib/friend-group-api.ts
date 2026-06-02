/**
 * 找人 / 找群 API —— 搜索用户/群（分页）、发起加好友/加群申请、群资料。
 *
 * 这些都走 Next 代理路由：
 *  - 用户搜索: GET /api/user/search?name|phone|email=&page=&size=
 *  - 加好友:   POST /api/social/friend/putIn { user_uid, req_msg? }
 *  - 群搜索:   GET /api/social/group/search?keyword=&page=&size=
 *  - 加群:     POST /api/social/group/putIn { group_id, req_msg?, join_source:1 }
 *  - 群资料:   GET /api/social/group/detail?group_id=
 */

import type { BackendUser } from './api-client';

export interface UserSearchResult {
  users: BackendUser[];
  total: number;
}

export interface GroupSearchItem {
  id: string;
  name: string;
  icon: string;
  description: string;
  creator_uid: string;
  member_count: number;
}

export interface GroupSearchResult {
  list: GroupSearchItem[];
  total: number;
}

/** 判断搜索词类型：手机号 / 邮箱 / 昵称（默认） */
export function detectUserQueryKind(q: string): 'phone' | 'email' | 'name' {
  if (/^1[3-9]\d{2,10}$/.test(q)) return 'phone';
  if (/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(q)) return 'email';
  return 'name';
}

/**
 * 搜索用户。手机号/邮箱为精确匹配（单结果，分页参数无效但仍传递），昵称为模糊分页。
 */
export async function searchUsers(
  token: string,
  query: string,
  page = 1,
  size = 20,
): Promise<UserSearchResult> {
  const q = query.trim();
  if (!q) return { users: [], total: 0 };

  const kind = detectUserQueryKind(q);
  const params = new URLSearchParams();
  params.set(kind, q);
  params.set('page', String(page));
  params.set('size', String(size));

  try {
    const resp = await fetch(`/api/user/search?${params.toString()}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    const json = await resp.json();
    if (json.success && json.data) {
      return { users: json.data.users || [], total: json.data.total ?? 0 };
    }
    return { users: [], total: 0 };
  } catch {
    return { users: [], total: 0 };
  }
}

/** 搜索群（群号精确 + 群名模糊，分页） */
export async function searchGroups(
  token: string,
  keyword: string,
  page = 1,
  size = 20,
): Promise<GroupSearchResult> {
  const q = keyword.trim();
  if (!q) return { list: [], total: 0 };

  const params = new URLSearchParams({ keyword: q, page: String(page), size: String(size) });
  try {
    const resp = await fetch(`/api/social/group/search?${params.toString()}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    const json = await resp.json();
    if (json.success && json.data) {
      return { list: json.data.list || [], total: json.data.total ?? 0 };
    }
    return { list: [], total: 0 };
  } catch {
    return { list: [], total: 0 };
  }
}

/** 发起加好友申请 */
export async function sendFriendRequest(token: string, userUid: string, reqMsg?: string): Promise<boolean> {
  const resp = await fetch('/api/social/friend/putIn', {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
    body: JSON.stringify({ user_uid: userUid, ...(reqMsg ? { req_msg: reqMsg } : {}) }),
  });
  const json = await resp.json();
  return json.success === true;
}

/** 发起加群申请（join_source=1 申请入群） */
export async function sendGroupRequest(token: string, groupId: string, reqMsg?: string): Promise<boolean> {
  const resp = await fetch('/api/social/group/putIn', {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
    body: JSON.stringify({ group_id: groupId, join_source: 1, ...(reqMsg ? { req_msg: reqMsg } : {}) }),
  });
  const json = await resp.json();
  return json.success === true;
}

export interface GroupDetail {
  group: {
    id: string;
    name: string;
    icon: string;
    description: string;
    create_uid: string;
    [k: string]: unknown;
  };
  members: unknown[];
}

/** 群资料（含成员列表） */
export async function getGroupDetail(token: string, groupId: string): Promise<GroupDetail | null> {
  try {
    const resp = await fetch(`/api/social/group/detail?group_id=${encodeURIComponent(groupId)}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    const json = await resp.json();
    if (json.success && json.data) return json.data as GroupDetail;
    return null;
  } catch {
    return null;
  }
}
