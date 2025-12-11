"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { supabase } from "@/lib/supabase";
import { api } from "@/lib/api";
import type { Card as CardType, SubmitAnswerResponse } from "@/types";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Brain, ArrowLeft } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import SubmissionReport from "@/components/SubmissionReport";
import FeedbackOverlay from "@/components/FeedbackOverlay";
import QuestionCard from "@/components/QuestionCard";
import AnswerInput from "@/components/AnswerInput";
import { useVoiceRecorder } from "@/hooks/useVoiceRecorder";
import { Toast } from "@/components/ui/toast";
import { useSolutionLoading } from "@/hooks/useSolutionLoading";

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
    const [cardLoadTime, setCardLoadTime] = useState<number>(Date.now());
    const [showFeedbackOverlay, setShowFeedbackOverlay] = useState(false);
    const [feedbackType, setFeedbackType] = useState<'success' | 'failure' | null>(null);
    
    // Toast state
    const [toastMessage, setToastMessage] = useState("");
    const [toastType, setToastType] = useState<"success" | "error" | "info" | "loading">("info");
    const [showToast, setShowToast] = useState(false);
    
    // Solution loading state
    const { isLoading: solutionLoading, progress: solutionProgress, startLoading, completeLoading, resetLoading } = useSolutionLoading();

    // Voice recording hook
    const { isRecording, isTranscribing, startRecording, stopRecording } = useVoiceRecorder(
        (transcribedText) => {
            setAnswer(prev => prev ? `${prev}\n\n${transcribedText}` : transcribedText);
        }
    );

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
                setCardLoadTime(Date.now());
            } else {
                router.push("/");
            }
        } catch (err) {
            console.error("Failed to load card:", err);
        } finally {
            setLoading(false);
        }
    };

    const handleSubmit = async () => {
        if (!card || !answer.trim()) return;

        const timeSpentSeconds = Math.floor((Date.now() - cardLoadTime) / 1000);

        setSubmitting(true);
        setToastMessage("Analyzing your answer...");
        setToastType("loading");
        setShowToast(true);
        
        try {
            const response = await api.submitAnswer(card.question.id, answer, timeSpentSeconds);
            
            // Store the initial response (without solution breakdown or with cached one)
            setResult(response);
            
            // Update toast to success
            setToastMessage("Analysis complete!");
            setToastType("success");
            setTimeout(() => setShowToast(false), 2000);
            
            // Determine feedback type and show overlay
            setFeedbackType(response.score >= 4 ? 'success' : 'failure');
            setShowFeedbackOverlay(true);
            
            // Auto-dismiss overlay after 3 seconds
            setTimeout(() => {
                setShowFeedbackOverlay(false);
                setTimeout(() => {
                    setShowResult(true);
                    
                    // If solution breakdown is not in the response, fetch it separately
                    if (!response.solution_breakdown) {
                        fetchSolutionBreakdown(card.question.id);
                    }
                }, 300);
            }, 3000);
        } catch (err) {
            console.error("Failed to submit answer:", err);
            
            setToastMessage("Failed to submit answer. Please try again.");
            setToastType("error");
            setTimeout(() => setShowToast(false), 3000);
            
            alert("Failed to submit answer. Please try again.");
        } finally {
            setSubmitting(false);
        }
    };

    const fetchSolutionBreakdown = async (questionId: string) => {
        // Start loading animation
        startLoading();
        setToastMessage("Loading solution breakdown...");
        setToastType("loading");
        setShowToast(true);
        
        try {
            const { solution_breakdown, cached } = await api.getSolutionBreakdown(questionId);
            
            // Update the result with the solution breakdown
            setResult(prev => prev ? {
                ...prev,
                solution_breakdown
            } : null);
            
            // Complete loading
            completeLoading();
            
            // Update toast
            setToastMessage(cached ? "Solution loaded!" : "Solution generated!");
            setToastType("success");
            setTimeout(() => setShowToast(false), 2000);
        } catch (err) {
            console.error("Failed to fetch solution breakdown:", err);
            resetLoading();
            setToastMessage("Failed to load solution breakdown");
            setToastType("error");
            setTimeout(() => setShowToast(false), 3000);
        }
    };

    const handleSkip = async () => {
        if (!card) return;

        setSkipping(true);
        try {
            await api.skipCard(card.question.id);
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
        resetLoading();
        loadCard();
    };

    const handleFeedbackDismiss = () => {
        setShowFeedbackOverlay(false);
        setTimeout(() => setShowResult(true), 300);
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

    // If showing result, render the feedback page
    if (showResult && result) {
        return (
            <>
                <Toast 
                    show={showToast} 
                    message={toastMessage} 
                    type={toastType}
                    onClose={() => setShowToast(false)}
                />
                <div className="p-4 min-h-screen bg-gray-50">
                    <SubmissionReport 
                        result={result} 
                        userAnswer={answer} 
                        question={card.question} 
                        onNext={handleNext}
                        nextLabel="Continue Learning"
                        onClose={() => router.push("/")}
                        solutionLoading={solutionLoading}
                        solutionProgress={solutionProgress}
                    />
                </div>
            </>
        );
    }

    // Original question view
    return (
        <>
            <Toast 
                show={showToast} 
                message={toastMessage} 
                type={toastType}
                onClose={() => setShowToast(false)}
            />
            {result && feedbackType && (
                <FeedbackOverlay 
                    show={showFeedbackOverlay}
                    score={result.score}
                    isSuccess={feedbackType === 'success'}
                    onDismiss={handleFeedbackDismiss}
                />
            )}
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
                            <QuestionCard 
                                question={card.question}
                                cardState={card.review.card_state}
                                showTopics={showTopics}
                            />

                            {/* Answer Input */}
                            <AnswerInput 
                                answer={answer}
                                onAnswerChange={setAnswer}
                                onSubmit={handleSubmit}
                                onSkip={handleSkip}
                                submitting={submitting}
                                skipping={skipping}
                                isRecording={isRecording}
                                isTranscribing={isTranscribing}
                                onStartRecording={startRecording}
                                onStopRecording={stopRecording}
                            />
                        </motion.div>
                    </AnimatePresence>
                </div>
            </div>
        </>
    );
}