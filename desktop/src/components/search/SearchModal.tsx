import { useState, useEffect, useRef, useCallback } from "react";
import { apiFetch } from "../../services/api";
import { useChannelStore } from "../../stores/channelStore";
import type { Message } from "../../types/message";
import Modal from "../common/Modal";
import Avatar from "../common/Avatar";

interface SearchResult {
  messages: Message[];
  total_count: number;
}

interface Props {
  onClose: () => void;
}

export default function SearchModal({ onClose }: Props) {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<Message[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [isSearching, setIsSearching] = useState(false);
  const [filter, setFilter] = useState<string>("");
  const { setActiveChannel } = useChannelStore();
  const inputRef = useRef<HTMLInputElement>(null);
  const debounceRef = useRef<ReturnType<typeof setTimeout>>();

  useEffect(() => {
    inputRef.current?.focus();
  }, []);

  const doSearch = useCallback(async (q: string, has: string) => {
    if (!q.trim()) {
      setResults([]);
      setTotalCount(0);
      return;
    }

    setIsSearching(true);
    try {
      const params = new URLSearchParams({ q, limit: "20" });
      if (has) params.set("has", has);

      const result = await apiFetch<SearchResult>(`/search?${params}`);
      setResults(result.messages || []);
      setTotalCount(result.total_count);
    } catch {
      setResults([]);
    } finally {
      setIsSearching(false);
    }
  }, []);

  useEffect(() => {
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => {
      doSearch(query, filter);
    }, 300);
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current);
    };
  }, [query, filter, doSearch]);

  const handleNavigate = (msg: Message) => {
    setActiveChannel(msg.channel_id);
    onClose();
  };

  return (
    <Modal onClose={onClose}>
      <div className="p-4">
        <input
          ref={inputRef}
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Search messages..."
          className="w-full rounded border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none dark:border-gray-600 dark:bg-surface-secondary dark:text-gray-100"
        />

        <div className="mt-2 flex gap-2">
          {["", "link", "code"].map((f) => (
            <button
              key={f}
              onClick={() => setFilter(f)}
              className={`rounded px-2 py-1 text-xs ${
                filter === f
                  ? "bg-blue-600 text-white"
                  : "bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-gray-700 dark:text-gray-300"
              }`}
            >
              {f === "" ? "All" : `has:${f}`}
            </button>
          ))}
        </div>

        <div className="mt-3 max-h-96 overflow-y-auto">
          {isSearching && (
            <div className="py-4 text-center text-sm text-gray-500">Searching...</div>
          )}

          {!isSearching && query && results.length === 0 && (
            <div className="py-4 text-center text-sm text-gray-500">No results found</div>
          )}

          {results.map((msg) => (
            <button
              key={msg.id}
              onClick={() => handleNavigate(msg)}
              className="flex w-full items-start gap-2 rounded p-2 text-left hover:bg-gray-50 dark:hover:bg-gray-800"
            >
              <Avatar name={msg.user?.name || "?"} size="sm" />
              <div className="flex-1 overflow-hidden">
                <div className="flex items-baseline">
                  <span className="text-xs font-bold text-gray-900 dark:text-gray-100">
                    {msg.user?.name}
                  </span>
                  <span className="ml-2 text-[10px] text-gray-500">
                    {new Date(msg.created_at).toLocaleDateString()}
                  </span>
                </div>
                <p className="truncate text-xs text-gray-600 dark:text-gray-400">
                  {msg.content}
                </p>
              </div>
            </button>
          ))}

          {totalCount > 20 && (
            <div className="py-2 text-center text-xs text-gray-500">
              Showing 20 of {totalCount} results
            </div>
          )}
        </div>
      </div>
    </Modal>
  );
}
