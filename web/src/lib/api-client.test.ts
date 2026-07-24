import { afterEach, describe, expect, test } from 'bun:test';

import { markAllNotificationsRead, markBusinessNotificationsRead, markNotificationIdsRead } from './api-client';

const originalFetch = globalThis.fetch;

afterEach(() => { globalThis.fetch = originalFetch; });

describe('notification read API contract', () => {
  test('keeps notification IDs as strings and maps unread_count', async () => {
    const bodies: unknown[] = [];
    globalThis.fetch = (async (_input, init) => {
      bodies.push(JSON.parse(String(init?.body)));
      return new Response(JSON.stringify({ affected: 1, unread_count: 4 }), { status: 200 });
    }) as typeof fetch;

    await expect(markNotificationIdsRead('token', ['9007199254740993'])).resolves.toEqual({ affected: 1, unreadCount: 4 });
    expect(bodies).toEqual([{ ids: ['9007199254740993'] }]);
  });

  test('sends exact paired business targets and an explicit all request', async () => {
    const bodies: unknown[] = [];
    globalThis.fetch = (async (_input, init) => {
      bodies.push(JSON.parse(String(init?.body)));
      return new Response(JSON.stringify({ affected: 2, unread_count: 0 }), { status: 200 });
    }) as typeof fetch;

    await markBusinessNotificationsRead('token', [
      { notify_type: 'friend.apply', biz_id: 'friend:1:apply' },
      { notify_type: 'group.invite', biz_id: 'group_invite:2:invite' },
    ]);
    await markAllNotificationsRead('token');

    expect(bodies).toEqual([
      { targets: [
        { notify_type: 'friend.apply', biz_id: 'friend:1:apply' },
        { notify_type: 'group.invite', biz_id: 'group_invite:2:invite' },
      ] },
      {},
    ]);
  });

  test('rejects empty explicit ID and target lists without fetching', () => {
    let called = false;
    globalThis.fetch = (() => { called = true; }) as unknown as typeof fetch;
    expect(() => markNotificationIdsRead('token', [])).toThrow('notification ids must not be empty');
    expect(() => markBusinessNotificationsRead('token', [])).toThrow('notification targets must not be empty');
    expect(called).toBe(false);
  });

  test('batches more than 100 exact business targets', async () => {
    const sizes: number[] = [];
    globalThis.fetch = (async (_input, init) => {
      const body = JSON.parse(String(init?.body));
      sizes.push(body.targets.length);
      return new Response(JSON.stringify({ affected: body.targets.length, unread_count: 3 }), { status: 200 });
    }) as typeof fetch;
    const targets = Array.from({ length: 205 }, (_, i) => ({ notify_type: 'friend.apply', biz_id: `friend:${i + 1}:apply` }));
    await expect(markBusinessNotificationsRead('token', targets)).resolves.toEqual({ affected: 205, unreadCount: 3 });
    expect(sizes).toEqual([100, 100, 5]);
  });
});
