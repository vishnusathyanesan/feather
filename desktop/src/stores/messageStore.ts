import { create } from "zustand";
import type { Message, CreateMessageRequest, ReactionGroup } from "../types/message";
import { apiFetch } from "../services/api";

interface MessageState {
  messagesByChannel: Record<string, Message[]>;
  threadMessages: Record<string, Message[]>;
  hasMore: Record<string, boolean>;
  isLoading: boolean;

  fetchMessages: (channelId: string, before?: string) => Promise<void>;
  fetchThread: (messageId: string) => Promise<void>;
  sendMessage: (channelId: string, req: CreateMessageRequest) => Promise<Message>;
  editMessage: (channelId: string, messageId: string, content: string) => Promise<void>;
  deleteMessage: (channelId: string, messageId: string) => Promise<void>;
  addMessage: (channelId: string, message: Message) => void;
  updateMessage: (channelId: string, message: Message) => void;
  removeMessage: (channelId: string, messageId: string) => void;
  addReaction: (messageId: string, emoji: string) => Promise<void>;
  removeReaction: (messageId: string, emoji: string) => Promise<void>;
  updateReactionInPlace: (channelId: string, messageId: string, emoji: string, userId: string, added: boolean) => void;
}

const fetchingChannels = new Set<string>();

export const useMessageStore = create<MessageState>((set, get) => ({
  messagesByChannel: {},
  threadMessages: {},
  hasMore: {},
  isLoading: false,

  fetchMessages: async (channelId, before) => {
    const key = `${channelId}:${before || ""}`;
    if (fetchingChannels.has(key)) return;
    fetchingChannels.add(key);

    set({ isLoading: true });
    try {
      const params = new URLSearchParams({ limit: "50" });
      if (before) params.set("before", before);

      const messages = await apiFetch<Message[]>(
        `/channels/${channelId}/messages?${params}`
      );

      set((state) => {
        const existing = before ? state.messagesByChannel[channelId] || [] : [];
        const merged = before ? [...messages, ...existing] : messages;
        return {
          messagesByChannel: {
            ...state.messagesByChannel,
            [channelId]: merged,
          },
          hasMore: {
            ...state.hasMore,
            [channelId]: messages.length === 50,
          },
          isLoading: false,
        };
      });
    } catch {
      set({ isLoading: false });
    } finally {
      fetchingChannels.delete(key);
    }
  },

  fetchThread: async (messageId) => {
    const messages = await apiFetch<Message[]>(`/messages/${messageId}/thread`);
    set((state) => ({
      threadMessages: { ...state.threadMessages, [messageId]: messages },
    }));
  },

  sendMessage: async (channelId, req) => {
    const message = await apiFetch<Message>(`/channels/${channelId}/messages`, {
      method: "POST",
      body: JSON.stringify(req),
    });
    return message;
  },

  editMessage: async (channelId, messageId, content) => {
    const message = await apiFetch<Message>(
      `/channels/${channelId}/messages/${messageId}`,
      {
        method: "PATCH",
        body: JSON.stringify({ content }),
      }
    );
    get().updateMessage(channelId, message);
  },

  deleteMessage: async (channelId, messageId) => {
    await apiFetch(`/channels/${channelId}/messages/${messageId}`, {
      method: "DELETE",
    });
  },

  addMessage: (channelId, message) =>
    set((state) => {
      const existing = state.messagesByChannel[channelId] || [];
      // Deduplicate: skip if message already exists
      if (existing.some((m) => m.id === message.id)) return state;
      return {
        messagesByChannel: {
          ...state.messagesByChannel,
          [channelId]: [...existing, message],
        },
      };
    }),

  updateMessage: (channelId, message) =>
    set((state) => {
      const existing = state.messagesByChannel[channelId] || [];
      return {
        messagesByChannel: {
          ...state.messagesByChannel,
          [channelId]: existing.map((m) => (m.id === message.id ? message : m)),
        },
      };
    }),

  removeMessage: (channelId, messageId) =>
    set((state) => {
      const existing = state.messagesByChannel[channelId] || [];
      return {
        messagesByChannel: {
          ...state.messagesByChannel,
          [channelId]: existing.filter((m) => m.id !== messageId),
        },
      };
    }),

  addReaction: async (messageId, emoji) => {
    await apiFetch(`/messages/${messageId}/reactions`, {
      method: "POST",
      body: JSON.stringify({ emoji }),
    });
  },

  removeReaction: async (messageId, emoji) => {
    await apiFetch(`/messages/${messageId}/reactions/${encodeURIComponent(emoji)}`, {
      method: "DELETE",
    });
  },

  updateReactionInPlace: (channelId, messageId, emoji, userId, added) =>
    set((state) => {
      const existing = state.messagesByChannel[channelId];
      if (!existing) return state;

      return {
        messagesByChannel: {
          ...state.messagesByChannel,
          [channelId]: existing.map((m) => {
            if (m.id !== messageId) return m;
            const reactions = [...(m.reactions || [])];

            if (added) {
              const idx = reactions.findIndex((r) => r.emoji === emoji);
              if (idx !== -1) {
                const group = reactions[idx];
                if (!group.users.includes(userId)) {
                  reactions[idx] = { ...group, count: group.count + 1, users: [...group.users, userId] };
                }
              } else {
                reactions.push({ emoji, count: 1, users: [userId] });
              }
            } else {
              const idx = reactions.findIndex((r) => r.emoji === emoji);
              if (idx !== -1) {
                const group = reactions[idx];
                const newUsers = group.users.filter((u) => u !== userId);
                if (newUsers.length === 0) {
                  reactions.splice(idx, 1);
                } else {
                  reactions[idx] = { ...group, count: newUsers.length, users: newUsers };
                }
              }
            }

            return { ...m, reactions };
          }),
        },
      };
    }),
}));
