import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { decodeFrame, Signaling } from './signaling';
import { encodeFrame, encodeMessage, Msg } from '../wire/codec';

class MockWebSocket {
  static OPEN = 1;
  readyState = 0;
  binaryType = 'arraybuffer';
  sent: ArrayBuffer[] = [];
  onopen: (() => void) | null = null;
  onerror: (() => void) | null = null;
  onmessage: ((ev: { data: ArrayBuffer }) => void) | null = null;
  onclose: (() => void) | null = null;

  constructor(public url: string) {
    queueMicrotask(() => {
      this.readyState = MockWebSocket.OPEN;
      this.onopen?.();
    });
  }

  send(data: ArrayBuffer) {
    this.sent.push(data);
  }

  close() {
    this.readyState = 3;
    this.onclose?.();
  }

  simulateMessage(type: string, payload: unknown) {
    let frame: Uint8Array;
    if (type === 'joined') {
      const p = payload as { peerId: string };
      const body = encodeJoinedTestPayload(p.peerId);
      frame = encodeFrame(Msg.JOINED, body);
    } else {
      frame = encodeMessage(type, payload);
    }
    this.onmessage?.({
      data: frame.buffer.slice(frame.byteOffset, frame.byteOffset + frame.byteLength),
    });
  }
}

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

describe('Signaling', () => {
  let instances: MockWebSocket[];

  beforeEach(() => {
    instances = [];
    class WebSocketMock extends MockWebSocket {
      static OPEN = MockWebSocket.OPEN;
      constructor(url: string) {
        super(url);
        instances.push(this);
      }
    }
    vi.stubGlobal('WebSocket', WebSocketMock);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('connects to the provided websocket url', async () => {
    const signaling = new Signaling('ws://localhost:8080/ws');
    await signaling.connect();
    expect(instances[0]?.url).toBe('ws://localhost:8080/ws');
  });

  it('dispatches typed handlers and supports once', async () => {
    const signaling = new Signaling('ws://localhost:8080/ws');
    await signaling.connect();
    const ws = instances[0]!;

    const seen: string[] = [];
    signaling.on('joined', (payload) => {
      seen.push((payload as { peerId: string }).peerId);
    });
    signaling.once('joined', () => {
      seen.push('once');
    });

    ws.simulateMessage('joined', { peerId: 'peer-a' });
    ws.simulateMessage('joined', { peerId: 'peer-b' });

    await Promise.resolve();
    expect(seen).toEqual(['peer-a', 'once', 'peer-b']);
  });

  it('serializes outbound messages as binary frames', async () => {
    const signaling = new Signaling('ws://localhost:8080/ws');
    await signaling.connect();
    await Promise.resolve();
    signaling.send('join', { roomId: 'room1', name: 'Ada' });
    const sent = instances[0]?.sent[0];
    expect(sent).toBeTruthy();
    const frame = decodeFrame(sent!);
    expect(frame.type).toBe(Msg.JOIN);
    expect(frame.payload.length).toBeGreaterThan(0);
  });
});
