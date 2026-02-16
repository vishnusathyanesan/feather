import type { User } from "./user";
import type { Message } from "./message";

export interface UserGroup {
  id: string;
  name: string;
  description: string;
  creator_id: string;
  created_at: string;
  updated_at: string;
  members?: User[];
}

export interface Mention {
  id: string;
  message_id: string;
  channel_id: string;
  mentioned_user_id?: string;
  mentioned_group_id?: string;
  mention_type: "user" | "group" | "channel" | "here";
  is_read: boolean;
  created_at: string;
  message?: Message;
}
