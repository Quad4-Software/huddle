import { describe, expect, it } from 'vitest';
import { captureKeyFromEvent, formatKeyCode, isModifierKey, isTypingTarget } from './keybind';

describe('keybind helpers', () => {
  it('detects typing targets', () => {
    const input = document.createElement('input');
    expect(isTypingTarget(input)).toBe(true);
    expect(isTypingTarget(document.createElement('div'))).toBe(false);
  });

  it('formats digit and letter codes', () => {
    expect(formatKeyCode('Digit5')).toBe('5');
    expect(formatKeyCode('KeyA')).toBe('A');
    expect(formatKeyCode('')).toBe('Not set');
  });

  it('identifies modifier keys', () => {
    expect(isModifierKey('ShiftLeft')).toBe(true);
    expect(isModifierKey('KeyA')).toBe(false);
  });

  it('captures keys from keyboard events', () => {
    expect(captureKeyFromEvent({ code: 'Escape' } as KeyboardEvent)).toBe('cancel');
    expect(captureKeyFromEvent({ code: 'Backspace' } as KeyboardEvent)).toBe('clear');
    expect(captureKeyFromEvent({ code: 'ShiftLeft' } as KeyboardEvent)).toBe(null);
    expect(captureKeyFromEvent({ code: 'KeyM' } as KeyboardEvent)).toEqual({ code: 'KeyM' });
  });
});
