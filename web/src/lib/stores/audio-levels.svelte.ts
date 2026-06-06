class AudioLevelsStore {
  levels = $state<Record<string, number>>({});

  level(peerId: string): number {
    return this.levels[peerId] ?? 0;
  }

  setBatch(next: Record<string, number>) {
    this.levels = next;
  }

  reset() {
    this.levels = {};
  }
}

export const audioLevels = new AudioLevelsStore();
