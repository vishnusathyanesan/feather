export type EventType =
  | "message.new"
  | "message.updated"
  | "message.deleted"
  | "reaction.added"
  | "reaction.removed"
  | "typing"
  | "presence.update"
  | "channel.created"
  | "channel.updated"
  | "channel.deleted"
  | "member.joined"
  | "member.left";

export interface WebSocketEvent {
  type: EventType;
  channel_id?: string;
  payload: unknown;
}

export interface TypingPayload {
  user_id: string;
  channel_id: string;
  user_name: string;
}

export interface PresencePayload {
  user_id: string;
  online: boolean;
}
