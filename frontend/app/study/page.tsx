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
        try {
            const response = await api.submitAnswer(card.question.id, answer, timeSpentSeconds);
            setResult(response);
            
            // Determine feedback type and show overlay
            setFeedbackType(response.score >= 4 ? 'success' : 'failure');
            setShowFeedbackOverlay(true);
            
            // Auto-dismiss overlay after 3 seconds
            setTimeout(() => {
                setShowFeedbackOverlay(false);
                setTimeout(() => {
                    setShowResult(true);
                }, 300);
            }, 3000);
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
            <div className="p-4 min-h-screen bg-gray-50">
                <SubmissionReport 
                    result={result} 
                    userAnswer={answer} 
                    question={card.question} 
                    onNext={handleNext}
                    nextLabel="Continue Learning"
                    onClose={() => router.push("/")}
                />
            </div>
        );
    }

    // Original question view
    return (
        <>
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