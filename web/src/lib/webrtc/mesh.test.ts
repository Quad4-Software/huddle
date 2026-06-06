import { describe, expect, it, vi } from 'vitest';
import { importRoomKey } from '../crypto/e2e';
import type { AttachmentMeta } from '../types';
import { Mesh } from './mesh';

const roomKey = 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA';

describe('Mesh broadcastFile', () => {
  it('passes attachment meta to onAttachment for the sender', async () => {
    const key = await importRoomKey(roomKey);
    let attachedMeta: AttachmentMeta | null = null;

    const mesh = new Mesh('peer-a', 'Alice', key, vi.fn(), {
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
});
