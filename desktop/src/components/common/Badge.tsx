interface Props {
  count: number;
}

export default function Badge({ count }: Props) {
  if (count <= 0) return null;

  return (
    <span className="ml-auto flex h-5 min-w-[20px] items-center justify-center rounded-full bg-red-500 px-1.5 text-[10px] font-bold text-white">
      {count > 99 ? "99+" : count}
    </span>
  );
}
