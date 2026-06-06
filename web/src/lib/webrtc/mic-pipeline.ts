export class MicPipeline {
  readonly analyser: AnalyserNode;
  readonly outputStream: MediaStream;
  private ctx: AudioContext;
  private inputGain: GainNode;

  constructor(input: MediaStream) {
    this.ctx = new AudioContext();
    const source = this.ctx.createMediaStreamSource(input);
    this.analyser = this.ctx.createAnalyser();
    this.analyser.fftSize = 256;

    this.inputGain = this.ctx.createGain();

    const dest = this.ctx.createMediaStreamDestination();

    source.connect(this.analyser);
    source.connect(this.inputGain);
    this.inputGain.connect(dest);

    this.outputStream = dest.stream;
  }

  setInputVolume(percent: number) {
    this.inputGain.gain.value = percent / 100;
  }

  close() {
    this.outputStream.getTracks().forEach((t) => t.stop());
    void this.ctx.close();
  }
}

export function createMicPipeline(input: MediaStream) {
  const pipeline = new MicPipeline(input);
  return {
    pipeline,
    analyser: pipeline.analyser,
    stream: pipeline.outputStream,
  };
}
