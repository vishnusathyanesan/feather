import type { Channel } from "../../types/channel";
import { initiateCall } from "../../services/webrtc";
import type { CallType } from "../../types/call";

interface Props {
  channel: Channel;
}

export default function Header({ channel }: Props) {
  const handleCall = (type: CallType) => {
    initiateCall(channel.id, type);
  };

  return (
    <div className="flex items-center border-b border-gray-200 px-4 py-3 dark:border-gray-700">
      <div className="flex-1">
        <div className="flex items-center">
          <span className="text-gray-500 dark:text-gray-400">#</span>
          <h2 className="ml-1 text-sm font-bold text-gray-900 dark:text-gray-100">
            {channel.name}
          </h2>
          {channel.type === "private" && (
            <span className="ml-2 rounded bg-gray-200 px-1.5 py-0.5 text-[10px] text-gray-600 dark:bg-gray-700 dark:text-gray-400">
              private
            </span>
          )}
        </div>
        {channel.topic && (
          <p className="mt-0.5 truncate text-xs text-gray-500 dark:text-gray-400">
            {channel.topic}
          </p>
        )}
      </div>
      <div className="flex items-center space-x-2 text-xs text-gray-500">
        <button
          onClick={() => handleCall("audio")}
          className="rounded p-1.5 hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-gray-700 dark:hover:text-gray-300"
          title="Audio call"
        >
          <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" />
          </svg>
        </button>
        <button
          onClick={() => handleCall("video")}
          className="rounded p-1.5 hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-gray-700 dark:hover:text-gray-300"
          title="Video call"
        >
          <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
          </svg>
        </button>
        {channel.member_count !== undefined && (
          <span>{channel.member_count} members</span>
        )}
      </div>
    </div>
  );
}
