import type { Message } from "../../types/message";
import MetadataCollapsible from "./MetadataCollapsible";
import MarkdownRenderer from "../common/MarkdownRenderer";

interface Props {
  message: Message;
}

const severityStyles = {
  info: {
    border: "border-l-blue-500",
    badge: "bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400",
    bg: "bg-blue-50/50 dark:bg-blue-950/20",
  },
  warning: {
    border: "border-l-yellow-500",
    badge: "bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400",
    bg: "bg-yellow-50/50 dark:bg-yellow-950/20",
  },
  critical: {
    border: "border-l-red-500",
    badge: "bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400",
    bg: "bg-red-50/50 dark:bg-red-950/20",
  },
};

export default function AlertMessage({ message }: Props) {
  const severity = (message.alert_severity || "info") as keyof typeof severityStyles;
  const styles = severityStyles[severity] || severityStyles.info;

  const timestamp = new Date(message.created_at).toLocaleTimeString([], {
    hour: "2-digit",
    minute: "2-digit",
  });

  return (
    <div className={`my-2 rounded border-l-4 ${styles.border} ${styles.bg} p-3`}>
      <div className="flex items-center gap-2">
        <span className={`rounded px-1.5 py-0.5 text-[10px] font-bold uppercase ${styles.badge}`}>
          {severity}
        </span>
        <span className="text-xs text-gray-500">{timestamp}</span>
      </div>
      <div className="mt-1 text-sm text-gray-800 dark:text-gray-200">
        <MarkdownRenderer content={message.content} />
      </div>
      {message.alert_metadata && Object.keys(message.alert_metadata).length > 0 && (
        <MetadataCollapsible metadata={message.alert_metadata as Record<string, unknown>} />
      )}
    </div>
  );
}
