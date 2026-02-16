import { create } from "zustand";
import type {
  WorkspaceInvitation,
  CreateInvitationRequest,
} from "../types/invitation";
import { apiFetch } from "../services/api";

interface InvitationState {
  invitations: WorkspaceInvitation[];
  isLoading: boolean;

  fetchInvitations: () => Promise<void>;
  createInvitation: (
    req: CreateInvitationRequest
  ) => Promise<WorkspaceInvitation>;
  revokeInvitation: (id: string) => Promise<void>;
}

export const useInvitationStore = create<InvitationState>((set) => ({
  invitations: [],
  isLoading: false,

  fetchInvitations: async () => {
    set({ isLoading: true });
    try {
      const invitations = await apiFetch<WorkspaceInvitation[]>("/invitations");
      set({ invitations: invitations || [], isLoading: false });
    } catch {
      set({ isLoading: false });
    }
  },

  createInvitation: async (req) => {
    const inv = await apiFetch<WorkspaceInvitation>("/invitations", {
      method: "POST",
      body: JSON.stringify(req),
    });
    set((state) => ({ invitations: [inv, ...state.invitations] }));
    return inv;
  },

  revokeInvitation: async (id) => {
    await apiFetch(`/invitations/${id}`, { method: "DELETE" });
    set((state) => ({
      invitations: state.invitations.filter((inv) => inv.id !== id),
    }));
  },
}));
