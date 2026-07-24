import { afterEach, describe, expect, test } from 'bun:test';
import { buildFriendRequestListURL, buildGroupRequestListURL, clearMatchingPublicNotificationTargets, friendRequestIDFromBizID, friendRequestNotificationTargets, groupInvitationAcceptPlan, groupRequestNotificationTargets, groupRequestTargetFromBizID, handleGroupInvitation, handleGroupRequest, mapFriendRequest, mapGroupInvitation, mapGroupRequest, markFriendRequestsRead, markGroupInvitationsRead, markGroupRequestsRead, notificationNavigationTarget, samePublicNotificationTargets } from './social-request-api';

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

  test('maps committed receipts to exact deduplicated public targets', () => {
    const received = mapFriendRequest({ request_id: '1', handle_result: 1 }, 'received');
    const accepted = mapFriendRequest({ request_id: '2', handle_result: 1 }, 'sent');
    const pending = mapFriendRequest({ request_id: '3', handle_result: 0 }, 'sent');
    expect(friendRequestNotificationTargets([received, accepted, accepted, pending])).toEqual([
      { notify_type: 'friend.apply', biz_id: 'friend:1:apply' },
      { notify_type: 'friend.accept', biz_id: 'friend:2:accept' },
    ]);
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
    const request = mapGroupRequest({ request_id: '9007199254740993', applicant_uid: '8', inviter_user_id: '7', group_id: '9', user: { nickname: 'Alice', avatar: '/alice.png' }, group: { name: 'Builders', icon: '/group.png' }, handle_result: 3, handle_msg: 'not now', invalid_reason: 'group closed', actionable: false, read_state: 1, req_time: 2 }, 'received');
    expect(request).toMatchObject({ id: '9007199254740993', applicantUid: '8', applicantName: 'Alice', applicantAvatar: '/alice.png', inviterUid: '7', groupName: 'Builders', groupIcon: '/group.png', status: 'invalidated', handleMessage: 'not now', invalidReason: 'group closed', actionable: false, read: true });
    const invitation = mapGroupInvitation({ id: '9007199254740995', inviter_uid: '7', invitee_uid: '8', group_id: '9', status: 4, reject_reason: 'busy', actionable: false, read_state: 0, created_at: 3 });
    expect(invitation).toMatchObject({ id: '9007199254740995', status: 'expired', rejectReason: 'busy', actionable: false, read: false });
  });

  test('locates exact cross-page notification targets', () => {
    expect(groupRequestTargetFromBizID('group:9007199254740993:apply')).toEqual({ tab: 'received', itemId: '9007199254740993' });
    expect(groupRequestTargetFromBizID('group:2:accept')).toEqual({ tab: 'sent', itemId: '2' });
    expect(groupRequestTargetFromBizID('group_invite:3:invite')).toEqual({ tab: 'invitations', itemId: '3' });
    expect(groupRequestTargetFromBizID('group:3:invite:extra')).toBeUndefined();
  });

  test('maps each group receipt class to its exact public target', () => {
    const received = mapGroupRequest({ request_id: '1' }, 'received');
    const rejected = mapGroupRequest({ request_id: '2', handle_result: 2 }, 'sent');
    const pending = mapGroupRequest({ request_id: '3' }, 'sent');
    const invitation = mapGroupInvitation({ id: '4' });
    expect(groupRequestNotificationTargets([received, rejected, pending, invitation, invitation])).toEqual([
      { notify_type: 'group.apply', biz_id: 'group:1:apply' },
      { notify_type: 'group.reject', biz_id: 'group:2:reject' },
      { notify_type: 'group.invite', biz_id: 'group_invite:4:invite' },
    ]);
  });

  test('routes only request notifications to request panels', () => {
    expect(notificationNavigationTarget('friend.accept', 'friend:2:accept')).toEqual({ kind: 'friendRequest', tab: 'sent', requestId: '2' });
    expect(notificationNavigationTarget('group.invite', 'group_invite:4:invite')).toEqual({ kind: 'groupRequest', tab: 'invitations', itemId: '4' });
    expect(notificationNavigationTarget('group.admin.set', undefined, '9')).toEqual({ kind: 'groupDetail', groupId: '9' });
    expect(notificationNavigationTarget('group.removed')).toBeUndefined();
    expect(notificationNavigationTarget('group.removed', undefined, '99')).toBeUndefined();
    expect(notificationNavigationTarget('group.unknown', 'group:1:apply', '9')).toBeUndefined();
  });

  test('matches only the exact submitted retry snapshot', () => {
    const submitted = [{ notify_type: 'group.apply', biz_id: 'group:1:apply' }];
    const changed = [...submitted, { notify_type: 'group.apply', biz_id: 'group:2:apply' }];
    expect(samePublicNotificationTargets(submitted, [...submitted])).toBe(true);
    expect(samePublicNotificationTargets(submitted, changed)).toBe(false);
    expect(clearMatchingPublicNotificationTargets(submitted, [...submitted])).toEqual([]);
    expect(clearMatchingPublicNotificationTargets(changed, submitted)).toBe(changed);
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

  test('sends an optional trimmed group rejection reason', async () => {
    let body = '';
    globalThis.fetch = (async (_input, init) => {
      body = String(init?.body);
      return new Response(JSON.stringify({ success: true, data: {} }), { status: 200 });
    }) as typeof fetch;
    await handleGroupRequest('token', '1', 2, '  group is full  ');
    expect(JSON.parse(body)).toEqual({ request_id: '1', handle_result: 2, handle_msg: 'group is full' });
  });
});
