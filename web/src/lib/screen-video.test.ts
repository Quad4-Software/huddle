import { describe, expect, it } from 'vitest';
import { hasScreenAudio } from './screen-video';

describe('screen-video', () => {
  it('detects live screen audio tracks', () => {
    const stream = {
      getAudioTracks: () => [{ readyState: 'live' }],
    } as MediaStream;
    expect(hasScreenAudio(stream)).toBe(true);
  });

  it('ignores ended screen audio tracks', () => {
    const stream = {
      getAudioTracks: () => [{ readyState: 'ended' }],
    } as MediaStream;
    expect(hasScreenAudio(stream)).toBe(false);
  });
});
