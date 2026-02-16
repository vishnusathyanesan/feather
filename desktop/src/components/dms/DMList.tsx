import { useDMStore } from "../../stores/dmStore";
import { useAuthStore } from "../../stores/authStore";
import { useChannelStore } from "../../stores/channelStore";
import type { Channel } from "../../types/channel";

function getDMDisplayName(dm: Channel, currentUserId: string): string {
  if (dm.type === "group_dm") {
    const names =
      dm.members
        ?.filter((m) => m.id !== currentUserId)
        .map((m) => m.name)
        .join(", ") || "Group DM";
    return names;
  }
  const other = dm.members?.find((m) => m.id !== currentUserId);
  return other?.name || "Direct Message";
}

export default function DMList() {
  const { dms } = useDMStore();
  const { user } = useAuthStore();
  const { activeChannelId, setActiveChannel } = useChannelStore();

  if (dms.length === 0) return null;

  return (
    <div className="space-y-0.5 px-2">
      {dms.map((dm) => {
        const isActive = dm.id === activeChannelId;
        const displayName = getDMDisplayName(dm, user?.id || "");
        const initial =
          dm.type === "group_dm"
            ? dm.members?.length?.toString() || "G"
            : dm.members?.find((m) => m.id !== user?.id)?.name?.charAt(0).toUpperCase() || "?";

        return (
          <button
            key={dm.id}
            onClick={() => setActiveChannel(dm.id)}
            className={`flex w-full items-center gap-2 rounded px-2 py-1 text-left text-sm ${
              isActive
                ? "bg-blue-600/20 text-white"
                : "text-gray-400 hover:bg-gray-700/50 hover:text-gray-200"
            }`}
          >
            <div className="flex h-6 w-6 items-center justify-center rounded-full bg-gray-600 text-xs font-bold text-white">
              {initial}
            </div>
            <span className="flex-1 truncate">{displayName}</span>
            {(dm.unread_count ?? 0) > 0 && (
              <span className="rounded-full bg-blue-600 px-1.5 py-0.5 text-[10px] font-bold text-white">
                {dm.unread_count}
              </span>
            )}
          </button>
        );
      })}
    </div>
  );
}
