import { beforeEach, describe, expect, it } from 'vitest';
import { session } from './session.svelte';

describe('session store', () => {
  beforeEach(() => {
    session.reset();
  });

  it('filters messages by active channel and dedupes', () => {
    const msg = {
      id: '1',
      channelId: 'general',
      authorId: 'a',
      authorName: 'Ada',
      text: 'hello',
      timestamp: 1,
    };
    session.addMessage(msg);
    session.addMessage(msg);
    expect(session.messages).toHaveLength(1);
    expect(session.messagesForChannel('general')).toHaveLength(1);
  });

  it('tracks screen shares when live video tracks exist', () => {
    const video = {
      getVideoTracks: () => [{ kind: 'video', readyState: 'live' }],
      getAudioTracks: () => [],
    } as MediaStream;

    session.addRemoteStream('peer-b', video);
    expect(session.screenShares).toHaveLength(1);
    expect(session.screenShares[0]?.peerId).toBe('peer-b');
  });

  it('patches members without reorder flicker', () => {
    session.setRoom({
      id: 'r1',
      name: 'Room',
      channels: [{ id: 'general', name: 'general' }],
      members: [
        { id: 'a', name: 'Ada', muted: false, deafened: false, speaking: false },
        { id: 'b', name: 'Bob', muted: false, deafened: false, speaking: false },
      ],
    });
    session.patchMember('b', { speaking: true });
    expect(session.sortedMembers.map((m) => m.name)).toEqual(['Ada', 'Bob']);
    expect(session.sortedMembers.find((m) => m.id === 'b')?.speaking).toBe(true);
  });

  it('toggles reactions per peer without duplicates', () => {
    session.toggleReaction('m1', '👍', 'a', true);
    session.toggleReaction('m1', '👍', 'a', true);
    session.toggleReaction('m1', '👍', 'b', true);
    expect(session.reactions['m1'][0].peerIds).toEqual(['a', 'b']);
    expect(session.hasReacted('m1', '👍', 'a')).toBe(true);

    session.toggleReaction('m1', '👍', 'a', false);
    expect(session.reactions['m1'][0].peerIds).toEqual(['b']);

    session.toggleReaction('m1', '👍', 'b', false);
    expect(session.reactions['m1']).toEqual([]);
  });

  it('tracks all active shares including local screen', () => {
    const video = {
      getVideoTracks: () => [{ kind: 'video', readyState: 'live' }],
      getAudioTracks: () => [],
    } as MediaStream;

    session.peerId = 'self';
    session.sharing = true;
    session.localScreen = video;
    session.addRemoteStream('peer-b', video);

    expect(session.allActiveShares).toHaveLength(2);
    expect(session.allActiveShares[0]?.peerId).toBe('self');
  });

  it('hides and shows the screen panel for the session', () => {
    expect(session.screenPanelVisible).toBe(true);
    session.hideScreenPanel();
    expect(session.screenPanelVisible).toBe(false);
    session.showScreenPanel('peer-a');
    expect(session.screenPanelVisible).toBe(true);
    expect(session.focusedShare).toBe('peer-a');
  });

  it('tracks watchers per share', () => {
    session.setWatchers('share-x', 'a', true);
    session.setWatchers('share-x', 'a', true);
    session.setWatchers('share-x', 'b', true);
    expect(session.watchers['share-x']).toEqual(['a', 'b']);
    session.setWatchers('share-x', 'a', false);
    expect(session.watchers['share-x']).toEqual(['b']);
  });
});
