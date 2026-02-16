import type { RTCConfig, CallType } from "../types/call";
import { apiFetch } from "./api";
import { wsService } from "./websocket";
import { useCallStore } from "../stores/callStore";

let peerConnection: RTCPeerConnection | null = null;
let rtcConfig: RTCConfig | null = null;

export async function fetchRTCConfig(): Promise<RTCConfig> {
  if (rtcConfig) return rtcConfig;
  rtcConfig = await apiFetch<RTCConfig>("/rtc/config");
  return rtcConfig;
}

export async function initiatePeerConnection(
  callId: string,
  channelId: string,
  callType: CallType,
  isInitiator: boolean,
  targetUserId: string
): Promise<void> {
  const config = await fetchRTCConfig();

  peerConnection = new RTCPeerConnection({
    iceServers: config.ice_servers.map((s) => ({
      urls: s.urls,
      username: s.username,
      credential: s.credential,
    })),
  });

  // Get local media
  const constraints: MediaStreamConstraints = {
    audio: true,
    video: callType === "video",
  };

  try {
    const stream = await navigator.mediaDevices.getUserMedia(constraints);
    useCallStore.getState().setLocalStream(stream);
    stream.getTracks().forEach((track) => {
      peerConnection!.addTrack(track, stream);
    });
  } catch (err) {
    console.error("Failed to get user media:", err);
    throw err;
  }

  // Handle remote stream
  peerConnection.ontrack = (event) => {
    const [remoteStream] = event.streams;
    useCallStore.getState().setRemoteStream(remoteStream);
  };

  // ICE candidate handling
  peerConnection.onicecandidate = (event) => {
    if (event.candidate) {
      wsService.send({
        type: "call.ice_candidate",
        payload: {
          call_id: callId,
          to_user: targetUserId,
          data: event.candidate,
        },
      });
    }
  };

  if (isInitiator) {
    try {
      const offer = await peerConnection.createOffer();
      await peerConnection.setLocalDescription(offer);
      wsService.send({
        type: "call.offer",
        payload: {
          call_id: callId,
          to_user: targetUserId,
          data: offer,
        },
      });
    } catch (err) {
      console.error("Failed to create offer:", err);
      closePeerConnection();
      throw err;
    }
  }
}

export async function handleOffer(
  callId: string,
  offer: RTCSessionDescriptionInit,
  targetUserId: string
): Promise<void> {
  if (!peerConnection) {
    console.warn("Received offer but peer connection not initialized, ignoring");
    return;
  }
  try {
    await peerConnection.setRemoteDescription(new RTCSessionDescription(offer));
    const answer = await peerConnection.createAnswer();
    await peerConnection.setLocalDescription(answer);
    wsService.send({
      type: "call.answer",
      payload: {
        call_id: callId,
        to_user: targetUserId,
        data: answer,
      },
    });
  } catch (err) {
    console.error("Failed to handle offer:", err);
    closePeerConnection();
  }
}

export async function handleAnswer(
  answer: RTCSessionDescriptionInit
): Promise<void> {
  if (!peerConnection) return;
  await peerConnection.setRemoteDescription(new RTCSessionDescription(answer));
}

export async function handleICECandidate(
  candidate: RTCIceCandidateInit
): Promise<void> {
  if (!peerConnection) return;
  try {
    await peerConnection.addIceCandidate(new RTCIceCandidate(candidate));
  } catch (err) {
    console.error("Failed to add ICE candidate:", err);
  }
}

export function closePeerConnection(): void {
  peerConnection?.close();
  peerConnection = null;
  useCallStore.getState().reset();
}

export function initiateCall(
  channelId: string,
  callType: CallType
): void {
  wsService.send({
    type: "call.initiate",
    payload: {
      channel_id: channelId,
      call_type: callType,
    },
  });
}

export function acceptCall(callId: string): void {
  wsService.send({
    type: "call.accept",
    payload: { call_id: callId },
  });
}

export function declineCall(callId: string): void {
  wsService.send({
    type: "call.decline",
    payload: { call_id: callId },
  });
}

export function hangupCall(callId: string): void {
  wsService.send({
    type: "call.hangup",
    payload: { call_id: callId },
  });
  closePeerConnection();
}
