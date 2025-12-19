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
    Zap,
    CheckCircle2,
    Timer,
    Flame,
    Settings,
    Menu,
    X,
    Coins,
    Landmark,
    Gem,
    Currency,
    Star,
    Minus,
    Plus,
} from "lucide-react";
import { Input } from "@/components/ui/input";
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogDescription,
    DialogFooter,
} from "@/components/ui/dialog";
import { motion } from "framer-motion";
import HistoryTable from "@/components/HistoryTable";
import { LeetCoin } from "@/components/LeetCoin";
import { StreakFlame } from "@/components/StreakFlame";
import { StreakDisplay } from "@/components/StreakDisplay";

export default function DashboardPage() {
    const router = useRouter();
    const [loading, setLoading] = useState(true);
    const [dashboard, setDashboard] = useState<DashboardData | null>(null);
    const [isSettingsOpen, setIsSettingsOpen] = useState(false);
    const [dailyLimit, setDailyLimit] = useState(5);
    const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
    const [isDesktopMenuOpen, setIsDesktopMenuOpen] = useState(false);

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
            if (data.stats.new_cards_limit) {
                setDailyLimit(data.stats.new_cards_limit);
            }
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
            <div className="flex min-h-screen items-center justify-center bg-gray-50">
                <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-600"></div>
            </div>
        );
    }

    if (!dashboard) {
        return (
            <div className="flex min-h-screen items-center justify-center bg-gray-50">
                <div className="text-center">
                    <p className="text-gray-500 mb-4">Failed to load dashboard</p>
                    <Button onClick={() => window.location.reload()}>Retry</Button>
                </div>
            </div>
        );
    }

    const totalDue = dashboard.due_counts.reviews_due + dashboard.due_counts.learning_due;
    const hasCardsAvailable = totalDue > 0 || dashboard.due_counts.new_available > 0;
    const isReviewMode = totalDue > 0;

    const stats = [
        {
            label: "Total",
            value: dashboard.stats.total_cards,
            icon: BookOpen,
            color: "text-blue-600",
            bg: "bg-blue-100",
        },
        {
            label: "New",
            value: dashboard.stats.new_cards,
            icon: Target,
            color: "text-green-600",
            bg: "bg-green-100",
        },
        {
            label: "Learning",
            value: dashboard.stats.learning_cards,
            icon: TrendingUp,
            color: "text-yellow-600",
            bg: "bg-yellow-100",
        },
        {
            label: "Review",
            value: dashboard.stats.review_cards,
            icon: Clock,
            color: "text-orange-600",
            bg: "bg-orange-100",
        },
        {
            label: "Mature",
            value: dashboard.stats.mature_cards,
            icon: Brain,
            color: "text-purple-600",
            bg: "bg-purple-100",
        },
    ];

    return (
        <div className="min-h-screen bg-gray-50 pb-20 md:pb-8">
            {/* Mobile Header */}
            <div className="bg-white border-b sticky top-0 z-50 md:hidden">
                <div className="flex items-center justify-between p-4">
                    <div className="flex items-center gap-2">
                        <div className="bg-blue-600 p-1.5 rounded-lg">
                            <Brain className="h-5 w-5 text-white" />
                        </div>
                        <span className="hidden sm:inline font-bold text-lg text-gray-900">LeetAnki</span>
                    </div>
                    <div className="flex items-center gap-2">
                        <StreakDisplay count={dashboard.stats.current_streak} active={dashboard.today_stats.reviews_done > 0} />

                        <div className="h-9 px-1 rounded-md flex items-center gap-1">
                            <LeetCoin />
                            <span className="text-lg font-extrabold text-orange-400">{dashboard.stats.coins}</span>
                        </div>
                        <div className="relative">
                            <Button
                                variant="ghost"
                                size="icon"
                                onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
                            >
                                {isMobileMenuOpen ? (
                                    <X className="h-5 w-5 text-gray-500" />
                                ) : (
                                    <Menu className="h-5 w-5 text-gray-500" />
                                )}
                            </Button>
                        </div>

                        {/* Mobile Dropdown Menu */}
                        {isMobileMenuOpen && (
                            <motion.div
                                initial={{ opacity: 0, scale: 0.95, y: -10 }}
                                animate={{ opacity: 1, scale: 1, y: 0 }}
                                exit={{ opacity: 0, scale: 0.95, y: -10 }}
                                className="absolute right-0 top-full mt-2 w-56 rounded-xl bg-white shadow-lg ring-1 ring-black/5 p-2 z-50"
                            >
                                <div className="space-y-1">
                                    <button
                                        onClick={() => {
                                            setIsMobileMenuOpen(false);
                                            setIsSettingsOpen(true);
                                        }}
                                        className="w-full text-left px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg flex items-center gap-2 font-medium"
                                    >
                                        <Settings className="h-4 w-4" />
                                        Set Daily Limit
                                    </button>
                                    <button
                                        onClick={handleLogout}
                                        className="w-full text-left px-3 py-2 text-sm text-red-600 hover:bg-red-50 rounded-lg flex items-center gap-2 font-medium"
                                    >
                                        <LogOut className="h-4 w-4" />
                                        Logout
                                    </button>
                                </div>
                            </motion.div>
                        )}
                    </div>
                </div>
            </div>

            <div className="max-w-4xl mx-auto p-4 space-y-6">
                {/* Desktop Header */}
                <div className="hidden md:flex items-center justify-between mb-8">
                    <div className="flex items-center gap-3">
                        <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-blue-600 shadow-lg shadow-blue-200">
                            <Brain className="h-6 w-6 text-white" />
                        </div>
                        <div>
                            <h1 className="text-2xl font-bold text-gray-900">LeetCode Anki</h1>
                            <p className="text-sm text-gray-500">Master algorithms efficiently</p>
                        </div>
                    </div>
                    <div className="flex items-center gap-4">
                        <StreakDisplay count={dashboard.stats.current_streak} active={dashboard.today_stats.reviews_done > 0} />
                        <div className="h-9 px-2 rounded-md flex items-center gap-1">
                            <LeetCoin />
                            <span className="text-lg font-extrabold text-orange-400">{dashboard.stats.coins}</span>
                        </div>
                        <div className="relative">
                            <Button
                                variant="outline"
                                size="icon"
                                onClick={() => setIsDesktopMenuOpen(!isDesktopMenuOpen)}
                                className="bg-white"
                            >
                                {isDesktopMenuOpen ? (
                                    <X className="h-4 w-4 text-gray-600" />
                                ) : (
                                    <Menu className="h-4 w-4 text-gray-600" />
                                )}
                            </Button>

                            {/* Desktop Dropdown Menu */}
                            {isDesktopMenuOpen && (
                                <motion.div
                                    initial={{ opacity: 0, scale: 0.95, y: -10 }}
                                    animate={{ opacity: 1, scale: 1, y: 0 }}
                                    exit={{ opacity: 0, scale: 0.95, y: -10 }}
                                    className="absolute right-0 top-full mt-2 w-56 rounded-xl bg-white shadow-lg ring-1 ring-black/5 p-2 z-50"
                                >
                                    <div className="space-y-1">
                                        <button
                                            onClick={() => {
                                                setIsDesktopMenuOpen(false);
                                                setIsSettingsOpen(true);
                                            }}
                                            className="w-full text-left px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg flex items-center gap-2 font-medium"
                                        >
                                            <Settings className="h-4 w-4" />
                                            Set Daily Limit
                                        </button>
                                        <button
                                            onClick={handleLogout}
                                            className="w-full text-left px-3 py-2 text-sm text-red-600 hover:bg-red-50 rounded-lg flex items-center gap-2 font-medium"
                                        >
                                            <LogOut className="h-4 w-4" />
                                            Logout
                                        </button>
                                    </div>
                                </motion.div>
                            )}
                        </div>
                    </div>
                </div>

                {/* Hero / Action Section */}
                <motion.div
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    className="grid gap-6 md:grid-cols-2"
                >
                    {/* Main Action Card */}
                    <Card className={`border-0 shadow-lg overflow-hidden relative ${
                        isReviewMode 
                            ? "bg-gradient-to-br from-blue-600 to-blue-700 text-white" 
                            : "bg-white text-gray-900"
                    }`}>
                        <div className="absolute top-0 right-0 p-32 bg-white/5 rounded-full -mr-16 -mt-16 blur-3xl" />
                        
                        <CardHeader className="relative z-10 pb-2">
                            <CardTitle className="flex items-center gap-2 text-lg font-medium opacity-90">
                                <Zap className={`h-5 w-5 ${isReviewMode ? "text-yellow-300" : "text-yellow-500"}`} />
                                {isReviewMode ? "Review Session" : "Study Session"}
                            </CardTitle>
                        </CardHeader>
                        
                        <CardContent className="relative z-10 space-y-6">
                            <div>
                                <div className="text-5xl font-bold tracking-tight mb-1">
                                    {totalDue > 0 ? totalDue : dashboard.due_counts.new_available}
                                </div>
                                <p className={`text-sm font-medium ${isReviewMode ? "text-blue-100" : "text-gray-500"}`}>
                                    {totalDue > 0 ? "Cards due for review" : "New cards available"}
                                </p>
                            </div>

                            {!isReviewMode && dashboard.next_card_due_at && (
                                <div className={`flex items-center gap-2 text-sm ${isReviewMode ? "bg-white/10" : "bg-gray-100"} p-3 rounded-lg w-fit`}>
                                    <Clock className="h-4 w-4" />
                                    <span>Next review: {new Date(dashboard.next_card_due_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</span>
                                </div>
                            )}

                            <Button 
                                size="lg" 
                                className={`w-full h-12 text-base font-semibold shadow-md transition-all hover:scale-[1.02] active:scale-[0.98] ${
                                    isReviewMode 
                                        ? "bg-white text-blue-600 hover:bg-blue-50" 
                                        : "bg-blue-600 text-white hover:bg-blue-700"
                                }`}
                                onClick={() => router.push("/study")}
                                disabled={!hasCardsAvailable}
                            >
                                <div className="flex items-center gap-2">
                                    <span>{isReviewMode ? "Start Review Session" : "Learn New Cards"}</span>
                                    <div className="flex items-center gap-1 h-9">
                                        <span className="text-lg font-bold text-orange-400">+</span>
                                        <LeetCoin size="sm" />
                                    </div>
                                </div>
                            </Button>
                        </CardContent>
                    </Card>

                    {/* Today's Progress - Compact Grid */}
                    <div className="grid grid-cols-2 gap-4">
                        <Card className="border-0 shadow-sm bg-white">
                            <CardContent className="px-6 flex flex-col justify-between h-full">
                                <div className="bg-green-100 p-2 w-fit rounded-lg mb-4">
                                    <CheckCircle2 className="h-5 w-5 text-green-600" />
                                </div>
                                <div>
                                    <div className="text-2xl font-bold text-gray-900">{dashboard.today_stats.reviews_done}</div>
                                    <div className="text-sm text-gray-500">Reviews Done</div>
                                </div>
                            </CardContent>
                        </Card>
                        <Card className="border-0 shadow-sm bg-white">
                            <CardContent className="px-6 flex flex-col justify-between h-full">
                                <div className="bg-orange-100 p-2 w-fit rounded-lg mb-4">
                                    <Flame className="h-5 w-5 text-orange-600" />
                                </div>
                                <div>
                                    <div className="text-2xl font-bold text-gray-900">{dashboard.today_stats.new_cards_done}</div>
                                    <div className="text-sm text-gray-500">New Learned</div>
                                </div>
                            </CardContent>
                        </Card>
                        <Card className="col-span-2 border-0 shadow-sm bg-white">
                            <CardContent className="px-6 flex items-center justify-between">
                                <div>
                                    <div className="text-sm text-gray-500 mb-1">Time Spent Today</div>
                                    <div className="text-2xl font-bold text-gray-900">
                                        {dashboard.today_stats.time_spent_minutes} <span className="text-sm font-normal text-gray-400">min</span>
                                    </div>
                                </div>
                                <div className="bg-purple-100 p-3 rounded-full">
                                    <Timer className="h-6 w-6 text-purple-600" />
                                </div>
                            </CardContent>
                        </Card>
                    </div>
                </motion.div>

                {/* Stats Overview - Scrollable on mobile */}
                <div>
                    <h2 className="text-lg font-semibold text-gray-900 mb-4 px-1">Overview</h2>
                    <div className="flex overflow-x-auto pb-4 gap-4 -mx-4 px-4 md:grid md:grid-cols-5 md:overflow-visible md:pb-0 md:px-0 md:mx-0 snap-x">
                        {stats.map((stat, index) => (
                            <motion.div
                                key={stat.label}
                                initial={{ opacity: 0, y: 20 }}
                                animate={{ opacity: 1, y: 0 }}
                                transition={{ delay: index * 0.05 }}
                                className="snap-center shrink-0 w-32 md:w-auto"
                            >
                                <div className="bg-white p-5 rounded-xl shadow-sm border border-gray-100 flex flex-col items-center text-center h-full justify-center space-y-2">
                                    <div className={`${stat.bg} p-2 rounded-full`}>
                                        <stat.icon className={`h-4 w-4 ${stat.color}`} />
                                    </div>
                                    <div>
                                        <div className="text-xl font-bold text-gray-900">{stat.value}</div>
                                        <div className="text-xs font-medium text-gray-500">{stat.label}</div>
                                    </div>
                                </div>
                            </motion.div>
                        ))}
                    </div>
                </div>

                {/* History Section */}
                <div>
                    <HistoryTable />
                </div>
            </div>

            {/* Settings Dialog */}
            <Dialog open={isSettingsOpen} onOpenChange={setIsSettingsOpen}>
                <DialogContent className="md:w-[400px]">
                    <DialogHeader className="text-left">
                        <DialogTitle>Study Settings</DialogTitle>
                        <DialogDescription>
                            Customize your daily learning goals.
                        </DialogDescription>
                    </DialogHeader>
                    <div className="flex flex-col gap-4">
                            <div className="flex flex-col gap-2">
                                <label className="text-sm font-medium text-gray-700">
                                    New cards per day
                                </label>
                                <div className="flex items-center gap-3">
                                    <Button
                                        variant="outline"
                                        size="icon"
                                        onClick={() => setDailyLimit(Math.max(0, dailyLimit - 1))}
                                        disabled={dailyLimit <= 0}
                                    >
                                        <Minus className="h-4 w-4" />
                                    </Button>
                                    <div className="flex-1 text-center">
                                        <span className="text-2xl font-bold">{dailyLimit}</span>
                                    </div>
                                    <Button
                                        variant="outline"
                                        size="icon"
                                        onClick={() => setDailyLimit(Math.min(100, dailyLimit + 1))}
                                        disabled={dailyLimit >= 100}
                                    >
                                        <Plus className="h-4 w-4" />
                                    </Button>
                                </div>
                                <p className="text-xs text-gray-500 text-center">
                                    Maximum new cards to show each day
                                </p>
                            </div>
                        </div>
                    <DialogFooter>
                        <Button
                            className="bg-gradient-to-br from-blue-600 to-blue-700"
                            onClick={async () => {
                                try {
                                    await api.updateDailyLimit(dailyLimit);
                                    setIsSettingsOpen(false);
                                    // Refresh dashboard to reflect changes (though local state is already updated)
                                    loadDashboard();
                                } catch (error) {
                                    console.error("Failed to update limit:", error);
                                    alert("Failed to update limit. Please try again.");
                                }
                            }}
                        >
                            Save Changes
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
}
