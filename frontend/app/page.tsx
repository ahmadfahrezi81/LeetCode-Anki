"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { supabase } from "@/lib/supabase";
import { api } from "@/lib/api";
import type { DashboardData } from "@/types";
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
    BookOpen,
    Target,
    TrendingUp,
    Clock,
    LogOut,
} from "lucide-react";
import { motion } from "framer-motion";

export default function DashboardPage() {
    const router = useRouter();
    const [loading, setLoading] = useState(true);
    const [dashboard, setDashboard] = useState<DashboardData | null>(null);

    useEffect(() => {
        checkAuth();
    }, []);

    const checkAuth = async () => {
        const { data } = await supabase.auth.getSession();
        if (!data.session) {
            router.push("/login");
            return;
        }
        loadDashboard();
    };

    const loadDashboard = async () => {
        try {
            const data = await api.getDashboard();
            setDashboard(data);
        } catch (err) {
            console.error("Failed to load dashboard:", err);
        } finally {
            setLoading(false);
        }
    };

    const handleLogout = async () => {
        await supabase.auth.signOut();
        router.push("/login");
    };

    if (loading) {
        return (
            <div className="flex min-h-screen items-center justify-center">
                <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary"></div>
            </div>
        );
    }

    if (!dashboard) {
        return (
            <div className="flex min-h-screen items-center justify-center">
                <p className="text-muted-foreground">
                    Failed to load dashboard
                </p>
            </div>
        );
    }

    const stats = [
        {
            label: "Total Cards",
            value: dashboard.stats.total_cards,
            icon: BookOpen,
            color: "bg-blue-500",
        },
        {
            label: "New",
            value: dashboard.stats.new_cards,
            icon: Target,
            color: "bg-green-500",
        },
        {
            label: "Learning",
            value: dashboard.stats.learning_cards,
            icon: TrendingUp,
            color: "bg-yellow-500",
        },
        {
            label: "Review",
            value: dashboard.stats.review_cards,
            icon: Clock,
            color: "bg-orange-500",
        },
        {
            label: "Mature",
            value: dashboard.stats.mature_cards,
            icon: Brain,
            color: "bg-purple-500",
        },
    ];

    return (
        <div className="min-h-screen p-4 md:p-8">
            <div className="mx-auto max-w-6xl">
                {/* Header */}
                <div className="mb-8 flex items-center justify-between">
                    <motion.div
                        initial={{ opacity: 0, y: -20 }}
                        animate={{ opacity: 1, y: 0 }}
                        className="flex items-center gap-3"
                    >
                        <div className="flex h-12 w-12 items-center justify-center rounded-full bg-primary">
                            <Brain className="h-6 w-6 text-primary-foreground" />
                        </div>
                        <div>
                            <h1 className="text-3xl font-bold tracking-tight">
                                LeetCode Anki
                            </h1>
                            <p className="text-sm text-muted-foreground">
                                Master algorithms with spaced repetition
                            </p>
                        </div>
                    </motion.div>
                    <Button variant="ghost" onClick={handleLogout}>
                        <LogOut className="mr-2 h-4 w-4" />
                        Logout
                    </Button>
                </div>

                {/* Stats Grid */}
                <div className="mb-8 grid gap-4 md:grid-cols-5">
                    {stats.map((stat, index) => (
                        <motion.div
                            key={stat.label}
                            initial={{ opacity: 0, y: 20 }}
                            animate={{ opacity: 1, y: 0 }}
                            transition={{ delay: index * 0.1 }}
                        >
                            <Card>
                                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                                    <CardTitle className="text-sm font-medium">
                                        {stat.label}
                                    </CardTitle>
                                    <div
                                        className={`${stat.color} rounded-full p-2`}
                                    >
                                        <stat.icon className="h-4 w-4 text-white" />
                                    </div>
                                </CardHeader>
                                <CardContent>
                                    <div className="text-2xl font-bold">
                                        {stat.value}
                                    </div>
                                </CardContent>
                            </Card>
                        </motion.div>
                    ))}
                </div>

                {/* Study Card */}
                <motion.div
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ delay: 0.5 }}
                >
                    <Card className="border-2">
                        <CardHeader>
                            <CardTitle className="text-2xl">
                                Ready to Study?
                            </CardTitle>
                            <CardDescription>
                                {dashboard.due_reviews > 0
                                    ? `You have ${dashboard.due_reviews} card${
                                          dashboard.due_reviews === 1 ? "" : "s"
                                      } due for review`
                                    : dashboard.available_new_card
                                    ? "Learn a new algorithm problem"
                                    : "No cards available right now. Check back later!"}
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <Button
                                size="lg"
                                className="w-full md:w-auto"
                                onClick={() => router.push("/study")}
                                disabled={
                                    dashboard.due_reviews === 0 &&
                                    !dashboard.available_new_card
                                }
                            >
                                <Brain className="mr-2 h-5 w-5" />
                                {dashboard.due_reviews > 0
                                    ? "Start Review"
                                    : "Learn New Card"}
                            </Button>
                        </CardContent>
                    </Card>
                </motion.div>
            </div>
        </div>
    );
}
