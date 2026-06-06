import { fireEvent, render, screen } from '@testing-library/svelte';
import { describe, expect, it, vi } from 'vitest';
import Landing from './Landing.svelte';

vi.mock('../session-controller', () => ({
  createRoom: vi.fn(),
}));

import { createRoom } from '../session-controller';

describe('Landing', () => {
  it('requires a display name before creating', async () => {
    render(Landing, { onSettings: () => {} });
    await fireEvent.click(screen.getByRole('button', { name: 'Create room' }));
    expect(screen.getByText('Enter your name')).toBeInTheDocument();
    expect(createRoom).not.toHaveBeenCalled();
  });

  it('requires a room name before creating', async () => {
    render(Landing, { onSettings: () => {} });
    await fireEvent.input(screen.getByPlaceholderText('Alex'), {
      target: { value: 'Ada' },
    });
    await fireEvent.click(screen.getByRole('button', { name: 'Create room' }));
    expect(screen.getByText('Enter a room name')).toBeInTheDocument();
    expect(createRoom).not.toHaveBeenCalled();
  });

  it('fills a random room name from the icon button', async () => {
    render(Landing, { onSettings: () => {} });
    const input = screen.getByPlaceholderText('Team sync') as HTMLInputElement;
    expect(input.value).toBe('');
    await fireEvent.click(screen.getByRole('button', { name: 'Random room name' }));
    expect(input.value.length).toBeGreaterThan(0);
  });

  it('submits trimmed room details', async () => {
    vi.mocked(createRoom).mockResolvedValueOnce();
    render(Landing, { onSettings: () => {} });

    await fireEvent.input(screen.getByPlaceholderText('Alex'), {
      target: { value: 'Ada' },
    });
    await fireEvent.input(screen.getByPlaceholderText('Team sync'), {
      target: { value: '  Sprint  ' },
    });
    await fireEvent.input(screen.getByPlaceholderText('Leave empty for invite only'), {
      target: { value: 'secret' },
    });
    await fireEvent.click(screen.getByRole('button', { name: 'Create room' }));

    expect(createRoom).toHaveBeenCalledWith('Sprint', 'secret');
  });
});
