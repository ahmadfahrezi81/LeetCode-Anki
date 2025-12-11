"use client";

import { useState, Children, cloneElement, isValidElement } from "react";
import { Button } from "@/components/ui/button";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";
import {
    Brain,
    Trophy,
    Target,
    BookOpen,
    Lightbulb,
    CheckCircle2,
    XCircle,
    TrendingUp,
    Code,
    Zap,
    AlertTriangle,
    ExternalLink,
} from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import type { SubmitAnswerResponse, History, Card as CardType } from "@/types";
import { SolutionSkeleton } from "@/components/SolutionSkeleton";

interface SubmissionReportProps {
    result: SubmitAnswerResponse | History;
    userAnswer: string;
    question?: CardType["question"];
    onNext?: () => void;
    nextLabel?: string;
    onClose?: () => void;
    solutionLoading?: boolean;
    solutionProgress?: number;
}

export default function SubmissionReport({ 
    result, 
    userAnswer, 
    question, 
    onNext,
    nextLabel = "Next Card",
    onClose,
    solutionLoading = false,
    solutionProgress = 0,
}: SubmissionReportProps) {
    const [showQuestionInReport, setShowQuestionInReport] = useState(true);

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
        return null;
    };

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

    return (
        <div className="min-h-screen bg-gray-50 md:p-8">
            <div className="mx-auto max-w-6xl">
                {/* Header */}
                <div className="mb-6 flex items-center justify-between">
                    <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
                        <Trophy className="h-7 w-7 text-yellow-500" />
                        Your Learning Report
                    </h1>
                    {onClose && (
                        <Button 
                            variant="ghost" 
                            size="icon"
                            onClick={onClose}
                            className="w-12 h-12 text-gray-600 hover:text-gray-700"
                        >
                            <XCircle style={{ width: 30, height: 30 }} />
                        </Button>
                    )}
                </div>

                {/* Overall Score Card */}
                <Card className={`mb-6 border-2 ${getScoreBgColor(result.score)}`}>
                    <CardContent>
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm text-gray-600 mb-1">Overall Score</p>
                                <p className={`text-6xl font-bold ${getScoreColor(result.score)}`}>
                                    {result.score}/5
                                </p>
                                <p className="text-sm text-gray-600 mt-2">
                                    {result.score >= 4 ? "Excellent! 🎉" : result.score >= 3 ? "Good job! 👍" : "Keep practicing! 💪"}
                                </p>
                            </div>
                            
                            {/* Next Review Info */}
                            <div className="text-right">
                                <p className="text-sm text-gray-600 mb-1">Next Review</p>
                                <p className="text-xl font-semibold text-gray-900">
                                    {formatInterval(result.interval_minutes, result.card_state) ? (
                                        <>in {formatInterval(result.interval_minutes, result.card_state)}</>
                                    ) : (
                                        <>{new Date(result.next_review_at).toLocaleDateString()}</>
                                    )}
                                </p>
                                <p className="text-xs text-gray-500 mt-1">
                                    {result.card_state.charAt(0).toUpperCase() + result.card_state.slice(1)}
                                </p>
                            </div>
                        </div>
                    </CardContent>
                </Card>

                {/* Sub-Scores */}
                {result.sub_scores && (
                    <Card className="mb-6">
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                <Target className="h-5 w-5 text-blue-600" />
                                Detailed Breakdown
                            </CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            {renderSubScoreBar(result.sub_scores.pattern_recognition, "Pattern Recognition")}
                            {renderSubScoreBar(result.sub_scores.algorithmic_correctness, "Algorithmic Correctness")}
                            {renderSubScoreBar(result.sub_scores.complexity_understanding, "Complexity Understanding")}
                            {renderSubScoreBar(result.sub_scores.edge_case_awareness, "Edge Case Awareness")}
                        </CardContent>
                    </Card>
                )}

                {/* Two Column Layout: Your Answer vs Feedback */}
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
                    {/* Your Answer */}
                    <Card className="gap-2">
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2 text-lg">
                                <Lightbulb className="h-5 w-5 text-yellow-500" />
                                Your Answer
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="bg-gray-50 border border-gray-200 rounded-lg p-4">
                                <p className="text-gray-700 whitespace-pre-wrap leading-relaxed">
                                    {userAnswer}
                                </p>
                            </div>
                        </CardContent>
                    </Card>

                    {/* Detailed Feedback */}
                    <Card className="gap-2">
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2 text-lg">
                                <CheckCircle2 className="h-5 w-5 text-green-600" />
                                Detailed Feedback
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="prose prose-sm max-w-none">
                                <ReactMarkdown
                                    remarkPlugins={[remarkGfm]}
                                    components={{
                                        p: ({children}) => <p className="mb-3 text-gray-700 leading-relaxed">{children}</p>,
                                        strong: ({children}) => <strong className="font-semibold text-gray-900">{children}</strong>,
                                    }}
                                >
                                    {result.feedback}
                                </ReactMarkdown>
                            </div>
                        </CardContent>
                    </Card>
                </div>

                {/* Collapsible Question Card */}
                {question && (
                    <Card className="mb-6">
                        <CardHeader 
                            className="cursor-pointer !flex !items-center !h-auto !py-0"
                            onClick={() => setShowQuestionInReport(!showQuestionInReport)}
                        >
                            <div className="flex items-center justify-between gap-2">
                                <CardTitle className="flex items-center gap-2 text-lg">
                                    <BookOpen className="h-5 w-5 text-blue-600" />
                                    Original Question
                                </CardTitle>
                                <motion.div
                                    animate={{ rotate: showQuestionInReport ? 180 : 0 }}
                                    transition={{ duration: 0.2 }}
                                >
                                    <svg
                                        className="h-6 w-6 text-gray-500"
                                        fill="none"
                                        strokeLinecap="round"
                                        strokeLinejoin="round"
                                        strokeWidth="2"
                                        viewBox="0 0 24 24"
                                        stroke="currentColor"
                                    >
                                        <path d="M19 9l-7 7-7-7"></path>
                                    </svg>
                                </motion.div>
                            </div>
                        </CardHeader>
                        <AnimatePresence>
                            {showQuestionInReport && (
                                <motion.div
                                    initial={{ height: 0, opacity: 0 }}
                                    animate={{ height: "auto", opacity: 1 }}
                                    exit={{ height: 0, opacity: 0 }}
                                    transition={{ duration: 0.2 }}
                                >
                                    <CardContent className="pt-0">
                                        <div className="mb-3 flex items-center gap-2 flex-wrap">
                                            <span
                                                className={`rounded-full px-3 py-1 text-xs font-medium ${
                                                    difficultyColors[question.difficulty]
                                                }`}
                                            >
                                                {question.difficulty}
                                            </span>
                                            {question.topics.map((topic) => (
                                                <span
                                                    key={topic}
                                                    className="rounded-full bg-gray-100 px-3 py-1 text-xs font-medium text-gray-700 border border-gray-300"
                                                >
                                                    {topic}
                                                </span>
                                            ))}
                                        </div>
                                        <h3 className="text-xl font-semibold text-gray-900 mb-4 flex items-center gap-2">
                                            <span>
                                                {question.leetcode_id}. {question.title}
                                            </span>
                                            <a 
                                                href={`https://leetcode.com/problems/${question.slug}/description/`}
                                                target="_blank"
                                                rel="noopener noreferrer"
                                                className="text-gray-400 hover:text-blue-600 transition-colors"
                                                title="View on LeetCode"
                                            >
                                                <ExternalLink className="h-5 w-5" />
                                            </a>
                                        </h3>
                                        <div className="text-gray-700 leading-relaxed">
                                            <ReactMarkdown 
                                                remarkPlugins={[remarkGfm]}
                                                components={{
                                                    pre: ({ children }) => {
                                                        const codeElement = Children.toArray(children).find(
                                                            (child) => isValidElement(child) && child.type === "code"
                                                        ) as React.ReactElement<{ children: React.ReactNode }> | undefined;

                                                        if (codeElement && codeElement.props.children) {
                                                            const text = String(codeElement.props.children);
                                                            if (text.includes("Input:") && text.includes("Output:")) {
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
                                                {question.description_markdown}
                                            </ReactMarkdown>
                                        </div>
                                    </CardContent>
                                </motion.div>
                            )}
                        </AnimatePresence>
                    </Card>
                )}

                {/* Solution Breakdown */}
                {solutionLoading ? (
                    <SolutionSkeleton progress={solutionProgress} />
                ) : result.solution_breakdown ? (
                    <motion.div
                        initial={{ opacity: 0, y: 10 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ duration: 0.5, ease: "easeOut" }}
                    >
                        <Card className="mb-6">
                            <CardHeader>
                                <CardTitle className="flex items-center gap-2 text-lg">
                                    <Code className="h-5 w-5 text-purple-600" />
                                    Complete Solution Breakdown
                                </CardTitle>
                                <CardDescription>
                                    Learn the optimal approach step-by-step
                                </CardDescription>
                            </CardHeader>
                            <CardContent className="space-y-6">
                                {/* Pattern */}
                                <div>
                                    <div className="flex items-center gap-2 mb-2">
                                        <Zap className="h-4 w-4 text-blue-600" />
                                        <h3 className="font-semibold text-gray-900">Pattern</h3>
                                    </div>
                                    <p className="font-medium text-blue-600 bg-blue-50 px-3 py-2 rounded-lg inline-block text-xs border-1 border-blue-600">
                                        {result.solution_breakdown.pattern}
                                    </p>
                                </div>

                                {/* Why This Pattern */}
                                <div>
                                    <h3 className="font-semibold text-gray-900 mb-2">Why This Pattern?</h3>
                                    <p className="text-gray-700 leading-relaxed">
                                        {result.solution_breakdown.why_this_pattern}
                                    </p>
                                </div>

                                {/* Approach Steps */}
                                <div>
                                    <h3 className="font-semibold text-gray-900 mb-3">Algorithm Steps</h3>
                                    <ol className="space-y-2">
                                        {result.solution_breakdown.approach_steps.map((step, index) => (
                                            <li key={index} className="flex gap-3">
                                                <span className="flex-shrink-0 flex items-center justify-center w-6 h-6 rounded-full bg-blue-100 text-blue-700 text-sm font-semibold">
                                                    {index + 1}
                                                </span>
                                                <span className="text-gray-700 leading-relaxed pt-0.5">{step}</span>
                                            </li>
                                        ))}
                                    </ol>
                                </div>

                                {/* Pseudocode */}
                                <div>
                                    <h3 className="font-semibold text-gray-900 mb-3">Pseudocode</h3>
                                    <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg overflow-x-auto text-sm leading-relaxed">
                                        <code>{result.solution_breakdown.pseudocode}</code>
                                    </pre>
                                </div>

                                {/* Complexity Analysis */}
                                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                    <div className="bg-purple-50 border border-purple-200 rounded-lg p-4">
                                        <p className="text-sm text-purple-700 font-medium mb-1">Time Complexity</p>
                                        <p className="text-2xl font-bold text-purple-900">{result.solution_breakdown.time_complexity}</p>
                                    </div>
                                    <div className="bg-indigo-50 border border-indigo-200 rounded-lg p-4">
                                        <p className="text-sm text-indigo-700 font-medium mb-1">Space Complexity</p>
                                        <p className="text-2xl font-bold text-indigo-900">{result.solution_breakdown.space_complexity}</p>
                                    </div>
                                </div>
                                <div className="bg-gray-50 border border-gray-200 rounded-lg p-4">
                                    <p className="text-sm font-medium text-gray-900 mb-2">Complexity Explanation</p>
                                    <p className="text-gray-700 leading-relaxed">
                                        {result.solution_breakdown.complexity_explanation}
                                    </p>
                                </div>

                                {/* Key Insights */}
                                {result.solution_breakdown.key_insights.length > 0 && (
                                    <div>
                                        <div className="flex items-center gap-2 mb-3">
                                            <TrendingUp className="h-4 w-4 text-green-600" />
                                            <h3 className="font-semibold text-gray-900">Key Insights</h3>
                                        </div>
                                        <ul className="space-y-2">
                                            {result.solution_breakdown.key_insights.map((insight, index) => (
                                                <li key={index} className="flex gap-2 items-start">
                                                    <CheckCircle2 className="h-5 w-5 text-green-600 flex-shrink-0 mt-0.5" />
                                                    <span className="text-gray-700 leading-relaxed">{insight}</span>
                                                </li>
                                            ))}
                                        </ul>
                                    </div>
                                )}

                                {/* Common Pitfalls */}
                                {result.solution_breakdown.common_pitfalls.length > 0 && (
                                    <div>
                                        <div className="flex items-center gap-2 mb-3">
                                            <AlertTriangle className="h-4 w-4 text-orange-600" />
                                            <h3 className="font-semibold text-gray-900">Common Pitfalls</h3>
                                        </div>
                                        <ul className="space-y-2">
                                            {result.solution_breakdown.common_pitfalls.map((pitfall, index) => (
                                                <li key={index} className="flex gap-2 items-start">
                                                    <XCircle className="h-5 w-5 text-orange-600 flex-shrink-0 mt-0.5" />
                                                    <span className="text-gray-700 leading-relaxed">{pitfall}</span>
                                                </li>
                                            ))}
                                        </ul>
                                    </div>
                                )}
                            </CardContent>
                        </Card>
                    </motion.div>
                ) : null}

                {/* Next Card Button */}
                {onNext && (
                    <div className="flex justify-center">
                        <Button 
                            onClick={onNext} 
                            size="lg"
                            className="bg-blue-600 hover:bg-blue-700 text-white px-8"
                        >
                            <Brain className="mr-2 h-5 w-5" />
                            {nextLabel}
                        </Button>
                    </div>
                )}
            </div>
        </div>
    );
}
