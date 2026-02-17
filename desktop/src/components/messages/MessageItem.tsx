import { memo, useState } from "react";
import type { Message } from "../../types/message";
import { useAuthStore } from "../../stores/authStore";
import { useMessageStore } from "../../stores/messageStore";
import Avatar from "../common/Avatar";
import MarkdownRenderer from "../common/MarkdownRenderer";
import AlertMessage from "../alerts/AlertMessage";
import ReactionPicker from "./ReactionPicker";

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

interface Props {
  message: Message;
  onOpenThread: (messageId: string) => void;
}

function MessageItemInner({ message, onOpenThread }: Props) {
  const { user: currentUser } = useAuthStore();
  const { editMessage, deleteMessage, addReaction, removeReaction } = useMessageStore();
  const [isEditing, setIsEditing] = useState(false);
  const [editContent, setEditContent] = useState(message.content);
  const [showActions, setShowActions] = useState(false);
  const [showReactionPicker, setShowReactionPicker] = useState(false);

  const isOwner = currentUser?.id === message.user_id;
  const isAdmin = currentUser?.role === "admin";

  if (message.is_alert) {
    return <AlertMessage message={message} />;
  }

  const handleEdit = async () => {
    if (editContent.trim() && editContent !== message.content) {
      await editMessage(message.channel_id, message.id, editContent);
    }
    setIsEditing(false);
  };

  const handleDelete = async () => {
    if (confirm("Delete this message?")) {
      await deleteMessage(message.channel_id, message.id);
    }
  };

  const handleReaction = async (emoji: string) => {
    await addReaction(message.id, emoji);
    setShowReactionPicker(false);
  };

  const timestamp = new Date(message.created_at).toLocaleTimeString([], {
    hour: "2-digit",
    minute: "2-digit",
  });

  return (
    <div
      className="group relative flex px-1 py-2 hover:bg-gray-50 dark:hover:bg-gray-800/50"
      onMouseEnter={() => setShowActions(true)}
      onMouseLeave={() => {
        setShowActions(false);
        setShowReactionPicker(false);
      }}
    >
      <Avatar
        name={message.user?.name || "?"}
        url={message.user?.avatar_url}
        size="md"
      />
      <div className="ml-2 flex-1 overflow-hidden">
        <div className="flex items-baseline">
          <span className="text-sm font-bold text-gray-900 dark:text-gray-100">
            {message.user?.name || "Unknown"}
          </span>
          <span className="ml-2 text-xs text-gray-500">{timestamp}</span>
          {message.edited_at && (
            <span className="ml-1 text-xs text-gray-400">(edited)</span>
          )}
        </div>

        {isEditing ? (
          <div className="mt-1">
            <textarea
              value={editContent}
              onChange={(e) => setEditContent(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter" && !e.shiftKey) {
                  e.preventDefault();
                  handleEdit();
                }
                if (e.key === "Escape") setIsEditing(false);
              }}
              className="w-full rounded border border-gray-300 px-2 py-1 text-sm focus:border-blue-500 focus:outline-none dark:border-gray-600 dark:bg-surface-secondary dark:text-gray-100"
              rows={2}
              autoFocus
            />
            <div className="mt-1 flex items-center space-x-2 text-xs text-gray-500">
              <span>Enter to save</span>
              <span>Esc to cancel</span>
            </div>
          </div>
        ) : (
          <div className="text-sm text-gray-800 dark:text-gray-200">
            <MarkdownRenderer content={message.content} />
          </div>
        )}

        {/* File attachments */}
        {message.attachments && message.attachments.length > 0 && (
          <div className="mt-2 flex flex-wrap gap-2">
            {message.attachments.map((att) => {
              const isImage = att.content_type.startsWith("image/");
              const downloadUrl = `${import.meta.env.VITE_API_URL || "http://localhost:8080/api/v1"}/files/${att.id}/download`;
              if (isImage) {
                return (
                  <a
                    key={att.id}
                    href={downloadUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="block max-w-xs overflow-hidden rounded border border-gray-200 dark:border-gray-700"
                  >
                    <img
                      src={downloadUrl}
                      alt={att.filename}
                      className="max-h-48 object-contain"
                      loading="lazy"
                    />
                    <div className="px-2 py-1 text-xs text-gray-500">{att.filename}</div>
                  </a>
                );
              }
              return (
                <a
                  key={att.id}
                  href={downloadUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center gap-2 rounded border border-gray-200 px-3 py-2 text-sm hover:bg-gray-50 dark:border-gray-700 dark:hover:bg-gray-700/50"
                >
                  <svg className="h-5 w-5 flex-shrink-0 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                  </svg>
                  <div className="min-w-0">
                    <div className="truncate text-blue-600 dark:text-blue-400">{att.filename}</div>
                    <div className="text-xs text-gray-400">{formatFileSize(att.size_bytes)}</div>
                  </div>
                </a>
              );
            })}
          </div>
        )}

        {/* Reactions */}
        {message.reactions && message.reactions.length > 0 && (
          <div className="mt-1 flex flex-wrap gap-1">
            {message.reactions.map((r) => {
              const hasReacted = currentUser ? r.users.includes(currentUser.id) : false;
              return (
                <button
                  key={r.emoji}
                  onClick={() => hasReacted ? removeReaction(message.id, r.emoji) : addReaction(message.id, r.emoji)}
                  className={`flex items-center rounded-full border px-2 py-0.5 text-xs hover:bg-gray-100 dark:hover:bg-gray-700 ${hasReacted ? "border-blue-400 bg-blue-50 dark:border-blue-500 dark:bg-blue-900/30" : "border-gray-200 dark:border-gray-600"}`}
                >
                  <span>{r.emoji}</span>
                  <span className="ml-1 text-gray-500">{r.count}</span>
                </button>
              );
            })}
          </div>
        )}

        {/* Thread indicator */}
        {message.reply_count > 0 && (
          <button
            onClick={() => onOpenThread(message.id)}
            className="mt-1 text-xs text-blue-600 hover:underline dark:text-blue-400"
          >
            {message.reply_count} {message.reply_count === 1 ? "reply" : "replies"}
          </button>
        )}
      </div>

      {/* Hover actions */}
      {showActions && !isEditing && (
        <div className="absolute -top-3 right-2 flex rounded border border-gray-200 bg-white shadow-sm dark:border-gray-600 dark:bg-gray-800">
          <button
            onClick={() => setShowReactionPicker(true)}
            className="px-2 py-1 text-xs text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-700"
            title="React"
          >
            😀
          </button>
          <button
            onClick={() => onOpenThread(message.id)}
            className="px-2 py-1 text-xs text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-700"
            title="Reply in thread"
          >
            💬
          </button>
          {isOwner && (
            <button
              onClick={() => {
                setEditContent(message.content);
                setIsEditing(true);
              }}
              className="px-2 py-1 text-xs text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-700"
              title="Edit"
            >
              ✏️
            </button>
          )}
          {(isOwner || isAdmin) && (
            <button
              onClick={handleDelete}
              className="px-2 py-1 text-xs text-red-500 hover:bg-gray-100 dark:hover:bg-gray-700"
              title="Delete"
            >
              🗑
            </button>
          )}
        </div>
      )}

      {showReactionPicker && (
        <div className="absolute -top-10 right-2 z-10">
          <ReactionPicker onSelect={handleReaction} />
        </div>
      )}
    </div>
  );
}

const MessageItem = memo(MessageItemInner);
export default MessageItem;
