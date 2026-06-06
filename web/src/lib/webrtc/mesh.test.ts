import { describe, expect, it, vi } from 'vitest';
import { importRoomKey, importSigningKey } from '../crypto/e2e';
import type { AttachmentMeta } from '../types';
import { Mesh } from './mesh';

const roomKey = 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA';

describe('Mesh broadcastFile', () => {
  it('passes attachment meta to onAttachment for the sender', async () => {
    const key = await importRoomKey(roomKey);
    const signingKey = await importSigningKey(roomKey);
    let attachedMeta: AttachmentMeta | null = null;

    const mesh = new Mesh('peer-a', 'Alice', key, signingKey, [], vi.fn(), {
      onMessage: vi.fn(),
      onAttachment: (meta, blob) => {
        attachedMeta = meta;
        expect(blob.type).toBe('image/gif');
      },
      onControl: vi.fn(),
      onTrack: vi.fn(),
      onTrackRemoved: vi.fn(),
      onPeerConnected: vi.fn(),
      onMeshReady: vi.fn(),
    });

    const file = new File([new Uint8Array([1, 2, 3])], '1.gif', { type: 'image/gif' });
    await mesh.broadcastFile('general', file);

    expect(attachedMeta?.id).toBeTruthy();
    expect(attachedMeta?.name).toBe('1.gif');
  });

  it('infers image mime from filename when the browser omits file.type', async () => {
    const key = await importRoomKey(roomKey);
    const signingKey = await importSigningKey(roomKey);
    let attachedMeta: AttachmentMeta | null = null;

    const mesh = new Mesh('peer-a', 'Alice', key, signingKey, [], vi.fn(), {
      onMessage: vi.fn(),
      onAttachment: (meta, blob) => {
        attachedMeta = meta;
        expect(blob.type).toBe('image/webp');
      },
      onControl: vi.fn(),
      onTrack: vi.fn(),
      onTrackRemoved: vi.fn(),
      onPeerConnected: vi.fn(),
      onMeshReady: vi.fn(),
    });

    const file = new File([new Uint8Array([1, 2, 3])], 'photo.webp', { type: '' });
    await mesh.broadcastFile('general', file);

    expect(attachedMeta?.mime).toBe('image/webp');
  });
});
