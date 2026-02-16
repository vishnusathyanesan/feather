import { useCallStore } from "../../stores/callStore";
import { acceptCall, declineCall, initiatePeerConnection } from "../../services/webrtc";

export default function IncomingCallModal() {
  const { incomingCall } = useCallStore();

  if (!incomingCall) return null;

  const handleAccept = async () => {
    // Initialize peer connection as callee (wait for offer from initiator)
    try {
      await initiatePeerConnection(
        incomingCall.id,
        incomingCall.channel_id,
        incomingCall.call_type,
        false,
        incomingCall.initiator_id
      );
    } catch {
      return; // Media denied or failed
    }
    acceptCall(incomingCall.id);
    useCallStore.getState().setActiveCall(incomingCall);
    useCallStore.getState().setIncomingCall(null);
  };

  const handleDecline = () => {
    declineCall(incomingCall.id);
    useCallStore.getState().setIncomingCall(null);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="w-full max-w-sm rounded-lg bg-white p-6 text-center shadow-xl dark:bg-surface-secondary">
        <div className="mb-4">
          <div className="mx-auto mb-3 flex h-16 w-16 items-center justify-center rounded-full bg-blue-100 dark:bg-blue-900/30">
            <svg
              className="h-8 w-8 text-blue-600"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z"
              />
            </svg>
          </div>
          <h3 className="text-lg font-bold text-gray-900 dark:text-gray-100">
            Incoming {incomingCall.call_type === "video" ? "Video" : "Audio"}{" "}
            Call
          </h3>
        </div>
        <div className="flex justify-center gap-4">
          <button
            onClick={handleDecline}
            className="flex h-14 w-14 items-center justify-center rounded-full bg-red-600 text-white hover:bg-red-700"
          >
            <svg
              className="h-6 w-6"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
          <button
            onClick={handleAccept}
            className="flex h-14 w-14 items-center justify-center rounded-full bg-green-600 text-white hover:bg-green-700"
          >
            <svg
              className="h-6 w-6"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z"
              />
            </svg>
          </button>
        </div>
      </div>
    </div>
  );
}
