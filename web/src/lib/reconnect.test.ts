import { describe, expect, it } from 'vitest';
import { reconnectDelayMs } from './reconnect';

describe('reconnectDelayMs', () => {
  it('uses exponential backoff capped at the max delay', () => {
    expect(reconnectDelayMs(1)).toBe(1000);
    expect(reconnectDelayMs(2)).toBe(2000);
    expect(reconnectDelayMs(3)).toBe(4000);
    expect(reconnectDelayMs(10)).toBe(15000);
  });
});
