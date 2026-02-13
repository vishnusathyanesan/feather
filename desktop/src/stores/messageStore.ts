import { create } from "zustand";
import type { Message, CreateMessageRequest } from "../types/message";
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
}

export const useMessageStore = create<MessageState>((set, get) => ({
  messagesByChannel: {},
  threadMessages: {},
  hasMore: {},
  isLoading: false,

  fetchMessages: async (channelId, before) => {
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
}));
