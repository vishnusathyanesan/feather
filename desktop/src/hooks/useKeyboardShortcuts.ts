import { useEffect } from "react";

interface Shortcut {
  key: string;
  meta?: boolean;
  ctrl?: boolean;
  handler: () => void;
}

export function useKeyboardShortcuts(shortcuts: Shortcut[]) {
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      for (const shortcut of shortcuts) {
        const metaMatch = shortcut.meta ? e.metaKey : true;
        const ctrlMatch = shortcut.ctrl ? e.ctrlKey : true;
        if (e.key === shortcut.key && metaMatch && ctrlMatch) {
          e.preventDefault();
          shortcut.handler();
          return;
        }
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [shortcuts]);
}
