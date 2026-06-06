const EXT_MIME: Record<string, string> = {
  avif: 'image/avif',
  bmp: 'image/bmp',
  gif: 'image/gif',
  heic: 'image/heic',
  heif: 'image/heif',
  jfif: 'image/jpeg',
  jpeg: 'image/jpeg',
  jpg: 'image/jpeg',
  png: 'image/png',
  svg: 'image/svg+xml',
  webp: 'image/webp',
  mp4: 'video/mp4',
  webm: 'video/webm',
  mov: 'video/quicktime',
  mkv: 'video/x-matroska',
};

export function mimeFromFilename(name: string): string {
  const ext = name.split('.').pop()?.toLowerCase() ?? '';
  return EXT_MIME[ext] ?? '';
}

export function resolveAttachmentMime(type: string, name: string): string {
  const normalized = type.trim().toLowerCase();
  if (normalized && normalized !== 'application/octet-stream') return normalized;
  return mimeFromFilename(name) || normalized || 'application/octet-stream';
}

export function isImageMime(mime: string): boolean {
  return mime.startsWith('image/');
}

export function isVideoMime(mime: string): boolean {
  return mime.startsWith('video/');
}

export function isGifMime(mime: string, name: string): boolean {
  const resolved = resolveAttachmentMime(mime, name);
  return resolved === 'image/gif' || name.toLowerCase().endsWith('.gif');
}

export function isPreviewable(mime: string, name: string): boolean {
  const resolved = resolveAttachmentMime(mime, name);
  return isImageMime(resolved) || isVideoMime(resolved);
}
