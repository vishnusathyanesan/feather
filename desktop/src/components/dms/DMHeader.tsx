import type { Channel } from "../../types/channel";
import { useAuthStore } from "../../stores/authStore";
import { usePresenceStore } from "../../stores/presenceStore";

interface Props {
  channel: Channel;
  onStartCall?: (type: "audio" | "video") => void;
}

export default function DMHeader({ channel, onStartCall }: Props) {
  const { user: currentUser } = useAuthStore();
  const { isOnline: checkOnline } = usePresenceStore();

  const isDM = channel.type === "dm";
  const otherUser = isDM
    ? channel.members?.find((m) => m.id !== currentUser?.id)
    : null;
  const isOnline = otherUser ? checkOnline(otherUser.id) : false;

  const displayName = isDM
    ? otherUser?.name || "Direct Message"
    : channel.members
        ?.filter((m) => m.id !== currentUser?.id)
        .map((m) => m.name)
        .join(", ") || "Group DM";

  return (
    <div className="flex items-center border-b border-gray-200 px-4 py-3 dark:border-gray-700">
      <div className="flex-1">
        <div className="flex items-center gap-2">
          {isDM && otherUser && (
            <div className="relative">
              <div className="flex h-8 w-8 items-center justify-center rounded-full bg-gray-300 text-xs font-bold text-gray-700 dark:bg-gray-600 dark:text-gray-300">
                {otherUser.name.charAt(0).toUpperCase()}
              </div>
              <span
                className={`absolute bottom-0 right-0 h-2.5 w-2.5 rounded-full border-2 border-white dark:border-gray-800 ${
                  isOnline ? "bg-green-500" : "bg-gray-400"
                }`}
              />
            </div>
          )}
          <div>
            <h2 className="text-sm font-bold text-gray-900 dark:text-gray-100">
              {displayName}
            </h2>
            {isDM && (
              <p className="text-xs text-gray-500">
                {isOnline ? "Online" : "Offline"}
              </p>
            )}
          </div>
        </div>
      </div>
      {onStartCall && (
        <div className="flex items-center gap-2">
          <button
            onClick={() => onStartCall("audio")}
            className="rounded p-2 text-gray-500 hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-gray-700 dark:hover:text-gray-300"
            title="Audio call"
          >
            <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" />
            </svg>
          </button>
          <button
            onClick={() => onStartCall("video")}
            className="rounded p-2 text-gray-500 hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-gray-700 dark:hover:text-gray-300"
            title="Video call"
          >
            <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
            </svg>
          </button>
        </div>
      )}
    </div>
  );
}
