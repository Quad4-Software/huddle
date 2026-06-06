import { describe, expect, it } from 'vitest';
import { decrypt, decryptText, encrypt, importRoomKey } from './e2e';

const roomKey = 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA';

async function testKey() {
  return importRoomKey(roomKey);
}

describe('e2e crypto', () => {
  it('round-trips text through AES-GCM', async () => {
    const key = await testKey();
    const encoded = await encrypt(key, 'huddle-message');
    const plain = await decryptText(key, encoded);
    expect(plain).toBe('huddle-message');
  });

  it('round-trips binary payloads', async () => {
    const key = await testKey();
    const bytes = new Uint8Array([0, 127, 255, 42]);
    const encoded = await encrypt(key, bytes);
    const decoded = await decrypt(key, encoded);
    expect(Array.from(decoded)).toEqual([0, 127, 255, 42]);
  });

  it('rejects tampered ciphertext', async () => {
    const key = await testKey();
    const encoded = await encrypt(key, 'secret');
    const raw = Uint8Array.from(atob(encoded), (c) => c.charCodeAt(0));
    raw[raw.length - 1] ^= 0xff;
    const tampered = btoa(String.fromCharCode(...raw));
    await expect(decryptText(key, tampered)).rejects.toThrow();
  });

  it('rejects wrong room key', async () => {
    const key = await testKey();
    const other = await importRoomKey('BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB');
    const encoded = await encrypt(key, 'locked');
    await expect(decryptText(other, encoded)).rejects.toThrow();
  });
});
