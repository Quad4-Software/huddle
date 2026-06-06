export const STORAGE_KEY = 'huddle-settings';

export interface Settings {
  inputDeviceId: string;
  outputDeviceId: string;
  displayName: string;
}

export const defaultSettings: Settings = {
  inputDeviceId: '',
  outputDeviceId: '',
  displayName: '',
};

export function loadSettings(): Settings {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) return { ...defaultSettings, ...JSON.parse(raw) };
  } catch {}
  return { ...defaultSettings };
}

export function saveSettings(s: Settings): void {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(s));
}
