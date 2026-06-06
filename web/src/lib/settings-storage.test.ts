import { describe, expect, it } from 'vitest';
import { STORAGE_KEY, defaultSettings, loadSettings, saveSettings } from './settings-storage';

describe('settings storage', () => {
  it('returns defaults when storage is empty', () => {
    expect(loadSettings()).toEqual(defaultSettings);
  });

  it('persists and reloads settings', () => {
    saveSettings({
      inputDeviceId: 'mic-1',
      outputDeviceId: 'spk-1',
      displayName: 'Ada',
    });
    expect(loadSettings()).toEqual({
      inputDeviceId: 'mic-1',
      outputDeviceId: 'spk-1',
      displayName: 'Ada',
    });
  });

  it('falls back to defaults for corrupt storage', () => {
    localStorage.setItem(STORAGE_KEY, '{not-json');
    expect(loadSettings()).toEqual(defaultSettings);
  });
});
