import { describe, expect, it } from 'vitest';
import { memberStatus, sortMembers } from './members';
import type { Member } from './types';

const sample: Member[] = [
  { id: 'b', name: 'Zoe', muted: false, deafened: false, speaking: false },
  { id: 'a', name: 'Ada', muted: false, deafened: false, speaking: true },
  { id: 'c', name: 'Ada', muted: false, deafened: false, speaking: false },
];

describe('members', () => {
  it('sorts members by name then id', () => {
    const sorted = sortMembers(sample);
    expect(sorted.map((m) => m.id)).toEqual(['a', 'c', 'b']);
  });

  it('reports connecting before online', () => {
    expect(memberStatus(sample[0], false, true)).toBe('connecting');
    expect(memberStatus(sample[0], true, false)).toBe('online');
    expect(memberStatus({ ...sample[0], speaking: true }, true, false)).toBe('speaking');
  });
});
