"use client";

import { motion, AnimatePresence } from "framer-motion";
import { Trophy, AlertTriangle, CheckCircle2 } from "lucide-react";

interface FeedbackOverlayProps {
    show: boolean;
    score: number;
    isSuccess: boolean;
    onDismiss: () => void;
}

export default function FeedbackOverlay({ 
    show, 
    score, 
    isSuccess, 
    onDismiss 
}: FeedbackOverlayProps) {
    return (
        <AnimatePresence>
            {show && (
                <motion.div
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    exit={{ opacity: 0 }}
                    transition={{ duration: 0.3 }}
                    className="fixed inset-0 z-50 flex items-center justify-center bg-black/20 backdrop-blur-sm"
                    onClick={onDismiss}
                >
                    <motion.div
                        initial={{ scale: 0.8, opacity: 0 }}
                        animate={{ scale: 1, opacity: 1 }}
                        exit={{ scale: 0.9, opacity: 0 }}
                        transition={{ 
                            type: "spring", 
                            stiffness: 200, 
                            damping: 20 
                        }}
                        className={`relative rounded-3xl p-12 text-center shadow-2xl max-w-md mx-4 ${
                            isSuccess 
                                ? 'bg-gradient-to-br from-green-50 to-emerald-50 border-2 border-green-200' 
                                : 'bg-gradient-to-br from-orange-50 to-amber-50 border-2 border-orange-200'
                        }`}
                    >
                        {/* Icon */}
                        <motion.div
                            initial={{ scale: 0 }}
                            animate={{ scale: 1 }}
                            transition={{ 
                                delay: 0.1,
                                type: "spring", 
                                stiffness: 200, 
                                damping: 15 
                            }}
                            className="mb-6 flex justify-center"
                        >
                            {isSuccess ? (
                                <div className="rounded-full bg-green-100 p-6">
                                    <Trophy className="h-16 w-16 text-green-600" />
                                </div>
                            ) : (
                                <div className="rounded-full bg-orange-100 p-6">
                                    <AlertTriangle className="h-16 w-16 text-orange-600" />
                                </div>
                            )}
                        </motion.div>

                        {/* Message */}
                        <motion.div
                            initial={{ y: 10, opacity: 0 }}
                            animate={{ y: 0, opacity: 1 }}
                            transition={{ delay: 0.2 }}
                        >
                            <h2 className={`text-3xl font-bold mb-2 ${
                                isSuccess ? 'text-green-700' : 'text-orange-700'
                            }`}>
                                {isSuccess ? 'Excellent!' : 'Keep Trying!'}
                            </h2>
                            <p className="text-gray-600 mb-4">
                                {isSuccess 
                                    ? 'Great understanding of the problem!' 
                                    : 'Review the feedback to improve'}
                            </p>
                        </motion.div>

                        {/* Score */}
                        <div className="flex justify-center gap-3">
                            <motion.div
                                initial={{ scale: 0.8, opacity: 0 }}
                                animate={{ scale: 1, opacity: 1 }}
                                transition={{ delay: 0.3 }}
                                className={`inline-flex items-center gap-2 px-6 py-3 rounded-full text-2xl font-bold ${
                                    isSuccess 
                                        ? 'bg-green-100 text-green-700 border-2 border-green-300' 
                                        : 'bg-orange-100 text-orange-700 border-2 border-orange-300'
                                }`}
                            >
                                <CheckCircle2 className={`h-6 w-6 ${isSuccess ? 'text-green-600' : 'text-orange-600'}`} />
                                {score}/5
                            </motion.div>
                        </div>

                        {/* Tap to continue hint */}
                        <motion.p
                            initial={{ opacity: 0 }}
                            animate={{ opacity: 1 }}
                            transition={{ delay: 0.8 }}
                            className="text-xs text-gray-400 mt-6"
                        >
                            Tap anywhere to continue
                        </motion.p>
                    </motion.div>
                </motion.div>
            )}
        </AnimatePresence>
    );
}
