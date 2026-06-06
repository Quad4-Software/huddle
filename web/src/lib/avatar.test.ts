import { describe, expect, it } from 'vitest';
import { avatarColors, avatarGradient, avatarInitials } from './avatar';

describe('avatar', () => {
  it('derives initials from one or two names', () => {
    expect(avatarInitials('Ada')).toBe('AD');
    expect(avatarInitials('Ada Lovelace')).toBe('AL');
    expect(avatarInitials('')).toBe('?');
    expect(avatarInitials('  grace  hopper ')).toBe('GH');
  });

  it('is deterministic for the same name', () => {
    expect(avatarColors('Ada')).toEqual(avatarColors('Ada'));
    expect(avatarGradient('Bob')).toBe(avatarGradient('Bob'));
  });

  it('returns a gradient string', () => {
    expect(avatarGradient('Ada')).toMatch(/^linear-gradient/);
  });
});
