import type { ReactNode } from "react";

interface Props {
  children: ReactNode;
}

export default function AppLayout({ children }: Props) {
  return (
    <div className="flex h-full bg-surface text-gray-900 dark:text-gray-100">
      {children}
    </div>
  );
}
