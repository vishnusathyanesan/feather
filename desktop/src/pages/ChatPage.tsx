import { useEffect, useState, useCallback } from "react";
import { useChannelStore } from "../stores/channelStore";
import { useMessageStore } from "../stores/messageStore";
import { usePresenceStore } from "../stores/presenceStore";
import { wsService } from "../services/websocket";
import { initNotifications, notify } from "../services/notifications";
import type { WebSocketEvent, PresencePayload, TypingPayload } from "../types/websocket";
import type { Message } from "../types/message";
import type { Channel } from "../types/channel";
import AppLayout from "../components/layout/AppLayout";
import Sidebar from "../components/layout/Sidebar";
import Header from "../components/layout/Header";
import MessageList from "../components/messages/MessageList";
import MessageInput from "../components/messages/MessageInput";
import ThreadPanel from "../components/messages/ThreadPanel";
import ChannelSwitcher from "../components/channels/ChannelSwitcher";
import SearchModal from "../components/search/SearchModal";

export default function ChatPage() {
  const { channels, activeChannelId, fetchChannels, setActiveChannel, updateUnreadCount } = useChannelStore();
  const { addMessage, updateMessage, removeMessage } = useMessageStore();
  const { setOnline, setOffline, setTyping } = usePresenceStore();
  const [threadParentId, setThreadParentId] = useState<string | null>(null);
  const [showChannelSwitcher, setShowChannelSwitcher] = useState(false);
  const [showSearch, setShowSearch] = useState(false);

  const activeChannel = channels.find((c) => c.id === activeChannelId);

  useEffect(() => {
    fetchChannels();
    initNotifications();
    wsService.connect();

    return () => {
      wsService.disconnect();
    };
  }, [fetchChannels]);

  // Auto-select first channel
  useEffect(() => {
    if (!activeChannelId && channels.length > 0) {
      setActiveChannel(channels[0].id);
    }
  }, [channels, activeChannelId, setActiveChannel]);

  // WebSocket event handlers
  useEffect(() => {
    const unsubNew = wsService.on("message.new", (event: WebSocketEvent) => {
      const msg = event.payload as Message;
      addMessage(msg.channel_id, msg);

      if (msg.channel_id !== activeChannelId) {
        const ch = channels.find((c) => c.id === msg.channel_id);
        updateUnreadCount(msg.channel_id, (ch?.unread_count || 0) + 1);
        notify(
          `#${ch?.name || "channel"} - ${msg.user?.name || "Someone"}`,
          msg.content.substring(0, 100)
        );
      }
    });

    const unsubUpdated = wsService.on("message.updated", (event: WebSocketEvent) => {
      const msg = event.payload as Message;
      updateMessage(msg.channel_id, msg);
    });

    const unsubDeleted = wsService.on("message.deleted", (event: WebSocketEvent) => {
      const msg = event.payload as Message;
      removeMessage(msg.channel_id, msg.id);
    });

    const unsubPresence = wsService.on("presence.update", (event: WebSocketEvent) => {
      const p = event.payload as PresencePayload;
      if (p.online) setOnline(p.user_id);
      else setOffline(p.user_id);
    });

    const unsubTyping = wsService.on("typing", (event: WebSocketEvent) => {
      const t = event.payload as TypingPayload;
      setTyping(t.channel_id, t.user_id, t.user_name);
    });

    const handleReactionChange = (event: WebSocketEvent) => {
      const chId = event.channel_id;
      if (chId) {
        const { fetchMessages } = useMessageStore.getState();
        fetchMessages(chId);
      }
    };
    const unsubReactionAdded = wsService.on("reaction.added", handleReactionChange);
    const unsubReactionRemoved = wsService.on("reaction.removed", handleReactionChange);

    const unsubChannelCreated = wsService.on("channel.created", () => {
      fetchChannels();
    });

    return () => {
      unsubNew();
      unsubUpdated();
      unsubDeleted();
      unsubReactionAdded();
      unsubReactionRemoved();
      unsubPresence();
      unsubTyping();
      unsubChannelCreated();
    };
  }, [activeChannelId, channels, addMessage, updateMessage, removeMessage, setOnline, setOffline, setTyping, updateUnreadCount, fetchChannels]);

  // Keyboard shortcuts
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault();
        setShowChannelSwitcher(true);
      }
      if ((e.metaKey || e.ctrlKey) && e.key === "/") {
        e.preventDefault();
        setShowSearch(true);
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, []);

  const handleOpenThread = useCallback((messageId: string) => {
    setThreadParentId(messageId);
  }, []);

  return (
    <AppLayout>
      <Sidebar onOpenChannelSwitcher={() => setShowChannelSwitcher(true)} />

      <div className="flex flex-1 flex-col overflow-hidden">
        {activeChannel ? (
          <>
            <Header channel={activeChannel} />
            <MessageList
              channelId={activeChannel.id}
              onOpenThread={handleOpenThread}
            />
            <MessageInput channelId={activeChannel.id} />
          </>
        ) : (
          <div className="flex flex-1 items-center justify-center text-gray-500">
            Select a channel to start chatting
          </div>
        )}
      </div>

      {threadParentId && (
        <ThreadPanel
          parentId={threadParentId}
          onClose={() => setThreadParentId(null)}
        />
      )}

      {showChannelSwitcher && (
        <ChannelSwitcher onClose={() => setShowChannelSwitcher(false)} />
      )}

      {showSearch && (
        <SearchModal onClose={() => setShowSearch(false)} />
      )}
    </AppLayout>
  );
}
