import { create } from "zustand";
import type { Mention } from "../types/mention";
import { apiFetch } from "../services/api";

interface MentionState {
  unreadMentions: Mention[];
  isLoading: boolean;

  fetchUnreadMentions: () => Promise<void>;
  markMentionsRead: (channelId: string) => Promise<void>;
  addMention: (mention: Mention) => void;
}

export const useMentionStore = create<MentionState>((set) => ({
  unreadMentions: [],
  isLoading: false,

  fetchUnreadMentions: async () => {
    set({ isLoading: true });
    try {
      const mentions = await apiFetch<Mention[]>("/mentions");
      set({ unreadMentions: mentions || [], isLoading: false });
    } catch {
      set({ isLoading: false });
    }
  },

  markMentionsRead: async (channelId) => {
    try {
      await apiFetch("/mentions/read", {
        method: "POST",
        body: JSON.stringify({ channel_id: channelId }),
      });
      set((state) => ({
        unreadMentions: state.unreadMentions.filter(
          (m) => m.channel_id !== channelId
        ),
      }));
    } catch {
      // ignore
    }
  },

  addMention: (mention) =>
    set((state) => ({
      unreadMentions: [mention, ...state.unreadMentions],
    })),
}));
