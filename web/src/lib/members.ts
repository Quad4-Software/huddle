import type { Member } from './types';

export function sortMembers(members: Member[]): Member[] {
  return [...members].sort((a, b) => {
    const name = a.name.localeCompare(b.name);
    if (name !== 0) return name;
    return a.id.localeCompare(b.id);
  });
}

export function memberStatus(member: Member, online: boolean, connecting = false): string {
  if (connecting) return 'connecting';
  if (!online) return 'offline';
  if (member.deafened) return 'deafened';
  if (member.muted) return 'muted';
  if (member.speaking) return 'speaking';
  return 'online';
}
