import { Flame, FlameKindling } from "lucide-react";
import { cn } from "@/lib/utils";

interface StreakFlameProps {
  className?: string;
  iconClassName?: string;
  size?: "sm" | "base" | "lg";
  active?: boolean;
}

const SIZE_MAP = {
  sm: { iconSize: 20 },
  base: { iconSize: 28 },
  lg: { iconSize: 36 },
} as const;

export function StreakFlame({
  className,
  iconClassName,
  size = "base",
  active = false,
}: StreakFlameProps) {
  const { iconSize } = SIZE_MAP[size];

  return (
    <div
      className={cn(
        "inline-flex items-center justify-center flex-none",
        className
      )}
    >
      <Flame
        size={iconSize}
        className={cn(
          "flex-none transition-colors",
          // Active state: Orange outline with Yellow-200 fill
          active && "text-orange-500 fill-yellow-300", 
          // Inactive state: Gray outline with no fill (transparent)
          !active && "text-gray-400 fill-gray-300",
          iconClassName
        )}
        strokeWidth={3}
      />
    </div>
  );
}