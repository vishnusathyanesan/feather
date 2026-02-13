import { useEffect } from "react";
import { useMessageStore } from "../stores/messageStore";

export function useMessages(channelId: string) {
  const { messagesByChannel, hasMore, isLoading, fetchMessages } = useMessageStore();
  const messages = messagesByChannel[channelId] || [];

  useEffect(() => {
    if (channelId) {
      fetchMessages(channelId);
    }
  }, [channelId, fetchMessages]);

  const loadMore = () => {
    if (messages.length > 0 && hasMore[channelId]) {
      fetchMessages(channelId, messages[0].id);
    }
  };

  return { messages, isLoading, hasMore: hasMore[channelId] || false, loadMore };
}
