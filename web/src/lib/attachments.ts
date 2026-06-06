export function isImageMime(mime: string): boolean {
  return mime.startsWith('image/');
}

export function isVideoMime(mime: string): boolean {
  return mime.startsWith('video/');
}

export function isGifMime(mime: string, name: string): boolean {
  return mime === 'image/gif' || name.toLowerCase().endsWith('.gif');
}

export function isPreviewable(mime: string, name: string): boolean {
  return isImageMime(mime) || isVideoMime(mime) || isGifMime(mime, name);
}
