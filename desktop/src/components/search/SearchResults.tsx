import type { Message } from "../../types/message";
import Avatar from "../common/Avatar";

interface Props {
  messages: Message[];
  onNavigate: (msg: Message) => void;
}

export default function SearchResults({ messages, onNavigate }: Props) {
  return (
    <div className="space-y-1">
      {messages.map((msg) => (
        <button
          key={msg.id}
          onClick={() => onNavigate(msg)}
          className="flex w-full items-start gap-2 rounded p-2 text-left hover:bg-gray-50 dark:hover:bg-gray-800"
        >
          <Avatar name={msg.user?.name || "?"} size="sm" />
          <div className="flex-1 overflow-hidden">
            <div className="flex items-baseline">
              <span className="text-xs font-bold text-gray-900 dark:text-gray-100">
                {msg.user?.name}
              </span>
              <span className="ml-2 text-[10px] text-gray-500">
                {new Date(msg.created_at).toLocaleDateString()}
              </span>
            </div>
            <p className="truncate text-xs text-gray-600 dark:text-gray-400">
              {msg.content}
            </p>
          </div>
        </button>
      ))}
    </div>
  );
}
