import { describe, expect, it } from 'vitest';
import { computeJitter, computeQuality } from './connection-health';

describe('connection health', () => {
  it('computes jitter from ping samples', () => {
    expect(computeJitter([])).toBeNull();
    expect(computeJitter([40])).toBeNull();
    expect(computeJitter([40, 60, 50])).toBe(15);
  });

  it('rates connection quality from ping, jitter, and ice', () => {
    expect(computeQuality(true, true, 1, 30, 10, 'connected')).toBe('good');
    expect(computeQuality(true, false, 1, 120, 20, 'checking')).toBe('fair');
    expect(computeQuality(false, false, 1, 30, 10, 'connected')).toBe('poor');
    expect(computeQuality(true, true, 1, 250, 10, 'connected')).toBe('poor');
    expect(computeQuality(true, true, 1, 30, 10, 'failed')).toBe('poor');
  });
});
