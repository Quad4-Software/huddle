import { describe, expect, it } from 'vitest';
import { decodeFrame, decodeMessage, encodeFrame, encodeMessage, Msg } from './codec';

describe('wire codec', () => {
  it('round-trips join frame header', () => {
    const frame = encodeMessage('join', {
      roomId: 'room1',
      invite: 'token',
      name: 'Ada',
    });
    const decoded = decodeFrame(frame);
    expect(decoded.type).toBe(Msg.JOIN);
    expect(decoded.payload.length).toBeGreaterThan(0);
  });

  it('round-trips ping as fixed 8 bytes', () => {
    const body = new Uint8Array(8);
    const view = new DataView(body.buffer);
    view.setBigUint64(0, BigInt(12345));
    const frame = encodeFrame(Msg.PING, body);
    const decoded = decodeFrame(frame);
    expect(decoded.type).toBe(Msg.PING);
    expect(decoded.payload.length).toBe(8);
  });

  it('decodes joined payloads', () => {
    const body = encodeJoinedTestPayload('peer-a');
    const frame = encodeFrame(Msg.JOINED, body);
    const msg = decodeMessage(frame);
    expect(msg.type).toBe('joined');
    expect((msg.payload as { peerId: string }).peerId).toBe('peer-a');
  });
});

function encodeJoinedTestPayload(peerId: string) {
  const enc = new TextEncoder();
  const parts: number[] = [];
  const append = (s: string) => {
    const b = enc.encode(s);
    parts.push(
      (b.length >>> 24) & 255,
      (b.length >>> 16) & 255,
      (b.length >>> 8) & 255,
      b.length & 255,
      ...b,
    );
  };
  append(peerId);
  append('');
  append('room');
  append('room');
  append('');
  parts.push(0, 0, 0, 0, 0, 0);
  return Uint8Array.from(parts);
}
