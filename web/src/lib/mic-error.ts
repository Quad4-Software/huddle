export function micErrorMessage(err: unknown): string {
  if (err instanceof DOMException) {
    if (err.name === 'NotAllowedError') return 'Microphone access denied';
    if (err.name === 'NotFoundError') return 'No microphone found';
    if (err.name === 'NotReadableError') return 'Microphone is in use by another app';
    if (err.name === 'OverconstrainedError') return 'Selected microphone is unavailable';
  }
  return 'Could not access microphone';
}
