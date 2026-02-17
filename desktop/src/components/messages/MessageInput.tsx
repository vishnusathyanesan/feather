import { useState, useRef, useCallback } from "react";
import { useMessageStore } from "../../stores/messageStore";
import { wsService } from "../../services/websocket";
import { apiUpload } from "../../services/api";
import type { FileAttachment } from "../../types/message";
import MentionAutocomplete from "../mentions/MentionAutocomplete";

interface Props {
  channelId: string;
  parentId?: string;
  placeholder?: string;
}

interface PendingFile {
  file: File;
  uploading: boolean;
  attachment?: FileAttachment;
  error?: string;
}

export default function MessageInput({ channelId, parentId, placeholder }: Props) {
  const [content, setContent] = useState("");
  const [sendError, setSendError] = useState(false);
  const [mentionQuery, setMentionQuery] = useState("");
  const [showMentions, setShowMentions] = useState(false);
  const [pendingFiles, setPendingFiles] = useState<PendingFile[]>([]);
  const { sendMessage } = useMessageStore();
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const lastTypingSent = useRef(0);

  const uploadFile = useCallback(async (file: File): Promise<FileAttachment | null> => {
    const formData = new FormData();
    formData.append("file", file);
    try {
      return await apiUpload<FileAttachment>(`/channels/${channelId}/files`, formData);
    } catch {
      return null;
    }
  }, [channelId]);

  const handleFilesSelected = useCallback(async (files: FileList | File[]) => {
    const fileArray = Array.from(files).slice(0, 5); // Max 5 files
    const newPending: PendingFile[] = fileArray.map((f) => ({ file: f, uploading: true }));
    setPendingFiles((prev) => [...prev, ...newPending]);

    for (let i = 0; i < fileArray.length; i++) {
      const attachment = await uploadFile(fileArray[i]);
      setPendingFiles((prev) => {
        const updated = [...prev];
        const idx = updated.findIndex((p) => p.file === fileArray[i] && p.uploading);
        if (idx !== -1) {
          updated[idx] = attachment
            ? { file: fileArray[i], uploading: false, attachment }
            : { file: fileArray[i], uploading: false, error: "Upload failed" };
        }
        return updated;
      });
    }
  }, [uploadFile]);

  const removePendingFile = useCallback((index: number) => {
    setPendingFiles((prev) => prev.filter((_, i) => i !== index));
  }, []);

  const handleSend = useCallback(async () => {
    const trimmed = content.trim();
    const attachmentIds = pendingFiles
      .filter((p) => p.attachment)
      .map((p) => p.attachment!.id);

    if (!trimmed && attachmentIds.length === 0) return;

    setSendError(false);
    try {
      await sendMessage(channelId, {
        content: trimmed || " ",
        parent_id: parentId,
        attachment_ids: attachmentIds.length > 0 ? attachmentIds : undefined,
      });
      setContent("");
      setPendingFiles([]);
      if (textareaRef.current) {
        textareaRef.current.style.height = "auto";
      }
    } catch {
      setSendError(true);
    }
  }, [content, channelId, parentId, sendMessage, pendingFiles]);

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
    const atMatch = textBeforeCursor.match(/@([\w.\-]*)$/);
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

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    if (e.dataTransfer.files.length > 0) {
      handleFilesSelected(e.dataTransfer.files);
    }
  }, [handleFilesSelected]);

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
  }, []);

  const formatSize = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  };

  return (
    <div
      className="border-t border-gray-200 px-2 py-2 dark:border-gray-700 sm:px-4 sm:py-3"
      onDrop={handleDrop}
      onDragOver={handleDragOver}
    >
      {sendError && (
        <div className="mb-2 text-xs text-red-500">Failed to send message. Please try again.</div>
      )}

      {/* Pending file previews */}
      {pendingFiles.length > 0 && (
        <div className="mb-2 flex flex-wrap gap-2">
          {pendingFiles.map((pf, i) => (
            <div
              key={i}
              className="flex items-center gap-1.5 rounded border border-gray-200 bg-gray-50 px-2 py-1 text-xs dark:border-gray-600 dark:bg-gray-800"
            >
              {pf.uploading ? (
                <span className="text-gray-400">Uploading...</span>
              ) : pf.error ? (
                <span className="text-red-500">{pf.error}</span>
              ) : (
                <>
                  <span className="max-w-[120px] truncate text-gray-700 dark:text-gray-300" title={pf.file.name}>
                    {pf.file.name}
                  </span>
                  <span className="text-gray-400">{formatSize(pf.file.size)}</span>
                </>
              )}
              <button
                onClick={() => removePendingFile(i)}
                className="ml-1 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"
              >
                x
              </button>
            </div>
          ))}
        </div>
      )}

      <div className="relative flex items-end rounded border border-gray-300 bg-white dark:border-gray-600 dark:bg-surface-secondary">
        <MentionAutocomplete
          query={mentionQuery}
          onSelect={handleMentionSelect}
          onClose={() => setShowMentions(false)}
          visible={showMentions}
        />

        {/* File upload button */}
        <button
          onClick={() => fileInputRef.current?.click()}
          className="px-2 py-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
          title="Attach file"
        >
          <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
          </svg>
        </button>
        <input
          ref={fileInputRef}
          type="file"
          multiple
          className="hidden"
          onChange={(e) => {
            if (e.target.files && e.target.files.length > 0) {
              handleFilesSelected(e.target.files);
              e.target.value = "";
            }
          }}
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
          disabled={!content.trim() && pendingFiles.filter((p) => p.attachment).length === 0}
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
