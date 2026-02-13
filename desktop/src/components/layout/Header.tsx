import type { Channel } from "../../types/channel";

interface Props {
  channel: Channel;
}

export default function Header({ channel }: Props) {
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
        {channel.member_count !== undefined && (
          <span>{channel.member_count} members</span>
        )}
      </div>
    </div>
  );
}
