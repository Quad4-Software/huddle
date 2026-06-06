import { getMicStream, setOutputDevice } from './audio';

export class MicPreview {
  private stream: MediaStream | null = null;
  private ctx: AudioContext | null = null;
  private audio: HTMLAudioElement | null = null;

  get active() {
    return this.stream !== null;
  }

  async start(deviceId: string, inputVolume: number, outputDeviceId: string) {
    await this.stop();
    this.stream = await getMicStream(deviceId || undefined);
    this.ctx = new AudioContext();
    const source = this.ctx.createMediaStreamSource(this.stream);
    const gain = this.ctx.createGain();
    gain.gain.value = inputVolume / 100;
    const dest = this.ctx.createMediaStreamDestination();
    source.connect(gain);
    gain.connect(dest);

    this.audio = new Audio();
    this.audio.srcObject = dest.stream;
    this.audio.autoplay = true;
    if (outputDeviceId) {
      await setOutputDevice(this.audio, outputDeviceId).catch(() => {});
    }
    await this.audio.play();
  }

  async restart(deviceId: string, inputVolume: number, outputDeviceId: string) {
    if (!this.active) return;
    await this.start(deviceId, inputVolume, outputDeviceId);
  }

  async stop() {
    this.stream?.getTracks().forEach((t) => t.stop());
    this.stream = null;
    if (this.audio) {
      this.audio.srcObject = null;
      this.audio = null;
    }
    if (this.ctx) {
      await this.ctx.close().catch(() => {});
      this.ctx = null;
    }
  }
}

export const micPreview = new MicPreview();
