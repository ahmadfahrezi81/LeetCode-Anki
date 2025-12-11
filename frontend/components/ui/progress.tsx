"use client";

import { motion } from "framer-motion";

interface ProgressBarProps {
    progress: number; // 0-100
    className?: string;
    showPercentage?: boolean;
}

export function ProgressBar({ progress, className = "", showPercentage = false }: ProgressBarProps) {
    return (
        <div className={`w-full ${className}`}>
            <div className="flex items-center justify-between mb-1">
                {showPercentage && (
                    <span className="text-xs font-medium text-gray-600">
                        {Math.round(progress)}%
                    </span>
                )}
            </div>
            <div className="h-2 w-full overflow-hidden rounded-full bg-gray-200">
                <motion.div
                    className="h-full bg-gradient-to-r from-blue-500 to-blue-600 rounded-full"
                    initial={{ width: 0 }}
                    animate={{ width: `${progress}%` }}
                    transition={{ duration: 0.3, ease: "easeOut" }}
                />
            </div>
        </div>
    );
}
