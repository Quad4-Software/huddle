import { decodeMessage, encodeMessage, encodeFrame, decodeFrame, Msg } from '../wire/codec';

type Handler = (payload: unknown) => void;

export class Signaling {
  private ws: WebSocket | null = null;
  private handlers = new Map<string, Handler[]>();
  private url: string;

  constructor(url?: string) {
    const proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
    this.url = url ?? `${proto}//${location.host}/ws`;
  }

  isOpen(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  disconnect() {
    this.ws?.close();
    this.ws = null;
  }

  connect(): Promise<void> {
    if (this.isOpen()) {
      return Promise.resolve();
    }
    this.disconnect();
    return new Promise((resolve, reject) => {
      this.ws = new WebSocket(this.url);
      this.ws.binaryType = 'arraybuffer';
      this.ws.onopen = () => resolve();
      this.ws.onerror = () => reject(new Error('connection failed'));
      this.ws.onmessage = (ev) => {
        void (async () => {
          try {
            const data = await toArrayBuffer(ev.data);
            const msg = decodeMessage(data);
            for (const fn of this.handlers.get(msg.type) ?? []) {
              fn(msg.payload);
            }
          } catch {}
        })();
      };
      this.ws.onclose = () => {
        for (const fn of this.handlers.get('close') ?? []) {
          fn(null);
        }
      };
    });
  }

  on(type: string, fn: Handler) {
    const list = this.handlers.get(type) ?? [];
    list.push(fn);
    this.handlers.set(type, list);
  }

  once(type: string, fn: Handler) {
    const wrapper: Handler = (payload) => {
      this.off(type, wrapper);
      fn(payload);
    };
    this.on(type, wrapper);
  }

  off(type: string, fn: Handler) {
    const list = this.handlers.get(type) ?? [];
    this.handlers.set(
      type,
      list.filter((h) => h !== fn),
    );
  }

  send(type: string, payload?: unknown) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(encodeMessage(type, payload));
    }
  }

  close() {
    this.disconnect();
    this.handlers.clear();
  }
}

async function toArrayBuffer(data: unknown): Promise<ArrayBuffer> {
  if (data instanceof ArrayBuffer) return data;
  if (data instanceof Blob) return data.arrayBuffer();
  throw new Error('unsupported websocket payload');
}

export { decodeMessage, encodeMessage, encodeFrame, decodeFrame, Msg };
