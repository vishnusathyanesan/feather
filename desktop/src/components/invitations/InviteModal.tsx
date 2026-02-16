import { useState } from "react";
import { useInvitationStore } from "../../stores/invitationStore";

interface Props {
  onClose: () => void;
}

export default function InviteModal({ onClose }: Props) {
  const { createInvitation } = useInvitationStore();
  const [email, setEmail] = useState("");
  const [maxUses, setMaxUses] = useState(10);
  const [ttlDays, setTtlDays] = useState(7);
  const [inviteURL, setInviteURL] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [copied, setCopied] = useState(false);

  const handleCreate = async () => {
    setIsSubmitting(true);
    try {
      const inv = await createInvitation({
        email: email || undefined,
        max_uses: maxUses,
        ttl_days: ttlDays,
      });
      setInviteURL(inv.invite_url);
    } catch {
      // error handled
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCopy = () => {
    if (inviteURL) {
      navigator.clipboard.writeText(inviteURL);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="w-full max-w-md rounded-lg bg-white p-6 shadow-xl dark:bg-surface-secondary">
        <h2 className="mb-4 text-lg font-bold text-gray-900 dark:text-gray-100">
          Invite People
        </h2>

        {inviteURL ? (
          <div className="space-y-4">
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Share this link to invite people to Feather:
            </p>
            <div className="flex items-center gap-2">
              <input
                readOnly
                value={inviteURL}
                className="flex-1 rounded border border-gray-300 bg-gray-50 px-3 py-2 text-sm dark:border-gray-600 dark:bg-gray-800 dark:text-gray-100"
              />
              <button
                onClick={handleCopy}
                className="rounded bg-blue-600 px-3 py-2 text-sm font-medium text-white hover:bg-blue-700"
              >
                {copied ? "Copied!" : "Copy"}
              </button>
            </div>
            <button
              onClick={onClose}
              className="w-full rounded border border-gray-300 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 dark:border-gray-600 dark:text-gray-300"
            >
              Done
            </button>
          </div>
        ) : (
          <div className="space-y-4">
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">
                Email (optional)
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full rounded border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none dark:border-gray-600 dark:bg-surface-secondary dark:text-gray-100"
                placeholder="user@example.com"
              />
            </div>
            <div className="flex gap-4">
              <div className="flex-1">
                <label className="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">
                  Max uses
                </label>
                <input
                  type="number"
                  min={1}
                  max={1000}
                  value={maxUses}
                  onChange={(e) => setMaxUses(Number(e.target.value))}
                  className="w-full rounded border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none dark:border-gray-600 dark:bg-surface-secondary dark:text-gray-100"
                />
              </div>
              <div className="flex-1">
                <label className="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">
                  Expires in (days)
                </label>
                <input
                  type="number"
                  min={1}
                  max={30}
                  value={ttlDays}
                  onChange={(e) => setTtlDays(Number(e.target.value))}
                  className="w-full rounded border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none dark:border-gray-600 dark:bg-surface-secondary dark:text-gray-100"
                />
              </div>
            </div>
            <div className="flex justify-end gap-2">
              <button
                onClick={onClose}
                className="rounded border border-gray-300 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 dark:border-gray-600 dark:text-gray-300"
              >
                Cancel
              </button>
              <button
                onClick={handleCreate}
                disabled={isSubmitting}
                className="rounded bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              >
                {isSubmitting ? "Creating..." : "Create Invite Link"}
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
