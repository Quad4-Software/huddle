import { Signaling } from './ws/signaling';
import { Mesh } from './webrtc/mesh';
import { importRoomKey, importSigningKey } from './crypto/e2e';
import { getMicStream, createAnalyser, isSpeaking } from './webrtc/audio';
import { getScreenStream } from './webrtc/screen';
import { session } from './stores/session.svelte';
import { settings } from './stores/settings.svelte';
import { loading } from './stores/loading.svelte';
import { buildInviteUrl } from './invite';
import { randomDisplayName } from './random-name';
import { solvePow } from './pow';
import type { RoomState, ControlMessage } from './types';
import {
  cleanBounded,
  MAX_CHAT_MESSAGE_LENGTH,
  MAX_DISPLAY_NAME_LENGTH,
  MAX_FILE_SIZE,
  MAX_PASSWORD_LENGTH,
  MAX_ROOM_NAME_LENGTH,
} from './validation';

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
let meshPeerId: string | null = null;
let intentionalLeave = false;
let reconnecting = false;

function peerResumeKey(roomId: string) {
  return `huddle:peer:${roomId}`;
}

function peerResumeTokenKey(roomId: string) {
  return `huddle:resume:${roomId}`;
}

function peerPasswordKey(roomId: string) {
  return `huddle:pw:${roomId}`;
}

function applyJoinedState(peerId: string, room: RoomState, resumeToken?: string) {
  session.peerId = peerId;
  session.setRoom(room);
  session.connected = true;
  session.error = '';
  sessionStorage.setItem(peerResumeKey(room.id), peerId);
  if (resumeToken) {
    sessionStorage.setItem(peerResumeTokenKey(room.id), resumeToken);
  }
}

async function ensureSignaling() {
  if (!signaling) {
    if (loading.active) {
      loading.setPhase('connecting');
      loading.advanceTo(5);
    }
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
    const p = payload as {
      peerId?: string;
      resumeToken?: string;
      room?: RoomState;
      peers?: string[] | null;
      iceServers?: RTCIceServer[];
    };
    if (!p?.peerId || !p?.room) return;

    applyJoinedState(p.peerId, p.room, p.resumeToken);
    const peers = Array.isArray(p.peers) ? p.peers : [];

    if (meshPeerId === p.peerId && mesh) {
      return;
    }

    let cryptoKey: CryptoKey;
    let signingKey: CryptoKey;
    try {
      const roomKey = session.roomKey || location.hash.match(/key=([^&]+)/)?.[1] || '';
      cryptoKey = await importRoomKey(roomKey);
      signingKey = await importSigningKey(roomKey);
    } catch {
      session.error = 'Could not unlock room encryption';
      return;
    }

    mesh?.destroy();
    meshPeerId = p.peerId;
    session.setPeerOnline(p.peerId, true);

    mesh = new Mesh(
      p.peerId,
      cleanBounded(settings.displayName || 'Guest', MAX_DISPLAY_NAME_LENGTH) || 'Guest',
      cryptoKey,
      signingKey,
      p.iceServers ?? [],
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
    const p = payload as { from: string; sdp: string; nonce: string; sig: string };
    await mesh?.handleOffer(p.from, p.sdp, p.nonce, p.sig);
  });

  signaling.on('answer', async (payload) => {
    const p = payload as { from: string; sdp: string; nonce: string; sig: string };
    await mesh?.handleAnswer(p.from, p.sdp, p.nonce, p.sig);
  });

  signaling.on('ice', async (payload) => {
    const p = payload as {
      from: string;
      candidate: RTCIceCandidateInit;
      nonce: string;
      sig: string;
    };
    await mesh?.handleICE(p.from, p.candidate, p.nonce, p.sig);
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
    session.meshReady = false;
    session.ping = null;
    if (intentionalLeave || reconnecting) {
      if (intentionalLeave) {
        session.connected = false;
      }
      return;
    }
    const roomId = location.pathname.match(/^\/r\/([^/]+)/)?.[1];
    const invite = new URLSearchParams(location.search).get('t');
    const key = location.hash.match(/key=([^&]+)/)?.[1];
    if (!roomId || !invite || !key || !session.room) {
      session.connected = false;
      return;
    }
    session.invite = invite;
    session.roomKey = key;
    void reconnectSession(roomId, invite);
  });
}

async function reconnectSession(roomId: string, invite: string) {
  if (reconnecting) return;
  reconnecting = true;
  const password = (sessionStorage.getItem(peerPasswordKey(roomId)) ?? '').slice(
    0,
    MAX_PASSWORD_LENGTH,
  );
  const name =
    settings.displayName ||
    session.room?.members.find((m) => m.id === session.peerId)?.name ||
    'Guest';
  try {
    mesh?.destroy();
    mesh = null;
    meshPeerId = null;
    signaling = null;
    signalingWired = false;
    await ensureSignaling();
    await joinRoom(roomId, invite, password, name);
  } catch {
    session.connected = false;
  } finally {
    reconnecting = false;
  }
}

export async function createRoom(name: string, password: string): Promise<void> {
  loading.start('creating');
  try {
    await ensureSignaling();
    const roomName = cleanBounded(name, MAX_ROOM_NAME_LENGTH);
    const roomPassword = password.slice(0, MAX_PASSWORD_LENGTH);

    return new Promise((resolve, reject) => {
      const fail = (err: unknown) => {
        loading.stop();
        reject(err instanceof Error ? err : new Error(String(err)));
      };

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
        joinRoom(
          p.roomId,
          p.invite,
          roomPassword,
          cleanBounded(settings.displayName || 'Host', MAX_DISPLAY_NAME_LENGTH) || 'Host',
          { powStart: 45, powEnd: 85 },
        )
          .then(() => {
            loading.stop();
            resolve();
          })
          .catch(fail);
      });
      signaling!.once('error', (payload) => {
        fail(new Error((payload as { message: string }).message));
      });
      loading.beginPow(5, 40);
      solvePow('create', (progress) => loading.setPowProgress(progress))
        .then((pow) => {
          loading.advanceTo(40);
          loading.setPhase('creating');
          signaling!.send('create_room', {
            name: roomName,
            password: roomPassword || undefined,
            pow,
          });
        })
        .catch(fail);
    });
  } catch (err) {
    loading.stop();
    throw err;
  }
}

export async function joinFromUrl(): Promise<boolean> {
  const params = new URLSearchParams(location.search);
  const roomId = location.pathname.match(/^\/r\/([^/]+)/)?.[1];
  const invite = params.get('t');
  const key = location.hash.match(/key=([^&]+)/)?.[1];
  if (!roomId || !invite || !key) return false;

  loading.start('joining');
  session.invite = invite;
  session.roomKey = key;
  let name = settings.displayName.trim();
  if (!name) {
    name = randomDisplayName();
  }
  const cleanName = cleanBounded(name, MAX_DISPLAY_NAME_LENGTH) || randomDisplayName();
  settings.setName(cleanName);

  try {
    await joinRoom(roomId, invite, '', cleanName);
  } catch {
    loading.stop();
    const password = (prompt('Enter room password if required') ?? '').slice(
      0,
      MAX_PASSWORD_LENGTH,
    );
    if (!password) return false;
    await joinRoom(roomId, invite, password, cleanName);
  }
  return true;
}

export async function joinRoom(
  roomId: string,
  invite: string,
  password: string,
  name: string,
  options?: { powStart?: number; powEnd?: number },
): Promise<void> {
  if (!loading.active) loading.start('joining');
  await ensureSignaling();
  const powStart = options?.powStart ?? 5;
  const powEnd = options?.powEnd ?? 85;
  const cleanName = cleanBounded(name, MAX_DISPLAY_NAME_LENGTH);
  const cleanPassword = password.slice(0, MAX_PASSWORD_LENGTH);

  let resumePeerId = sessionStorage.getItem(peerResumeKey(roomId)) ?? undefined;
  let resumeToken = sessionStorage.getItem(peerResumeTokenKey(roomId)) ?? undefined;
  if (resumePeerId && !resumeToken) {
    sessionStorage.removeItem(peerResumeKey(roomId));
    resumePeerId = undefined;
  }
  if (!resumePeerId || !resumeToken) {
    resumePeerId = undefined;
    resumeToken = undefined;
  }

  return new Promise((resolve, reject) => {
    const fail = (err: unknown) => {
      loading.stop();
      reject(err instanceof Error ? err : new Error(String(err)));
    };

    signaling!.once('joined', (payload) => {
      const p = payload as { peerId?: string; room?: RoomState; resumeToken?: string };
      if (p.peerId && p.room) {
        applyJoinedState(p.peerId, p.room, p.resumeToken);
      }
      if (cleanPassword) {
        sessionStorage.setItem(peerPasswordKey(roomId), cleanPassword);
      }
      loading.stop();
      resolve();
    });
    signaling!.once('error', (payload) => {
      fail(new Error((payload as { message: string }).message));
    });
    loading.beginPow(powStart, powEnd);
    solvePow('join', (progress) => loading.setPowProgress(progress))
      .then((pow) => {
        loading.advanceTo(powEnd);
        loading.setPhase('joining');
        signaling!.send('join', {
          roomId,
          invite,
          password: cleanPassword || undefined,
          name: cleanName,
          pow,
          resumePeerId,
          resumeToken,
        });
      })
      .catch(fail);
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
  const clean = text.trim().slice(0, MAX_CHAT_MESSAGE_LENGTH);
  if (!clean) return;
  await mesh.broadcastMessage(session.activeChannel, clean);
}

export async function sendFile(file: File) {
  if (!mesh) return;
  if (file.size > MAX_FILE_SIZE) {
    session.error = 'File is too large';
    return;
  }
  await mesh.broadcastFile(session.activeChannel, file);
}

export function toggleReaction(messageId: string, emoji: string) {
  const add = !session.hasReacted(messageId, emoji, session.peerId);
  session.toggleReaction(messageId, emoji, session.peerId, add);
  mesh?.broadcastControl({ kind: 'reaction', messageId, emoji, peerId: session.peerId, add });
}

export function changeName(name: string) {
  const clean = cleanBounded(name, MAX_DISPLAY_NAME_LENGTH);
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
  meshPeerId = null;
  signaling = null;
  signalingWired = false;
  micStream = null;
  lastSpeaking = false;
}

function clearPeerResume(roomId?: string) {
  if (roomId) {
    sessionStorage.removeItem(peerResumeKey(roomId));
  }
}

function forceLeaveSession(message: string) {
  intentionalLeave = true;
  clearPeerResume(session.room?.id);
  cleanupSession();
  session.reset();
  session.error = message;
  history.replaceState(null, '', '/');
  intentionalLeave = false;
}

export function leaveSession() {
  intentionalLeave = true;
  clearPeerResume(session.room?.id);
  signaling?.send('leave');
  cleanupSession();
  session.reset();
  history.replaceState(null, '', '/');
  intentionalLeave = false;
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
