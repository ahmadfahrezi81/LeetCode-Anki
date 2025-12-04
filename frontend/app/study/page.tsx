"use client";

import { useEffect, useState, Children, cloneElement, isValidElement } from "react";
import { useRouter } from "next/navigation";
import { supabase } from "@/lib/supabase";
import { api } from "@/lib/api";
import type { Card as CardType, SubmitAnswerResponse } from "@/types";
import { Button } from "@/components/ui/button";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";
import { Textarea } from "@/components/ui/textarea";
import {
    Brain,
    ArrowLeft,
    Send,
    Trophy,
    Target,
    BookOpen,
    Lightbulb,
    SkipForward,
    Clock,
    CheckCircle2,
    XCircle,
    TrendingUp,
    Code,
    Zap,
    AlertTriangle,
    Mic,
    ExternalLink,
} from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import SubmissionReport from "@/components/SubmissionReport";
import { div } from "framer-motion/client";


export default function StudyPage() {
    const router = useRouter();
    const [loading, setLoading] = useState(true);
    const [card, setCard] = useState<CardType | null>(null);
    const [answer, setAnswer] = useState("");
    const [submitting, setSubmitting] = useState(false);
    const [result, setResult] = useState<SubmitAnswerResponse | null>(null);
    const [showResult, setShowResult] = useState(false);
    const [skipping, setSkipping] = useState(false);
    const [showTopics, setShowTopics] = useState(false);
    const [showQuestionInReport, setShowQuestionInReport] = useState(true);
    const [isRecording, setIsRecording] = useState(false);
    const [isTranscribing, setIsTranscribing] = useState(false);
    const [mediaRecorder, setMediaRecorder] = useState<MediaRecorder | null>(null);

    useEffect(() => {
        let ignore = false;

        const checkAuth = async () => {
            const { data } = await supabase.auth.getSession();
            if (ignore) return;

            if (!data.session) {
                router.push("/login");
                return;
            }
            loadCard();
        };

        checkAuth();

        return () => {
            ignore = true;
        };
    }, []);

    const loadCard = async () => {
        try {
            const data = await api.getNextCard();
            if (data.card) {
                setCard(data.card);
            } else {
                // No cards available
                router.push("/");
            }
        } catch (err) {
            console.error("Failed to load card:", err);
        } finally {
            setLoading(false);
        }
    };

    const startRecording = async () => {
        try {
            const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
            const recorder = new MediaRecorder(stream);
            const chunks: Blob[] = [];

            recorder.ondataavailable = (e) => {
                if (e.data.size > 0) {
                    chunks.push(e.data);
                }
            };

            recorder.onstop = async () => {
                setIsTranscribing(true);
                const audioBlob = new Blob(chunks, { type: 'audio/webm' });
                
                // Send to backend for transcription
                const formData = new FormData();
                formData.append('audio', audioBlob, 'recording.webm');

                try {
                    const { data: { session } } = await supabase.auth.getSession();
                    const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/transcribe`, {
                        method: 'POST',
                        headers: {
                            'Authorization': `Bearer ${session?.access_token}`,
                        },
                        body: formData,
                    });

                    if (!response.ok) {
                        throw new Error('Transcription failed');
                    }

                    const result = await response.json();
                    // Append transcribed text to answer
                    setAnswer(prev => prev ? `${prev}\n\n${result.text}` : result.text);
                } catch (err) {
                    console.error('Transcription error:', err);
                    alert('Failed to transcribe audio. Please try again.');
                } finally {
                    setIsTranscribing(false);
                }

                // Stop all tracks
                stream.getTracks().forEach(track => track.stop());
            };

            recorder.start();
            setMediaRecorder(recorder);
            setIsRecording(true);
        } catch (err) {
            console.error('Failed to start recording:', err);
            alert('Failed to access microphone. Please check permissions.');
        }
    };

    const stopRecording = () => {
        if (mediaRecorder && mediaRecorder.state !== 'inactive') {
            mediaRecorder.stop();
            setIsRecording(false);
            setMediaRecorder(null);
        }
    };

    const handleSubmit = async () => {
        if (!card || !answer.trim()) return;

        setSubmitting(true);
        try {
            const response = await api.submitAnswer(card.question.id, answer);
            console.log("🔍 Full response from backend:", response);
            console.log("📊 Sub-scores:", response.sub_scores);
            console.log("💻 Solution breakdown:", response.solution_breakdown);
            setResult(response);
            setShowResult(true);
        } catch (err) {
            console.error("Failed to submit answer:", err);
            alert("Failed to submit answer. Please try again.");
        } finally {
            setSubmitting(false);
        }
    };

    const handleSkip = async () => {
        if (!card) return;

        setSkipping(true);
        try {
            await api.skipCard(card.question.id);
            // Load next card
            setAnswer("");
            loadCard();
        } catch (err) {
            console.error("Failed to skip card:", err);
            alert("Failed to skip card. Please try again.");
        } finally {
            setSkipping(false);
        }
    };

    const handleNext = () => {
        setAnswer("");
        setResult(null);
        setShowResult(false);
        loadCard();
    };


    if (loading) {
        return (
            <div className="flex min-h-screen items-center justify-center bg-gray-50">
                <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-600"></div>
            </div>
        );
    }

    if (!card) {
        return (
            <div className="flex min-h-screen items-center justify-center bg-gray-50">
                <Card className="max-w-md">
                    <CardHeader>
                        <CardTitle>No Cards Available</CardTitle>
                        <CardDescription>
                            Check back later for more problems!
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <Button onClick={() => router.push("/")}>
                            <ArrowLeft className="mr-2 h-4 w-4" />
                            Back to Dashboard
                        </Button>
                    </CardContent>
                </Card>
            </div>
        );
    }

    const difficultyColors = {
        Easy: "bg-green-100 text-green-800 border border-green-300",
        Medium: "bg-yellow-100 text-yellow-800 border border-yellow-300",
        Hard: "bg-red-100 text-red-800 border border-red-300",
    };

    const getScoreColor = (score: number) => {
        if (score >= 4) return "text-green-600";
        if (score >= 3) return "text-yellow-600";
        return "text-red-600";
    };

    const getScoreBgColor = (score: number) => {
        if (score >= 4) return "bg-green-50 border-green-200";
        if (score >= 3) return "bg-yellow-50 border-yellow-200";
        return "bg-red-50 border-red-200";
    };

    // Format interval for display
    const formatInterval = (intervalMinutes: number, cardState: string) => {
        if (cardState === "learning" || cardState === "relearning") {
            if (intervalMinutes < 60) {
                return `${intervalMinutes} minute${intervalMinutes !== 1 ? "s" : ""}`;
            } else if (intervalMinutes < 1440) {
                const hours = Math.round(intervalMinutes / 60);
                return `${hours} hour${hours !== 1 ? "s" : ""}`;
            } else {
                const days = Math.round(intervalMinutes / 1440);
                return `${days} day${days !== 1 ? "s" : ""}`;
            }
        }
        return null; // Use date for review cards
    };



    // Render sub-score bar
    const renderSubScoreBar = (score: number, label: string) => {
        const percentage = (score / 5) * 100;
        const color = score >= 4 ? "bg-green-500" : score >= 3 ? "bg-yellow-500" : "bg-red-500";
        
        return (
            <div className="space-y-1">
                <div className="flex items-center justify-between text-sm">
                    <span className="text-gray-700 font-medium">{label}</span>
                    <span className={`font-bold ${getScoreColor(score)}`}>{score}/5</span>
                </div>
                <div className="h-2 w-full overflow-hidden rounded-full bg-gray-200">
                    <div
                        className={`h-full ${color} transition-all duration-500`}
                        style={{ width: `${percentage}%` }}
                    />
                </div>
            </div>
        );
    };

    // If showing result, render the feedback page
    if (showResult && result) {
        return (
            <div className="p-4 min-h-screen bg-gray-50">
                <SubmissionReport 
                    result={result} 
                    userAnswer={answer} 
                    question={card.question} 
                    onNext={handleNext}
                    nextLabel="Continue Learning"
                />
            </div>
        
        );
    }

    // Original question view
    return (
        <div className="min-h-screen bg-gray-50 p-4 md:p-8">
            <div className="mx-auto max-w-4xl">
                {/* Header */}
                <div className="mb-6 flex items-center justify-between">
                    <Button variant="ghost" onClick={() => router.push("/")}>
                        <ArrowLeft className="mr-2 h-4 w-4" />
                        Dashboard
                    </Button>
                    <div className="flex items-center gap-2">
                        <span className="text-sm text-gray-600">
                            Card State:
                        </span>
                        <span className="rounded-full bg-blue-100 px-3 py-1 text-xs font-medium text-blue-800 border border-blue-300">
                            {card.review.card_state.charAt(0).toUpperCase() +
                                card.review.card_state.slice(1)}
                        </span>
                    </div>
                </div>

                {/* Problem Card */}
                <AnimatePresence mode="wait">
                    <motion.div
                        key={card.question.id}
                        initial={{ opacity: 0, y: 20 }}
                        animate={{ opacity: 1, y: 0 }}
                        exit={{ opacity: 0, y: -20 }}
                    >
                        <Card className="mb-6 bg-white shadow-sm">
                            <CardHeader>
                                <div className="flex items-start justify-between">
                                    <div className="flex-1">
                                        <div className="mb-2 flex items-center gap-2">
                                            <span
                                                className={`rounded-full px-3 py-1 text-xs font-medium ${
                                                    difficultyColors[
                                                        card.question.difficulty
                                                    ]
                                                }`}
                                            >
                                                {card.question.difficulty}
                                            </span>
                                            {showTopics && card.question.topics.map(
                                                (topic) => (
                                                    <span
                                                        key={topic}
                                                        className="rounded-full bg-gray-100 px-3 py-1 text-xs font-medium text-gray-700 border border-gray-300"
                                                    >
                                                        {topic}
                                                    </span>
                                                )
                                            )}
                                        </div>
                                        <CardTitle className="text-2xl text-gray-900 flex items-center gap-2">
                                            <span>
                                                {card.question.leetcode_id}. {card.question.title}
                                            </span>
                                            <a 
                                                href={`https://leetcode.com/problems/${card.question.slug}/description/`}
                                                target="_blank"
                                                rel="noopener noreferrer"
                                                className="text-gray-400 hover:text-blue-600 transition-colors"
                                                title="View on LeetCode"
                                            >
                                                <ExternalLink className="h-5 w-5" />
                                            </a>
                                        </CardTitle>
                                    </div>
                                    <BookOpen className="h-6 w-6 text-gray-400" />
                                </div>
                            </CardHeader>
                            <CardContent>
                                <div className="text-gray-700 leading-relaxed">
                                    <ReactMarkdown 
                                        remarkPlugins={[remarkGfm]}
                                        components={{
                                            pre: ({ children }) => {
                                                // Check if this is an example block (contains Input/Output)
                                                const codeElement = Children.toArray(children).find(
                                                    (child) => isValidElement(child) && child.type === "code"
                                                ) as React.ReactElement<{ children: React.ReactNode }> | undefined;

                                                if (codeElement && codeElement.props.children) {
                                                    const text = String(codeElement.props.children);
                                                    // Check for LeetCode example pattern
                                                    if (text.includes("Input:") && text.includes("Output:")) {
                                                        // Parse the example text
                                                        const parts = text.split(/(Input:|Output:|Explanation:)/).filter(Boolean);
                                                        const items = [];
                                                        for (let i = 0; i < parts.length; i += 2) {
                                                            if (i + 1 < parts.length) {
                                                                items.push({
                                                                    label: parts[i].replace(":", ""),
                                                                    value: parts[i + 1].trim()
                                                                });
                                                            }
                                                        }

                                                        return (
                                                            <div className="my-4 rounded-lg bg-gray-50 p-4 border border-gray-200 text-sm">
                                                                {items.map((item, index) => (
                                                                    <div key={index} className="mb-1 last:mb-0">
                                                                        <span className="font-semibold text-gray-900">{item.label}:</span>{" "}
                                                                        <span className="text-gray-700 font-mono">{item.value}</span>
                                                                    </div>
                                                                ))}
                                                            </div>
                                                        );
                                                    }
                                                }

                                                return (
                                                    <div className="relative my-4">
                                                        <pre className="bg-gray-100 text-gray-800 p-4 rounded-lg overflow-x-auto border border-gray-300 whitespace-pre-wrap">
                                                            {Children.map(children, (child) =>
                                                                isValidElement(child)
                                                                    ? cloneElement(child as React.ReactElement, { isBlock: true } as any)
                                                                    : child
                                                            )}
                                                        </pre>
                                                    </div>
                                                );
                                            },
                                            code: ({ className, children, isBlock, ...props }: any) => {
                                                const match = /language-(\w+)/.exec(className || "");
                                                const isInline = !match && !className && !isBlock;

                                                if (isInline) {
                                                    return (
                                                        <code
                                                            className="bg-gray-100 text-gray-800 px-1 py-0.5 rounded text-[0.9em] font-mono border border-gray-200 whitespace-nowrap"
                                                            {...props}
                                                        >
                                                            {children}
                                                        </code>
                                                    );
                                                }

                                                return (
                                                    <code className={className} {...props}>
                                                        {children}
                                                    </code>
                                                );
                                            },
                                            p: ({children}) => <p className="mb-4 leading-7">{children}</p>,
                                            ul: ({children}) => <ul className="list-disc pl-6 mb-4 space-y-1">{children}</ul>,
                                            ol: ({children}) => <ol className="list-decimal pl-6 mb-4 space-y-1">{children}</ol>,
                                            li: ({children}) => <li className="pl-1">{children}</li>,
                                            h1: ({children}) => <h1 className="text-2xl font-bold mb-4 mt-6 text-gray-900">{children}</h1>,
                                            h2: ({children}) => <h2 className="text-xl font-semibold mb-3 mt-5 text-gray-900">{children}</h2>,
                                            h3: ({children}) => <h3 className="text-lg font-semibold mb-2 mt-4 text-gray-900">{children}</h3>,
                                            blockquote: ({children}) => <blockquote className="border-l-4 border-blue-400 pl-4 italic my-4 text-gray-600 bg-gray-50 py-2 pr-2 rounded-r">{children}</blockquote>,
                                            strong: ({children}) => <strong className="font-semibold text-gray-900">{children}</strong>,
                                            a: ({children, href}) => <a href={href} className="text-blue-600 hover:underline underline-offset-4" target="_blank" rel="noopener noreferrer">{children}</a>,
                                            table: ({children}) => <div className="overflow-x-auto my-4 rounded-lg border border-gray-300"><table className="w-full text-sm text-left bg-white">{children}</table></div>,
                                            thead: ({children}) => <thead className="bg-gray-100 text-gray-700 uppercase">{children}</thead>,
                                            tbody: ({children}) => <tbody className="divide-y divide-gray-200">{children}</tbody>,
                                            tr: ({children}) => <tr className="bg-white hover:bg-gray-50 transition-colors">{children}</tr>,
                                            th: ({children}) => <th className="px-4 py-3 font-medium">{children}</th>,
                                            td: ({children}) => <td className="px-4 py-3">{children}</td>,
                                            img: ({src, alt}) => <img src={src} alt={alt || ''} className="max-w-full h-auto rounded-lg my-4" />,
                                        }}
                                    >
                                        {card.question.description_markdown}
                                    </ReactMarkdown>
                                </div>
                            </CardContent>
                        </Card>

                        {/* Answer Input */}
                        <Card className="bg-white shadow-sm">
                            <CardHeader>
                                <CardTitle className="flex items-center gap-2 text-gray-900">
                                    <Lightbulb className="h-5 w-5 text-yellow-500" />
                                    Explain Your Approach
                                </CardTitle>
                                <CardDescription className="text-gray-600">
                                    Describe the algorithm and data structures
                                    you would use to solve this problem
                                </CardDescription>
                            </CardHeader>
                            <CardContent className="space-y-6">
                                <div className="relative">
                                    <Textarea
                                        placeholder="Example: I would use a hashmap to store complements. For each number, I check if target minus the current number exists in the map..."
                                        value={answer}
                                        onChange={(e) => setAnswer(e.target.value)}
                                        className="max-h-60 min-h-32 resize-none bg-white border-gray-300 pr-12"
                                        disabled={isTranscribing}
                                    />
                                    <Button
                                        type="button"
                                        variant={isRecording ? "destructive" : "outline"}
                                        size="icon"
                                        className={`absolute top-2 right-2 ${isRecording ? 'animate-pulse' : ''}`}
                                        onClick={isRecording ? stopRecording : startRecording}
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
                                        onClick={handleSkip}
                                        disabled={skipping || submitting || isTranscribing}
                                        className="flex-[0.4]"
                                    >
                                        {skipping ? (
                                            <>Processing...</>
                                        ) : (
                                            <>
                                                <SkipForward className="mr-2 h-4 w-4" />
                                                Skip Card (10 min)
                                            </>
                                        )}
                                    </Button>
                                    <Button
                                        onClick={handleSubmit}
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
                    </motion.div>
                </AnimatePresence>
            </div>
        </div>
    );
}