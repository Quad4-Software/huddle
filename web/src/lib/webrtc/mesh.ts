import { createPeer } from './peer';
import { encrypt, decrypt, decryptText } from '../crypto/e2e';
import type { ChatMessage, AttachmentMeta, ControlMessage } from '../types';

export type SignalSend = (
  type: 'offer' | 'answer' | 'ice',
  to: string,
  data: { sdp?: string; candidate?: RTCIceCandidateInit },
) => void;

export type MeshEvents = {
  onMessage: (msg: ChatMessage) => void;
  onAttachment: (meta: AttachmentMeta, blob: Blob) => void;
  onControl: (msg: ControlMessage) => void;
  onTrack: (peerId: string, stream: MediaStream) => void;
  onTrackRemoved: (peerId: string, stream: MediaStream | null) => void;
  onPeerConnected: (peerId: string, connected: boolean) => void;
  onMeshReady: () => void;
};

const CHUNK_SIZE = 16384;

type PeerChannels = { chat?: RTCDataChannel; files?: RTCDataChannel; control?: RTCDataChannel };

export class Mesh {
  private peers = new Map<string, RTCPeerConnection>();
  private channels = new Map<string, PeerChannels>();
  private pendingICE = new Map<string, RTCIceCandidateInit[]>();
  private chatQueue: string[] = [];
  private fileQueue: string[] = [];
  private controlQueue: string[] = [];
  private fileBuffers = new Map<string, { meta: AttachmentMeta; chunks: Uint8Array[] }>();
  private localId: string;
  private localName: string;
  private cryptoKey: CryptoKey;
  private localStream: MediaStream | null = null;
  private screenStream: MediaStream | null = null;
  private signal: SignalSend;
  private events: MeshEvents;

  constructor(
    localId: string,
    localName: string,
    cryptoKey: CryptoKey,
    signal: SignalSend,
    events: MeshEvents,
  ) {
    this.localId = localId;
    this.localName = localName;
    this.cryptoKey = cryptoKey;
    this.signal = signal;
    this.events = events;
  }

  hasOpenChannels(): boolean {
    for (const [, chs] of this.channels) {
      if (chs.chat?.readyState === 'open') return true;
    }
    return this.channels.size === 0;
  }

  async addLocalAudio(stream: MediaStream) {
    this.localStream = stream;
    for (const [, pc] of this.peers) {
      this.attachLocalTracks(pc);
    }
  }

  async connectTo(peerId: string) {
    if (this.peers.has(peerId)) return;
    const pc = createPeer();
    this.registerPeer(peerId, pc);
    this.bindLocalChannels(peerId, pc);
    if (this.localStream) this.attachLocalTracks(pc);
    if (this.screenStream) this.attachScreenTrack(pc);
    const offer = await pc.createOffer();
    await pc.setLocalDescription(offer);
    this.signal('offer', peerId, { sdp: offer.sdp ?? undefined });
  }

  async handleOffer(from: string, sdp: string) {
    let pc = this.peers.get(from);
    const isRenegotiation = !!pc?.currentRemoteDescription;
    if (!pc) {
      pc = createPeer();
      this.registerPeer(from, pc);
      if (this.localStream) this.attachLocalTracks(pc);
      if (this.screenStream) this.attachScreenTrack(pc);
    }
    await pc.setRemoteDescription({ type: 'offer', sdp });
    const answer = await pc.createAnswer();
    await pc.setLocalDescription(answer);
    this.signal('answer', from, { sdp: answer.sdp ?? undefined });
    this.flushICE(from, pc);
    if (isRenegotiation) return;
  }

  async handleAnswer(from: string, sdp: string) {
    const pc = this.peers.get(from);
    if (!pc) return;
    await pc.setRemoteDescription({ type: 'answer', sdp });
    this.flushICE(from, pc);
  }

  async handleICE(from: string, candidate: RTCIceCandidateInit) {
    const pc = this.peers.get(from);
    if (!pc || !pc.remoteDescription) {
      const q = this.pendingICE.get(from) ?? [];
      q.push(candidate);
      this.pendingICE.set(from, q);
      return;
    }
    await pc.addIceCandidate(candidate);
  }

  private flushICE(peerId: string, pc: RTCPeerConnection) {
    const q = this.pendingICE.get(peerId) ?? [];
    this.pendingICE.delete(peerId);
    for (const c of q) {
      pc.addIceCandidate(c).catch(() => {});
    }
  }

  private registerPeer(peerId: string, pc: RTCPeerConnection) {
    this.peers.set(peerId, pc);
    this.channels.set(peerId, {});

    pc.onicecandidate = (ev) => {
      if (ev.candidate) {
        this.signal('ice', peerId, { candidate: ev.candidate.toJSON() });
      }
    };

    pc.ontrack = (ev) => {
      const stream = ev.streams[0] ?? new MediaStream([ev.track]);
      ev.track.onended = () => {
        const live = stream.getTracks().filter((t) => t.readyState === 'live');
        if (live.length === 0) {
          this.events.onTrackRemoved(peerId, stream);
          return;
        }
        this.events.onTrack(peerId, stream);
      };
      this.events.onTrack(peerId, stream);
    };

    pc.ondatachannel = (ev) => {
      this.wireChannel(peerId, ev.channel);
    };

    pc.onconnectionstatechange = () => {
      const connected = pc.connectionState === 'connected';
      if (pc.connectionState === 'failed' || pc.connectionState === 'closed') {
        this.removePeer(peerId);
        this.events.onPeerConnected(peerId, false);
        return;
      }
      this.events.onPeerConnected(peerId, connected);
    };
  }

  private bindLocalChannels(peerId: string, pc: RTCPeerConnection) {
    const chat = pc.createDataChannel('chat', { ordered: true });
    const files = pc.createDataChannel('files', { ordered: true });
    const control = pc.createDataChannel('control', { ordered: true });
    this.wireChannel(peerId, chat);
    this.wireChannel(peerId, files);
    this.wireChannel(peerId, control);
  }

  private wireChannel(peerId: string, ch: RTCDataChannel) {
    const entry = this.channels.get(peerId) ?? {};
    if (ch.label === 'chat') entry.chat = ch;
    if (ch.label === 'files') entry.files = ch;
    if (ch.label === 'control') entry.control = ch;
    this.channels.set(peerId, entry);

    ch.onopen = () => {
      if (ch.label === 'chat') {
        this.flushChatQueue();
        this.events.onMeshReady();
      }
      if (ch.label === 'files') this.flushFileQueue();
      if (ch.label === 'control') this.flushControlQueue();
    };

    if (ch.label === 'chat') {
      ch.onmessage = async (ev) => {
        try {
          const plain = await decryptText(this.cryptoKey, ev.data as string);
          const msg = JSON.parse(plain) as ChatMessage;
          this.events.onMessage(msg);
        } catch {}
      };
    }

    if (ch.label === 'files') {
      ch.onmessage = async (ev) => {
        await this.handleFileMessage(ev.data as string);
      };
    }

    if (ch.label === 'control') {
      ch.onmessage = async (ev) => {
        try {
          const plain = await decryptText(this.cryptoKey, ev.data as string);
          const msg = JSON.parse(plain) as ControlMessage;
          this.events.onControl(msg);
        } catch {}
      };
    }
  }

  private flushChatQueue() {
    if (this.chatQueue.length === 0) return;
    const queued = [...this.chatQueue];
    this.chatQueue = [];
    for (const payload of queued) {
      this.sendToAllChat(payload, true);
    }
  }

  private flushFileQueue() {
    if (this.fileQueue.length === 0) return;
    const queued = [...this.fileQueue];
    this.fileQueue = [];
    for (const payload of queued) {
      this.sendToAllFiles(payload, true);
    }
  }

  private flushControlQueue() {
    if (this.controlQueue.length === 0) return;
    const queued = [...this.controlQueue];
    this.controlQueue = [];
    for (const payload of queued) {
      this.sendToAllControl(payload, true);
    }
  }

  private attachLocalTracks(pc: RTCPeerConnection) {
    if (!this.localStream) return;
    for (const track of this.localStream.getTracks()) {
      const existing = pc
        .getSenders()
        .find((s) => s.track?.kind === track.kind && track.kind === 'audio');
      if (existing) {
        existing.replaceTrack(track);
      } else {
        pc.addTrack(track, this.localStream);
      }
    }
  }

  private attachScreenTrack(pc: RTCPeerConnection) {
    if (!this.screenStream) return;
    for (const track of this.screenStream.getTracks()) {
      const existing = pc.getSenders().find((s) => s.track?.id === track.id);
      if (existing) {
        existing.replaceTrack(track);
      } else {
        pc.addTrack(track, this.screenStream);
      }
    }
  }

  async broadcastMessage(channelId: string, text: string) {
    const msg: ChatMessage = {
      id: crypto.randomUUID(),
      channelId,
      authorId: this.localId,
      authorName: this.localName,
      text,
      timestamp: Date.now(),
    };
    const payload = await encrypt(this.cryptoKey, JSON.stringify(msg));
    this.sendToAllChat(payload);
    this.events.onMessage(msg);
  }

  async broadcastFile(channelId: string, file: File) {
    const meta: AttachmentMeta = {
      id: crypto.randomUUID(),
      name: file.name,
      mime: file.type || 'application/octet-stream',
      size: file.size,
    };
    const msg: ChatMessage = {
      id: crypto.randomUUID(),
      channelId,
      authorId: this.localId,
      authorName: this.localName,
      text: '',
      timestamp: Date.now(),
      attachment: meta,
    };
    const buf = new Uint8Array(await file.arrayBuffer());
    const blob = new Blob([buf], { type: meta.mime });
    this.events.onAttachment(meta, blob);

    const msgPayload = await encrypt(this.cryptoKey, JSON.stringify(msg));
    this.sendToAllChat(msgPayload);
    this.events.onMessage(msg);
    const header = await encrypt(this.cryptoKey, JSON.stringify({ type: 'file-start', meta }));
    this.sendToAllFiles(header);

    for (let i = 0; i < buf.length; i += CHUNK_SIZE) {
      const chunk = buf.slice(i, i + CHUNK_SIZE);
      const enc = await encrypt(this.cryptoKey, chunk);
      const chunkHeader = await encrypt(
        this.cryptoKey,
        JSON.stringify({ type: 'file-chunk', id: meta.id }),
      );
      this.sendToAllFiles(chunkHeader + '|' + enc);
    }

    const end = await encrypt(this.cryptoKey, JSON.stringify({ type: 'file-end', id: meta.id }));
    this.sendToAllFiles(end);
  }

  private async handleFileMessage(data: string) {
    if (data.includes('|')) {
      const sep = data.indexOf('|');
      const hdr = data.slice(0, sep);
      const body = data.slice(sep + 1);
      const headerPlain = await decryptText(this.cryptoKey, hdr);
      const header = JSON.parse(headerPlain) as { type: string; id: string };
      if (header.type === 'file-chunk') {
        const bytes = await decrypt(this.cryptoKey, body);
        const buf = this.fileBuffers.get(header.id);
        if (buf) buf.chunks.push(bytes);
      }
      return;
    }
    const plain = await decryptText(this.cryptoKey, data);
    const parsed = JSON.parse(plain) as {
      type: string;
      meta?: AttachmentMeta;
      id?: string;
    };
    if (parsed.type === 'file-start' && parsed.meta) {
      this.fileBuffers.set(parsed.meta.id, { meta: parsed.meta, chunks: [] });
    }
    if (parsed.type === 'file-end' && parsed.id) {
      const buf = this.fileBuffers.get(parsed.id);
      if (!buf) return;
      const total = buf.chunks.reduce((n, c) => n + c.length, 0);
      const merged = new Uint8Array(total);
      let offset = 0;
      for (const c of buf.chunks) {
        merged.set(c, offset);
        offset += c.length;
      }
      const blob = new Blob([merged], { type: buf.meta.mime });
      this.events.onAttachment(buf.meta, blob);
      this.fileBuffers.delete(parsed.id);
    }
  }

  private sendToAllChat(payload: string, fromQueue = false) {
    let sent = false;
    for (const [, chs] of this.channels) {
      if (chs.chat?.readyState === 'open') {
        chs.chat.send(payload);
        sent = true;
      }
    }
    if (!sent && this.channels.size > 0 && !fromQueue) {
      this.chatQueue.push(payload);
    }
  }

  private sendToAllFiles(payload: string, fromQueue = false) {
    let sent = false;
    for (const [, chs] of this.channels) {
      if (chs.files?.readyState === 'open') {
        chs.files.send(payload);
        sent = true;
      }
    }
    if (!sent && this.channels.size > 0 && !fromQueue) {
      this.fileQueue.push(payload);
    }
  }

  async broadcastControl(msg: ControlMessage) {
    const payload = await encrypt(this.cryptoKey, JSON.stringify(msg));
    this.sendToAllControl(payload);
  }

  private sendToAllControl(payload: string, fromQueue = false) {
    let sent = false;
    for (const [, chs] of this.channels) {
      if (chs.control?.readyState === 'open') {
        chs.control.send(payload);
        sent = true;
      }
    }
    if (!sent && this.channels.size > 0 && !fromQueue) {
      this.controlQueue.push(payload);
    }
  }

  async startScreenShare(stream: MediaStream) {
    this.screenStream = stream;
    for (const [peerId, pc] of this.peers) {
      this.attachScreenTrack(pc);
      const offer = await pc.createOffer();
      await pc.setLocalDescription(offer);
      this.signal('offer', peerId, { sdp: offer.sdp ?? undefined });
    }
    stream.getVideoTracks()[0]?.addEventListener('ended', () => this.stopScreenShare());
  }

  stopScreenShare() {
    if (!this.screenStream) return;
    const trackIds = new Set(this.screenStream.getTracks().map((t) => t.id));
    this.screenStream.getTracks().forEach((t) => t.stop());
    this.screenStream = null;
    for (const [, pc] of this.peers) {
      for (const sender of pc.getSenders()) {
        if (sender.track && trackIds.has(sender.track.id)) {
          sender.replaceTrack(null);
        }
      }
    }
  }

  setMuted(muted: boolean) {
    this.localStream?.getAudioTracks().forEach((t) => (t.enabled = !muted));
  }

  setLocalName(name: string) {
    this.localName = name;
  }

  removePeer(peerId: string) {
    const pc = this.peers.get(peerId);
    if (pc) {
      pc.close();
      this.peers.delete(peerId);
      this.channels.delete(peerId);
      this.events.onTrackRemoved(peerId, null);
    }
  }

  destroy() {
    for (const [, pc] of this.peers) pc.close();
    this.peers.clear();
    this.channels.clear();
    this.chatQueue = [];
    this.fileQueue = [];
    this.controlQueue = [];
    this.localStream = null;
    this.screenStream = null;
  }
}
