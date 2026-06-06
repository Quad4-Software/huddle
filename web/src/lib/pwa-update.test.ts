import { describe, expect, it, beforeEach } from 'vitest';
import { applyPendingUpdate, pwaUpdate, setPwaInSession } from './pwa-update.svelte';

describe('pwa session tracking', () => {
  beforeEach(() => {
    sessionStorage.clear();
    pwaUpdate.pending = false;
  });

  it('marks active voice sessions in session storage', () => {
    setPwaInSession(true);
    expect(sessionStorage.getItem('huddle:pwa:in-session')).toBe('1');
    setPwaInSession(false);
    expect(sessionStorage.getItem('huddle:pwa:in-session')).toBeNull();
  });

  it('does not reload when no update is pending', () => {
    pwaUpdate.pending = false;
    applyPendingUpdate();
    expect(pwaUpdate.pending).toBe(false);
  });
});
