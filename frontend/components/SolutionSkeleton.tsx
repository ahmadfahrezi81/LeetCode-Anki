"use client";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Code } from "lucide-react";
import { ProgressBar } from "@/components/ui/progress";

interface SolutionSkeletonProps {
    progress: number;
}

export function SolutionSkeleton({ progress }: SolutionSkeletonProps) {
    return (
        <Card className="mb-6 animate-pulse-subtle">
            <CardHeader>
                <CardTitle className="flex items-center gap-2 text-lg">
                    <Code className="h-5 w-5 text-purple-600" />
                    Complete Solution Breakdown
                </CardTitle>
                <CardDescription>
                    Generating optimal solution analysis...
                </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
                {/* Progress Bar */}
                <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                    <div className="flex items-center justify-between mb-2">
                        <p className="text-sm font-medium text-blue-800">
                            {progress < 30 && "Analyzing problem patterns..."}
                            {progress >= 30 && progress < 60 && "Generating solution approach..."}
                            {progress >= 60 && progress < 85 && "Creating detailed breakdown..."}
                            {progress >= 85 && "Finalizing analysis..."}
                        </p>
                        <span className="text-xs font-medium text-blue-600">
                            {Math.round(progress)}%
                        </span>
                    </div>
                    <ProgressBar progress={progress} />
                </div>

                {/* Skeleton Content */}
                <div className="space-y-4">
                    {/* Pattern Skeleton */}
                    <div>
                        <div className="h-4 w-20 bg-gray-200 rounded mb-2"></div>
                        <div className="h-8 w-48 bg-gray-200 rounded"></div>
                    </div>

                    {/* Why This Pattern Skeleton */}
                    <div>
                        <div className="h-4 w-32 bg-gray-200 rounded mb-2"></div>
                        <div className="space-y-2">
                            <div className="h-3 w-full bg-gray-200 rounded"></div>
                            <div className="h-3 w-5/6 bg-gray-200 rounded"></div>
                        </div>
                    </div>

                    {/* Steps Skeleton */}
                    <div>
                        <div className="h-4 w-36 bg-gray-200 rounded mb-3"></div>
                        <div className="space-y-3">
                            {[1, 2, 3, 4].map((i) => (
                                <div key={i} className="flex gap-3">
                                    <div className="w-6 h-6 bg-gray-200 rounded-full flex-shrink-0"></div>
                                    <div className="flex-1 space-y-2">
                                        <div className="h-3 w-full bg-gray-200 rounded"></div>
                                        <div className="h-3 w-4/5 bg-gray-200 rounded"></div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>

                    {/* Pseudocode Skeleton */}
                    <div>
                        <div className="h-4 w-24 bg-gray-200 rounded mb-3"></div>
                        <div className="h-32 w-full bg-gray-200 rounded"></div>
                    </div>

                    {/* Complexity Skeleton */}
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div className="h-24 bg-gray-200 rounded-lg"></div>
                        <div className="h-24 bg-gray-200 rounded-lg"></div>
                    </div>
                </div>
            </CardContent>
        </Card>
    );
}
