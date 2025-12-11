"use client";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Textarea } from "@/components/ui/textarea";
import { Button } from "@/components/ui/button";
import { Lightbulb, Send, SkipForward, Mic } from "lucide-react";

interface AnswerInputProps {
    answer: string;
    onAnswerChange: (value: string) => void;
    onSubmit: () => void;
    onSkip: () => void;
    submitting: boolean;
    skipping: boolean;
    isRecording: boolean;
    isTranscribing: boolean;
    onStartRecording: () => void;
    onStopRecording: () => void;
}

export default function AnswerInput({
    answer,
    onAnswerChange,
    onSubmit,
    onSkip,
    submitting,
    skipping,
    isRecording,
    isTranscribing,
    onStartRecording,
    onStopRecording,
}: AnswerInputProps) {
    return (
        <Card className="bg-white shadow-sm py-4 md:py-6">
            <CardHeader className="px-4 md:px-6">
                <CardTitle className="flex items-center gap-2 text-gray-900">
                    <Lightbulb className="h-5 w-5 text-yellow-500" />
                    Explain Your Approach
                </CardTitle>
                <CardDescription className="text-gray-600">
                    Describe the algorithm and data structures you would use to solve this problem
                </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6 px-4 md:px-6">
                <div className="relative">
                    <Textarea
                        placeholder="Example: I would use a hashmap to store complements. For each number, I check if target minus the current number exists in the map..."
                        value={answer}
                        onChange={(e) => onAnswerChange(e.target.value)}
                        className="max-h-60 min-h-32 resize-none bg-white border-gray-300 pr-12"
                        disabled={isTranscribing}
                    />
                    <Button
                        type="button"
                        variant={isRecording ? "destructive" : "outline"}
                        size="icon"
                        className={`absolute top-2 right-2 ${isRecording ? 'animate-pulse' : ''}`}
                        onClick={isRecording ? onStopRecording : onStartRecording}
                        disabled={submitting || skipping || isTranscribing}
                    >
                        <Mic className={`h-4 w-4 ${isRecording ? 'text-white' : ''}`} />
                    </Button>
                    {isRecording && (
                        <div className="absolute top-2 right-14 flex items-center gap-2 bg-red-100 text-red-700 px-3 py-1 rounded-full text-xs font-medium border border-red-300">
                            <div className="w-2 h-2 bg-red-600 rounded-full animate-pulse" />
                            Recording...
                        </div>
                    )}
                    {isTranscribing && (
                        <div className="absolute top-2 right-14 flex items-center gap-2 bg-blue-100 text-blue-700 px-3 py-1 rounded-full text-xs font-medium border border-blue-300">
                            <div className="w-3 h-3 border-2 border-blue-600 border-t-transparent rounded-full animate-spin" />
                            Transcribing...
                        </div>
                    )}
                </div>
                <div className="flex gap-2">
                    <Button
                        variant="outline"
                        onClick={onSkip}
                        disabled={skipping || submitting || isTranscribing}
                        className="flex-[0.4]"
                    >
                        {skipping ? (
                            <>Processing...</>
                        ) : (
                            <>
                                <SkipForward className="mr-2 h-4 w-4" />
                                Skip (30 min)
                            </>
                        )}
                    </Button>
                    <Button
                        onClick={onSubmit}
                        disabled={submitting || !answer.trim() || skipping || isTranscribing}
                        className="flex-[1.6] bg-blue-600 hover:bg-blue-700 text-white"
                    >
                        {submitting ? (
                            <>Processing...</>
                        ) : (
                            <>
                                <Send className="mr-2 h-4 w-4" />
                                Submit Answer
                            </>
                        )}
                    </Button>
                </div>
            </CardContent>
        </Card>
    );
}
