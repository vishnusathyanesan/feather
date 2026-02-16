import { useState, useEffect, useRef } from "react";
import { useChannelStore } from "../../stores/channelStore";
import Modal from "../common/Modal";

interface Props {
  onClose: () => void;
}

export default function ChannelSwitcher({ onClose }: Props) {
  const { channels, setActiveChannel } = useChannelStore();
  const [query, setQuery] = useState("");
  const [selectedIndex, setSelectedIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);

  const filtered = channels.filter((ch) =>
    (ch.name || "").toLowerCase().includes(query.toLowerCase())
  );

  useEffect(() => {
    inputRef.current?.focus();
  }, []);

  useEffect(() => {
    setSelectedIndex(0);
  }, [query]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setSelectedIndex((i) => Math.min(i + 1, filtered.length - 1));
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setSelectedIndex((i) => Math.max(i - 1, 0));
    } else if (e.key === "Enter" && filtered[selectedIndex]) {
      setActiveChannel(filtered[selectedIndex].id);
      onClose();
    }
  };

  return (
    <Modal onClose={onClose}>
      <div className="p-4">
        <input
          ref={inputRef}
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Switch to channel..."
          className="w-full rounded border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none dark:border-gray-600 dark:bg-surface-secondary dark:text-gray-100"
        />
        <div className="mt-2 max-h-64 overflow-y-auto">
          {filtered.map((ch, i) => (
            <button
              key={ch.id}
              onClick={() => {
                setActiveChannel(ch.id);
                onClose();
              }}
              className={`flex w-full items-center rounded px-3 py-2 text-sm ${
                i === selectedIndex
                  ? "bg-blue-600 text-white"
                  : "text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700"
              }`}
            >
              <span className="mr-2 text-gray-400">
                {ch.type === "private" ? "🔒" : "#"}
              </span>
              {ch.name}
              {ch.topic && (
                <span className="ml-2 truncate text-xs text-gray-500">
                  {ch.topic}
                </span>
              )}
            </button>
          ))}
          {filtered.length === 0 && (
            <div className="px-3 py-2 text-sm text-gray-500">No channels found</div>
          )}
        </div>
      </div>
    </Modal>
  );
}
