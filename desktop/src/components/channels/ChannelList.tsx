import { useChannelStore } from "../../stores/channelStore";
import Badge from "../common/Badge";

export default function ChannelList({ onSelect }: { onSelect?: () => void }) {
  const { channels, activeChannelId, setActiveChannel } = useChannelStore();

  return (
    <div className="space-y-0.5 px-2">
      {channels.map((channel) => (
        <button
          key={channel.id}
          onClick={() => { setActiveChannel(channel.id); onSelect?.(); }}
          className={`flex w-full items-center rounded px-2 py-1 text-sm ${
            activeChannelId === channel.id
              ? "bg-sidebar-active text-white"
              : "text-gray-400 hover:bg-sidebar-hover hover:text-gray-200"
          }`}
        >
          <span className="mr-1.5 text-gray-500">
            {channel.type === "private" ? "🔒" : "#"}
          </span>
          <span className="flex-1 truncate text-left">{channel.name}</span>
          <Badge count={channel.unread_count || 0} />
        </button>
      ))}
    </div>
  );
}
