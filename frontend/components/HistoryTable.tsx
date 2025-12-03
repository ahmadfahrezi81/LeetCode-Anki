"use client";

import { useEffect, useState } from "react";
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
import SubmissionReport from "@/components/SubmissionReport";
import { 
    ChevronLeft, 
    ChevronRight, 
    History as HistoryIcon, 
    Star, 
    Calendar,
    Eye
} from "lucide-react";

export default function HistoryTable() {
    const [history, setHistory] = useState<History[]>([]);
    const [loading, setLoading] = useState(true);
    const [page, setPage] = useState(0);
    const limit = 10;
    const [hasMore, setHasMore] = useState(true);

    const [selectedItem, setSelectedItem] = useState<History | null>(null);
    const [selectedQuestion, setSelectedQuestion] = useState<Question | null>(null);
    const [loadingQuestion, setLoadingQuestion] = useState(false);

    useEffect(() => {
        loadHistory();
    }, [page]);

    const loadHistory = async () => {
        setLoading(true);
        try {
            const res = await api.getHistory(limit, page * limit);
            // Ensure we have an array
            const data = res.data || [];
            setHistory(data);
            setHasMore(data.length === limit);
        } catch (err) {
            console.error("Failed to load history:", err);
        } finally {
            setLoading(false);
        }
    };

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

    return (
        <>
            <Card className="border-0 shadow-sm bg-white">
                <CardHeader className="flex flex-row items-center justify-between pb-2">
                    <CardTitle className="text-lg font-semibold text-gray-900 flex items-center gap-2">
                        <HistoryIcon className="h-5 w-5 text-gray-500" />
                        Recent History
                    </CardTitle>
                    <div className="flex items-center gap-2">
                        <Button
                            variant="outline"
                            size="icon"
                            onClick={() => setPage(p => Math.max(0, p - 1))}
                            disabled={page === 0 || loading}
                            className="h-8 w-8"
                        >
                            <ChevronLeft className="h-4 w-4" />
                        </Button>
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
                </CardHeader>
                <CardContent>
                    <div className="relative overflow-x-auto">
                        <table className="w-full text-sm text-left">
                            <thead className="text-xs text-gray-500 uppercase bg-gray-50/50">
                                <tr>
                                    <th className="px-4 py-3 font-medium">Date</th>
                                    <th className="px-4 py-3 font-medium">Question</th>
                                    <th className="px-4 py-3 font-medium">Score</th>
                                    <th className="px-4 py-3 font-medium">State</th>
                                    <th className="px-4 py-3 font-medium">Next Review</th>
                                    <th className="px-4 py-3 font-medium text-right">Actions</th>
                                </tr>
                            </thead>
                            <tbody>
                                {loading ? (
                                    Array.from({ length: 5 }).map((_, i) => (
                                        <tr key={i} className="border-b border-gray-50">
                                            <td className="px-4 py-4"><div className="h-4 bg-gray-100 rounded w-24 animate-pulse" /></td>
                                            <td className="px-4 py-4"><div className="h-4 bg-gray-100 rounded w-48 animate-pulse" /></td>
                                            <td className="px-4 py-4"><div className="h-6 bg-gray-100 rounded w-12 animate-pulse" /></td>
                                            <td className="px-4 py-4"><div className="h-4 bg-gray-100 rounded w-20 animate-pulse" /></td>
                                            <td className="px-4 py-4"><div className="h-4 bg-gray-100 rounded w-24 animate-pulse" /></td>
                                            <td className="px-4 py-4"><div className="h-8 bg-gray-100 rounded w-8 animate-pulse ml-auto" /></td>
                                        </tr>
                                    ))
                                ) : history.length === 0 ? (
                                    <tr>
                                        <td colSpan={6} className="px-4 py-8 text-center text-gray-500">
                                            No history found
                                        </td>
                                    </tr>
                                ) : (
                                    history.map((item) => (
                                        <tr key={item.id} className="border-b border-gray-50 last:border-0 hover:bg-gray-50/50 transition-colors">
                                            <td className="px-4 py-4 whitespace-nowrap text-gray-500">
                                                <div className="flex items-center gap-2">
                                                    <Calendar className="h-3 w-3" />
                                                    {formatDate(item.submitted_at)}
                                                </div>
                                            </td>
                                            <td className="px-4 py-4 font-medium text-gray-900">
                                                {item.question_leetcode_id ? `${item.question_leetcode_id}. ` : ''}{item.question_title || 'Unknown Question'}
                                            </td>
                                            <td className="px-4 py-4">
                                                <div className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium border ${getScoreColor(item.score)}`}>
                                                    <Star className="h-3 w-3 fill-current" />
                                                    {item.score}/5
                                                </div>
                                            </td>
                                            <td className="px-4 py-4">
                                                <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium capitalize
                                                    ${item.card_state === 'learning' || item.card_state === 'relearning' ? 'bg-blue-50 text-blue-700 border border-blue-200' : 
                                                      item.card_state === 'review' ? 'bg-green-50 text-green-700 border border-green-200' : 
                                                      'bg-gray-100 text-gray-700 border border-gray-200'}`}>
                                                    {item.card_state}
                                                </span>
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
                </CardContent>
            </Card>

            <Drawer open={!!selectedItem} onOpenChange={handleCloseDrawer} >
                <DrawerContent
                    className="
                        h-[95vh] mt-0 rounded-t-[10px] bg-gray-50
                        [&>*:first-child]:bg-gray-500  
                    "
                >
                    <DrawerHeader className="text-left px-6 pt-6">
                        <DrawerTitle>Submission Details</DrawerTitle>
                        <DrawerDescription>
                            Review your answer, score, and detailed feedback.
                        </DrawerDescription>
                    </DrawerHeader>
                    <div className="flex-1 overflow-y-auto px-6 pb-10">
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
