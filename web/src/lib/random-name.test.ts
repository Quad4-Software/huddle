import { describe, expect, it } from 'vitest';
import { randomDisplayName } from './random-name';
import { MAX_DISPLAY_NAME_LENGTH } from './validation';

describe('randomDisplayName', () => {
  it('returns a two-word name within display limit', () => {
    const name = randomDisplayName();
    expect(name.length).toBeLessThanOrEqual(MAX_DISPLAY_NAME_LENGTH);
    expect(name.split(' ')).toHaveLength(2);
  });
});
