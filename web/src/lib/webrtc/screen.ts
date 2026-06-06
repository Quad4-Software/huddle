export async function getScreenStream(): Promise<MediaStream> {
  return navigator.mediaDevices.getDisplayMedia({
    video: { frameRate: 30 },
    audio: true,
  });
}

export function stopScreenStream(stream: MediaStream | null) {
  stream?.getTracks().forEach((t) => t.stop());
}
