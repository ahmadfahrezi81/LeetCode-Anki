import { cn } from "@/lib/utils";
import { StreakFlame } from "./StreakFlame";

interface StreakDisplayProps {
  count: number;
  active?: boolean;
  className?: string;
  size?: "sm" | "base" | "lg";
}

export function StreakDisplay({
  count,
  active = false,
  className,
  size = "base",
}: StreakDisplayProps) {
  return (
    <div
      className={cn(
        "h-9 px-2 rounded-md flex items-center gap-0.5",
        className
      )}
    >
      <StreakFlame active={active} size={size} />
      <span
        className={cn(
          "text-lg font-extrabold transition-colors",
          active && "text-red-500",
          !active && "text-gray-400"
        )}
      >
        {count}
      </span>
    </div>
  );
}