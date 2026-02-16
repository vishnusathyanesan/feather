import { create } from "zustand";
import type { Call, CallType } from "../types/call";

interface CallState {
  activeCall: Call | null;
  incomingCall: Call | null;
  localStream: MediaStream | null;
  remoteStream: MediaStream | null;
  isMuted: boolean;
  isCameraOff: boolean;

  setActiveCall: (call: Call | null) => void;
  setIncomingCall: (call: Call | null) => void;
  setLocalStream: (stream: MediaStream | null) => void;
  setRemoteStream: (stream: MediaStream | null) => void;
  toggleMute: () => void;
  toggleCamera: () => void;
  reset: () => void;
}

export const useCallStore = create<CallState>((set, get) => ({
  activeCall: null,
  incomingCall: null,
  localStream: null,
  remoteStream: null,
  isMuted: false,
  isCameraOff: false,

  setActiveCall: (call) => set({ activeCall: call }),
  setIncomingCall: (call) => set({ incomingCall: call }),
  setLocalStream: (stream) => set({ localStream: stream }),
  setRemoteStream: (stream) => set({ remoteStream: stream }),

  toggleMute: () => {
    const { localStream, isMuted } = get();
    if (localStream) {
      localStream.getAudioTracks().forEach((t) => {
        t.enabled = isMuted;
      });
    }
    set({ isMuted: !isMuted });
  },

  toggleCamera: () => {
    const { localStream, isCameraOff } = get();
    if (localStream) {
      localStream.getVideoTracks().forEach((t) => {
        t.enabled = isCameraOff;
      });
    }
    set({ isCameraOff: !isCameraOff });
  },

  reset: () => {
    const { localStream, remoteStream } = get();
    localStream?.getTracks().forEach((t) => t.stop());
    remoteStream?.getTracks().forEach((t) => t.stop());
    set({
      activeCall: null,
      incomingCall: null,
      localStream: null,
      remoteStream: null,
      isMuted: false,
      isCameraOff: false,
    });
  },
}));
