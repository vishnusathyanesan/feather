import { create } from "zustand";
import type { Channel, CreateChannelRequest } from "../types/channel";
import { apiFetch } from "../services/api";

interface ChannelState {
  channels: Channel[];
  activeChannelId: string | null;
  isLoading: boolean;

  fetchChannels: () => Promise<void>;
  setActiveChannel: (id: string) => void;
  createChannel: (req: CreateChannelRequest) => Promise<Channel>;
  joinChannel: (id: string) => Promise<void>;
  leaveChannel: (id: string) => Promise<void>;
  markRead: (id: string) => Promise<void>;
  updateChannel: (channel: Channel) => void;
  addChannel: (channel: Channel) => void;
  removeChannel: (id: string) => void;
  updateUnreadCount: (channelId: string, count: number) => void;
}

export const useChannelStore = create<ChannelState>((set, get) => ({
  channels: [],
  activeChannelId: null,
  isLoading: false,

  fetchChannels: async () => {
    set({ isLoading: true });
    try {
      const channels = await apiFetch<Channel[]>("/channels");
      set({ channels, isLoading: false });
    } catch {
      set({ isLoading: false });
    }
  },

  setActiveChannel: (id) => {
    set({ activeChannelId: id });
    // Mark as read when switching
    get().markRead(id);
  },

  createChannel: async (req) => {
    const channel = await apiFetch<Channel>("/channels", {
      method: "POST",
      body: JSON.stringify(req),
    });
    set((state) => ({ channels: [...state.channels, channel] }));
    return channel;
  },

  joinChannel: async (id) => {
    await apiFetch(`/channels/${id}/join`, { method: "POST" });
    get().fetchChannels();
  },

  leaveChannel: async (id) => {
    await apiFetch(`/channels/${id}/leave`, { method: "POST" });
    set((state) => ({
      channels: state.channels.filter((c) => c.id !== id),
      activeChannelId: state.activeChannelId === id ? null : state.activeChannelId,
    }));
  },

  markRead: async (id) => {
    try {
      await apiFetch(`/channels/${id}/read`, { method: "POST" });
      set((state) => ({
        channels: state.channels.map((c) =>
          c.id === id ? { ...c, unread_count: 0 } : c
        ),
      }));
    } catch {
      // ignore
    }
  },

  updateChannel: (channel) =>
    set((state) => ({
      channels: state.channels.map((c) => (c.id === channel.id ? channel : c)),
    })),

  addChannel: (channel) =>
    set((state) => ({
      channels: [...state.channels, channel],
    })),

  removeChannel: (id) =>
    set((state) => ({
      channels: state.channels.filter((c) => c.id !== id),
    })),

  updateUnreadCount: (channelId, count) =>
    set((state) => ({
      channels: state.channels.map((c) =>
        c.id === channelId ? { ...c, unread_count: count } : c
      ),
    })),
}));
