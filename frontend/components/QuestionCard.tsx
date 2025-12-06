"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { BookOpen, ExternalLink } from "lucide-react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { Children, cloneElement, isValidElement } from "react";
import type { Question } from "@/types";

interface QuestionCardProps {
    question: Question;
    cardState: string;
    showTopics: boolean;
}

export default function QuestionCard({ question, cardState, showTopics }: QuestionCardProps) {
    const difficultyColors = {
        Easy: "bg-green-100 text-green-800 border border-green-300",
        Medium: "bg-yellow-100 text-yellow-800 border border-yellow-300",
        Hard: "bg-red-100 text-red-800 border border-red-300",
    };

    return (
        <Card className="mb-6 bg-white shadow-sm">
            <CardHeader>
                <div className="flex items-start justify-between">
                    <div className="flex-1">
                        <div className="mb-2 flex items-center gap-2">
                            <span
                                className={`rounded-full px-3 py-1 text-xs font-medium ${
                                    difficultyColors[question.difficulty]
                                }`}
                            >
                                {question.difficulty}
                            </span>
                            {showTopics && question.topics.map((topic) => (
                                <span
                                    key={topic}
                                    className="rounded-full bg-gray-100 px-3 py-1 text-xs font-medium text-gray-700 border border-gray-300"
                                >
                                    {topic}
                                </span>
                            ))}
                        </div>
                        <CardTitle className="text-2xl text-gray-900 flex items-center gap-2">
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
                        </CardTitle>
                    </div>
                    <BookOpen className="h-6 w-6 text-gray-400" />
                </div>
            </CardHeader>
            <CardContent>
                <div className="text-gray-700 leading-relaxed">
                    <ReactMarkdown 
                        remarkPlugins={[remarkGfm]}
                        components={{
                            pre: ({ children }) => {
                                // Check if this is an example block (contains Input/Output)
                                const codeElement = Children.toArray(children).find(
                                    (child) => isValidElement(child) && child.type === "code"
                                ) as React.ReactElement<{ children: React.ReactNode }> | undefined;

                                if (codeElement && codeElement.props.children) {
                                    const text = String(codeElement.props.children);
                                    // Check for LeetCode example pattern
                                    if (text.includes("Input:") && text.includes("Output:")) {
                                        // Parse the example text
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
        </Card>
    );
}
