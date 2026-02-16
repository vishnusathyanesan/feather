import { create } from "zustand";
import type { Channel } from "../types/channel";
import { apiFetch } from "../services/api";

interface DMState {
  dms: Channel[];
  activeDMId: string | null;
  isLoading: boolean;

  fetchDMs: () => Promise<void>;
  createDM: (userId: string) => Promise<Channel>;
  createGroupDM: (userIds: string[]) => Promise<Channel>;
  setActiveDM: (id: string | null) => void;
  addDM: (dm: Channel) => void;
}

export const useDMStore = create<DMState>((set, get) => ({
  dms: [],
  activeDMId: null,
  isLoading: false,

  fetchDMs: async () => {
    set({ isLoading: true });
    try {
      const dms = await apiFetch<Channel[]>("/dms");
      set({ dms: dms || [], isLoading: false });
    } catch {
      set({ isLoading: false });
    }
  },

  createDM: async (userId) => {
    const dm = await apiFetch<Channel>("/dms", {
      method: "POST",
      body: JSON.stringify({ user_id: userId }),
    });
    const existing = get().dms.find((d) => d.id === dm.id);
    if (!existing) {
      set((state) => ({ dms: [dm, ...state.dms] }));
    }
    return dm;
  },

  createGroupDM: async (userIds) => {
    const dm = await apiFetch<Channel>("/dms/group", {
      method: "POST",
      body: JSON.stringify({ user_ids: userIds }),
    });
    set((state) => ({ dms: [dm, ...state.dms] }));
    return dm;
  },

  setActiveDM: (id) => set({ activeDMId: id }),

  addDM: (dm) => {
    const existing = get().dms.find((d) => d.id === dm.id);
    if (!existing) {
      set((state) => ({ dms: [dm, ...state.dms] }));
    }
  },
}));
