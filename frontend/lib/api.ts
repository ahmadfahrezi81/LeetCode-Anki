import { supabase } from "./supabase";

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
    getDashboard: () => apiRequest("/api/dashboard"),

    // Study Session
    getNextCard: () => apiRequest("/api/card/next"),

    submitAnswer: (questionId: string, answer: string) =>
        apiRequest("/api/review/submit", {
            method: "POST",
            body: JSON.stringify({ question_id: questionId, answer }),
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
};

export type { DashboardData, Card, SubmitAnswerResponse } from "@/types";
