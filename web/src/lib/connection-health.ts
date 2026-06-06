export type ConnectionQuality = 'good' | 'fair' | 'poor' | 'unknown';

export function computeJitter(samples: number[]): number | null {
  if (samples.length < 2) return null;
  let sum = 0;
  for (let i = 1; i < samples.length; i++) {
    sum += Math.abs(samples[i]! - samples[i - 1]!);
  }
  return Math.round(sum / (samples.length - 1));
}

export function computeQuality(
  connected: boolean,
  meshReady: boolean,
  peerCount: number,
  ping: number | null,
  jitter: number | null,
  iceState: string,
): ConnectionQuality {
  if (!connected) return 'poor';
  if (iceState === 'failed' || iceState === 'disconnected') return 'poor';
  if (ping !== null && ping > 200) return 'poor';
  if (jitter !== null && jitter > 80) return 'poor';
  if (peerCount > 0 && !meshReady) return 'fair';
  if (ping !== null && ping > 80) return 'fair';
  if (jitter !== null && jitter > 40) return 'fair';
  if (iceState === 'checking' || iceState === 'new') return 'fair';
  if (connected && (meshReady || peerCount === 0)) return 'good';
  return 'unknown';
}

export function qualityColor(quality: ConnectionQuality): string {
  if (quality === 'good') return 'bg-online';
  if (quality === 'fair') return 'bg-away';
  if (quality === 'poor') return 'bg-danger';
  return 'bg-offline';
}

export function qualityLabel(quality: ConnectionQuality): string {
  if (quality === 'good') return 'Good connection';
  if (quality === 'fair') return 'Fair connection';
  if (quality === 'poor') return 'Poor connection';
  return 'Connection unknown';
}

export function iceStateLabel(state: string): string {
  if (state === 'connected' || state === 'completed') return 'ICE connected';
  if (state === 'checking' || state === 'new') return 'ICE connecting';
  if (state === 'disconnected') return 'ICE disconnected';
  if (state === 'failed') return 'ICE failed';
  return '';
}
