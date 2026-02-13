interface Props {
  name: string;
  url?: string;
  size?: "sm" | "md" | "lg";
}

const sizeMap = {
  sm: "h-6 w-6 text-[10px]",
  md: "h-8 w-8 text-xs",
  lg: "h-10 w-10 text-sm",
};

const colors = [
  "bg-blue-600", "bg-green-600", "bg-purple-600",
  "bg-orange-600", "bg-pink-600", "bg-teal-600",
];

function getColor(name: string) {
  let hash = 0;
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash);
  }
  return colors[Math.abs(hash) % colors.length];
}

export default function Avatar({ name, url, size = "md" }: Props) {
  if (url) {
    return (
      <img
        src={url}
        alt={name}
        className={`${sizeMap[size]} rounded object-cover`}
      />
    );
  }

  return (
    <div
      className={`${sizeMap[size]} ${getColor(name)} flex items-center justify-center rounded font-bold text-white`}
    >
      {name.charAt(0).toUpperCase()}
    </div>
  );
}
