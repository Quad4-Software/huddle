export const SITE_NAME = 'Huddle';

export const SITE_TAGLINE = 'Self-hosted voice, video, and chat for small groups';

export const SITE_DESCRIPTION =
  'Create a room, share a link, talk. Voice, screen sharing, and end-to-end encrypted chat for small teams.';

export const SITE_KEYWORDS =
  'voice chat, video chat, webrtc, self-hosted, encrypted chat, screen sharing, huddle';

export type SeoView = 'landing' | 'room' | 'settings';

export function pageTitle(view: SeoView, roomName?: string): string {
  if (view === 'settings') return `Settings | ${SITE_NAME}`;
  if (view === 'room' && roomName) return `${roomName} | ${SITE_NAME}`;
  return SITE_NAME;
}

export function pageDescription(view: SeoView, roomName?: string): string {
  if (view === 'settings') {
    return `Audio, display, and identity settings for ${SITE_NAME}.`;
  }
  if (view === 'room' && roomName) {
    return `Join ${roomName} on ${SITE_NAME}. Voice, screen sharing, and encrypted chat with your group.`;
  }
  return SITE_DESCRIPTION;
}

export function robotsDirective(pathname: string): string {
  if (/^\/r\/[^/]+/.test(pathname)) {
    return 'noindex, nofollow';
  }
  return 'index, follow';
}

export function canonicalUrl(pathname: string, search = ''): string {
  const path = pathname || '/';
  const query = search.startsWith('?') ? search : search ? `?${search}` : '';
  return `${location.origin}${path}${query}`;
}
