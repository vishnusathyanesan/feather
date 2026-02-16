import { useEffect, useState, useCallback, lazy, Suspense } from "react";
import { useChannelStore } from "../stores/channelStore";
import { useMessageStore } from "../stores/messageStore";
import { usePresenceStore } from "../stores/presenceStore";
import { useDMStore } from "../stores/dmStore";
import { useMentionStore } from "../stores/mentionStore";
import { useCallStore } from "../stores/callStore";
import { wsService } from "../services/websocket";
import { initNotifications, notify } from "../services/notifications";
import {
  initiatePeerConnection,
  handleOffer,
  handleAnswer,
  handleICECandidate,
  closePeerConnection,
} from "../services/webrtc";
import type { WebSocketEvent, PresencePayload, TypingPayload } from "../types/websocket";
import type { Message } from "../types/message";
import type { Channel } from "../types/channel";
import type { Call } from "../types/call";
import type { Mention } from "../types/mention";
import AppLayout from "../components/layout/AppLayout";
import Sidebar from "../components/layout/Sidebar";
import Header from "../components/layout/Header";
import DMHeader from "../components/dms/DMHeader";
import MessageList from "../components/messages/MessageList";
import MessageInput from "../components/messages/MessageInput";
import { useAuthStore } from "../stores/authStore";

const ThreadPanel = lazy(() => import("../components/messages/ThreadPanel"));
const ChannelSwitcher = lazy(() => import("../components/channels/ChannelSwitcher"));
const SearchModal = lazy(() => import("../components/search/SearchModal"));
const IncomingCallModal = lazy(() => import("../components/calls/IncomingCallModal"));
const CallView = lazy(() => import("../components/calls/CallView"));

export default function ChatPage() {
  const { channels, activeChannelId, fetchChannels, setActiveChannel } = useChannelStore();
  const { fetchDMs, addDM, dms } = useDMStore();
  const { fetchUnreadMentions } = useMentionStore();
  const { user: currentUser } = useAuthStore();
  const { activeCall, incomingCall } = useCallStore();
  const [threadParentId, setThreadParentId] = useState<string | null>(null);
  const [showChannelSwitcher, setShowChannelSwitcher] = useState(false);
  const [showSearch, setShowSearch] = useState(false);
  const [sidebarOpen, setSidebarOpen] = useState(false);

  // Find active channel across both channels and DMs
  const activeChannel =
    channels.find((c) => c.id === activeChannelId) ||
    dms.find((d) => d.id === activeChannelId);

  const isDM =
    activeChannel?.type === "dm" || activeChannel?.type === "group_dm";

  useEffect(() => {
    fetchChannels();
    fetchDMs();
    fetchUnreadMentions();
    initNotifications();
    wsService.connect();

    return () => {
      wsService.disconnect();
    };
  }, [fetchChannels, fetchDMs, fetchUnreadMentions]);

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
      useMessageStore.getState().addMessage(msg.channel_id, msg);

      const { activeChannelId: currentChannelId, channels: currentChannels, updateUnreadCount: updateUnread } = useChannelStore.getState();
      if (msg.channel_id !== currentChannelId) {
        const ch = currentChannels.find((c) => c.id === msg.channel_id);
        updateUnread(msg.channel_id, (ch?.unread_count || 0) + 1);
        notify(
          `#${ch?.name || "channel"} - ${msg.user?.name || "Someone"}`,
          msg.content.substring(0, 100)
        );
      }
    });

    const unsubUpdated = wsService.on("message.updated", (event: WebSocketEvent) => {
      const msg = event.payload as Message;
      useMessageStore.getState().updateMessage(msg.channel_id, msg);
    });

    const unsubDeleted = wsService.on("message.deleted", (event: WebSocketEvent) => {
      const msg = event.payload as Message;
      useMessageStore.getState().removeMessage(msg.channel_id, msg.id);
    });

    const unsubPresence = wsService.on("presence.update", (event: WebSocketEvent) => {
      const p = event.payload as PresencePayload;
      const store = usePresenceStore.getState();
      if (p.online) store.setOnline(p.user_id);
      else store.setOffline(p.user_id);
    });

    const unsubTyping = wsService.on("typing", (event: WebSocketEvent) => {
      const t = event.payload as TypingPayload;
      usePresenceStore.getState().setTyping(t.channel_id, t.user_id, t.user_name);
    });

    const unsubReactionAdded = wsService.on("reaction.added", (event: WebSocketEvent) => {
      const p = event.payload as { message_id: string; user_id: string; emoji: string };
      if (event.channel_id) {
        useMessageStore.getState().updateReactionInPlace(event.channel_id, p.message_id, p.emoji, p.user_id, true);
      }
    });
    const unsubReactionRemoved = wsService.on("reaction.removed", (event: WebSocketEvent) => {
      const p = event.payload as { message_id: string; user_id: string; emoji: string };
      if (event.channel_id) {
        useMessageStore.getState().updateReactionInPlace(event.channel_id, p.message_id, p.emoji, p.user_id, false);
      }
    });

    const unsubChannelCreated = wsService.on("channel.created", () => {
      useChannelStore.getState().fetchChannels();
    });

    // DM events
    const unsubDMCreated = wsService.on("dm.created", (event: WebSocketEvent) => {
      const dm = event.payload as Channel;
      useDMStore.getState().addDM(dm);
    });

    // Mention events
    const unsubMentionNew = wsService.on("mention.new", (event: WebSocketEvent) => {
      const mention = event.payload as Mention;
      useMentionStore.getState().addMention(mention);
    });

    // Call events
    const unsubCallRinging = wsService.on("call.ringing", (event: WebSocketEvent) => {
      const call = event.payload as Call;
      if (call.initiator_id === currentUser?.id) {
        // I'm the caller — show "Calling..." UI
        useCallStore.getState().setActiveCall(call);
      } else {
        // I'm being called
        useCallStore.getState().setIncomingCall(call);
      }
    });

    const unsubCallAccepted = wsService.on("call.accepted", (event: WebSocketEvent) => {
      const call = event.payload as Call;
      useCallStore.getState().setActiveCall(call);
      useCallStore.getState().setIncomingCall(null);

      // If I'm the initiator, set up peer connection and create the SDP offer
      if (call.initiator_id === currentUser?.id && call.accepted_by) {
        initiatePeerConnection(
          call.id,
          call.channel_id,
          call.call_type,
          true,
          call.accepted_by
        ).catch(() => closePeerConnection());
      }
    });

    const unsubCallDeclined = wsService.on("call.declined", () => {
      useCallStore.getState().setIncomingCall(null);
      closePeerConnection();
    });

    const unsubCallEnded = wsService.on("call.ended", () => {
      closePeerConnection();
    });

    const unsubCallMissed = wsService.on("call.missed", () => {
      useCallStore.getState().setIncomingCall(null);
      closePeerConnection();
    });

    const unsubCallOffer = wsService.on("call.offer", (event: WebSocketEvent) => {
      const p = event.payload as { call_id: string; from_user: string; data: RTCSessionDescriptionInit };
      handleOffer(p.call_id, p.data, p.from_user);
    });

    const unsubCallAnswer = wsService.on("call.answer", (event: WebSocketEvent) => {
      const p = event.payload as { data: RTCSessionDescriptionInit };
      handleAnswer(p.data);
    });

    const unsubICE = wsService.on("call.ice_candidate", (event: WebSocketEvent) => {
      const p = event.payload as { data: RTCIceCandidateInit };
      handleICECandidate(p.data);
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
      unsubDMCreated();
      unsubMentionNew();
      unsubCallRinging();
      unsubCallAccepted();
      unsubCallDeclined();
      unsubCallEnded();
      unsubCallMissed();
      unsubCallOffer();
      unsubCallAnswer();
      unsubICE();
    };
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

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

  const handleStartCall = useCallback(
    (type: "audio" | "video") => {
      if (!activeChannelId) return;
      wsService.send({
        type: "call.initiate",
        payload: { channel_id: activeChannelId, call_type: type },
      });
    },
    [activeChannelId]
  );

  const openSidebar = useCallback(() => setSidebarOpen(true), []);
  const closeSidebar = useCallback(() => setSidebarOpen(false), []);

  return (
    <AppLayout>
      <Sidebar
        onOpenChannelSwitcher={() => setShowChannelSwitcher(true)}
        open={sidebarOpen}
        onClose={closeSidebar}
      />

      <div className="flex min-w-0 flex-1 flex-col overflow-hidden">
        {activeChannel ? (
          <>
            {isDM ? (
              <DMHeader channel={activeChannel} onStartCall={handleStartCall} onOpenSidebar={openSidebar} />
            ) : (
              <Header channel={activeChannel} onOpenSidebar={openSidebar} />
            )}
            <MessageList
              channelId={activeChannel.id}
              onOpenThread={handleOpenThread}
            />
            <MessageInput channelId={activeChannel.id} />
          </>
        ) : (
          <div className="flex flex-1 items-center justify-center px-4 text-center text-gray-500">
            <div>
              <p>Select a channel to start chatting</p>
              <button
                onClick={openSidebar}
                className="mt-2 text-blue-500 hover:text-blue-600 md:hidden"
              >
                Open channels
              </button>
            </div>
          </div>
        )}
      </div>

      <Suspense fallback={null}>
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

        {incomingCall && <IncomingCallModal />}
        {activeCall && <CallView />}
      </Suspense>
    </AppLayout>
  );
}
