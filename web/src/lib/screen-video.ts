export function bindScreenVideo(el: HTMLVideoElement, stream: MediaStream) {
  let detach = attachStream(el, stream);

  return {
    update(next: MediaStream) {
      detach();
      detach = attachStream(el, next);
    },
    destroy() {
      detach();
      el.removeAttribute('src');
      el.srcObject = null;
    },
  };
}

function attachStream(el: HTMLVideoElement, stream: MediaStream) {
  el.srcObject = stream;
  playWhenReady(el);

  const cleanups: Array<() => void> = [];

  const retry = () => playWhenReady(el);
  for (const track of stream.getVideoTracks()) {
    track.addEventListener('unmute', retry);
    track.addEventListener('mute', retry);
    cleanups.push(() => {
      track.removeEventListener('unmute', retry);
      track.removeEventListener('mute', retry);
    });
  }

  stream.addEventListener('addtrack', retry);
  cleanups.push(() => stream.removeEventListener('addtrack', retry));

  return () => cleanups.forEach((fn) => fn());
}

function playWhenReady(el: HTMLVideoElement) {
  const tryPlay = () => {
    void el.play().catch(() => {});
  };
  if (el.readyState >= HTMLMediaElement.HAVE_METADATA) {
    tryPlay();
    return;
  }
  el.addEventListener('loadedmetadata', tryPlay, { once: true });
  el.addEventListener('canplay', tryPlay, { once: true });
}

export function hasScreenAudio(stream: MediaStream): boolean {
  return stream.getAudioTracks().some((t) => t.readyState === 'live');
}
