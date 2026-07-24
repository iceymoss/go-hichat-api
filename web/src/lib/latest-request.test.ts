import { describe, expect, test } from 'bun:test';
import { createLatestRequest } from './latest-request';

describe('createLatestRequest', () => {
  test('only the newest generation remains current', () => {
    const requests = createLatestRequest();
    const first = requests.begin();
    const second = requests.begin();
    expect(first()).toBe(false);
    expect(second()).toBe(true);
  });

  test('invalidate rejects the active generation', () => {
    const requests = createLatestRequest();
    const active = requests.begin();
    requests.invalidate();
    expect(active()).toBe(false);
  });
});
