export interface Question {
    id: string;
    leetcode_id: number;
    title: string;
    slug: string;
    difficulty: "Easy" | "Medium" | "Hard";
    description_markdown: string;
    topics: string[];
    correct_approach: string;
    created_at: string;
}

export interface Review {
    id: string;
    user_id: string;
    question_id: string;
    card_state: "new" | "learning" | "review" | "relearning";
    quality: number | null;
    easiness_factor: number;
    interval_days: number;
    repetitions: number;
    next_review_at: string;
    last_reviewed_at: string | null;
    total_reviews: number;
    total_lapses: number;
    created_at: string;
}

export interface Card {
    question: Question;
    review: Review;
}

export interface UserStats {
    user_id: string;
    total_cards: number;
    new_cards: number;
    learning_cards: number;
    review_cards: number;
    mature_cards: number;
    updated_at: string;
}

export interface DashboardData {
    stats: UserStats;
    due_reviews: number;
    available_new_card: boolean;
}

export interface SubmitAnswerResponse {
    score: number;
    feedback: string;
    correct_approach: string;
    next_review_at: string;
    card_state: string;
}
