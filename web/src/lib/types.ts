export interface Channel {
  id: string;
  name: string;
}

export interface Member {
  id: string;
  name: string;
  muted: boolean;
  deafened: boolean;
  speaking: boolean;
}

export interface RoomState {
  id: string;
  name: string;
  hostId?: string;
  channels: Channel[];
  members: Member[];
}

export interface PowPayload {
  id: string;
  nonce: number;
}

export interface ChatMessage {
  id: string;
  channelId: string;
  authorId: string;
  authorName: string;
  text: string;
  timestamp: number;
  attachment?: AttachmentMeta;
}

export interface ReactionState {
  emoji: string;
  peerIds: string[];
}

export interface ControlReaction {
  kind: 'reaction';
  messageId: string;
  emoji: string;
  peerId: string;
  add: boolean;
}

export interface ControlWatch {
  kind: 'watch';
  shareId: string;
  peerId: string;
  watching: boolean;
}

export type ControlMessage = ControlReaction | ControlWatch;

export interface AttachmentMeta {
  id: string;
  name: string;
  mime: string;
  size: number;
}

export interface ScreenShare {
  peerId: string;
  stream: MediaStream;
}

export type View = 'landing' | 'room' | 'settings';
