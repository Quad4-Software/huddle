import { describe, expect, it } from 'vitest';
import { pageDescription, pageTitle, robotsDirective } from './seo';

describe('seo', () => {
  it('builds landing title and description', () => {
    expect(pageTitle('landing')).toBe('Huddle');
    expect(pageDescription('landing')).toContain('encrypted chat');
  });

  it('builds room-specific title and description', () => {
    expect(pageTitle('room', 'War Room')).toBe('War Room | Huddle');
    expect(pageDescription('room', 'War Room')).toContain('War Room');
  });

  it('noindexes invite room routes', () => {
    expect(robotsDirective('/r/abc123')).toBe('noindex, nofollow');
    expect(robotsDirective('/')).toBe('index, follow');
  });
});
