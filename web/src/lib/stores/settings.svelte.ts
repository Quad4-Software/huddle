import {
  defaultSettings,
  loadSettings,
  saveSettings,
  type InputMode,
  type Settings,
} from '../settings-storage';

class SettingsStore {
  inputDeviceId = $state(loadSettings().inputDeviceId);
  outputDeviceId = $state(loadSettings().outputDeviceId);
  displayName = $state(loadSettings().displayName);
  inputMode = $state<InputMode>(loadSettings().inputMode);
  voiceActivationThreshold = $state(loadSettings().voiceActivationThreshold);
  inputVolume = $state(loadSettings().inputVolume);
  outputVolume = $state(loadSettings().outputVolume);
  pushToTalkKey = $state(loadSettings().pushToTalkKey);
  toggleMuteKey = $state(loadSettings().toggleMuteKey);
  toggleDeafenKey = $state(loadSettings().toggleDeafenKey);

  private snapshot(): Settings {
    return {
      inputDeviceId: this.inputDeviceId,
      outputDeviceId: this.outputDeviceId,
      displayName: this.displayName,
      inputMode: this.inputMode,
      voiceActivationThreshold: this.voiceActivationThreshold,
      inputVolume: this.inputVolume,
      outputVolume: this.outputVolume,
      pushToTalkKey: this.pushToTalkKey,
      toggleMuteKey: this.toggleMuteKey,
      toggleDeafenKey: this.toggleDeafenKey,
    };
  }

  persist() {
    saveSettings(this.snapshot());
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

  setInputMode(mode: InputMode) {
    this.inputMode = mode;
    this.persist();
  }

  setVoiceActivationThreshold(value: number) {
    this.voiceActivationThreshold = value;
    this.persist();
  }

  setInputVolume(value: number) {
    this.inputVolume = value;
    this.persist();
  }

  setOutputVolume(value: number) {
    this.outputVolume = value;
    this.persist();
  }

  setPushToTalkKey(code: string) {
    this.pushToTalkKey = code;
    this.persist();
  }

  setToggleMuteKey(code: string) {
    this.toggleMuteKey = code;
    this.persist();
  }

  setToggleDeafenKey(code: string) {
    this.toggleDeafenKey = code;
    this.persist();
  }

  reset() {
    const d = { ...defaultSettings, displayName: this.displayName };
    this.inputDeviceId = d.inputDeviceId;
    this.outputDeviceId = d.outputDeviceId;
    this.inputMode = d.inputMode;
    this.voiceActivationThreshold = d.voiceActivationThreshold;
    this.inputVolume = d.inputVolume;
    this.outputVolume = d.outputVolume;
    this.pushToTalkKey = d.pushToTalkKey;
    this.toggleMuteKey = d.toggleMuteKey;
    this.toggleDeafenKey = d.toggleDeafenKey;
    this.persist();
  }
}

export const settings = new SettingsStore();

export function outputGain(deafened: boolean): number {
  if (deafened) return 0;
  return settings.outputVolume / 100;
}

export function voiceActivationThreshold(): number {
  return Math.round(5 + (100 - settings.voiceActivationThreshold) * 0.45);
}

export function voiceActivationLevelThreshold(): number {
  return Math.min(1, voiceActivationThreshold() / 72);
}
