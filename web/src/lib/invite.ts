export function buildInviteUrl(roomId: string, invite: string, roomKey: string): string {
  return `/r/${roomId}?t=${encodeURIComponent(invite)}#key=${roomKey}`;
}

export function buildFullInviteUrl(
  origin: string,
  roomId: string,
  invite: string,
  roomKey: string,
): string {
  return `${origin}${buildInviteUrl(roomId, invite, roomKey)}`;
}

export function parseInviteLocation(pathname: string, search: string, hash: string) {
  const roomId = pathname.match(/^\/r\/([^/]+)/)?.[1] ?? null;
  const invite = new URLSearchParams(search).get('t');
  const roomKey = hash.match(/key=([^&]+)/)?.[1] ?? null;
  return { roomId, invite, roomKey };
}
