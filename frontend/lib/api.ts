import { supabase } from "./supabase";
import type { DashboardData, Card, SubmitAnswerResponse, History, Question } from "@/types";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";


async function getAuthToken(): Promise<string | null> {
    const { data } = await supabase.auth.getSession();
    return data.session?.access_token || null;
}

async function apiRequest<T>(
    endpoint: string,
    options: RequestInit = {}
): Promise<T> {
    const token = await getAuthToken();

    if (!token) {
        throw new Error("Not authenticated");
    }

    const response = await fetch(`${API_URL}${endpoint}`, {
        ...options,
        headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${token}`,
            ...options.headers,
        },
    });

    if (!response.ok) {
        const error = await response
            .json()
            .catch(() => ({ error: "Request failed" }));
        throw new Error(error.error || `HTTP ${response.status}`);
    }

    return response.json();
}

export const api = {
    // Dashboard
    getDashboard: (): Promise<DashboardData> =>
        apiRequest<DashboardData>("/api/dashboard"),

    // Study Session
    getNextCard: (): Promise<{ card: Card | null; message?: string }> =>
        apiRequest<{ card: Card | null; message?: string }>("/api/card/next"),

    submitAnswer: (
        questionId: string,
        answer: string,
        timeSpentSeconds: number = 0
    ): Promise<SubmitAnswerResponse> =>
        apiRequest<SubmitAnswerResponse>("/api/review/submit", {
            method: "POST",
            body: JSON.stringify({ 
                question_id: questionId, 
                answer,
                time_spent_seconds: timeSpentSeconds 
            }),
        }),

    skipCard: (questionId: string): Promise<{
        message: string;
        next_review_at: string;
        card_state: string;
        interval_minutes: number;
    }> =>
        apiRequest<{
            message: string;
            next_review_at: string;
            card_state: string;
            interval_minutes: number;
        }>("/api/review/skip", {
            method: "POST",
            body: JSON.stringify({ question_id: questionId }),
        }),

    // Admin
    refreshProblems: (easy = 5, medium = 10, hard = 5) =>
        apiRequest(
            `/api/admin/refresh-problems?easy=${easy}&medium=${medium}&hard=${hard}`,
            {
                method: "POST",
            }
        ),

    getProblemStats: () => apiRequest("/api/admin/problem-stats"),

    // History
    getHistory: (
        limit = 20, 
        offset = 0,
        filters?: {
            difficulties?: string[];
            minScore?: number;
            maxScore?: number;
            states?: string[];
        }
    ): Promise<{ data: History[]; limit: number; offset: number }> => {
        let url = `/api/history?limit=${limit}&offset=${offset}`;
        
        if (filters) {
            if (filters.difficulties && filters.difficulties.length > 0) {
                url += `&difficulty=${filters.difficulties.join(',')}`;
            }
            if (filters.minScore !== undefined) {
                url += `&minScore=${filters.minScore}`;
            }
            if (filters.maxScore !== undefined) {
                url += `&maxScore=${filters.maxScore}`;
            }
            if (filters.states && filters.states.length > 0) {
                url += `&state=${filters.states.join(',')}`;
            }
        }
        
        return apiRequest<{ data: History[]; limit: number; offset: number }>(url);
    },

    // Questions
    getQuestionById: (questionId: string): Promise<{ question: Question }> =>
        apiRequest<{ question: Question }>(`/api/questions/${questionId}`),
};
