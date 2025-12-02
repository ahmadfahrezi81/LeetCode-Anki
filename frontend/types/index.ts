// Types shared between the backend (Go) and the frontend (TS)

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

// Review – mirrors models.Review
export interface Review {
  id: string;
  user_id: string;
  question_id: string;
  card_state: "new" | "learning" | "review" | "relearning";
  quality: number | null;
  easiness_factor: number;
  interval_days: number;
  interval_minutes: number; // precise interval
  current_step: number; // learning step (0‑3)
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

// User‑level aggregates
export interface UserStats {
  user_id: string;
  total_cards: number;
  new_cards: number;
  learning_cards: number;
  review_cards: number;
  mature_cards: number;
  updated_at: string;
}

// Due counts – Anki‑style
export interface DueCounts {
  learning_due: number;
  reviews_due: number;
  new_available: number;
  new_studied_today: number;
}

// Today's session stats
export interface TodayStats {
  reviews_done: number;
  new_cards_done: number;
  time_spent_minutes: number;
}

// Full dashboard payload returned by GET /api/dashboard
export interface DashboardData {
  stats: UserStats;
  due_counts: DueCounts;
  today_stats: TodayStats;
  next_card_due_at: string | null; // ISO timestamp (or null)
  all_cards_studied: boolean;
}

// New “progress summary” types (admin view)
export interface DifficultyProgress {
  total: number;
  mastered: number;
  learning: number;
  unseen: number;
}

export interface TopicProgress {
  total: number;
  mastered: number;
  learning: number;
  unseen: number;
}

export interface ProgressSummary {
  by_difficulty: Record<string, DifficultyProgress>;
  by_topic: Record<string, TopicProgress>;
}

// Question‑level statistics (admin view)
export interface QuestionStats {
  total_attempts: number;
  avg_easiness: number;
  total_failures: number;
  avg_reviews: number;
  mastered_count: number;
  difficulty_score?: number;
}

// Existing answer‑related types
export interface SubmitAnswerResponse {
  score: number;
  feedback: string;
  correct_approach: string;
  next_review_at: string;
  card_state: string;
}
