import { useAuthStore } from "../../stores/authStore";
import { useChannelStore } from "../../stores/channelStore";
import ChannelList from "../channels/ChannelList";
import CreateChannelModal from "../channels/CreateChannelModal";
import { useState } from "react";

interface Props {
  onOpenChannelSwitcher: () => void;
}

export default function Sidebar({ onOpenChannelSwitcher }: Props) {
  const { user, logout } = useAuthStore();
  const [showCreateChannel, setShowCreateChannel] = useState(false);

  return (
    <div className="flex w-60 flex-col bg-sidebar text-gray-300">
      {/* Workspace header */}
      <div className="flex items-center justify-between border-b border-gray-700 px-4 py-3">
        <h1 className="text-lg font-bold text-white">Feather</h1>
      </div>

      {/* Quick switcher button */}
      <button
        onClick={onOpenChannelSwitcher}
        className="mx-3 mt-3 flex items-center rounded border border-gray-600 px-3 py-1.5 text-xs text-gray-400 hover:border-gray-500 hover:text-gray-300"
      >
        <span className="flex-1 text-left">Switch channel...</span>
        <kbd className="ml-2 text-[10px] text-gray-500">⌘K</kbd>
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
          ⏏
        </button>
      </div>

      {showCreateChannel && (
        <CreateChannelModal onClose={() => setShowCreateChannel(false)} />
      )}
    </div>
  );
}
