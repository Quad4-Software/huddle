import { measureAudioLevel } from './audio';
import { audioLevels } from '../stores/audio-levels.svelte';

type PeerSource = {
  analyser: AnalyserNode;
  source?: MediaStreamAudioSourceNode;
};

export class AudioLevelMonitor {
  private peers = new Map<string, PeerSource>();
  private smoothed = new Map<string, number>();
  private ctx: AudioContext | null = null;
  private raf = 0;

  registerAnalyser(peerId: string, analyser: AnalyserNode) {
    this.unregister(peerId);
    analyser.smoothingTimeConstant = 0.65;
    this.peers.set(peerId, { analyser });
    this.start();
  }

  registerStream(peerId: string, stream: MediaStream) {
    this.unregister(peerId);
    if (!this.ctx) this.ctx = new AudioContext();
    const source = this.ctx.createMediaStreamSource(stream);
    const analyser = this.ctx.createAnalyser();
    analyser.fftSize = 256;
    analyser.smoothingTimeConstant = 0.65;
    source.connect(analyser);
    this.peers.set(peerId, { analyser, source });
    this.start();
  }

  unregister(peerId: string) {
    const peer = this.peers.get(peerId);
    if (!peer) return;
    peer.source?.disconnect();
    this.peers.delete(peerId);
    this.smoothed.delete(peerId);
    if (this.peers.size === 0) this.stop();
  }

  destroy() {
    this.stop();
    for (const peerId of [...this.peers.keys()]) {
      this.unregister(peerId);
    }
    void this.ctx?.close();
    this.ctx = null;
    audioLevels.reset();
  }

  private start() {
    if (this.raf) return;
    this.raf = requestAnimationFrame(this.tick);
  }

  private stop() {
    if (!this.raf) return;
    cancelAnimationFrame(this.raf);
    this.raf = 0;
  }

  private tick = () => {
    const next: Record<string, number> = {};
    for (const [peerId, { analyser }] of this.peers) {
      const raw = measureAudioLevel(analyser);
      const prev = this.smoothed.get(peerId) ?? 0;
      const level = raw > prev ? raw : prev * 0.82;
      this.smoothed.set(peerId, level);
      next[peerId] = level;
    }
    audioLevels.setBatch(next);
    this.raf = this.peers.size > 0 ? requestAnimationFrame(this.tick) : 0;
  };
}
