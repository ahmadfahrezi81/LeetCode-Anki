"use client";

import { motion, AnimatePresence } from "framer-motion";
import { CheckCircle2, Loader2, XCircle, Info } from "lucide-react";

export type ToastType = "success" | "error" | "info" | "loading";

interface ToastProps {
    show: boolean;
    message: string;
    type?: ToastType;
    onClose?: () => void;
}

export function Toast({ show, message, type = "info", onClose }: ToastProps) {
    const icons = {
        success: <CheckCircle2 className="h-5 w-5 text-green-600" />,
        error: <XCircle className="h-5 w-5 text-red-600" />,
        info: <Info className="h-5 w-5 text-blue-600" />,
        loading: <Loader2 className="h-5 w-5 text-blue-600 animate-spin" />,
    };

    const styles = {
        success: "bg-green-50 border-green-200 text-green-800",
        error: "bg-red-50 border-red-200 text-red-800",
        info: "bg-blue-50 border-blue-200 text-blue-800",
        loading: "bg-blue-50 border-blue-200 text-blue-800",
    };

    return (
        <AnimatePresence>
            {show && (
                <motion.div
                    initial={{ opacity: 0, y: -50, scale: 0.95 }}
                    animate={{ opacity: 1, y: 0, scale: 1 }}
                    exit={{ opacity: 0, y: -50, scale: 0.95 }}
                    transition={{ duration: 0.2, ease: "easeOut" }}
                    className="fixed top-4 right-4 z-50 max-w-md"
                >
                    <div
                        className={`flex items-center gap-3 px-4 py-3 rounded-lg border shadow-lg ${styles[type]}`}
                    >
                        {icons[type]}
                        <p className="text-sm font-medium flex-1">{message}</p>
                        {onClose && type !== "loading" && (
                            <button
                                onClick={onClose}
                                className="text-gray-500 hover:text-gray-700 transition-colors"
                            >
                                <XCircle className="h-4 w-4" />
                            </button>
                        )}
                    </div>
                </motion.div>
            )}
        </AnimatePresence>
    );
}
