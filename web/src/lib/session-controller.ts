import { Signaling } from './ws/signaling';
import { Mesh } from './webrtc/mesh';
import { importRoomKey, importSigningKey } from './crypto/e2e';
import { getMicStream } from './webrtc/audio';
import { createMicPipeline, type MicPipeline } from './webrtc/mic-pipeline';
import { getScreenStream } from './webrtc/screen';
import { session } from './stores/session.svelte';
import { settings, voiceActivationThreshold } from './stores/settings.svelte';
import { loading } from './stores/loading.svelte';
import { connection } from './stores/connection.svelte';
import { buildInviteUrl } from './invite';
import { randomDisplayName } from './random-name';
import { solvePow } from './pow';
import { RECONNECT_MAX_ATTEMPTS, reconnectDelayMs, sleep } from './reconnect';
import type { RoomState, ControlMessage } from './types';
import { isTypingTarget } from './keybind';
import { isSpeaking } from './webrtc/audio';
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
let micPipeline: MicPipeline | null = null;
let analyser: AnalyserNode | null = null;
let speakTimer: ReturnType<typeof setInterval> | null = null;
let pingTimer: ReturnType<typeof setInterval> | null = null;
let lastSpeaking = false;
let pttActive = false;
let pttListenersInstalled = false;
let signalingWired = false;
let meshPeerId: string | null = null;
let intentionalLeave = false;
let reconnecting = false;
let reconnectAbort = false;

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
  connection.setOnline();
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
      const processed = createMicPipeline(micStream);
      micPipeline = processed.pipeline;
      micPipeline.setInputVolume(settings.inputVolume);
      analyser = processed.analyser;
      await mesh.addLocalAudio(processed.stream);
      applyMicTransmit();
      startVoiceActivity();
      ensurePttListeners();
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
    if (p.peerId === session.peerId) {
      session.muted = p.muted;
      session.deafened = p.deafened;
      applyMicTransmit();
    }
  });

  signaling.on('pong', (payload) => {
    const p = payload as { t: number };
    session.ping = Math.max(0, Math.round(Date.now() - p.t));
  });

  signaling.on('close', () => {
    session.meshReady = false;
    session.ping = null;
    if (intentionalLeave) {
      session.connected = false;
      connection.reset();
      return;
    }
    if (reconnecting) return;

    const roomId = location.pathname.match(/^\/r\/([^/]+)/)?.[1];
    const invite = new URLSearchParams(location.search).get('t');
    const key = location.hash.match(/key=([^&]+)/)?.[1];
    if (!roomId || !invite || !key || !session.room) {
      session.connected = false;
      connection.reset();
      return;
    }

    session.invite = invite;
    session.roomKey = key;
    session.connected = false;
    void reconnectSession(roomId, invite);
  });
}

function teardownTransport() {
  if (speakTimer) clearInterval(speakTimer);
  if (pingTimer) clearInterval(pingTimer);
  speakTimer = null;
  pingTimer = null;
  micPipeline?.close();
  micPipeline = null;
  micStream?.getTracks().forEach((t) => t.stop());
  micStream = null;
  analyser = null;
  mesh?.destroy();
  mesh = null;
  meshPeerId = null;
  lastSpeaking = false;
  pttActive = false;
  session.meshReady = false;
  session.ping = null;
}

async function reconnectSession(roomId: string, invite: string) {
  if (intentionalLeave || reconnectAbort) return;
  if (reconnecting) return;

  reconnecting = true;
  reconnectAbort = false;
  connection.startReconnect();

  const password = (sessionStorage.getItem(peerPasswordKey(roomId)) ?? '').slice(
    0,
    MAX_PASSWORD_LENGTH,
  );
  const name =
    settings.displayName ||
    session.room?.members.find((m) => m.id === session.peerId)?.name ||
    'Guest';

  for (let attempt = 1; attempt <= RECONNECT_MAX_ATTEMPTS; attempt++) {
    if (intentionalLeave || reconnectAbort) break;

    connection.setAttempt(attempt);
    connection.setDetail(
      attempt === 1
        ? 'Restoring connection to the server...'
        : `Retrying connection (${attempt}/${RECONNECT_MAX_ATTEMPTS})...`,
    );

    try {
      teardownTransport();
      signaling?.close();
      signaling = null;
      signalingWired = false;

      connection.setDetail('Connecting to server...');
      await ensureSignaling();
      await joinRoom(roomId, invite, password, name, { silent: true });

      reconnecting = false;
      return;
    } catch {
      if (attempt < RECONNECT_MAX_ATTEMPTS && !intentionalLeave && !reconnectAbort) {
        const wait = reconnectDelayMs(attempt);
        connection.setDetail(`Waiting ${Math.ceil(wait / 1000)}s before retry...`);
        await sleep(wait);
      }
    }
  }

  if (!intentionalLeave && !reconnectAbort) {
    connection.setOffline();
    session.connected = false;
  }
  reconnecting = false;
}

export function retryConnection() {
  if (reconnecting) return;

  const roomId = session.room?.id ?? location.pathname.match(/^\/r\/([^/]+)/)?.[1];
  const invite = session.invite || new URLSearchParams(location.search).get('t') || '';
  const key = session.roomKey || location.hash.match(/key=([^&]+)/)?.[1] || '';
  if (!roomId || !invite || !key) {
    connection.setOffline();
    return;
  }

  session.invite = invite;
  session.roomKey = key;
  reconnectAbort = false;
  void reconnectSession(roomId, invite);
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
  options?: { powStart?: number; powEnd?: number; silent?: boolean },
): Promise<void> {
  const silent = options?.silent ?? false;
  if (!silent && !loading.active) loading.start('joining');
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
      if (!silent) loading.stop();
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
      if (!silent) loading.stop();
      else connection.setDetail('');
      resolve();
    });
    signaling!.once('error', (payload) => {
      fail(new Error((payload as { message: string }).message));
    });

    const onPowProgress = (progress: number) => {
      if (silent) {
        connection.setDetail(`Solving PoW... ${Math.round(progress)}%`);
      } else {
        loading.setPhase('pow');
        loading.setPowProgress(progress);
      }
    };

    if (silent) {
      connection.setDetail('Solving PoW...');
    } else {
      loading.beginPow(powStart, powEnd);
    }

    solvePow('join', onPowProgress)
      .then((pow) => {
        if (silent) {
          connection.setDetail('Rejoining room...');
        } else {
          loading.advanceTo(powEnd);
          loading.setPhase('joining');
        }
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

export function moderateMember(peerId: string, muted: boolean, deafened: boolean) {
  if (!session.isHost || peerId === session.peerId) return;
  const nextDeafened = deafened;
  const nextMuted = muted || nextDeafened;
  signaling?.send('moderate_member', { peerId, muted: nextMuted, deafened: nextDeafened });
}

function ensurePttListeners() {
  if (pttListenersInstalled) return;
  pttListenersInstalled = true;
  window.addEventListener('keydown', onPttKeyDown);
  window.addEventListener('keyup', onPttKeyUp);
  window.addEventListener('blur', onPttBlur);
}

function removePttListeners() {
  if (!pttListenersInstalled) return;
  pttListenersInstalled = false;
  window.removeEventListener('keydown', onPttKeyDown);
  window.removeEventListener('keyup', onPttKeyUp);
  window.removeEventListener('blur', onPttBlur);
  pttActive = false;
}

function onPttKeyDown(e: KeyboardEvent) {
  if (settings.inputMode !== 'pushToTalk') return;
  if (e.code !== settings.pushToTalkKey) return;
  if (e.repeat || isTypingTarget(e.target)) return;
  e.preventDefault();
  pttActive = true;
  applyMicTransmit();
}

function onPttKeyUp(e: KeyboardEvent) {
  if (settings.inputMode !== 'pushToTalk') return;
  if (e.code !== settings.pushToTalkKey) return;
  pttActive = false;
  applyMicTransmit();
}

function onPttBlur() {
  if (!pttActive) return;
  pttActive = false;
  applyMicTransmit();
}

function applyMicTransmit() {
  if (session.muted) {
    mesh?.setMuted(true);
    if (lastSpeaking) {
      lastSpeaking = false;
      sendMemberState(false);
    }
    return;
  }
  if (settings.inputMode === 'pushToTalk') {
    mesh?.setMuted(!pttActive);
    if (pttActive !== lastSpeaking) {
      lastSpeaking = pttActive;
      sendMemberState(pttActive);
    }
    return;
  }
  mesh?.setMuted(false);
}

export function applyAudioSettings() {
  micPipeline?.setInputVolume(settings.inputVolume);
  applyMicTransmit();
}

function startVoiceActivity() {
  if (!analyser) return;
  if (speakTimer) clearInterval(speakTimer);
  speakTimer = setInterval(() => {
    if (!analyser || session.muted || settings.inputMode === 'pushToTalk') {
      if (lastSpeaking) {
        lastSpeaking = false;
        sendMemberState(false);
      }
      return;
    }
    const speaking = isSpeaking(analyser, voiceActivationThreshold());
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
  applyMicTransmit();
  if (session.muted) lastSpeaking = false;
  sendMemberState(false);
}

export function toggleDeafen() {
  session.deafened = !session.deafened;
  if (session.deafened) session.muted = true;
  applyMicTransmit();
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
  speakTimer = null;
  pingTimer = null;
  micPipeline?.close();
  micPipeline = null;
  micStream?.getTracks().forEach((t) => t.stop());
  session.localScreen?.getTracks().forEach((t) => t.stop());
  mesh?.destroy();
  signaling?.close();
  mesh = null;
  meshPeerId = null;
  signaling = null;
  signalingWired = false;
  micStream = null;
  analyser = null;
  lastSpeaking = false;
  pttActive = false;
  removePttListeners();
}

function clearPeerResume(roomId?: string) {
  if (roomId) {
    sessionStorage.removeItem(peerResumeKey(roomId));
  }
}

function forceLeaveSession(message: string) {
  reconnectAbort = true;
  intentionalLeave = true;
  clearPeerResume(session.room?.id);
  cleanupSession();
  session.reset();
  session.error = message;
  connection.reset();
  history.replaceState(null, '', '/');
  intentionalLeave = false;
}

export function leaveSession() {
  reconnectAbort = true;
  intentionalLeave = true;
  clearPeerResume(session.room?.id);
  signaling?.send('leave');
  cleanupSession();
  session.reset();
  connection.reset();
  history.replaceState(null, '', '/');
  intentionalLeave = false;
}

export async function refreshMic() {
  micPipeline?.close();
  micPipeline = null;
  micStream?.getTracks().forEach((t) => t.stop());
  micStream = await getMicStream(settings.inputDeviceId || undefined);
  const processed = createMicPipeline(micStream);
  micPipeline = processed.pipeline;
  micPipeline.setInputVolume(settings.inputVolume);
  analyser = processed.analyser;
  await mesh?.addLocalAudio(processed.stream);
  applyMicTransmit();
  startVoiceActivity();
}
