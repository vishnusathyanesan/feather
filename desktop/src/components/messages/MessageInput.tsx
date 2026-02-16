import { useState, useRef, useCallback } from "react";
import { useMessageStore } from "../../stores/messageStore";
import { wsService } from "../../services/websocket";
import MentionAutocomplete from "../mentions/MentionAutocomplete";

interface Props {
  channelId: string;
  parentId?: string;
  placeholder?: string;
}

export default function MessageInput({ channelId, parentId, placeholder }: Props) {
  const [content, setContent] = useState("");
  const [sendError, setSendError] = useState(false);
  const [mentionQuery, setMentionQuery] = useState("");
  const [showMentions, setShowMentions] = useState(false);
  const { sendMessage } = useMessageStore();
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const lastTypingSent = useRef(0);

  const handleSend = useCallback(async () => {
    const trimmed = content.trim();
    if (!trimmed) return;

    setSendError(false);
    try {
      await sendMessage(channelId, {
        content: trimmed,
        parent_id: parentId,
      });
      setContent("");
      if (textareaRef.current) {
        textareaRef.current.style.height = "auto";
      }
    } catch {
      setSendError(true);
    }
  }, [content, channelId, parentId, sendMessage]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (showMentions) return; // Let autocomplete handle keys
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const value = e.target.value;
    setContent(value);

    // Auto-grow textarea
    const el = e.target;
    el.style.height = "auto";
    el.style.height = Math.min(el.scrollHeight, 200) + "px";

    // Check for @mention trigger
    const cursorPos = el.selectionStart;
    const textBeforeCursor = value.substring(0, cursorPos);
    const atMatch = textBeforeCursor.match(/@(\w*)$/);
    if (atMatch) {
      setMentionQuery(atMatch[1]);
      setShowMentions(true);
    } else {
      setShowMentions(false);
      setMentionQuery("");
    }

    // Debounced typing indicator
    const now = Date.now();
    if (now - lastTypingSent.current > 3000) {
      wsService.sendTyping(channelId);
      lastTypingSent.current = now;
    }
  };

  const handleMentionSelect = (name: string) => {
    const cursorPos = textareaRef.current?.selectionStart ?? content.length;
    const textBeforeCursor = content.substring(0, cursorPos);
    const atIndex = textBeforeCursor.lastIndexOf("@");
    if (atIndex === -1) return;

    const before = content.substring(0, atIndex);
    const after = content.substring(cursorPos);
    const newContent = `${before}@${name} ${after}`;
    setContent(newContent);
    setShowMentions(false);
    setMentionQuery("");

    // Restore focus
    setTimeout(() => {
      if (textareaRef.current) {
        const newPos = atIndex + name.length + 2;
        textareaRef.current.focus();
        textareaRef.current.setSelectionRange(newPos, newPos);
      }
    }, 0);
  };

  return (
    <div className="border-t border-gray-200 px-2 py-2 dark:border-gray-700 sm:px-4 sm:py-3">
      {sendError && (
        <div className="mb-2 text-xs text-red-500">Failed to send message. Please try again.</div>
      )}
      <div className="relative flex items-end rounded border border-gray-300 bg-white dark:border-gray-600 dark:bg-surface-secondary">
        <MentionAutocomplete
          query={mentionQuery}
          onSelect={handleMentionSelect}
          onClose={() => setShowMentions(false)}
          visible={showMentions}
        />
        <textarea
          ref={textareaRef}
          value={content}
          onChange={handleChange}
          onKeyDown={handleKeyDown}
          placeholder={placeholder || `Message #channel`}
          className="flex-1 resize-none bg-transparent px-3 py-2 text-sm focus:outline-none dark:text-gray-100"
          rows={1}
        />
        <button
          onClick={handleSend}
          disabled={!content.trim()}
          className="px-3 py-2 text-blue-600 hover:text-blue-700 disabled:text-gray-400 dark:text-blue-400"
        >
          <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 20 20">
            <path d="M10.894 2.553a1 1 0 00-1.788 0l-7 14a1 1 0 001.169 1.409l5-1.429A1 1 0 009 15.571V11a1 1 0 112 0v4.571a1 1 0 00.725.962l5 1.428a1 1 0 001.17-1.408l-7-14z" />
          </svg>
        </button>
      </div>
    </div>
  );
}
