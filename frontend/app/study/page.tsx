"use client";

import { useEffect, useState } from "react";
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
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import {
    Brain,
    ArrowLeft,
    Send,
    Trophy,
    Target,
    BookOpen,
    Lightbulb,
} from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";

export default function StudyPage() {
    const router = useRouter();
    const [loading, setLoading] = useState(true);
    const [card, setCard] = useState<CardType | null>(null);
    const [answer, setAnswer] = useState("");
    const [submitting, setSubmitting] = useState(false);
    const [result, setResult] = useState<SubmitAnswerResponse | null>(null);
    const [showResult, setShowResult] = useState(false);

    useEffect(() => {
        checkAuth();
    }, []);

    const checkAuth = async () => {
        const { data } = await supabase.auth.getSession();
        if (!data.session) {
            router.push("/login");
            return;
        }
        loadCard();
    };

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

    const handleSubmit = async () => {
        if (!card || !answer.trim()) return;

        setSubmitting(true);
        try {
            const response = await api.submitAnswer(card.question.id, answer);
            setResult(response);
            setShowResult(true);
        } catch (err) {
            console.error("Failed to submit answer:", err);
            alert("Failed to submit answer. Please try again.");
        } finally {
            setSubmitting(false);
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
            <div className="flex min-h-screen items-center justify-center">
                <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary"></div>
            </div>
        );
    }

    if (!card) {
        return (
            <div className="flex min-h-screen items-center justify-center">
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
        Easy: "bg-green-500 text-white",
        Medium: "bg-yellow-500 text-white",
        Hard: "bg-red-500 text-white",
    };

    const getScoreColor = (score: number) => {
        if (score >= 4) return "text-green-600";
        if (score >= 3) return "text-yellow-600";
        return "text-red-600";
    };

    return (
        <div className="min-h-screen p-4 md:p-8">
            <div className="mx-auto max-w-4xl">
                {/* Header */}
                <div className="mb-6 flex items-center justify-between">
                    <Button variant="ghost" onClick={() => router.push("/")}>
                        <ArrowLeft className="mr-2 h-4 w-4" />
                        Dashboard
                    </Button>
                    <div className="flex items-center gap-2">
                        <span className="text-sm text-muted-foreground">
                            Card State:
                        </span>
                        <span className="rounded-full bg-primary/10 px-3 py-1 text-xs font-medium text-primary">
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
                        <Card className="mb-6">
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
                                            {card.question.topics.map(
                                                (topic) => (
                                                    <span
                                                        key={topic}
                                                        className="rounded-full bg-secondary px-3 py-1 text-xs font-medium"
                                                    >
                                                        {topic}
                                                    </span>
                                                )
                                            )}
                                        </div>
                                        <CardTitle className="text-2xl">
                                            {card.question.title}
                                        </CardTitle>
                                    </div>
                                    <BookOpen className="h-6 w-6 text-muted-foreground" />
                                </div>
                            </CardHeader>
                            <CardContent>
                                <div className="prose prose-sm dark:prose-invert max-w-none">
                                    <div className="whitespace-pre-wrap rounded-lg bg-secondary/50 p-4 text-sm">
                                        {card.question.description_markdown}
                                    </div>
                                </div>
                            </CardContent>
                        </Card>

                        {/* Answer Input */}
                        <Card>
                            <CardHeader>
                                <CardTitle className="flex items-center gap-2">
                                    <Lightbulb className="h-5 w-5" />
                                    Explain Your Approach
                                </CardTitle>
                                <CardDescription>
                                    Describe the algorithm and data structures
                                    you would use to solve this problem
                                </CardDescription>
                            </CardHeader>
                            <CardContent className="space-y-4">
                                <Textarea
                                    placeholder="Example: I would use a hashmap to store complements. For each number, I check if target minus the current number exists in the map..."
                                    value={answer}
                                    onChange={(e) => setAnswer(e.target.value)}
                                    rows={8}
                                    className="resize-none"
                                />
                                <Button
                                    onClick={handleSubmit}
                                    disabled={submitting || !answer.trim()}
                                    className="w-full"
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
                            </CardContent>
                        </Card>
                    </motion.div>
                </AnimatePresence>

                {/* Result Modal */}
                <Dialog open={showResult} onOpenChange={setShowResult}>
                    <DialogContent className="sm:max-w-xl">
                        <DialogHeader>
                            <DialogTitle className="flex items-center gap-2">
                                <Trophy className="h-6 w-6 text-yellow-500" />
                                Your Result
                            </DialogTitle>
                        </DialogHeader>
                        {result && (
                            <div className="space-y-4">
                                {/* Score */}
                                <div className="rounded-lg border-2 border-primary/20 bg-primary/5 p-6 text-center">
                                    <p className="mb-2 text-sm text-muted-foreground">
                                        Score
                                    </p>
                                    <p
                                        className={`text-6xl font-bold ${getScoreColor(
                                            result.score
                                        )}`}
                                    >
                                        {result.score}/5
                                    </p>
                                </div>

                                {/* Feedback */}
                                <div className="rounded-lg bg-secondary/50 p-4">
                                    <p className="mb-1 text-sm font-medium">
                                        Feedback
                                    </p>
                                    <p className="text-sm text-muted-foreground">
                                        {result.feedback}
                                    </p>
                                </div>

                                {/* Correct Approach */}
                                <div className="rounded-lg bg-secondary/50 p-4">
                                    <p className="mb-1 text-sm font-medium">
                                        Correct Approach
                                    </p>
                                    <p className="text-sm text-muted-foreground">
                                        {result.correct_approach}
                                    </p>
                                </div>

                                {/* Next Review */}
                                <div className="flex items-center justify-between rounded-lg border p-4">
                                    <div className="flex items-center gap-2">
                                        <Target className="h-5 w-5 text-muted-foreground" />
                                        <span className="text-sm">
                                            Next Review
                                        </span>
                                    </div>
                                    <span className="text-sm font-medium">
                                        {new Date(
                                            result.next_review_at
                                        ).toLocaleDateString()}
                                    </span>
                                </div>

                                <Button onClick={handleNext} className="w-full">
                                    <Brain className="mr-2 h-4 w-4" />
                                    Next Card
                                </Button>
                            </div>
                        )}
                    </DialogContent>
                </Dialog>
            </div>
        </div>
    );
}
