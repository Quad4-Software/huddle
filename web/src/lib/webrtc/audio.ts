export interface AudioDevices {
  inputs: MediaDeviceInfo[];
  outputs: MediaDeviceInfo[];
}

export async function listAudioDevices(): Promise<AudioDevices> {
  const devices = await navigator.mediaDevices.enumerateDevices();
  return {
    inputs: devices.filter((d) => d.kind === 'audioinput'),
    outputs: devices.filter((d) => d.kind === 'audiooutput'),
  };
}

export async function getMicStream(deviceId?: string): Promise<MediaStream> {
  const constraints: MediaStreamConstraints = {
    audio: {
      deviceId: deviceId ? { exact: deviceId } : undefined,
      echoCancellation: true,
      noiseSuppression: true,
      autoGainControl: true,
      channelCount: 1,
    },
    video: false,
  };
  return navigator.mediaDevices.getUserMedia(constraints);
}

export function setOutputDevice(el: HTMLAudioElement, deviceId: string): Promise<void> {
  const sink = el as HTMLAudioElement & { setSinkId?: (id: string) => Promise<void> };
  if (sink.setSinkId) {
    return sink.setSinkId(deviceId);
  }
  return Promise.resolve();
}

export function createAnalyser(
  ctx: AudioContext,
  stream: MediaStream,
): { analyser: AnalyserNode; source: MediaStreamAudioSourceNode } {
  const source = ctx.createMediaStreamSource(stream);
  const analyser = ctx.createAnalyser();
  analyser.fftSize = 256;
  source.connect(analyser);
  return { analyser, source };
}

export function isSpeaking(analyser: AnalyserNode, threshold = 15): boolean {
  const data = new Uint8Array(analyser.frequencyBinCount);
  analyser.getByteFrequencyData(data);
  let sum = 0;
  for (const v of data) sum += v;
  return sum / data.length > threshold;
}

export function measureAudioLevel(analyser: AnalyserNode): number {
  const data = new Uint8Array(analyser.frequencyBinCount);
  analyser.getByteFrequencyData(data);
  let sum = 0;
  for (const v of data) sum += v;
  const avg = sum / data.length;
  return Math.min(1, avg / 72);
}
