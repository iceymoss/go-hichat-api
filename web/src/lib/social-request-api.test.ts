import { afterEach, describe, expect, test } from 'bun:test';
import { buildFriendRequestListURL, friendRequestIDFromBizID, mapFriendRequest, markFriendRequestsRead } from './social-request-api';

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
