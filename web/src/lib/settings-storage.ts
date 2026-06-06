export type InputMode = 'voiceActivation' | 'pushToTalk';

export const STORAGE_KEY = 'huddle-settings';

export interface Settings {
  inputDeviceId: string;
  outputDeviceId: string;
  displayName: string;
  inputMode: InputMode;
  voiceActivationThreshold: number;
  inputVolume: number;
  outputVolume: number;
  pushToTalkKey: string;
}

export const defaultSettings: Settings = {
  inputDeviceId: '',
  outputDeviceId: '',
  displayName: '',
  inputMode: 'voiceActivation',
  voiceActivationThreshold: 40,
  inputVolume: 100,
  outputVolume: 100,
  pushToTalkKey: 'Space',
};

function clamp(n: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, n));
}

function normalize(raw: Partial<Settings>): Settings {
  return {
    inputDeviceId: typeof raw.inputDeviceId === 'string' ? raw.inputDeviceId : '',
    outputDeviceId: typeof raw.outputDeviceId === 'string' ? raw.outputDeviceId : '',
    displayName: typeof raw.displayName === 'string' ? raw.displayName : '',
    inputMode: raw.inputMode === 'pushToTalk' ? 'pushToTalk' : 'voiceActivation',
    voiceActivationThreshold: clamp(Number(raw.voiceActivationThreshold) || 40, 1, 100),
    inputVolume: clamp(Number(raw.inputVolume) || 100, 0, 200),
    outputVolume: clamp(Number(raw.outputVolume) || 100, 0, 200),
    pushToTalkKey:
      typeof raw.pushToTalkKey === 'string' && raw.pushToTalkKey ? raw.pushToTalkKey : 'Space',
  };
}

export function loadSettings(): Settings {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) return normalize({ ...defaultSettings, ...JSON.parse(raw) });
  } catch {}
  return { ...defaultSettings };
}

export function saveSettings(s: Settings): void {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(normalize(s)));
}
