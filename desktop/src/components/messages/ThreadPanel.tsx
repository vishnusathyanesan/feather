import { useEffect, useState } from "react";
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
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    setIsLoading(true);
    fetchThread(parentId).finally(() => setIsLoading(false));
  }, [parentId, fetchThread]);

  const parent = messages[0];

  return (
    <>
      {/* Mobile backdrop */}
      <div className="fixed inset-0 z-30 bg-black/50 md:hidden" onClick={onClose} />

      <div className="fixed inset-0 z-40 flex flex-col bg-surface md:static md:z-auto md:w-[350px] md:border-l md:border-gray-200 dark:md:border-gray-700">
        <div className="flex items-center justify-between border-b border-gray-200 px-4 py-3 dark:border-gray-700">
          <h3 className="text-sm font-bold text-gray-900 dark:text-gray-100">Thread</h3>
          <button
            onClick={onClose}
            className="text-gray-500 hover:text-gray-700 dark:hover:text-gray-300"
          >
            <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div className="flex-1 overflow-y-auto px-2">
          {isLoading && messages.length === 0 ? (
            <div className="flex h-full items-center justify-center text-sm text-gray-500">
              Loading thread...
            </div>
          ) : (
            messages.map((msg) => (
              <MessageItem
                key={msg.id}
                message={msg}
                onOpenThread={() => {}}
              />
            ))
          )}
        </div>

        {parent && (
          <MessageInput
            channelId={parent.channel_id}
            parentId={parentId}
            placeholder="Reply in thread..."
          />
        )}
      </div>
    </>
  );
}
