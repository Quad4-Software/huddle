export type LoadingPhase = 'connecting' | 'pow' | 'creating' | 'joining';

class LoadingStore {
  active = $state(false);
  phase = $state<LoadingPhase>('connecting');
  progress = $state(0);
  detail = $state('');
  private powRangeStart = 0;
  private powRangeEnd = 0;

  start(phase: LoadingPhase, detail = '') {
    this.active = true;
    this.phase = phase;
    this.progress = 0;
    this.detail = detail;
    this.powRangeStart = 0;
    this.powRangeEnd = 0;
  }

  setPhase(phase: LoadingPhase, detail?: string) {
    this.phase = phase;
    if (detail !== undefined) this.detail = detail;
  }

  advanceTo(progress: number) {
    const next = Math.min(100, Math.max(0, progress));
    if (next > this.progress) this.progress = next;
  }

  beginPow(rangeStart: number, rangeEnd: number) {
    this.powRangeStart = rangeStart;
    this.powRangeEnd = rangeEnd;
    this.setPhase('pow');
    this.advanceTo(rangeStart);
  }

  setPowProgress(powPercent: number) {
    const t = Math.min(100, Math.max(0, powPercent)) / 100;
    this.advanceTo(this.powRangeStart + t * (this.powRangeEnd - this.powRangeStart));
  }

  stop() {
    this.active = false;
    this.progress = 0;
    this.detail = '';
    this.powRangeStart = 0;
    this.powRangeEnd = 0;
  }
}

export const loading = new LoadingStore();
