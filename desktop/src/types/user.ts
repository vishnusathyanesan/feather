export interface User {
  id: string;
  email: string;
  name: string;
  avatar_url: string;
  role: "admin" | "member" | "bot";
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface RegisterRequest {
  email: string;
  name: string;
  password: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface AuthResponse {
  user: User;
  access_token: string;
  refresh_token: string;
}

export interface UpdateUserRequest {
  name?: string;
  avatar_url?: string;
}

export interface GoogleOAuthRequest {
  credential: string;
}
