import { useEffect, useRef, useState } from "react";
import { useCallStore } from "../../stores/callStore";
import CallControls from "./CallControls";

export default function CallView() {
  const { activeCall, localStream, remoteStream } = useCallStore();
  const localVideoRef = useRef<HTMLVideoElement>(null);
  const remoteVideoRef = useRef<HTMLVideoElement>(null);
  const [duration, setDuration] = useState(0);

  useEffect(() => {
    if (localVideoRef.current && localStream) {
      localVideoRef.current.srcObject = localStream;
    }
  }, [localStream]);

  useEffect(() => {
    if (remoteVideoRef.current && remoteStream) {
      remoteVideoRef.current.srcObject = remoteStream;
    }
  }, [remoteStream]);

  // Call duration timer
  useEffect(() => {
    if (!activeCall || activeCall.status !== "in_progress") return;
    const interval = setInterval(() => setDuration((d) => d + 1), 1000);
    return () => clearInterval(interval);
  }, [activeCall]);

  if (!activeCall) return null;

  const isVideo = activeCall.call_type === "video";
  const mins = Math.floor(duration / 60);
  const secs = duration % 60;
  const timeStr = `${mins.toString().padStart(2, "0")}:${secs.toString().padStart(2, "0")}`;

  return (
    <div className="fixed inset-0 z-40 flex flex-col bg-gray-900">
      {/* Remote video / audio placeholder */}
      <div className="relative flex flex-1 items-center justify-center">
        {isVideo && remoteStream ? (
          <video
            ref={remoteVideoRef}
            autoPlay
            playsInline
            className="h-full w-full object-contain"
          />
        ) : (
          <div className="flex flex-col items-center gap-4">
            <div className="flex h-24 w-24 items-center justify-center rounded-full bg-gray-700 text-3xl font-bold text-white">
              ?
            </div>
            <p className="text-lg text-white">
              {activeCall.status === "ringing" ? "Calling..." : timeStr}
            </p>
          </div>
        )}

        {/* Local video (PIP) */}
        {isVideo && localStream && (
          <div className="absolute bottom-4 right-4 h-32 w-48 overflow-hidden rounded-lg border-2 border-white/20 shadow-lg">
            <video
              ref={localVideoRef}
              autoPlay
              playsInline
              muted
              className="h-full w-full object-cover"
            />
          </div>
        )}

        {/* Duration overlay */}
        {isVideo && activeCall.status === "in_progress" && (
          <div className="absolute top-4 left-1/2 -translate-x-1/2 rounded-full bg-black/50 px-4 py-1 text-sm text-white">
            {timeStr}
          </div>
        )}
      </div>

      {/* Controls */}
      <CallControls />
    </div>
  );
}
