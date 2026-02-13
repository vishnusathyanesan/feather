import { useEffect } from "react";
import { useMessageStore } from "../../stores/messageStore";
import MessageItem from "./MessageItem";
import MessageInput from "./MessageInput";

interface Props {
  parentId: string;
  onClose: () => void;
}

export default function ThreadPanel({ parentId, onClose }: Props) {
  const { threadMessages, fetchThread } = useMessageStore();
  const messages = threadMessages[parentId] || [];

  useEffect(() => {
    fetchThread(parentId);
  }, [parentId, fetchThread]);

  const parent = messages[0];

  return (
    <div className="flex w-[350px] flex-col border-l border-gray-200 bg-surface dark:border-gray-700">
      <div className="flex items-center justify-between border-b border-gray-200 px-4 py-3 dark:border-gray-700">
        <h3 className="text-sm font-bold text-gray-900 dark:text-gray-100">Thread</h3>
        <button
          onClick={onClose}
          className="text-gray-500 hover:text-gray-700 dark:hover:text-gray-300"
        >
          ✕
        </button>
      </div>

      <div className="flex-1 overflow-y-auto px-2">
        {messages.map((msg) => (
          <MessageItem
            key={msg.id}
            message={msg}
            onOpenThread={() => {}}
          />
        ))}
      </div>

      {parent && (
        <MessageInput
          channelId={parent.channel_id}
          parentId={parentId}
          placeholder="Reply in thread..."
        />
      )}
    </div>
  );
}
