import { afterEach, describe, expect, test } from 'bun:test';
import { buildFriendRequestListURL, buildGroupRequestListURL, friendRequestIDFromBizID, groupInvitationAcceptPlan, groupRequestTargetFromBizID, handleGroupInvitation, mapFriendRequest, mapGroupInvitation, mapGroupRequest, markFriendRequestsRead, markGroupInvitationsRead, markGroupRequestsRead } from './social-request-api';

const originalFetch = globalThis.fetch;
afterEach(() => { globalThis.fetch = originalFetch; });

describe('friend request API contract', () => {
  test('builds one paginated query and omits status for all', () => {
    expect(buildFriendRequestListURL('received', 'all', 2, 20)).toBe('/api/social/friend/putIns?class=1&page=2&size=20');
    expect(buildFriendRequestListURL('sent', 'accepted', 1, 20)).toBe('/api/social/friend/putIns?class=0&page=1&size=20&status=1');
  });

  test('preserves decimal IDs above Number.MAX_SAFE_INTEGER', () => {
    const request = mapFriendRequest({ request_id: '9007199254740993', peer_uid: '22', handle_result: 0, read_state: 0, actionable: true, req_time: 1, handled_at: 2 }, 'received');
    expect(request.id).toBe('9007199254740993');
    expect(request.peerUid).toBe('22');
    expect(request.actionable).toBe(true);
    expect(request.handledAt?.getTime()).toBe(2000);
  });

  test('uses direction-specific peer fallback', () => {
    expect(mapFriendRequest({ request_id: '1', user_id: '8' }, 'received').peerUid).toBe('8');
    expect(mapFriendRequest({ request_id: '2', req_uid: '9' }, 'sent').peerUid).toBe('9');
  });

  test('extracts an exact request ID from notification bizId', () => {
    expect(friendRequestIDFromBizID('friend:9007199254740993:accept')).toBe('9007199254740993');
    expect(friendRequestIDFromBizID('group:1:accept')).toBeUndefined();
  });

  test('never sends an empty receipt ID list', async () => {
    let called = false;
    globalThis.fetch = (() => { called = true; }) as unknown as typeof fetch;
    expect(markFriendRequestsRead('token', [])).rejects.toThrow('requestIds must not be empty');
    expect(called).toBe(false);
  });
});

describe('group request API contract', () => {
  test('builds one paginated query for each tab with backend classes', () => {
    expect(buildGroupRequestListURL('received', 'all', 2, 20)).toBe('/api/social/group/putInsByUid?class=2&page=2&size=20');
    expect(buildGroupRequestListURL('sent', 'accepted', 1, 10)).toBe('/api/social/group/putInsByUid?class=1&page=1&size=10&status=1');
    expect(buildGroupRequestListURL('invitations', 'pending', 3, 5)).toBe('/api/social/group/invitations?page=3&size=5&status=0');
    expect(buildGroupRequestListURL('received', 'invalidated', 1, 20)).toBe('/api/social/group/putInsByUid?class=2&page=1&size=20&status=3');
    expect(buildGroupRequestListURL('invitations', 'invalidated', 1, 20)).toBe('/api/social/group/invitations?page=1&size=20&status=3');
    expect(buildGroupRequestListURL('invitations', 'expired', 1, 20)).toBe('/api/social/group/invitations?page=1&size=20&status=4');
  });

  test('maps canonical request and invitation fields without rounding IDs', () => {
    const request = mapGroupRequest({ request_id: '9007199254740993', applicant_uid: '8', inviter_user_id: '7', group_id: '9', user: { nickname: 'Alice', avatar: '/alice.png' }, group: { name: 'Builders', icon: '/group.png' }, handle_result: 0, actionable: true, read_state: 1, req_time: 2 }, 'received');
    expect(request).toMatchObject({ id: '9007199254740993', applicantUid: '8', applicantName: 'Alice', applicantAvatar: '/alice.png', inviterUid: '7', groupName: 'Builders', groupIcon: '/group.png', status: 'pending', actionable: true, read: true });
    const invitation = mapGroupInvitation({ id: '9007199254740995', inviter_uid: '7', invitee_uid: '8', group_id: '9', status: 4, actionable: false, read_state: 0, created_at: 3 });
    expect(invitation).toMatchObject({ id: '9007199254740995', status: 'expired', actionable: false, read: false });
  });

  test('locates exact cross-page notification targets', () => {
    expect(groupRequestTargetFromBizID('group:9007199254740993:apply')).toEqual({ tab: 'received', itemId: '9007199254740993' });
    expect(groupRequestTargetFromBizID('group:2:accept')).toEqual({ tab: 'sent', itemId: '2' });
    expect(groupRequestTargetFromBizID('group_invite:3:invite')).toEqual({ tab: 'invitations', itemId: '3' });
    expect(groupRequestTargetFromBizID('group:3:invite:extra')).toBeUndefined();
  });

  test('reads exact visible IDs and returns categorized counts', async () => {
    const calls: Array<{ url: string; body: string }> = [];
    globalThis.fetch = (async (input, init) => {
      calls.push({ url: String(input), body: String(init?.body) });
      return new Response(JSON.stringify({ success: true, data: { count: 6, apply: 1, result: 2, invite: 3 } }), { status: 200 });
    }) as typeof fetch;
    await expect(markGroupRequestsRead('token', ['9007199254740993'])).resolves.toEqual({ total: 6, apply: 1, result: 2, invite: 3 });
    await expect(markGroupInvitationsRead('token', ['9007199254740995'])).resolves.toEqual({ total: 6, apply: 1, result: 2, invite: 3 });
    expect(calls).toEqual([
      { url: '/api/social/group/putIns/read', body: '{"request_ids":["9007199254740993"]}' },
      { url: '/api/social/group/invitations/read', body: '{"invitation_ids":["9007199254740995"]}' },
    ]);
  });

  test('never sends empty group receipt lists', () => {
    expect(markGroupRequestsRead('token', [])).rejects.toThrow('requestIds must not be empty');
    expect(markGroupInvitationsRead('token', [])).rejects.toThrow('invitationIds must not be empty');
  });

  test.each([
    ['joined', 'accepted'],
    ['pending_approval', 'accepted'],
    ['invalidated', 'invalidated'],
    ['expired', 'expired'],
  ] as const)('preserves invitation handle join state %s', async (joinState, status) => {
    const statusCode = status === 'accepted' ? 1 : status === 'invalidated' ? 3 : 4;
    globalThis.fetch = (async () => new Response(JSON.stringify({ success: true, data: {
      invitation_id: '9007199254740993', status: statusCode, join_state: joinState,
      group_request_id: joinState === 'pending_approval' ? '9007199254740995' : '', group_id: '8',
    } }), { status: 200 })) as unknown as typeof fetch;
    await expect(handleGroupInvitation('token', '9007199254740993', 1)).resolves.toEqual({
      invitationId: '9007199254740993', status, joinState,
      groupRequestId: joinState === 'pending_approval' ? '9007199254740995' : undefined, groupId: '8',
    });
  });

  test.each([
    ['joined', 'membership_joined', true],
    ['pending_approval', 'approval_pending', false],
    ['invalidated', 'terminal_invalidated', false],
    ['expired', 'terminal_expired', false],
    ['rejected', 'no_membership', false],
  ] as const)('plans invitation state %s without implying membership', (joinState, effect, refreshMembership) => {
    expect(groupInvitationAcceptPlan(joinState)).toEqual({
      effect, invalidateRequests: true, refreshGroups: refreshMembership, refreshConversations: refreshMembership,
    });
  });

  test('sends an optional trimmed invitation rejection reason only when nonempty', async () => {
    const bodies: string[] = [];
    globalThis.fetch = (async (_input, init) => {
      bodies.push(String(init?.body));
      return new Response(JSON.stringify({ success: true, data: {
        invitation_id: '1', status: 2, join_state: 'rejected', group_id: '2',
      } }), { status: 200 });
    }) as typeof fetch;
    await handleGroupInvitation('token', '1', 2, '  not now  ');
    await handleGroupInvitation('token', '1', 2, '   ');
    expect(bodies.map(body => JSON.parse(body))).toEqual([
      { result: 2, handle_msg: 'not now' },
      { result: 2 },
    ]);
  });
});
