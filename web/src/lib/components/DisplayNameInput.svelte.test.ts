import { fireEvent, render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import DisplayNameInput from './DisplayNameInput.svelte';
import { settings } from '../stores/settings.svelte';
import { STORAGE_KEY } from '../settings-storage';

describe('DisplayNameInput', () => {
  it('persists name to localStorage as the user types', async () => {
    localStorage.removeItem(STORAGE_KEY);
    settings.displayName = '';

    render(DisplayNameInput, { value: '' });

    await fireEvent.input(screen.getByPlaceholderText('Alex'), {
      target: { value: 'Ada' },
    });

    expect(settings.displayName).toBe('Ada');
    expect(JSON.parse(localStorage.getItem(STORAGE_KEY) ?? '{}').displayName).toBe('Ada');
  });

  it('fills a random name from the icon button', async () => {
    localStorage.removeItem(STORAGE_KEY);
    settings.displayName = '';

    render(DisplayNameInput, { value: '' });

    await fireEvent.click(screen.getByRole('button', { name: 'Random name' }));

    expect(settings.displayName.length).toBeGreaterThan(0);
    expect(settings.displayName.split(' ')).toHaveLength(2);
  });
});
