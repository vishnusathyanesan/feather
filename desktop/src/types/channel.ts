import { User } from "./user";

export interface Channel {
  id: string;
  name: string | null;
  topic: string;
  description: string;
  type: "public" | "private" | "system" | "dm" | "group_dm";
  is_readonly: boolean;
  creator_id?: string;
  created_at: string;
  updated_at: string;
  unread_count?: number;
  member_count?: number;
  members?: User[];
}

export interface CreateChannelRequest {
  name: string;
  topic?: string;
  description?: string;
  type: "public" | "private" | "system";
  is_readonly?: boolean;
}

export interface UpdateChannelRequest {
  name?: string;
  topic?: string;
  description?: string;
  is_readonly?: boolean;
}

export interface ChannelMember {
  channel_id: string;
  user_id: string;
  role: string;
  last_read_at: string;
  joined_at: string;
  user?: User;
}
