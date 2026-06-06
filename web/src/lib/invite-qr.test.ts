import { describe, expect, it } from 'vitest';
import { generateInviteQrDataUrl } from './invite-qr';

describe('invite QR', () => {
  it('generates a PNG data URL for an invite link', async () => {
    const url = 'https://huddle.example/r/room123?t=invite.token#key=room-key';
    const dataUrl = await generateInviteQrDataUrl(url);
    expect(dataUrl.startsWith('data:image/png;base64,')).toBe(true);
  });
});
