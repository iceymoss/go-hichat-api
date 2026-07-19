import { afterEach, describe, expect, test } from 'bun:test';
import { sendGroupRequest } from './friend-group-api';

const originalFetch = globalThis.fetch;
afterEach(() => { globalThis.fetch = originalFetch; });

describe('group join API contract', () => {
  test('omits join_source and returns structured server state', async () => {
    let requestBody = '';
    globalThis.fetch = (async (_input, init) => {
      requestBody = String(init?.body);
      return new Response(JSON.stringify({
        success: true,
        data: {
          group_id: '9007199254740993', is_pass: 1, request_id: '9007199254740995',
          already_pending: false, already_member: true,
        },
      }), { status: 200 });
    }) as typeof fetch;

    await expect(sendGroupRequest('token', '9007199254740993', 'hello')).resolves.toEqual({
      groupId: '9007199254740993', isPass: true, requestId: '9007199254740995',
      alreadyPending: false, alreadyMember: true,
    });
    expect(JSON.parse(requestBody)).toEqual({ group_id: '9007199254740993', req_msg: 'hello' });
  });
});
