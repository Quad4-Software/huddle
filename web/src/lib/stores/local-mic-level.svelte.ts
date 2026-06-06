class LocalMicLevelStore {
  level = $state(0);

  set(value: number) {
    this.level = value;
  }

  reset() {
    this.level = 0;
  }
}

export const localMicLevel = new LocalMicLevelStore();
