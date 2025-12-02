import { supabase } from "./supabase";
import type { DashboardData, Card, SubmitAnswerResponse, History } from "@/types";

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
        answer: string
    ): Promise<SubmitAnswerResponse> =>
        apiRequest<SubmitAnswerResponse>("/api/review/submit", {
            method: "POST",
            body: JSON.stringify({ question_id: questionId, answer }),
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
    getHistory: (limit = 20, offset = 0): Promise<{ data: History[]; limit: number; offset: number }> =>
        apiRequest<{ data: History[]; limit: number; offset: number }>(`/api/history?limit=${limit}&offset=${offset}`),
};
