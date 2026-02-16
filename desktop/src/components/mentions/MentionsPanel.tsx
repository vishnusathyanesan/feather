import { useEffect } from "react";
import { useMentionStore } from "../../stores/mentionStore";
import { useChannelStore } from "../../stores/channelStore";

interface Props {
  onClose: () => void;
}

export default function MentionsPanel({ onClose }: Props) {
  const { unreadMentions, fetchUnreadMentions } = useMentionStore();
  const { setActiveChannel } = useChannelStore();

  useEffect(() => {
    fetchUnreadMentions();
  }, [fetchUnreadMentions]);

  const handleClickMention = (channelId: string) => {
    setActiveChannel(channelId);
    onClose();
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="mx-2 w-full max-w-lg rounded-lg bg-white shadow-xl dark:bg-surface-secondary sm:mx-auto">
        <div className="flex items-center justify-between border-b border-gray-200 p-4 dark:border-gray-700">
          <h2 className="text-lg font-bold text-gray-900 dark:text-gray-100">
            Mentions
          </h2>
          <button
            onClick={onClose}
            className="text-gray-500 hover:text-gray-700 dark:hover:text-gray-300"
          >
            x
          </button>
        </div>

        <div className="max-h-96 overflow-y-auto p-4">
          {unreadMentions.length === 0 ? (
            <p className="text-center text-sm text-gray-500">
              No unread mentions
            </p>
          ) : (
            <div className="space-y-3">
              {unreadMentions.map((m) => (
                <button
                  key={m.id}
                  onClick={() => handleClickMention(m.channel_id)}
                  className="w-full rounded-lg border border-gray-200 p-3 text-left hover:bg-gray-50 dark:border-gray-700 dark:hover:bg-gray-700/50"
                >
                  <div className="flex items-center gap-2">
                    {m.message?.user && (
                      <div className="flex h-6 w-6 items-center justify-center rounded-full bg-blue-500 text-[10px] font-bold text-white">
                        {m.message.user.name?.charAt(0).toUpperCase()}
                      </div>
                    )}
                    <span className="text-sm font-medium text-gray-900 dark:text-gray-100">
                      {m.message?.user?.name || "Unknown"}
                    </span>
                    <span className="text-xs text-gray-500">
                      {new Date(m.created_at).toLocaleTimeString()}
                    </span>
                  </div>
                  <p className="mt-1 truncate text-sm text-gray-600 dark:text-gray-400">
                    {m.message?.content}
                  </p>
                </button>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
