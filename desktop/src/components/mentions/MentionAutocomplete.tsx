import { useState, useEffect, useRef, useCallback } from "react";
import { apiFetch } from "../../services/api";
import type { User } from "../../types/user";
import type { UserGroup } from "../../types/mention";

interface Props {
  query: string;
  onSelect: (name: string) => void;
  onClose: () => void;
  visible: boolean;
}

type MentionItem =
  | { type: "user"; user: User }
  | { type: "group"; group: UserGroup }
  | { type: "keyword"; name: string; label: string };

export default function MentionAutocomplete({
  query,
  onSelect,
  onClose,
  visible,
}: Props) {
  const [items, setItems] = useState<MentionItem[]>([]);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const listRef = useRef<HTMLDivElement>(null);
  const debounceRef = useRef<ReturnType<typeof setTimeout>>();
  const abortRef = useRef<AbortController>();

  useEffect(() => {
    if (!visible || !query) {
      setItems([]);
      return;
    }

    const q = query.toLowerCase();

    // Cancel previous request
    abortRef.current?.abort();

    // Debounce API calls
    clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(async () => {
      const results: MentionItem[] = [];

      // Built-in keywords (no API call needed)
      const keywords = [
        { name: "channel", label: "Notify all members" },
        { name: "here", label: "Notify online members" },
        { name: "everyone", label: "Notify all members" },
      ];
      for (const kw of keywords) {
        if (kw.name.startsWith(q)) {
          results.push({ type: "keyword", ...kw });
        }
      }

      try {
        abortRef.current = new AbortController();
        const params = `?q=${encodeURIComponent(q)}`;
        const [users, groups] = await Promise.all([
          apiFetch<User[]>(`/users${params}`, { signal: abortRef.current.signal }),
          apiFetch<UserGroup[]>(`/groups${params}`, { signal: abortRef.current.signal }),
        ]);

        for (const u of users || []) {
          results.push({ type: "user", user: u });
        }

        for (const g of groups || []) {
          results.push({ type: "group", group: g });
        }
      } catch {
        // ignore fetch errors and aborts
      }

      setItems(results.slice(0, 8));
      setSelectedIndex(0);
    }, 200);

    return () => {
      clearTimeout(debounceRef.current);
      abortRef.current?.abort();
    };
  }, [query, visible]);

  const handleSelect = useCallback((item: MentionItem) => {
    switch (item.type) {
      case "user":
        onSelect(item.user.name);
        break;
      case "group":
        onSelect(item.group.name);
        break;
      case "keyword":
        onSelect(item.name);
        break;
    }
  }, [onSelect]);

  useEffect(() => {
    if (!visible) return;

    const handler = (e: KeyboardEvent) => {
      if (e.key === "ArrowDown") {
        e.preventDefault();
        setSelectedIndex((i) => Math.min(i + 1, items.length - 1));
      } else if (e.key === "ArrowUp") {
        e.preventDefault();
        setSelectedIndex((i) => Math.max(i - 1, 0));
      } else if (e.key === "Enter" || e.key === "Tab") {
        e.preventDefault();
        if (items[selectedIndex]) {
          handleSelect(items[selectedIndex]);
        }
      } else if (e.key === "Escape") {
        onClose();
      }
    };

    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [visible, items, selectedIndex, onClose, handleSelect]);

  if (!visible || items.length === 0) return null;

  return (
    <div
      ref={listRef}
      className="absolute bottom-full left-0 mb-1 w-64 rounded-lg border border-gray-200 bg-white shadow-lg dark:border-gray-700 dark:bg-gray-800"
    >
      {items.map((item, i) => (
        <button
          key={i}
          onClick={() => handleSelect(item)}
          className={`flex w-full items-center gap-2 px-3 py-2 text-left text-sm ${
            i === selectedIndex
              ? "bg-blue-50 dark:bg-blue-900/20"
              : "hover:bg-gray-50 dark:hover:bg-gray-700/50"
          }`}
        >
          {item.type === "user" && (
            <>
              <div className="flex h-6 w-6 items-center justify-center rounded-full bg-blue-500 text-[10px] font-bold text-white">
                {item.user.name.charAt(0).toUpperCase()}
              </div>
              <span className="text-gray-900 dark:text-gray-100">
                {item.user.name}
              </span>
            </>
          )}
          {item.type === "group" && (
            <>
              <div className="flex h-6 w-6 items-center justify-center rounded-full bg-purple-500 text-[10px] font-bold text-white">
                G
              </div>
              <span className="text-gray-900 dark:text-gray-100">
                {item.group.name}
              </span>
            </>
          )}
          {item.type === "keyword" && (
            <>
              <div className="flex h-6 w-6 items-center justify-center rounded-full bg-orange-500 text-[10px] font-bold text-white">
                @
              </div>
              <div>
                <span className="text-gray-900 dark:text-gray-100">
                  {item.name}
                </span>
                <span className="ml-1 text-xs text-gray-500">
                  {item.label}
                </span>
              </div>
            </>
          )}
        </button>
      ))}
    </div>
  );
}
