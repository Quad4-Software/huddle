import { describe, expect, it } from 'vitest';
import { solveChallenge } from './pow';

describe('solveChallenge', () => {
  it('finds a nonce that satisfies the difficulty', async () => {
    const nonce = await solveChallenge('test-prefix', 10);
    const data = new TextEncoder().encode(`test-prefix:${nonce}`);
    const digest = new Uint8Array(await crypto.subtle.digest('SHA-256', data));
    let bits = 0;
    for (const byte of digest) {
      if (byte !== 0) {
        bits += Math.clz32(byte) - 24;
        break;
      }
      bits += 8;
    }
    expect(bits).toBeGreaterThanOrEqual(10);
  });
});
