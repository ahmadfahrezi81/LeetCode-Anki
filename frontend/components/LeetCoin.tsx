import { Star } from "lucide-react";
import { cn } from "@/lib/utils";

interface LeetCoinProps {
  className?: string;
  iconClassName?: string;
  size?: "sm" | "base" | "lg";
}

const SIZE_MAP = {
  sm: {
    containerSize: "w-6 h-6", // 20px total
    padding: "p-0.5",
    border: "border-3",
    iconSize: 12,
  },
  base: {
    containerSize: "w-7 h-7", // 28px total
    padding: "p-1",
    border: "border-[3px]",
    iconSize: 14,
  },
  lg: {
    containerSize: "w-9 h-9", // 36px total
    padding: "p-1.5",
    border: "border-4",
    iconSize: 18,
  },
} as const;

export function LeetCoin({
  className,
  iconClassName,
  size = "base",
}: LeetCoinProps) {
  const { containerSize, padding, border, iconSize } = SIZE_MAP[size];

  return (
    <div
      className={cn(
        // Fixed dimensions to prevent parent interference
        "inline-flex items-center justify-center rounded-full",
        "bg-orange-400 border-yellow-300",
        "flex-shrink-0 flex-grow-0", // Prevent flex container resizing
        containerSize,
        padding,
        border,
        className
      )}
      style={{
        // Inline styles for maximum specificity
        minWidth: "auto",
        minHeight: "auto",
      }}
    >
      <Star
        size={iconSize}
        className={cn(
          "text-yellow-300 fill-yellow-300",
          "flex-shrink-0 flex-grow-0", // Prevent icon from being resized
          iconClassName
        )}
        style={{ 
          width: `${iconSize}px`, 
          height: `${iconSize}px`,
          minWidth: `${iconSize}px`,
          minHeight: `${iconSize}px`,
          maxWidth: `${iconSize}px`,
          maxHeight: `${iconSize}px`,
        }}
        strokeWidth={2} // Remove stroke for cleaner look
      />
    </div>
  );
}