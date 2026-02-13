import { create } from "zustand";
import type { User, LoginRequest, RegisterRequest, AuthResponse } from "../types/user";
import { apiFetch, setTokens, clearTokens, setAuthFailureHandler } from "../services/api";

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;

  login: (req: LoginRequest) => Promise<void>;
  register: (req: RegisterRequest) => Promise<void>;
  logout: () => Promise<void>;
  loadUser: () => Promise<void>;
  setUser: (user: User) => void;
  clearError: () => void;
}

export const useAuthStore = create<AuthState>((set, get) => {
  setAuthFailureHandler(() => {
    set({ user: null, isAuthenticated: false });
    clearTokens();
    localStorage.removeItem("feather_refresh_token");
  });

  return {
    user: null,
    isAuthenticated: false,
    isLoading: true,
    error: null,

    login: async (req) => {
      try {
        set({ isLoading: true, error: null });
        const res = await apiFetch<AuthResponse>("/auth/login", {
          method: "POST",
          body: JSON.stringify(req),
        });
        setTokens(res.access_token, res.refresh_token);
        localStorage.setItem("feather_refresh_token", res.refresh_token);
        set({ user: res.user, isAuthenticated: true, isLoading: false });
      } catch (err) {
        set({ error: (err as Error).message, isLoading: false });
        throw err;
      }
    },

    register: async (req) => {
      try {
        set({ isLoading: true, error: null });
        const res = await apiFetch<AuthResponse>("/auth/register", {
          method: "POST",
          body: JSON.stringify(req),
        });
        setTokens(res.access_token, res.refresh_token);
        localStorage.setItem("feather_refresh_token", res.refresh_token);
        set({ user: res.user, isAuthenticated: true, isLoading: false });
      } catch (err) {
        set({ error: (err as Error).message, isLoading: false });
        throw err;
      }
    },

    logout: async () => {
      const refreshToken = localStorage.getItem("feather_refresh_token");
      if (refreshToken) {
        try {
          await apiFetch("/auth/logout", {
            method: "POST",
            body: JSON.stringify({ refresh_token: refreshToken }),
          });
        } catch {
          // ignore
        }
      }
      clearTokens();
      localStorage.removeItem("feather_refresh_token");
      set({ user: null, isAuthenticated: false });
    },

    loadUser: async () => {
      const refreshToken = localStorage.getItem("feather_refresh_token");
      if (!refreshToken) {
        set({ isLoading: false });
        return;
      }

      try {
        const res = await apiFetch<AuthResponse>("/auth/refresh", {
          method: "POST",
          body: JSON.stringify({ refresh_token: refreshToken }),
        });
        setTokens(res.access_token, res.refresh_token);
        localStorage.setItem("feather_refresh_token", res.refresh_token);
        set({ user: res.user, isAuthenticated: true, isLoading: false });
      } catch {
        localStorage.removeItem("feather_refresh_token");
        set({ isLoading: false });
      }
    },

    setUser: (user) => set({ user }),
    clearError: () => set({ error: null }),
  };
});
