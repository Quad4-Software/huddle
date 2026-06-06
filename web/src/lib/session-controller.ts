import { Signaling } from './ws/signaling';
import { Mesh } from './webrtc/mesh';
import { importRoomKey } from './crypto/e2e';
import { getMicStream, createAnalyser, isSpeaking } from './webrtc/audio';
import { getScreenStream } from './webrtc/screen';
import { session } from './stores/session.svelte';
import { settings } from './stores/settings.svelte';
import { buildInviteUrl } from './invite';
import { solvePow } from './pow';
import type { RoomState, ControlMessage } from './types';

export { buildInviteUrl };

let signaling: Signaling | null = null;
let mesh: Mesh | null = null;
let micStream: MediaStream | null = null;
let audioCtx: AudioContext | null = null;
let analyser: AnalyserNode | null = null;
let speakTimer: ReturnType<typeof setInterval> | null = null;
let pingTimer: ReturnType<typeof setInterval> | null = null;
let lastSpeaking = false;
let signalingWired = false;

async function ensureSignaling() {
  if (!signaling) {
    signaling = new Signaling();
    await signaling.connect();
  }
  if (!signalingWired) {
    wireSignaling();
    signalingWired = true;
  }
}

function handleControl(msg: ControlMessage) {
  if (msg.kind === 'reaction') {
    session.toggleReaction(msg.messageId, msg.emoji, msg.peerId, msg.add);
  } else if (msg.kind === 'watch') {
    session.setWatchers(msg.shareId, msg.peerId, msg.watching);
  }
}

function wireSignaling() {
  if (!signaling) return;

  signaling.on('joined', async (payload) => {
    const p = payload as { peerId?: string; room?: RoomState; peers?: string[] | null };
    if (!p?.peerId || !p?.room) return;

    const peers = Array.isArray(p.peers) ? p.peers : [];

    if (session.peerId === p.peerId && mesh) {
      session.setRoom(p.room);
      session.connected = true;
      return;
    }

    const cryptoKey = await importRoomKey(
      session.roomKey || location.hash.match(/key=([^&]+)/)?.[1] || '',
    );

    mesh?.destroy();
    session.peerId = p.peerId;
    session.setRoom(p.room);
    session.connected = true;
    session.setPeerOnline(p.peerId, true);

    mesh = new Mesh(
      p.peerId,
      settings.displayName || 'Guest',
      cryptoKey,
      (type, to, data) => {
        signaling?.send(type, { to, ...data });
      },
      {
        onMessage: (msg) => session.addMessage(msg),
        onAttachment: (meta, blob) => session.setAttachment(meta.id, blob),
        onControl: handleControl,
        onTrack: (peerId, stream) => {
          const wasShare = session.screenShares.some((s) => s.peerId === peerId);
          session.addRemoteStream(peerId, stream);
          const isShare = session.screenShares.some((s) => s.peerId === peerId);
          if (isShare && !wasShare && peerId !== session.peerId) {
            mesh?.broadcastControl({
              kind: 'watch',
              shareId: peerId,
              peerId: session.peerId,
              watching: !session.pausedShares[peerId],
            });
          }
        },
        onTrackRemoved: (peerId, stream) => session.removeRemoteStream(peerId, stream),
        onPeerConnected: (peerId, connected) => {
          session.setPeerOnline(peerId, connected);
        },
        onMeshReady: () => {
          session.meshReady = mesh?.hasOpenChannels() ?? false;
          broadcastWatchState();
        },
      },
    );

    try {
      micStream = await getMicStream(settings.inputDeviceId || undefined);
      await mesh.addLocalAudio(micStream);
      startVoiceActivity();
    } catch {
      session.error = 'Microphone access denied';
    }

    for (const peer of peers) {
      if (p.peerId < peer) {
        await mesh.connectTo(peer);
      }
    }
    session.meshReady = mesh.hasOpenChannels();
    startPing();
  });

  signaling.on('peer_joined', async (payload) => {
    const p = payload as { peerId: string };
    if (session.peerId && session.peerId < p.peerId) {
      await mesh?.connectTo(p.peerId);
    }
  });

  signaling.on('peer_left', (payload) => {
    const p = payload as { peerId: string };
    mesh?.removePeer(p.peerId);
    session.setPeerOnline(p.peerId, false);
    if (session.room) {
      session.setRoom({
        ...session.room,
        members: session.room.members.filter((m) => m.id !== p.peerId),
      });
    }
  });

  signaling.on('kicked', () => {
    forceLeaveSession('You were removed from the room');
  });

  signaling.on('offer', async (payload) => {
    const p = payload as { from: string; sdp: string };
    await mesh?.handleOffer(p.from, p.sdp);
  });

  signaling.on('answer', async (payload) => {
    const p = payload as { from: string; sdp: string };
    await mesh?.handleAnswer(p.from, p.sdp);
  });

  signaling.on('ice', async (payload) => {
    const p = payload as { from: string; candidate: RTCIceCandidateInit };
    await mesh?.handleICE(p.from, p.candidate);
  });

  signaling.on('room_state', (payload) => {
    const next = payload as RoomState;
    if (!session.room) {
      session.setRoom(next);
      return;
    }
    session.setRoom({ ...session.room, ...next, members: next.members });
  });

  signaling.on('member_update', (payload) => {
    const p = payload as { peerId: string; muted: boolean; deafened: boolean; speaking: boolean };
    session.patchMember(p.peerId, {
      muted: p.muted,
      deafened: p.deafened,
      speaking: p.speaking,
    });
  });

  signaling.on('pong', (payload) => {
    const p = payload as { t: number };
    session.ping = Math.max(0, Math.round(Date.now() - p.t));
  });

  signaling.on('close', () => {
    session.connected = false;
    session.meshReady = false;
    session.ping = null;
  });
}

export async function createRoom(name: string, password: string): Promise<void> {
  await ensureSignaling();

  return new Promise((resolve, reject) => {
    signaling!.once('created', (payload) => {
      const p = payload as {
        roomId: string;
        invite: string;
        roomKey: string;
        expiresAt: number;
      };
      session.invite = p.invite;
      session.roomKey = p.roomKey;
      const url = buildInviteUrl(p.roomId, p.invite, p.roomKey);
      history.replaceState(null, '', url);
      joinRoom(p.roomId, p.invite, password, settings.displayName || 'Host')
        .then(resolve)
        .catch(reject);
    });
    signaling!.once('error', (payload) => {
      reject(new Error((payload as { message: string }).message));
    });
    solvePow('create')
      .then((pow) => {
        signaling!.send('create_room', { name, password: password || undefined, pow });
      })
      .catch(reject);
  });
}

export async function joinFromUrl(): Promise<boolean> {
  const params = new URLSearchParams(location.search);
  const roomId = location.pathname.match(/^\/r\/([^/]+)/)?.[1];
  const invite = params.get('t');
  const key = location.hash.match(/key=([^&]+)/)?.[1];
  if (!roomId || !invite || !key) return false;

  session.invite = invite;
  session.roomKey = key;
  const name = settings.displayName || prompt('Display name') || 'Guest';
  settings.setName(name);

  let password = params.get('pw') ?? '';
  try {
    await joinRoom(roomId, invite, password, name);
  } catch (e) {
    if (e instanceof Error && e.message.includes('password')) {
      password = prompt('Room password') ?? '';
      if (!password) return false;
      await joinRoom(roomId, invite, password, name);
    } else {
      throw e;
    }
  }
  return true;
}

export async function joinRoom(
  roomId: string,
  invite: string,
  password: string,
  name: string,
): Promise<void> {
  await ensureSignaling();

  return new Promise((resolve, reject) => {
    signaling!.once('joined', () => resolve());
    signaling!.once('error', (payload) => {
      reject(new Error((payload as { message: string }).message));
    });
    solvePow('join')
      .then((pow) => {
        signaling!.send('join', { roomId, invite, password: password || undefined, name, pow });
      })
      .catch(reject);
  });
}

export function kickMember(peerId: string) {
  if (!session.isHost || peerId === session.peerId) return;
  signaling?.send('kick', { peerId });
}

function startVoiceActivity() {
  if (!micStream) return;
  audioCtx = new AudioContext();
  const { analyser: a } = createAnalyser(audioCtx, micStream);
  analyser = a;
  speakTimer = setInterval(() => {
    if (!analyser || session.muted) {
      if (lastSpeaking) {
        lastSpeaking = false;
        sendMemberState(false);
      }
      return;
    }
    const speaking = isSpeaking(analyser);
    if (speaking !== lastSpeaking) {
      lastSpeaking = speaking;
      sendMemberState(speaking);
    }
  }, 150);
}

function startPing() {
  if (pingTimer) clearInterval(pingTimer);
  const beat = () => signaling?.send('ping', { t: Date.now() });
  beat();
  pingTimer = setInterval(beat, 4000);
}

function sendMemberState(speaking: boolean) {
  signaling?.send('member_update', {
    peerId: session.peerId,
    muted: session.muted,
    deafened: session.deafened,
    speaking,
  });
  session.patchMember(session.peerId, {
    muted: session.muted,
    deafened: session.deafened,
    speaking,
  });
}

export function toggleMute() {
  session.muted = !session.muted;
  mesh?.setMuted(session.muted);
  if (session.muted) lastSpeaking = false;
  sendMemberState(false);
}

export function toggleDeafen() {
  session.deafened = !session.deafened;
  if (session.deafened) session.muted = true;
  mesh?.setMuted(session.muted);
  lastSpeaking = false;
  sendMemberState(false);
}

export async function sendMessage(text: string) {
  if (!mesh) {
    session.error = 'Not connected yet';
    return;
  }
  await mesh.broadcastMessage(session.activeChannel, text);
}

export async function sendFile(file: File) {
  if (!mesh) return;
  await mesh.broadcastFile(session.activeChannel, file);
}

export function toggleReaction(messageId: string, emoji: string) {
  const add = !session.hasReacted(messageId, emoji, session.peerId);
  session.toggleReaction(messageId, emoji, session.peerId, add);
  mesh?.broadcastControl({ kind: 'reaction', messageId, emoji, peerId: session.peerId, add });
}

export function changeName(name: string) {
  const clean = name.trim();
  if (!clean) return;
  settings.setName(clean);
  mesh?.setLocalName(clean);
  session.patchMember(session.peerId, { name: clean });
  signaling?.send('rename', { name: clean });
}

function broadcastWatchState() {
  for (const share of session.screenShares) {
    const watching = !session.pausedShares[share.peerId];
    mesh?.broadcastControl({
      kind: 'watch',
      shareId: share.peerId,
      peerId: session.peerId,
      watching,
    });
  }
}

export function toggleShareAudioMuted(shareId: string) {
  session.toggleShareAudioMuted(shareId);
}

export function setSharePaused(shareId: string, paused: boolean) {
  session.setPaused(shareId, paused);
  mesh?.broadcastControl({
    kind: 'watch',
    shareId,
    peerId: session.peerId,
    watching: !paused,
  });
}

export async function startShare() {
  const stream = await getScreenStream();
  session.sharing = true;
  session.localScreen = stream;
  session.focusedShare = session.peerId;
  await mesh?.startScreenShare(stream);
}

export function stopShare() {
  mesh?.stopScreenShare();
  session.sharing = false;
  session.localScreen = null;
  if (session.focusedShare === session.peerId) session.focusedShare = null;
}

function cleanupSession() {
  if (speakTimer) clearInterval(speakTimer);
  if (pingTimer) clearInterval(pingTimer);
  audioCtx?.close();
  micStream?.getTracks().forEach((t) => t.stop());
  session.localScreen?.getTracks().forEach((t) => t.stop());
  mesh?.destroy();
  signaling?.close();
  mesh = null;
  signaling = null;
  signalingWired = false;
  micStream = null;
  lastSpeaking = false;
}

function forceLeaveSession(message: string) {
  cleanupSession();
  session.reset();
  session.error = message;
  history.replaceState(null, '', '/');
}

export function leaveSession() {
  signaling?.send('leave');
  cleanupSession();
  session.reset();
  history.replaceState(null, '', '/');
}

export async function refreshMic() {
  micStream?.getTracks().forEach((t) => t.stop());
  micStream = await getMicStream(settings.inputDeviceId || undefined);
  await mesh?.addLocalAudio(micStream);
  if (speakTimer) {
    audioCtx?.close();
    startVoiceActivity();
  }
}
