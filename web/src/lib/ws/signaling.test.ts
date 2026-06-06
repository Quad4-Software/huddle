import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { Signaling } from './signaling';

class MockWebSocket {
  static OPEN = 1;
  readyState = 0;
  sent: string[] = [];
  onopen: (() => void) | null = null;
  onerror: (() => void) | null = null;
  onmessage: ((ev: { data: string }) => void) | null = null;
  onclose: (() => void) | null = null;

  constructor(public url: string) {
    queueMicrotask(() => {
      this.readyState = MockWebSocket.OPEN;
      this.onopen?.();
    });
  }

  send(data: string) {
    this.sent.push(data);
  }

  close() {
    this.readyState = 3;
    this.onclose?.();
  }

  simulateMessage(payload: unknown) {
    this.onmessage?.({ data: JSON.stringify(payload) });
  }
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

    ws.simulateMessage({ type: 'joined', payload: { peerId: 'peer-a' } });
    ws.simulateMessage({ type: 'joined', payload: { peerId: 'peer-b' } });

    expect(seen).toEqual(['peer-a', 'once', 'peer-b']);
  });

  it('serializes outbound messages', async () => {
    const signaling = new Signaling('ws://localhost:8080/ws');
    await signaling.connect();
    await Promise.resolve();
    signaling.send('join', { roomId: 'room1', name: 'Ada' });
    expect(instances[0]?.sent).toEqual([
      JSON.stringify({ type: 'join', payload: { roomId: 'room1', name: 'Ada' } }),
    ]);
  });
});
