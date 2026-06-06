type Handler = (payload: unknown) => void;

export class Signaling {
  private ws: WebSocket | null = null;
  private handlers = new Map<string, Handler[]>();
  private url: string;

  constructor(url?: string) {
    const proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
    this.url = url ?? `${proto}//${location.host}/ws`;
  }

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      this.ws = new WebSocket(this.url);
      this.ws.onopen = () => resolve();
      this.ws.onerror = () => reject(new Error('connection failed'));
      this.ws.onmessage = (ev) => {
        try {
          const msg = JSON.parse(ev.data as string) as { type: string; payload?: unknown };
          for (const fn of this.handlers.get(msg.type) ?? []) {
            fn(msg.payload);
          }
        } catch {}
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
      this.ws.send(JSON.stringify({ type, payload }));
    }
  }

  close() {
    this.ws?.close();
    this.ws = null;
    this.handlers.clear();
  }
}
