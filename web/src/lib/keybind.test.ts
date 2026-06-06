import { describe, expect, it } from 'vitest';
import { formatKeyCode, isTypingTarget } from './keybind';

describe('keybind helpers', () => {
  it('detects typing targets', () => {
    const input = document.createElement('input');
    expect(isTypingTarget(input)).toBe(true);
    expect(isTypingTarget(document.createElement('div'))).toBe(false);
  });

  it('formats digit and letter codes', () => {
    expect(formatKeyCode('Digit5')).toBe('5');
    expect(formatKeyCode('KeyA')).toBe('A');
  });
});
