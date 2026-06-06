import { loadSettings, saveSettings } from '../settings-storage';

class SettingsStore {
  inputDeviceId = $state(loadSettings().inputDeviceId);
  outputDeviceId = $state(loadSettings().outputDeviceId);
  displayName = $state(loadSettings().displayName);

  persist() {
    saveSettings({
      inputDeviceId: this.inputDeviceId,
      outputDeviceId: this.outputDeviceId,
      displayName: this.displayName,
    });
  }

  setInput(id: string) {
    this.inputDeviceId = id;
    this.persist();
  }

  setOutput(id: string) {
    this.outputDeviceId = id;
    this.persist();
  }

  setName(name: string) {
    this.displayName = name;
    this.persist();
  }
}

export const settings = new SettingsStore();
