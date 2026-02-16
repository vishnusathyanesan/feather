export type CallStatus = "ringing" | "in_progress" | "ended" | "missed" | "declined";
export type CallType = "audio" | "video";

export interface Call {
  id: string;
  channel_id: string;
  initiator_id: string;
  call_type: CallType;
  status: CallStatus;
  accepted_by?: string;
  started_at?: string;
  ended_at?: string;
  created_at: string;
}

export interface CallParticipant {
  call_id: string;
  user_id: string;
  joined_at?: string;
  left_at?: string;
}

export interface ICEServerConfig {
  urls: string[];
  username?: string;
  credential?: string;
}

export interface RTCConfig {
  ice_servers: ICEServerConfig[];
  enabled: boolean;
}
