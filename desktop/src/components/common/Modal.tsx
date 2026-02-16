import { useEffect, type ReactNode } from "react";

interface Props {
  children: ReactNode;
  onClose: () => void;
}

export default function Modal({ children, onClose }: Props) {
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [onClose]);

  return (
    <div
      className="fixed inset-0 z-50 flex items-end justify-center bg-black/50 sm:items-start sm:pt-[15vh]"
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div className="max-h-[90vh] w-full overflow-y-auto rounded-t-lg bg-white shadow-xl dark:bg-[#2a2a3e] sm:max-w-lg sm:rounded-lg">
        {children}
      </div>
    </div>
  );
}
