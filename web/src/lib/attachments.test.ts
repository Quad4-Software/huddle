import { describe, expect, it } from 'vitest';
import { isGifMime, isImageMime, isPreviewable, isVideoMime } from './attachments';

describe('attachments', () => {
  it('detects image and video mime types', () => {
    expect(isImageMime('image/png')).toBe(true);
    expect(isImageMime('image/gif')).toBe(true);
    expect(isVideoMime('video/mp4')).toBe(true);
    expect(isImageMime('application/pdf')).toBe(false);
    expect(isVideoMime('image/png')).toBe(false);
  });

  it('detects gifs by mime or extension', () => {
    expect(isGifMime('image/gif', 'wave.gif')).toBe(true);
    expect(isGifMime('application/octet-stream', 'wave.GIF')).toBe(true);
    expect(isGifMime('image/png', 'wave.png')).toBe(false);
  });

  it('groups previewable media', () => {
    expect(isPreviewable('image/jpeg', 'photo.jpg')).toBe(true);
    expect(isPreviewable('video/webm', 'clip.webm')).toBe(true);
    expect(isPreviewable('application/pdf', 'doc.pdf')).toBe(false);
  });
});
