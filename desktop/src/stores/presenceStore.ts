import { create } from "zustand";

interface PresenceState {
  onlineUsers: Set<string>;
  typingUsers: Record<string, { userName: string; timeout: ReturnType<typeof setTimeout> }>;

  setOnline: (userId: string) => void;
  setOffline: (userId: string) => void;
  setTyping: (channelId: string, userId: string, userName: string) => void;
  clearTyping: (channelId: string, userId: string) => void;
  isOnline: (userId: string) => boolean;
}

export const usePresenceStore = create<PresenceState>((set, get) => ({
  onlineUsers: new Set(),
  typingUsers: {},

  setOnline: (userId) =>
    set((state) => {
      const next = new Set(state.onlineUsers);
      next.add(userId);
      return { onlineUsers: next };
    }),

  setOffline: (userId) =>
    set((state) => {
      const next = new Set(state.onlineUsers);
      next.delete(userId);
      return { onlineUsers: next };
    }),

  setTyping: (channelId, userId, userName) =>
    set((state) => {
      const key = `${channelId}:${userId}`;
      const existing = state.typingUsers[key];
      if (existing) clearTimeout(existing.timeout);

      const timeout = setTimeout(() => {
        get().clearTyping(channelId, userId);
      }, 3000);

      return {
        typingUsers: { ...state.typingUsers, [key]: { userName, timeout } },
      };
    }),

  clearTyping: (channelId, userId) =>
    set((state) => {
      const key = `${channelId}:${userId}`;
      const { [key]: _, ...rest } = state.typingUsers;
      return { typingUsers: rest };
    }),

  isOnline: (userId) => get().onlineUsers.has(userId),
}));
