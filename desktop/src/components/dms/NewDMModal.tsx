import { useState, useEffect } from "react";
import { apiFetch } from "../../services/api";
import { useDMStore } from "../../stores/dmStore";
import { useChannelStore } from "../../stores/channelStore";
import { useAuthStore } from "../../stores/authStore";
import type { User } from "../../types/user";

interface Props {
  onClose: () => void;
}

export default function NewDMModal({ onClose }: Props) {
  const [users, setUsers] = useState<User[]>([]);
  const [search, setSearch] = useState("");
  const [selected, setSelected] = useState<User[]>([]);
  const [isCreating, setIsCreating] = useState(false);
  const { createDM, createGroupDM } = useDMStore();
  const { setActiveChannel } = useChannelStore();
  const { user: currentUser } = useAuthStore();

  useEffect(() => {
    apiFetch<User[]>("/users").then((data) => {
      setUsers((data || []).filter((u) => u.id !== currentUser?.id));
    });
  }, [currentUser?.id]);

  const filtered = users.filter(
    (u) =>
      u.name.toLowerCase().includes(search.toLowerCase()) ||
      u.email.toLowerCase().includes(search.toLowerCase())
  );

  const toggleUser = (user: User) => {
    setSelected((prev) =>
      prev.find((u) => u.id === user.id)
        ? prev.filter((u) => u.id !== user.id)
        : [...prev, user]
    );
  };

  const handleCreate = async () => {
    if (selected.length === 0) return;
    setIsCreating(true);
    try {
      let dm;
      if (selected.length === 1) {
        dm = await createDM(selected[0].id);
      } else {
        dm = await createGroupDM(selected.map((u) => u.id));
      }
      setActiveChannel(dm.id);
      onClose();
    } catch {
      // error handled
    } finally {
      setIsCreating(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="w-full max-w-md rounded-lg bg-white shadow-xl dark:bg-surface-secondary">
        <div className="border-b border-gray-200 p-4 dark:border-gray-700">
          <h2 className="text-lg font-bold text-gray-900 dark:text-gray-100">
            New Message
          </h2>
          <input
            autoFocus
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search users..."
            className="mt-2 w-full rounded border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none dark:border-gray-600 dark:bg-gray-800 dark:text-gray-100"
          />
          {selected.length > 0 && (
            <div className="mt-2 flex flex-wrap gap-1">
              {selected.map((u) => (
                <span
                  key={u.id}
                  className="flex items-center gap-1 rounded-full bg-blue-100 px-2 py-0.5 text-xs text-blue-800 dark:bg-blue-900/30 dark:text-blue-300"
                >
                  {u.name}
                  <button
                    onClick={() => toggleUser(u)}
                    className="ml-0.5 text-blue-600 hover:text-blue-800"
                  >
                    x
                  </button>
                </span>
              ))}
            </div>
          )}
        </div>

        <div className="max-h-64 overflow-y-auto p-2">
          {filtered.map((u) => {
            const isSelected = selected.some((s) => s.id === u.id);
            return (
              <button
                key={u.id}
                onClick={() => toggleUser(u)}
                className={`flex w-full items-center gap-3 rounded px-3 py-2 text-left text-sm ${
                  isSelected
                    ? "bg-blue-50 dark:bg-blue-900/20"
                    : "hover:bg-gray-50 dark:hover:bg-gray-700/50"
                }`}
              >
                <div className="flex h-8 w-8 items-center justify-center rounded-full bg-gray-300 text-xs font-bold text-gray-700 dark:bg-gray-600 dark:text-gray-300">
                  {u.name.charAt(0).toUpperCase()}
                </div>
                <div className="flex-1">
                  <div className="font-medium text-gray-900 dark:text-gray-100">
                    {u.name}
                  </div>
                  <div className="text-xs text-gray-500">{u.email}</div>
                </div>
                {isSelected && (
                  <span className="text-blue-600">&#10003;</span>
                )}
              </button>
            );
          })}
        </div>

        <div className="flex justify-end gap-2 border-t border-gray-200 p-4 dark:border-gray-700">
          <button
            onClick={onClose}
            className="rounded border border-gray-300 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 dark:border-gray-600 dark:text-gray-300"
          >
            Cancel
          </button>
          <button
            onClick={handleCreate}
            disabled={selected.length === 0 || isCreating}
            className="rounded bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
          >
            {isCreating ? "Creating..." : "Go"}
          </button>
        </div>
      </div>
    </div>
  );
}
