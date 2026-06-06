import { sortMembers } from '../members';
import type { RoomState, ChatMessage, Member, ScreenShare, ReactionState } from '../types';

class SessionStore {
  peerId = $state('');
  room = $state<RoomState | null>(null);
  invite = $state('');
  roomKey = $state('');
  activeChannel = $state('general');
  messages = $state<ChatMessage[]>([]);
  attachments = $state<Record<string, Blob>>({});
  reactions = $state<Record<string, ReactionState[]>>({});
  remoteVoiceStreams = $state<Record<string, MediaStream>>({});
  screenStreamByPeer = $state<Record<string, MediaStream>>({});
  screenShares = $state<ScreenShare[]>([]);
  pausedShares = $state<Record<string, boolean>>({});
  shareAudioMuted = $state<Record<string, boolean>>({});
  watchers = $state<Record<string, string[]>>({});
  peerOnline = $state<Record<string, boolean>>({});
  localScreen = $state<MediaStream | null>(null);
  focusedShare = $state<string | null>(null);
  screenPanelVisible = $state(true);
  muted = $state(false);
  deafened = $state(false);
  sharing = $state(false);
  connected = $state(false);
  meshReady = $state(false);
  ping = $state<number | null>(null);
  error = $state('');

  sortedMembers = $derived(sortMembers(this.room?.members ?? []));
  isHost = $derived(!!this.room?.hostId && this.room.hostId === this.peerId);
  allActiveShares = $derived.by(() => {
    const shares = [...this.screenShares];
    if (this.localScreen && this.sharing) {
      shares.unshift({ peerId: this.peerId, stream: this.localScreen });
    }
    return shares;
  });

  setRoom(r: RoomState) {
    this.room = { ...r, members: sortMembers(r.members) };
  }

  patchMember(peerId: string, patch: Partial<Member>) {
    if (!this.room) return;
    const members = this.room.members.map((m) => (m.id === peerId ? { ...m, ...patch } : m));
    this.room = { ...this.room, members: sortMembers(members) };
  }

  memberName(peerId: string): string {
    if (peerId === this.peerId) return 'You';
    return this.room?.members.find((m) => m.id === peerId)?.name ?? 'Peer';
  }

  setPeerOnline(peerId: string, online: boolean) {
    this.peerOnline = { ...this.peerOnline, [peerId]: online };
  }

  addMessage(msg: ChatMessage) {
    if (this.messages.some((m) => m.id === msg.id)) return;
    this.messages = [...this.messages, msg];
  }

  messagesForChannel(channelId: string) {
    return this.messages.filter((m) => m.channelId === channelId);
  }

  setAttachment(metaId: string, blob: Blob) {
    this.attachments = { ...this.attachments, [metaId]: blob };
  }

  toggleReaction(messageId: string, emoji: string, peerId: string, add: boolean) {
    const current = this.reactions[messageId] ?? [];
    const existing = current.find((r) => r.emoji === emoji);
    let next: ReactionState[];
    if (add) {
      if (existing) {
        if (existing.peerIds.includes(peerId)) return;
        next = current.map((r) =>
          r.emoji === emoji ? { ...r, peerIds: [...r.peerIds, peerId] } : r,
        );
      } else {
        next = [...current, { emoji, peerIds: [peerId] }];
      }
    } else {
      next = current
        .map((r) =>
          r.emoji === emoji ? { ...r, peerIds: r.peerIds.filter((p) => p !== peerId) } : r,
        )
        .filter((r) => r.peerIds.length > 0);
    }
    this.reactions = { ...this.reactions, [messageId]: next };
  }

  hasReacted(messageId: string, emoji: string, peerId: string): boolean {
    return (this.reactions[messageId] ?? []).some(
      (r) => r.emoji === emoji && r.peerIds.includes(peerId),
    );
  }

  addRemoteStream(peerId: string, stream: MediaStream) {
    const hasVideo = stream.getVideoTracks().length > 0;
    if (hasVideo) {
      this.screenStreamByPeer = { ...this.screenStreamByPeer, [peerId]: stream };
      if (!this.focusedShare) this.focusedShare = peerId;
      this.syncScreenShares();
      return;
    }
    if (stream.getAudioTracks().length > 0) {
      this.remoteVoiceStreams = { ...this.remoteVoiceStreams, [peerId]: stream };
    }
  }

  removeRemoteStream(peerId: string, stream: MediaStream | null) {
    if (!stream) {
      const voice = { ...this.remoteVoiceStreams };
      delete voice[peerId];
      this.remoteVoiceStreams = voice;
      const screens = { ...this.screenStreamByPeer };
      delete screens[peerId];
      this.screenStreamByPeer = screens;
      if (this.focusedShare === peerId) this.focusedShare = null;
      this.syncScreenShares();
      return;
    }

    if (stream.getVideoTracks().length > 0) {
      if (this.screenStreamByPeer[peerId] === stream) {
        const screens = { ...this.screenStreamByPeer };
        delete screens[peerId];
        this.screenStreamByPeer = screens;
      }
      if (this.focusedShare === peerId) {
        this.focusedShare = this.screenShares.find((s) => s.peerId !== peerId)?.peerId ?? null;
      }
      this.syncScreenShares();
      return;
    }

    if (this.remoteVoiceStreams[peerId] === stream) {
      const voice = { ...this.remoteVoiceStreams };
      delete voice[peerId];
      this.remoteVoiceStreams = voice;
    }
  }

  syncScreenShares() {
    const shares: ScreenShare[] = [];
    for (const [peerId, stream] of Object.entries(this.screenStreamByPeer)) {
      if (stream.getVideoTracks().some((t) => t.readyState === 'live')) {
        shares.push({ peerId, stream });
      }
    }
    this.screenShares = shares;
    if (this.focusedShare && !shares.some((s) => s.peerId === this.focusedShare)) {
      this.focusedShare = shares[0]?.peerId ?? null;
    }
  }

  toggleShareAudioMuted(peerId: string) {
    const next = !this.shareAudioMuted[peerId];
    this.shareAudioMuted = { ...this.shareAudioMuted, [peerId]: next };
  }

  isShareAudioMuted(peerId: string): boolean {
    return this.shareAudioMuted[peerId] === true;
  }

  setPaused(shareId: string, paused: boolean) {
    this.pausedShares = { ...this.pausedShares, [shareId]: paused };
  }

  showScreenPanel(peerId?: string) {
    this.screenPanelVisible = true;
    if (peerId) {
      this.focusedShare = peerId;
      return;
    }
    if (!this.focusedShare || !this.allActiveShares.some((s) => s.peerId === this.focusedShare)) {
      this.focusedShare = this.allActiveShares[0]?.peerId ?? null;
    }
  }

  hideScreenPanel() {
    this.screenPanelVisible = false;
  }

  setWatchers(shareId: string, peerId: string, watching: boolean) {
    const current = this.watchers[shareId] ?? [];
    const next = watching
      ? current.includes(peerId)
        ? current
        : [...current, peerId]
      : current.filter((p) => p !== peerId);
    this.watchers = { ...this.watchers, [shareId]: next };
  }

  reset() {
    this.peerId = '';
    this.room = null;
    this.invite = '';
    this.roomKey = '';
    this.activeChannel = 'general';
    this.messages = [];
    this.attachments = {};
    this.reactions = {};
    this.remoteVoiceStreams = {};
    this.screenStreamByPeer = {};
    this.screenShares = [];
    this.pausedShares = {};
    this.shareAudioMuted = {};
    this.watchers = {};
    this.peerOnline = {};
    this.localScreen = null;
    this.focusedShare = null;
    this.screenPanelVisible = true;
    this.muted = false;
    this.deafened = false;
    this.sharing = false;
    this.connected = false;
    this.meshReady = false;
    this.ping = null;
    this.error = '';
  }
}

export const session = new SessionStore();
