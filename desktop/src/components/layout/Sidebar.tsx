import { useAuthStore } from "../../stores/authStore";
import { useChannelStore } from "../../stores/channelStore";
import { useDMStore } from "../../stores/dmStore";
import { useMentionStore } from "../../stores/mentionStore";
import ChannelList from "../channels/ChannelList";
import DMList from "../dms/DMList";
import { useState, lazy, Suspense } from "react";

const CreateChannelModal = lazy(() => import("../channels/CreateChannelModal"));
const InviteModal = lazy(() => import("../invitations/InviteModal"));
const NewDMModal = lazy(() => import("../dms/NewDMModal"));
const MentionsPanel = lazy(() => import("../mentions/MentionsPanel"));

interface Props {
  onOpenChannelSwitcher: () => void;
}

export default function Sidebar({ onOpenChannelSwitcher }: Props) {
  const { user, logout } = useAuthStore();
  const [showCreateChannel, setShowCreateChannel] = useState(false);
  const [showInvite, setShowInvite] = useState(false);
  const [showNewDM, setShowNewDM] = useState(false);
  const [showMentions, setShowMentions] = useState(false);
  const { unreadMentions } = useMentionStore();

  return (
    <div className="flex w-60 flex-col bg-sidebar text-gray-300">
      {/* Workspace header */}
      <div className="flex items-center gap-2 border-b border-gray-700 px-4 py-3">
        <img src="/feather-logo.png" alt="Feather" className="h-7 w-7 rounded-full" />
        <h1 className="flex-1 text-lg font-bold text-white">Feather</h1>
        <button
          onClick={() => setShowInvite(true)}
          className="text-xs text-gray-500 hover:text-gray-300"
          title="Invite people"
        >
          +Invite
        </button>
      </div>

      {/* Quick switcher button */}
      <button
        onClick={onOpenChannelSwitcher}
        className="mx-3 mt-3 flex items-center rounded border border-gray-600 px-3 py-1.5 text-xs text-gray-400 hover:border-gray-500 hover:text-gray-300"
      >
        <span className="flex-1 text-left">Switch channel...</span>
        <kbd className="ml-2 text-[10px] text-gray-500">&#8984;K</kbd>
      </button>

      {/* Mentions button */}
      <button
        onClick={() => setShowMentions(true)}
        className="mx-3 mt-2 flex items-center rounded px-3 py-1.5 text-xs text-gray-400 hover:bg-gray-700/50 hover:text-gray-300"
      >
        <span className="flex-1 text-left">Mentions</span>
        {unreadMentions.length > 0 && (
          <span className="rounded-full bg-red-600 px-1.5 py-0.5 text-[10px] font-bold text-white">
            {unreadMentions.length}
          </span>
        )}
      </button>

      {/* Channel list */}
      <div className="mt-3 flex-1 overflow-y-auto">
        <div className="flex items-center justify-between px-4 py-1">
          <span className="text-xs font-semibold uppercase text-gray-500">Channels</span>
          <button
            onClick={() => setShowCreateChannel(true)}
            className="text-gray-500 hover:text-gray-300"
            title="Create channel"
          >
            +
          </button>
        </div>
        <ChannelList />

        {/* Direct Messages */}
        <div className="mt-4 flex items-center justify-between px-4 py-1">
          <span className="text-xs font-semibold uppercase text-gray-500">Direct Messages</span>
          <button
            onClick={() => setShowNewDM(true)}
            className="text-gray-500 hover:text-gray-300"
            title="New message"
          >
            +
          </button>
        </div>
        <DMList />
      </div>

      {/* User info */}
      <div className="flex items-center border-t border-gray-700 px-4 py-3">
        <div className="flex h-8 w-8 items-center justify-center rounded bg-blue-600 text-xs font-bold text-white">
          {user?.name?.charAt(0).toUpperCase()}
        </div>
        <div className="ml-2 flex-1 overflow-hidden">
          <div className="truncate text-sm font-medium text-white">{user?.name}</div>
          <div className="truncate text-xs text-gray-500">{user?.email}</div>
        </div>
        <button
          onClick={logout}
          className="ml-2 text-xs text-gray-500 hover:text-gray-300"
          title="Sign out"
        >
          &#9167;
        </button>
      </div>

      {showCreateChannel && (
        <Suspense fallback={null}>
          <CreateChannelModal onClose={() => setShowCreateChannel(false)} />
        </Suspense>
      )}

      {showInvite && (
        <Suspense fallback={null}>
          <InviteModal onClose={() => setShowInvite(false)} />
        </Suspense>
      )}

      {showNewDM && (
        <Suspense fallback={null}>
          <NewDMModal onClose={() => setShowNewDM(false)} />
        </Suspense>
      )}

      {showMentions && (
        <Suspense fallback={null}>
          <MentionsPanel onClose={() => setShowMentions(false)} />
        </Suspense>
      )}
    </div>
  );
}
