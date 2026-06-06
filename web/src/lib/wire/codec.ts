export const MAGIC = 0x48;

export const Msg = {
  CREATE_ROOM: 1,
  JOIN: 2,
  LEAVE: 3,
  OFFER: 4,
  ANSWER: 5,
  ICE: 6,
  ROOM_STATE: 7,
  MEMBER_UPDATE: 8,
  ADD_CHANNEL: 9,
  ERROR: 10,
  JOINED: 11,
  CREATED: 12,
  PEER_JOINED: 13,
  RENAME: 14,
  PING: 15,
  PONG: 16,
  KICK: 17,
  KICKED: 18,
  MODERATE_MEMBER: 19,
  PEER_LEFT: 20,
} as const;

export const TYPE_NAME: Record<number, string> = {
  [Msg.CREATE_ROOM]: 'create_room',
  [Msg.JOIN]: 'join',
  [Msg.LEAVE]: 'leave',
  [Msg.OFFER]: 'offer',
  [Msg.ANSWER]: 'answer',
  [Msg.ICE]: 'ice',
  [Msg.ROOM_STATE]: 'room_state',
  [Msg.MEMBER_UPDATE]: 'member_update',
  [Msg.ADD_CHANNEL]: 'add_channel',
  [Msg.ERROR]: 'error',
  [Msg.JOINED]: 'joined',
  [Msg.CREATED]: 'created',
  [Msg.PEER_JOINED]: 'peer_joined',
  [Msg.RENAME]: 'rename',
  [Msg.PING]: 'ping',
  [Msg.PONG]: 'pong',
  [Msg.KICK]: 'kick',
  [Msg.KICKED]: 'kicked',
  [Msg.MODERATE_MEMBER]: 'moderate_member',
  [Msg.PEER_LEFT]: 'peer_left',
};

export const NAME_TYPE: Record<string, number> = Object.fromEntries(
  Object.entries(TYPE_NAME).map(([code, name]) => [name, Number(code)]),
);

const HEADER = 6;
const MAX_STRING = 65536;

const FLAG_MUTED = 1;
const FLAG_DEAFENED = 2;
const FLAG_SPEAKING = 4;
const SIGNAL_SDP = 1;
const SIGNAL_CANDIDATE = 2;

const enc = new TextEncoder();
const dec = new TextDecoder();

function appendU32(buf: number[], n: number) {
  buf.push((n >>> 24) & 0xff, (n >>> 16) & 0xff, (n >>> 8) & 0xff, n & 0xff);
}

function readU32(data: Uint8Array, off: number) {
  return ((data[off] << 24) | (data[off + 1] << 16) | (data[off + 2] << 8) | data[off + 3]) >>> 0;
}

function appendString(buf: number[], s: string) {
  const bytes = enc.encode(s);
  if (bytes.length > MAX_STRING) throw new Error('string too long');
  appendU32(buf, bytes.length);
  for (const b of bytes) buf.push(b);
}

function readString(data: Uint8Array, off: number) {
  const n = readU32(data, off);
  off += 4;
  if (n > MAX_STRING || off + n > data.length) throw new Error('invalid string');
  return { value: dec.decode(data.subarray(off, off + n)), off: off + n };
}

function appendBytes(buf: number[], bytes: Uint8Array) {
  appendU32(buf, bytes.length);
  for (const b of bytes) buf.push(b);
}

function readBytes(data: Uint8Array, off: number) {
  const n = readU32(data, off);
  off += 4;
  if (off + n > data.length) throw new Error('invalid bytes');
  return { value: data.slice(off, off + n), off: off + n };
}

function appendPow(buf: number[], pow?: { id: string; nonce: number }) {
  if (!pow) {
    buf.push(0);
    return;
  }
  buf.push(1);
  appendString(buf, pow.id);
  const hi = Math.floor(pow.nonce / 0x100000000);
  const lo = pow.nonce >>> 0;
  buf.push(
    (hi >>> 24) & 0xff,
    (hi >>> 16) & 0xff,
    (hi >>> 8) & 0xff,
    hi & 0xff,
    (lo >>> 24) & 0xff,
    (lo >>> 16) & 0xff,
    (lo >>> 8) & 0xff,
    lo & 0xff,
  );
}

function readPow(data: Uint8Array, off: number) {
  if (off >= data.length) throw new Error('invalid pow');
  if (data[off] === 0) return { value: undefined, off: off + 1 };
  off++;
  const id = readString(data, off);
  off = id.off;
  if (off + 8 > data.length) throw new Error('invalid pow nonce');
  const hi = readU32(data, off);
  const lo = readU32(data, off + 4);
  return { value: { id: id.value, nonce: hi * 0x100000000 + lo }, off: off + 8 };
}

function memberFlags(m: { muted?: boolean; deafened?: boolean; speaking?: boolean }) {
  let f = 0;
  if (m.muted) f |= FLAG_MUTED;
  if (m.deafened) f |= FLAG_DEAFENED;
  if (m.speaking) f |= FLAG_SPEAKING;
  return f;
}

function readMember(id: string, name: string, flags: number) {
  return {
    id,
    name,
    muted: (flags & FLAG_MUTED) !== 0,
    deafened: (flags & FLAG_DEAFENED) !== 0,
    speaking: (flags & FLAG_SPEAKING) !== 0,
  };
}

export function encodeFrame(type: number, payload: Uint8Array): Uint8Array {
  const out = new Uint8Array(HEADER + payload.length);
  out[0] = MAGIC;
  out[1] = type;
  out[2] = (payload.length >>> 24) & 0xff;
  out[3] = (payload.length >>> 16) & 0xff;
  out[4] = (payload.length >>> 8) & 0xff;
  out[5] = payload.length & 0xff;
  out.set(payload, HEADER);
  return out;
}

export function decodeFrame(data: ArrayBuffer | Uint8Array): { type: number; payload: Uint8Array } {
  const view = data instanceof Uint8Array ? data : new Uint8Array(data);
  if (view.length < HEADER || view[0] !== MAGIC) throw new Error('invalid frame');
  const len = readU32(view, 2);
  if (HEADER + len > view.length) throw new Error('invalid frame length');
  return { type: view[1], payload: view.slice(HEADER, HEADER + len) };
}

function payloadBuf(): number[] {
  return [];
}

function toPayload(buf: number[]) {
  return Uint8Array.from(buf);
}

export function encodeMessage(typeName: string, payload?: unknown): Uint8Array {
  const type = NAME_TYPE[typeName];
  if (!type) throw new Error(`unknown message type: ${typeName}`);
  return encodeFrame(type, encodePayload(type, payload));
}

export function decodeMessage(data: ArrayBuffer | Uint8Array): { type: string; payload: unknown } {
  const frame = decodeFrame(data);
  const type = TYPE_NAME[frame.type];
  if (!type) throw new Error(`unknown wire type: ${frame.type}`);
  return { type, payload: decodePayload(frame.type, frame.payload) };
}

function encodePayload(type: number, payload: unknown): Uint8Array {
  switch (type) {
    case Msg.CREATE_ROOM:
      return encodeCreateRoom(payload as CreateRoomPayload);
    case Msg.JOIN:
      return encodeJoin(payload as JoinPayload);
    case Msg.LEAVE:
      return new Uint8Array(0);
    case Msg.OFFER:
    case Msg.ANSWER:
      return encodeSignal(payload as SignalPayload, SIGNAL_SDP);
    case Msg.ICE:
      return encodeSignal(payload as SignalPayload, SIGNAL_CANDIDATE);
    case Msg.PING:
      return encodePing((payload as PingPayload).t);
    case Msg.MEMBER_UPDATE:
      return encodeMemberUpdate(payload as MemberUpdatePayload);
    case Msg.RENAME:
      return encodeRename(payload as RenamePayload);
    case Msg.KICK:
    case Msg.PEER_JOINED:
    case Msg.PEER_LEFT:
      return encodePeerRef(payload as PeerRefPayload);
    case Msg.MODERATE_MEMBER:
      return encodeModerateMember(payload as ModerateMemberPayload);
    case Msg.ADD_CHANNEL:
      return encodeAddChannel(payload as AddChannelPayload);
    default:
      throw new Error(`encode not implemented for type ${type}`);
  }
}

function decodePayload(type: number, data: Uint8Array): unknown {
  switch (type) {
    case Msg.CREATED:
      return decodeCreated(data);
    case Msg.JOINED:
      return decodeJoined(data);
    case Msg.ROOM_STATE:
      return decodeRoom(data);
    case Msg.OFFER:
    case Msg.ANSWER:
      return decodeSignal(data, 'sdp');
    case Msg.ICE:
      return decodeSignal(data, 'candidate');
    case Msg.PONG:
      return decodePing(data);
    case Msg.ERROR:
      return decodeError(data);
    case Msg.MEMBER_UPDATE:
      return decodeMemberUpdate(data);
    case Msg.PEER_JOINED:
    case Msg.PEER_LEFT:
    case Msg.KICK:
      return decodePeerRef(data);
    case Msg.KICKED:
      return null;
    default:
      return null;
  }
}

type PowPayload = { id: string; nonce: number };
type CreateRoomPayload = { name: string; password?: string; pow?: PowPayload };
type JoinPayload = {
  roomId: string;
  invite: string;
  password?: string;
  name: string;
  resumePeerId?: string;
  resumeToken?: string;
  pow?: PowPayload;
};
type SignalPayload = {
  to: string;
  from?: string;
  sdp?: string;
  candidate?: RTCIceCandidateInit;
  nonce?: string;
  sig?: string;
};
type MemberUpdatePayload = {
  peerId: string;
  muted: boolean;
  deafened: boolean;
  speaking: boolean;
};
type RenamePayload = { name: string };
type PeerRefPayload = { peerId: string };
type ModerateMemberPayload = { peerId: string; muted: boolean; deafened: boolean };
type AddChannelPayload = { id: string; name: string };
type PingPayload = { t: number };

function encodeCreateRoom(p: CreateRoomPayload) {
  const buf = payloadBuf();
  appendString(buf, p.name);
  appendString(buf, p.password ?? '');
  appendPow(buf, p.pow ? { id: p.pow.id, nonce: p.pow.nonce } : undefined);
  return toPayload(buf);
}

function encodeJoin(p: JoinPayload) {
  const buf = payloadBuf();
  appendString(buf, p.roomId);
  appendString(buf, p.invite);
  appendString(buf, p.password ?? '');
  appendString(buf, p.name);
  appendString(buf, p.resumePeerId ?? '');
  appendString(buf, p.resumeToken ?? '');
  appendPow(buf, p.pow ? { id: p.pow.id, nonce: p.pow.nonce } : undefined);
  return toPayload(buf);
}

function encodeSignal(p: SignalPayload, kind: number) {
  const buf = payloadBuf();
  appendString(buf, p.to);
  appendString(buf, p.from ?? '');
  appendString(buf, p.nonce ?? '');
  appendString(buf, p.sig ?? '');
  buf.push(kind);
  if (kind === SIGNAL_CANDIDATE) {
    const body = enc.encode(JSON.stringify(p.candidate ?? null));
    appendBytes(buf, body);
  } else {
    appendBytes(buf, enc.encode(p.sdp ?? ''));
  }
  return toPayload(buf);
}

function encodeMemberUpdate(p: MemberUpdatePayload) {
  const buf = payloadBuf();
  appendString(buf, p.peerId);
  buf.push(memberFlags(p));
  return toPayload(buf);
}

function encodeRename(p: RenamePayload) {
  const buf = payloadBuf();
  appendString(buf, p.name);
  return toPayload(buf);
}

function encodePeerRef(p: PeerRefPayload) {
  const buf = payloadBuf();
  appendString(buf, p.peerId);
  return toPayload(buf);
}

function encodeModerateMember(p: ModerateMemberPayload) {
  const buf = payloadBuf();
  appendString(buf, p.peerId);
  let flags = 0;
  if (p.muted) flags |= FLAG_MUTED;
  if (p.deafened) flags |= FLAG_DEAFENED;
  buf.push(flags);
  return toPayload(buf);
}

function encodeAddChannel(p: AddChannelPayload) {
  const buf = payloadBuf();
  appendString(buf, p.id);
  appendString(buf, p.name);
  return toPayload(buf);
}

function encodePing(t: number) {
  const buf = new ArrayBuffer(8);
  const view = new DataView(buf);
  view.setBigUint64(0, BigInt(t));
  return new Uint8Array(buf);
}

function decodeRoomAt(data: Uint8Array, off: number) {
  const id = readString(data, off);
  off = id.off;
  const name = readString(data, off);
  off = name.off;
  const hostId = readString(data, off);
  off = hostId.off;
  const chCount = (data[off] << 8) | data[off + 1];
  off += 2;
  const channels: { id: string; name: string }[] = [];
  for (let i = 0; i < chCount; i++) {
    const chId = readString(data, off);
    off = chId.off;
    const chName = readString(data, off);
    off = chName.off;
    channels.push({ id: chId.value, name: chName.value });
  }
  const mCount = (data[off] << 8) | data[off + 1];
  off += 2;
  const members: {
    id: string;
    name: string;
    muted: boolean;
    deafened: boolean;
    speaking: boolean;
  }[] = [];
  for (let i = 0; i < mCount; i++) {
    const mid = readString(data, off);
    off = mid.off;
    const mname = readString(data, off);
    off = mname.off;
    members.push(readMember(mid.value, mname.value, data[off]));
    off++;
  }
  return {
    room: { id: id.value, name: name.value, hostId: hostId.value, channels, members },
    off,
  };
}

function decodeRoom(data: Uint8Array) {
  return decodeRoomAt(data, 0).room;
}

function decodeJoined(data: Uint8Array) {
  let off = 0;
  const peerId = readString(data, off);
  off = peerId.off;
  const resumeToken = readString(data, off);
  off = resumeToken.off;
  const roomPart = decodeRoomAt(data, off);
  off = roomPart.off;
  const peerCount = (data[off] << 8) | data[off + 1];
  off += 2;
  const peers: string[] = [];
  for (let i = 0; i < peerCount; i++) {
    const peer = readString(data, off);
    off = peer.off;
    peers.push(peer.value);
  }
  const iceCount = (data[off] << 8) | data[off + 1];
  off += 2;
  const iceServers: { urls: string[]; username?: string; credential?: string }[] = [];
  for (let i = 0; i < iceCount; i++) {
    const urlCount = (data[off] << 8) | data[off + 1];
    off += 2;
    const urls: string[] = [];
    for (let j = 0; j < urlCount; j++) {
      const url = readString(data, off);
      off = url.off;
      urls.push(url.value);
    }
    const username = readString(data, off);
    off = username.off;
    const credential = readString(data, off);
    off = credential.off;
    iceServers.push({
      urls,
      username: username.value || undefined,
      credential: credential.value || undefined,
    });
  }
  return {
    peerId: peerId.value,
    resumeToken: resumeToken.value,
    room: roomPart.room,
    peers,
    iceServers,
  };
}

function decodeCreated(data: Uint8Array) {
  let off = 0;
  const roomId = readString(data, off);
  off = roomId.off;
  const invite = readString(data, off);
  off = invite.off;
  const roomKey = readString(data, off);
  off = roomKey.off;
  const expiresAt = Number(readU32(data, off) >>> 0) * 0x100000000 + readU32(data, off + 4);
  return { roomId: roomId.value, invite: invite.value, roomKey: roomKey.value, expiresAt };
}

function decodeSignal(data: Uint8Array, field: 'sdp' | 'candidate') {
  let off = 0;
  const to = readString(data, off);
  off = to.off;
  const from = readString(data, off);
  off = from.off;
  const nonce = readString(data, off);
  off = nonce.off;
  const sig = readString(data, off);
  off = sig.off;
  const kind = data[off];
  off++;
  const body = readBytes(data, off);
  const out: SignalPayload = {
    to: to.value,
    from: from.value || undefined,
    nonce: nonce.value,
    sig: sig.value,
  };
  if (field === 'candidate') {
    out.candidate = JSON.parse(dec.decode(body.value)) as RTCIceCandidateInit;
  } else {
    out.sdp = dec.decode(body.value);
  }
  void kind;
  return out;
}

function decodeMemberUpdate(data: Uint8Array) {
  let off = 0;
  const peerId = readString(data, off);
  off = peerId.off;
  const flags = data[off];
  const m = readMember(peerId.value, '', flags);
  return { peerId: m.id, muted: m.muted, deafened: m.deafened, speaking: m.speaking };
}

function decodePeerRef(data: Uint8Array) {
  return { peerId: readString(data, 0).value };
}

function decodeError(data: Uint8Array) {
  return { message: readString(data, 0).value };
}

function decodePing(data: Uint8Array) {
  if (data.length !== 8) throw new Error('invalid ping');
  const view = new DataView(data.buffer, data.byteOffset, data.byteLength);
  return { t: Number(view.getBigUint64(0)) };
}
