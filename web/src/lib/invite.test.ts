import { describe, expect, it } from 'vitest';
import { buildInviteUrl, parseInviteLocation } from './invite';

describe('invite urls', () => {
  it('builds encoded invite links with fragment key', () => {
    const url = buildInviteUrl('room123', 'invite.token', 'room-key');
    expect(url).toBe('/r/room123?t=invite.token#key=room-key');
  });

  it('parses invite links from location parts', () => {
    const parsed = parseInviteLocation('/r/room123', '?t=invite.token&pw=secret', '#key=room-key');
    expect(parsed).toEqual({
      roomId: 'room123',
      invite: 'invite.token',
      roomKey: 'room-key',
    });
  });

  it('returns nulls for invalid locations', () => {
    expect(parseInviteLocation('/', '', '')).toEqual({
      roomId: null,
      invite: null,
      roomKey: null,
    });
  });
});
