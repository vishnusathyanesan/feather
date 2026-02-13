import { useEffect, useRef, useCallback } from "react";
import { useVirtualizer } from "@tanstack/react-virtual";
import { useMessageStore } from "../../stores/messageStore";
import MessageItem from "./MessageItem";

interface Props {
  channelId: string;
  onOpenThread: (messageId: string) => void;
}

export default function MessageList({ channelId, onOpenThread }: Props) {
  const { messagesByChannel, hasMore, isLoading, fetchMessages } = useMessageStore();
  const messages = messagesByChannel[channelId] || [];
  const parentRef = useRef<HTMLDivElement>(null);
  const isNearBottomRef = useRef(true);

  useEffect(() => {
    fetchMessages(channelId);
  }, [channelId, fetchMessages]);

  const virtualizer = useVirtualizer({
    count: messages.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 72,
    overscan: 10,
  });

  // Auto-scroll to bottom on new messages if near bottom
  useEffect(() => {
    if (isNearBottomRef.current && messages.length > 0) {
      requestAnimationFrame(() => {
        const el = parentRef.current;
        if (el) el.scrollTop = el.scrollHeight;
      });
    }
  }, [messages.length]);

  const handleScroll = useCallback(() => {
    const el = parentRef.current;
    if (!el) return;

    // Check if near bottom
    isNearBottomRef.current = el.scrollHeight - el.scrollTop - el.clientHeight < 100;

    // Load more on scroll to top
    if (el.scrollTop < 100 && !isLoading && hasMore[channelId] && messages.length > 0) {
      fetchMessages(channelId, messages[0].id);
    }
  }, [channelId, isLoading, hasMore, messages, fetchMessages]);

  return (
    <div
      ref={parentRef}
      onScroll={handleScroll}
      className="flex-1 overflow-y-auto px-4"
    >
      {isLoading && messages.length === 0 && (
        <div className="flex h-full items-center justify-center text-sm text-gray-500">
          Loading messages...
        </div>
      )}

      {!isLoading && messages.length === 0 && (
        <div className="flex h-full items-center justify-center text-sm text-gray-500">
          No messages yet. Start the conversation!
        </div>
      )}

      <div
        style={{
          height: `${virtualizer.getTotalSize()}px`,
          width: "100%",
          position: "relative",
        }}
      >
        {virtualizer.getVirtualItems().map((virtualRow) => {
          const message = messages[virtualRow.index];
          return (
            <div
              key={virtualRow.key}
              data-index={virtualRow.index}
              ref={virtualizer.measureElement}
              style={{
                position: "absolute",
                top: 0,
                left: 0,
                width: "100%",
                transform: `translateY(${virtualRow.start}px)`,
              }}
            >
              <MessageItem
                message={message}
                onOpenThread={onOpenThread}
              />
            </div>
          );
        })}
      </div>
    </div>
  );
}
