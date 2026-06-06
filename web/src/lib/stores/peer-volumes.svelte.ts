const STORAGE_KEY = 'huddle-peer-volumes';

function clamp(n: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, n));
}

function load(): Record<string, number> {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return {};
    const parsed = JSON.parse(raw) as Record<string, unknown>;
    const out: Record<string, number> = {};
    for (const [peerId, value] of Object.entries(parsed)) {
      if (typeof value === 'number' && Number.isFinite(value)) {
        out[peerId] = clamp(Math.round(value), 0, 200);
      }
    }
    return out;
  } catch {
    return {};
  }
}

function save(volumes: Record<string, number>) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(volumes));
}

class PeerVolumeStore {
  volumes = $state<Record<string, number>>(load());

  get(peerId: string): number {
    return this.volumes[peerId] ?? 100;
  }

  gain(peerId: string): number {
    return this.get(peerId) / 100;
  }

  set(peerId: string, value: number) {
    const next = clamp(Math.round(value), 0, 200);
    this.volumes = { ...this.volumes, [peerId]: next };
    save(this.volumes);
  }
}

export const peerVolumes = new PeerVolumeStore();
