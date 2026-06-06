import { measureAudioLevel, getMicStream } from './audio';
import { createMicPipeline } from './mic-pipeline';

export class SettingsMicSampler {
  private stream: MediaStream | null = null;
  private pipeline: ReturnType<typeof createMicPipeline>['pipeline'] | null = null;
  private raf = 0;
  private smoothed = 0;

  get active() {
    return this.stream !== null;
  }

  async ensureStarted(deviceId: string, inputVolume: number) {
    if (this.active) return;
    await this.start(deviceId, inputVolume);
  }

  async restart(deviceId: string, inputVolume: number) {
    await this.stop();
    await this.start(deviceId, inputVolume);
  }

  async start(deviceId: string, inputVolume: number) {
    await this.stop();
    this.stream = await getMicStream({ deviceId: deviceId || undefined });
    const processed = createMicPipeline(this.stream);
    this.pipeline = processed.pipeline;
    this.pipeline.setInputVolume(inputVolume);
    processed.analyser.smoothingTimeConstant = 0.65;
    this.loop(processed.analyser);
  }

  async stop() {
    if (this.raf) cancelAnimationFrame(this.raf);
    this.raf = 0;
    this.smoothed = 0;
    this.pipeline?.close();
    this.pipeline = null;
    this.stream?.getTracks().forEach((t) => t.stop());
    this.stream = null;
    localMicLevel.reset();
  }

  private loop(analyser: AnalyserNode) {
    const tick = () => {
      const raw = measureAudioLevel(analyser);
      this.smoothed = raw > this.smoothed ? raw : this.smoothed * 0.82;
      localMicLevel.set(this.smoothed);
      this.raf = requestAnimationFrame(tick);
    };
    this.raf = requestAnimationFrame(tick);
  }
}

export const settingsMicSampler = new SettingsMicSampler();
