import { getMicStream, setOutputDevice, measureAudioLevel } from './audio';
import { localMicLevel } from '../stores/local-mic-level.svelte';

export class MicPreview {
  private stream: MediaStream | null = null;
  private ctx: AudioContext | null = null;
  private audio: HTMLAudioElement | null = null;
  private analyser: AnalyserNode | null = null;
  private raf = 0;
  private smoothed = 0;

  get active() {
    return this.stream !== null;
  }

  async start(deviceId: string, inputVolume: number, outputDeviceId: string) {
    await this.stop();
    this.stream = await getMicStream({ deviceId: deviceId || undefined });
    this.ctx = new AudioContext();
    const source = this.ctx.createMediaStreamSource(this.stream);
    this.analyser = this.ctx.createAnalyser();
    this.analyser.fftSize = 256;
    this.analyser.smoothingTimeConstant = 0.65;
    const gain = this.ctx.createGain();
    gain.gain.value = inputVolume / 100;
    const dest = this.ctx.createMediaStreamDestination();
    source.connect(this.analyser);
    source.connect(gain);
    gain.connect(dest);

    this.audio = new Audio();
    this.audio.srcObject = dest.stream;
    this.audio.autoplay = true;
    if (outputDeviceId) {
      await setOutputDevice(this.audio, outputDeviceId).catch(() => {});
    }
    await this.audio.play();
    this.startLevelLoop();
  }

  async restart(deviceId: string, inputVolume: number, outputDeviceId: string) {
    if (!this.active) return;
    await this.start(deviceId, inputVolume, outputDeviceId);
  }

  async stop() {
    this.stopLevelLoop();
    this.stream?.getTracks().forEach((t) => t.stop());
    this.stream = null;
    this.analyser = null;
    if (this.audio) {
      this.audio.srcObject = null;
      this.audio = null;
    }
    if (this.ctx) {
      await this.ctx.close().catch(() => {});
      this.ctx = null;
    }
    localMicLevel.reset();
  }

  private startLevelLoop() {
    if (!this.analyser) return;
    const analyser = this.analyser;
    const tick = () => {
      if (!this.analyser) return;
      const raw = measureAudioLevel(analyser);
      this.smoothed = raw > this.smoothed ? raw : this.smoothed * 0.82;
      localMicLevel.set(this.smoothed);
      this.raf = requestAnimationFrame(tick);
    };
    this.raf = requestAnimationFrame(tick);
  }

  private stopLevelLoop() {
    if (this.raf) cancelAnimationFrame(this.raf);
    this.raf = 0;
    this.smoothed = 0;
  }
}

export const micPreview = new MicPreview();
