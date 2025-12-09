"use client";

import { useEffect, useState, useRef } from "react";
import { api } from "@/lib/api";
import type { History, Question } from "@/types";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
    Drawer,
    DrawerContent,
    DrawerHeader,
    DrawerTitle,
    DrawerDescription,
} from "@/components/ui/drawer";
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogDescription,
} from "@/components/ui/dialog";
import SubmissionReport from "@/components/SubmissionReport";
import { 
    ChevronLeft, 
    ChevronRight, 
    History as HistoryIcon, 
    Star, 
    Calendar,
    Eye,
    Filter,
    Loader2
} from "lucide-react";

interface Filters {
    difficulty: ("Easy" | "Medium" | "Hard")[];
    minScore: number | null;
    maxScore: number | null;
    state: string[];
}

export default function HistoryTable() {
    const [history, setHistory] = useState<History[]>([]);
    const [loading, setLoading] = useState(true);
    const [loadingMore, setLoadingMore] = useState(false);
    const [page, setPage] = useState(0);
    const limit = 10;
    const [hasMore, setHasMore] = useState(true);

    const [selectedItem, setSelectedItem] = useState<History | null>(null);
    const [selectedQuestion, setSelectedQuestion] = useState<Question | null>(null);
    const [loadingQuestion, setLoadingQuestion] = useState(false);

    const [showFilters, setShowFilters] = useState(false);
    const [filters, setFilters] = useState<Filters>({
        difficulty: [],
        minScore: null,
        maxScore: null,
        state: []
    });
    const [appliedFilters, setAppliedFilters] = useState<Filters>({
        difficulty: [],
        minScore: null,
        maxScore: null,
        state: []
    });

    const observerTarget = useRef<HTMLDivElement>(null);

    // Load history when page or applied filters change
    useEffect(() => {
        loadHistory(page === 0);
    }, [page, appliedFilters]);

    const loadHistory = async (reset: boolean) => {
        if (reset) {
            setLoading(true);
        } else {
            setLoadingMore(true);
        }
        
        try {
            const apiFilters: {
                difficulties?: string[];
                minScore?: number;
                maxScore?: number;
                states?: string[];
            } = {};

            if (appliedFilters.difficulty.length > 0) {
                apiFilters.difficulties = appliedFilters.difficulty;
            }
            if (appliedFilters.minScore !== null) {
                apiFilters.minScore = appliedFilters.minScore;
            }
            if (appliedFilters.maxScore !== null) {
                apiFilters.maxScore = appliedFilters.maxScore;
            }
            if (appliedFilters.state.length > 0) {
                apiFilters.states = appliedFilters.state;
            }

            const res = await api.getHistory(
                limit, 
                page * limit,
                Object.keys(apiFilters).length > 0 ? apiFilters : undefined
            );
            const data = res.data || [];
            
            if (reset) {
                setHistory(data);
            } else {
                setHistory(prev => [...prev, ...data]);
            }
            
            setHasMore(data.length === limit);
        } catch (err) {
            console.error("Failed to load history:", err);
        } finally {
            setLoading(false);
            setLoadingMore(false);
        }
    };

    const loadMore = () => {
        if (!loadingMore && hasMore && !loading) {
            setPage(p => p + 1);
        }
    };

    // Infinite scroll observer
    useEffect(() => {
        const currentTarget = observerTarget.current;
        if (!currentTarget) return;

        const observer = new IntersectionObserver(
            (entries) => {
                const entry = entries[0];
                if (entry.isIntersecting && hasMore && !loadingMore && !loading) {
                    loadMore();
                }
            },
            { 
                threshold: 0.1,
                rootMargin: '100px'
            }
        );

        observer.observe(currentTarget);

        return () => {
            observer.disconnect();
        };
    }, [hasMore, loadingMore, loading]);

    const formatDate = (dateString: string) => {
        return new Date(dateString).toLocaleDateString("en-US", {
            month: "short",
            day: "numeric",
            hour: "2-digit",
            minute: "2-digit"
        });
    };

    const fetchQuestion = async (questionId: string) => {
        setLoadingQuestion(true);
        try {
            const res = await api.getQuestionById(questionId);
            setSelectedQuestion(res.question);
        } catch (err) {
            console.error("Failed to load question:", err);
            setSelectedQuestion(null);
        } finally {
            setLoadingQuestion(false);
        }
    };

    const handleViewDetails = (item: History) => {
        setSelectedItem(item);
        fetchQuestion(item.question_id);
    };

    const handleCloseDrawer = (open: boolean) => {
        if (!open) {
            setSelectedItem(null);
            setSelectedQuestion(null);
        }
    };

    const getScoreColor = (score: number) => {
        if (score >= 4) return "text-green-600 bg-green-50 border-green-200";
        if (score >= 3) return "text-yellow-600 bg-yellow-50 border-yellow-200";
        return "text-red-600 bg-red-50 border-red-200";
    };

    const getDifficultyColor = (difficulty?: string) => {
        if (difficulty === 'Easy') return 'bg-green-50 text-green-700 border-green-200';
        if (difficulty === 'Medium') return 'bg-yellow-50 text-yellow-700 border-yellow-200';
        if (difficulty === 'Hard') return 'bg-red-50 text-red-700 border-red-200';
        return 'bg-gray-100 text-gray-700 border-gray-200';
    };

    const getStateColor = (state: string) => {
        if (state === 'learning' || state === 'relearning') return 'bg-blue-50 text-blue-700 border-blue-200';
        if (state === 'review') return 'bg-green-50 text-green-700 border-green-200';
        return 'bg-gray-100 text-gray-700 border-gray-200';
    };

    const toggleFilter = <K extends keyof Filters>(
        key: K,
        value: Filters[K] extends Array<infer U> ? U : never
    ) => {
        setFilters(prev => {
            const current = prev[key];
            if (Array.isArray(current)) {
                const newArray = current.includes(value as any)
                    ? current.filter(v => v !== value)
                    : [...current, value];
                return { ...prev, [key]: newArray };
            }
            return prev;
        });
    };

    const applyFilters = () => {
        setAppliedFilters(filters);
        setPage(0);
        setHistory([]);
        setShowFilters(false);
    };

    const clearFilters = () => {
        const emptyFilters = {
            difficulty: [],
            minScore: null,
            maxScore: null,
            state: []
        };
        setFilters(emptyFilters);
        setAppliedFilters(emptyFilters);
        setPage(0);
        setHistory([]);
    };

    const activeFilterCount = 
        appliedFilters.difficulty.length + 
        appliedFilters.state.length + 
        (appliedFilters.minScore !== null || appliedFilters.maxScore !== null ? 1 : 0);

    return (
        <>
            <Card className="border-0 shadow-sm bg-white py-4 gap-4">
                <CardHeader className="flex flex-row items-center justify-between">
                    <CardTitle className="text-lg font-semibold text-gray-900 flex items-center gap-2">
                        <HistoryIcon className="h-5 w-5 text-gray-500" />
                        Recent History
                    </CardTitle>
                    <div className="flex items-center gap-2">
                        <Button
                            variant="outline"
                            size="icon"
                            onClick={() => setShowFilters(true)}
                            className="h-8 w-8 relative"
                        >
                            <Filter className="h-4 w-4" />
                            {activeFilterCount > 0 && (
                                <span className="absolute -top-1 -right-1 h-4 w-4 rounded-full bg-blue-600 text-white text-[10px] flex items-center justify-center font-medium">
                                    {activeFilterCount}
                                </span>
                            )}
                        </Button>
                        
                        <div className="hidden md:flex items-center gap-2">
                            <Button
                                variant="outline"
                                size="icon"
                                onClick={() => {
                                    setPage(0);
                                    setHistory([]);
                                }}
                                disabled={page === 0 || loading}
                                className="h-8 w-8"
                            >
                                <ChevronLeft className="h-4 w-4" />
                            </Button>
                            <span className="text-sm text-gray-600 min-w-[60px] text-center">
                                Page {page + 1}
                            </span>
                            <Button
                                variant="outline"
                                size="icon"
                                onClick={() => setPage(p => p + 1)}
                                disabled={!hasMore || loading}
                                className="h-8 w-8"
                            >
                                <ChevronRight className="h-4 w-4" />
                            </Button>
                        </div>
                    </div>
                </CardHeader>
                <CardContent className="px-4">
                    {/* Desktop Table View */}
                    <div className="hidden md:block relative overflow-x-auto">
                        <table className="w-full text-sm text-left">
                            <thead className="text-xs text-gray-500 uppercase bg-gray-50/50">
                                <tr>
                                    <th className="px-4 py-3 font-medium">Question</th>
                                    <th className="px-4 py-3 font-medium">Difficulty</th>
                                    <th className="px-4 py-3 font-medium">Score</th>
                                    <th className="px-4 py-3 font-medium">State</th>
                                    <th className="px-4 py-3 font-medium">Date</th>
                                    <th className="px-4 py-3 font-medium">Next Review</th>
                                    <th className="px-4 py-3 font-medium text-right">Actions</th>
                                </tr>
                            </thead>
                            <tbody>
                                {loading ? (
                                    Array.from({ length: 5 }).map((_, i) => (
                                        <tr key={i} className="border-b border-gray-50">
                                            <td className="px-4 py-4"><div className="h-4 bg-gray-100 rounded w-48 animate-pulse" /></td>
                                            <td className="px-4 py-4"><div className="h-6 bg-gray-100 rounded w-16 animate-pulse" /></td>
                                            <td className="px-4 py-4"><div className="h-6 bg-gray-100 rounded w-12 animate-pulse" /></td>
                                            <td className="px-4 py-4"><div className="h-4 bg-gray-100 rounded w-20 animate-pulse" /></td>
                                            <td className="px-4 py-4"><div className="h-4 bg-gray-100 rounded w-24 animate-pulse" /></td>
                                            <td className="px-4 py-4"><div className="h-4 bg-gray-100 rounded w-24 animate-pulse" /></td>
                                            <td className="px-4 py-4"><div className="h-8 bg-gray-100 rounded w-8 animate-pulse ml-auto" /></td>
                                        </tr>
                                    ))
                                ) : history.length === 0 ? (
                                    <tr>
                                        <td colSpan={7} className="px-4 py-8 text-center text-gray-500">
                                            {activeFilterCount > 0 ? "No history matches your filters" : "No history found"}
                                        </td>
                                    </tr>
                                ) : (
                                    history.map((item) => (
                                        <tr key={item.id} className="border-b border-gray-50 last:border-0 hover:bg-gray-50/50 transition-colors">
                                            <td className="px-4 py-4 font-medium text-gray-900">
                                                {item.question_leetcode_id ? `${item.question_leetcode_id}. ` : ''}{item.question_title || 'Unknown Question'}
                                            </td>
                                            <td className="px-4 py-4">
                                                <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border ${getDifficultyColor(item.question_difficulty)}`}>
                                                    {item.question_difficulty || 'N/A'}
                                                </span>
                                            </td>
                                            <td className="px-4 py-4">
                                                <div className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium border ${getScoreColor(item.score)}`}>
                                                    <Star className="h-3 w-3 fill-current" />
                                                    {item.score}/5
                                                </div>
                                            </td>
                                            <td className="px-4 py-4">
                                                <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium capitalize border ${getStateColor(item.card_state)}`}>
                                                    {item.card_state}
                                                </span>
                                            </td>
                                            <td className="px-4 py-4 text-gray-500 text-xs">
                                                {formatDate(item.submitted_at)}
                                            </td>
                                            <td className="px-4 py-4 text-gray-500 text-xs">
                                                {formatDate(item.next_review_at)}
                                            </td>
                                            <td className="px-4 py-4 text-right">
                                                <Button
                                                    variant="ghost"
                                                    size="sm"
                                                    onClick={() => handleViewDetails(item)}
                                                    className="h-8 w-8 p-0"
                                                >
                                                    <Eye className="h-4 w-4 text-gray-500" />
                                                </Button>
                                            </td>
                                        </tr>
                                    ))
                                )}
                            </tbody>
                        </table>
                    </div>

                    {/* Mobile Card View */}
                    <div className="md:hidden space-y-4">
                        {loading ? (
                            Array.from({ length: 3 }).map((_, i) => (
                                <div key={i} className="bg-white border border-gray-100 rounded-xl p-4 animate-pulse">
                                    <div className="h-5 bg-gray-100 rounded w-3/4 mb-3" />
                                    <div className="flex gap-2 mb-3">
                                        <div className="h-6 bg-gray-100 rounded w-16" />
                                        <div className="h-6 bg-gray-100 rounded w-12" />
                                    </div>
                                    <div className="h-4 bg-gray-100 rounded w-1/2" />
                                </div>
                            ))
                        ) : history.length === 0 ? (
                            <div className="text-center py-12 text-gray-500">
                                {activeFilterCount > 0 ? "No history matches your filters" : "No history found"}
                            </div>
                        ) : (
                            <>
                                {history.map((item) => (
                                    <div
                                        key={item.id}
                                        className="bg-white border border-gray-100 rounded-xl p-4 shadow-sm hover:shadow-md transition-all duration-200 active:scale-[0.98]"
                                        onClick={() => handleViewDetails(item)}
                                    >
                                        <h3 className="font-semibold text-gray-900 mb-2 line-clamp-2">
                                            {item.question_leetcode_id ? `${item.question_leetcode_id}. ` : ''}
                                            {item.question_title || 'Unknown Question'}
                                        </h3>

                                        <div className="flex flex-wrap gap-2 mb-3">
                                            <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border ${getDifficultyColor(item.question_difficulty)}`}>
                                                {item.question_difficulty || 'N/A'}
                                            </span>
                                            <div className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium border ${getScoreColor(item.score)}`}>
                                                <Star className="h-3 w-3 fill-current" />
                                                {item.score}/5
                                            </div>
                                            <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium capitalize border ${getStateColor(item.card_state)}`}>
                                                {item.card_state}
                                            </span>
                                        </div>

                                        <div className="flex items-center justify-between text-xs text-gray-500">
                                            <div className="flex items-center gap-1">
                                                <Calendar className="h-3 w-3" />
                                                <span>{formatDate(item.submitted_at)}</span>
                                            </div>
                                            <div className="flex items-center gap-1">
                                                <span className="text-gray-400">Next:</span>
                                                <span>{formatDate(item.next_review_at)}</span>
                                            </div>
                                        </div>
                                    </div>
                                ))}

                                <div ref={observerTarget} className="py-4">
                                    {loadingMore && (
                                        <div className="flex items-center justify-center gap-2 text-gray-500">
                                            <Loader2 className="h-4 w-4 animate-spin" />
                                            <span className="text-sm">Loading more...</span>
                                        </div>
                                    )}
                                </div>

                                {hasMore && !loadingMore && !loading && (
                                    <Button
                                        onClick={loadMore}
                                        variant="outline"
                                        className="w-full"
                                    >
                                        Load More
                                    </Button>
                                )}

                                {!hasMore && history.length > 0 && (
                                    <div className="text-center py-4 text-sm text-gray-400">
                                        No more history to load
                                    </div>
                                )}
                            </>
                        )}
                    </div>
                </CardContent>
            </Card>

            {/* Filter Dialog */}
            <Dialog open={showFilters} onOpenChange={setShowFilters}>
                <DialogContent className="max-w-sm">
                    <DialogHeader>
                        <DialogTitle className="flex items-center justify-between">
                            <span>Filter History</span>
                            {activeFilterCount > 0 && (
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    onClick={clearFilters}
                                    className="text-blue-600 hover:text-blue-700"
                                >
                                    Clear All
                                </Button>
                            )}
                        </DialogTitle>
                        <DialogDescription className="text-left">
                            Filter your submission history by difficulty, score, and state
                        </DialogDescription>
                    </DialogHeader>

                    <div className="space-y-6 py-4">
                        <div>
                            <label className="text-sm font-medium text-gray-700 mb-2 block">
                                Difficulty
                            </label>
                            <div className="flex flex-wrap gap-2">
                                {(['Easy', 'Medium', 'Hard'] as const).map(diff => (
                                    <button
                                        key={diff}
                                        onClick={() => toggleFilter('difficulty', diff)}
                                        className={`px-3 py-1.5 rounded-lg text-sm font-medium border transition-all ${
                                            filters.difficulty.includes(diff)
                                                ? `${getDifficultyColor(diff)} ring-2 ring-offset-1 ${
                                                    diff === 'Easy' ? 'ring-green-400' :
                                                    diff === 'Medium' ? 'ring-yellow-400' :
                                                    'ring-red-400'
                                                }`
                                                : 'bg-gray-50 text-gray-600 border-gray-200 hover:bg-gray-100'
                                        }`}
                                    >
                                        {diff}
                                    </button>
                                ))}
                            </div>
                        </div>

                        <div>
                            <label className="text-sm font-medium text-gray-700 mb-2 block">
                                Score Range
                            </label>
                            <div className="flex flex-wrap gap-2">
                                {[1, 2, 3, 4, 5].map(score => (
                                    <button
                                        key={score}
                                        onClick={() => {
                                            if (filters.minScore === score && filters.maxScore === score) {
                                                setFilters(prev => ({ ...prev, minScore: null, maxScore: null }));
                                            } else {
                                                setFilters(prev => ({ ...prev, minScore: score, maxScore: score }));
                                            }
                                        }}
                                        className={`px-3 py-1.5 rounded-lg text-sm font-medium border transition-all flex items-center gap-1 ${
                                            filters.minScore === score && filters.maxScore === score
                                                ? `${getScoreColor(score)} ring-2 ring-offset-1 ${
                                                    score >= 4 ? 'ring-green-400' :
                                                    score >= 3 ? 'ring-yellow-400' :
                                                    'ring-red-400'
                                                }`
                                                : 'bg-gray-50 text-gray-600 border-gray-200 hover:bg-gray-100'
                                        }`}
                                    >
                                        <Star className="h-3 w-3 fill-current" />
                                        {score}
                                    </button>
                                ))}
                            </div>
                        </div>

                        <div>
                            <label className="text-sm font-medium text-gray-700 mb-2 block">
                                Card State
                            </label>
                            <div className="flex flex-wrap gap-2">
                                {['new', 'learning', 'review', 'relearning'].map(state => (
                                    <button
                                        key={state}
                                        onClick={() => toggleFilter('state', state)}
                                        className={`px-3 py-1.5 rounded-lg text-sm font-medium border capitalize transition-all ${
                                            filters.state.includes(state)
                                                ? `${getStateColor(state)} ring-2 ring-offset-1 ${
                                                    state === 'review' ? 'ring-green-400' :
                                                    state === 'learning' || state === 'relearning' ? 'ring-blue-400' :
                                                    'ring-gray-400'
                                                }`
                                                : 'bg-gray-50 text-gray-600 border-gray-200 hover:bg-gray-100'
                                        }`}
                                    >
                                        {state}
                                    </button>
                                ))}
                            </div>
                        </div>
                    </div>

                    <div className="flex gap-2">
                        <Button
                            variant="outline"
                            onClick={() => setShowFilters(false)}
                            className="flex-1"
                        >
                            Cancel
                        </Button>
                        <Button
                            onClick={applyFilters}
                            className="flex-1 bg-blue-600 hover:bg-blue-700"
                        >
                            Apply Filters
                        </Button>
                    </div>
                </DialogContent>
            </Dialog>

            {/* Submission Details Drawer */}
            <Drawer open={!!selectedItem} onOpenChange={handleCloseDrawer}>
                <DrawerContent
                    className="
                        h-[87vh] md:h-[95vh] mt-0 rounded-t-[10px] bg-gray-50
                        [&>*:first-child]:bg-gray-500
                    "
                >
                    <DrawerHeader className="text-left px-4 pt-6">
                        <DrawerTitle>Submission Details</DrawerTitle>
                        <DrawerDescription>
                            Review your answer, score, and detailed feedback.
                        </DrawerDescription>
                    </DrawerHeader>
                    <div className="flex-1 overflow-y-auto px-4 pb-10 select-text">
                        {loadingQuestion ? (
                            <div className="flex items-center justify-center py-12">
                                <div className="text-center">
                                    <div className="h-8 w-8 border-4 border-gray-200 border-t-gray-600 rounded-full animate-spin mx-auto mb-4" />
                                    <p className="text-gray-500">Loading question details...</p>
                                </div>
                            </div>
                        ) : selectedItem && selectedQuestion ? (
                            <SubmissionReport 
                                result={selectedItem} 
                                userAnswer={selectedItem.user_answer}
                                question={selectedQuestion}
                            />
                        ) : selectedItem ? (
                            <div className="flex items-center justify-center py-12">
                                <p className="text-red-500">Failed to load question details</p>
                            </div>
                        ) : null}
                    </div>
                </DrawerContent>
            </Drawer>
        </>
    );
}