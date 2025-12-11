"use client";

import { useState, useEffect, useCallback } from "react";

interface UseSolutionLoadingReturn {
    isLoading: boolean;
    progress: number;
    startLoading: () => void;
    completeLoading: () => void;
    resetLoading: () => void;
}

export function useSolutionLoading(): UseSolutionLoadingReturn {
    const [isLoading, setIsLoading] = useState(false);
    const [progress, setProgress] = useState(0);

    const startLoading = useCallback(() => {
        setIsLoading(true);
        setProgress(0);
    }, []);

    const completeLoading = useCallback(() => {
        setProgress(100);
        setTimeout(() => {
            setIsLoading(false);
            setProgress(0);
        }, 300);
    }, []);

    const resetLoading = useCallback(() => {
        setIsLoading(false);
        setProgress(0);
    }, []);

    useEffect(() => {
        if (!isLoading) return;

        // Simulate progress with a realistic curve
        // Fast initial progress, then slows down
        const intervals: NodeJS.Timeout[] = [];
        
        // 0-30% in first 2 seconds (fast start)
        const interval1 = setInterval(() => {
            setProgress((prev) => {
                if (prev >= 30) return prev;
                return prev + 3;
            });
        }, 200);
        intervals.push(interval1);

        // 30-60% in next 5 seconds (medium)
        setTimeout(() => {
            const interval2 = setInterval(() => {
                setProgress((prev) => {
                    if (prev >= 60) return prev;
                    return prev + 1.5;
                });
            }, 250);
            intervals.push(interval2);
        }, 2000);

        // 60-85% in next 8 seconds (slow)
        setTimeout(() => {
            const interval3 = setInterval(() => {
                setProgress((prev) => {
                    if (prev >= 85) return prev;
                    return prev + 0.5;
                });
            }, 400);
            intervals.push(interval3);
        }, 7000);

        // 85-95% very slow (to indicate we're waiting)
        setTimeout(() => {
            const interval4 = setInterval(() => {
                setProgress((prev) => {
                    if (prev >= 95) return prev;
                    return prev + 0.2;
                });
            }, 1000);
            intervals.push(interval4);
        }, 15000);

        return () => {
            intervals.forEach(clearInterval);
        };
    }, [isLoading]);

    return {
        isLoading,
        progress,
        startLoading,
        completeLoading,
        resetLoading,
    };
}
