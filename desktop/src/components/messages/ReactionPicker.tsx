interface Props {
  onSelect: (emoji: string) => void;
}

const quickEmojis = ["👍", "❤️", "😂", "🎉", "🤔", "👀", "🚀", "✅"];

export default function ReactionPicker({ onSelect }: Props) {
  return (
    <div className="flex gap-1 rounded-lg border border-gray-200 bg-white p-2 shadow-lg dark:border-gray-600 dark:bg-gray-800">
      {quickEmojis.map((emoji) => (
        <button
          key={emoji}
          onClick={() => onSelect(emoji)}
          className="rounded p-1 text-lg hover:bg-gray-100 dark:hover:bg-gray-700"
        >
          {emoji}
        </button>
      ))}
    </div>
  );
}
