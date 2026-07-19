export type FriendRequestUnread = { total: number; apply: number; result: number };
export type GroupRequestUnread = { total: number; apply: number; result: number; invite: number };

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
