import { User } from "./user";

export interface ReactionGroup {
  emoji: string;
  count: number;
  users: string[];
}

export interface FileAttachment {
  id: string;
  message_id?: string;
  channel_id: string;
  user_id: string;
  filename: string;
  content_type: string;
  size_bytes: number;
  download_url?: string;
  created_at: string;
}

export interface Message {
  id: string;
  channel_id: string;
  user_id: string;
  parent_id?: string;
  content: string;
  is_alert: boolean;
  alert_severity?: "info" | "warning" | "critical";
  alert_metadata?: Record<string, unknown>;
  edited_at?: string;
  deleted_at?: string;
  created_at: string;
  user?: User;
  reactions?: ReactionGroup[];
  attachments?: FileAttachment[];
  reply_count: number;
}

export interface CreateMessageRequest {
  content: string;
  parent_id?: string;
  attachment_ids?: string[];
}

export interface UpdateMessageRequest {
  content: string;
}
