import { useState } from "react";

interface Props {
  metadata: Record<string, unknown>;
}

export default function MetadataCollapsible({ metadata }: Props) {
  const [isOpen, setIsOpen] = useState(false);

  const entries = Object.entries(metadata);
  if (entries.length === 0) return null;

  return (
    <div className="mt-2">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="text-xs font-medium text-gray-600 hover:text-gray-800 dark:text-gray-400 dark:hover:text-gray-200"
      >
        {isOpen ? "▾" : "▸"} Metadata ({entries.length})
      </button>
      {isOpen && (
        <div className="mt-1 rounded bg-gray-100/50 p-2 dark:bg-gray-800/50">
          <table className="w-full text-xs">
            <tbody>
              {entries.map(([key, value]) => (
                <tr key={key}>
                  <td className="py-0.5 pr-3 font-medium text-gray-600 dark:text-gray-400">
                    {key}
                  </td>
                  <td className="py-0.5 text-gray-800 dark:text-gray-200">
                    {String(value)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
