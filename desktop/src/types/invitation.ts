export interface WorkspaceInvitation {
  id: string;
  inviter_id: string;
  email?: string;
  token: string;
  expires_at: string;
  accepted_at?: string;
  accepted_by?: string;
  max_uses: number;
  use_count: number;
  created_at: string;
  invite_url: string;
}

export interface CreateInvitationRequest {
  email?: string;
  max_uses: number;
  ttl_days: number;
}
