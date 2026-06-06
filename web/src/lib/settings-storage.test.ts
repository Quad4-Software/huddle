import { describe, expect, it } from 'vitest';
import { formatKeyCode } from './keybind';
import { defaultSettings, loadSettings, saveSettings } from './settings-storage';

describe('settings storage', () => {
  it('returns defaults when storage is empty', () => {
    expect(loadSettings()).toEqual(defaultSettings);
  });

  it('persists and reloads settings', () => {
    saveSettings({
      ...defaultSettings,
      inputDeviceId: 'mic-1',
      outputDeviceId: 'spk-1',
      displayName: 'Ada',
      inputMode: 'pushToTalk',
      pushToTalkKey: 'KeyV',
      toggleMuteKey: 'KeyM',
      toggleDeafenKey: 'KeyD',
    });
    expect(loadSettings()).toEqual({
      ...defaultSettings,
      inputDeviceId: 'mic-1',
      outputDeviceId: 'spk-1',
      displayName: 'Ada',
      inputMode: 'pushToTalk',
      pushToTalkKey: 'KeyV',
      toggleMuteKey: 'KeyM',
      toggleDeafenKey: 'KeyD',
    });
  });

  it('falls back to defaults for corrupt storage', () => {
    localStorage.setItem('huddle-settings', '{not-json');
    expect(loadSettings()).toEqual(defaultSettings);
  });
});

describe('formatKeyCode', () => {
  it('formats common key codes', () => {
    expect(formatKeyCode('Space')).toBe('Space');
    expect(formatKeyCode('KeyV')).toBe('V');
  });
});
