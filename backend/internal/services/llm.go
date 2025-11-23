package services

import (
	"context"
	"fmt"
	"leetcode-anki/backend/config"
	"strconv"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type LLMService struct {
	client *openai.Client
}

func NewLLMService() *LLMService {
	client := openai.NewClient(config.AppConfig.OpenAIKey)
	return &LLMService{client: client}
}

// ScoreAnswer uses GPT to score the user's explanation
func (l *LLMService) ScoreAnswer(ctx context.Context, questionTitle, questionDescription, correctApproach, userAnswer string) (int, string, error) {
	prompt := l.buildScoringPrompt(questionTitle, questionDescription, correctApproach, userAnswer)

	resp, err := l.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini, // Using GPT-4o-mini for cost efficiency
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are an expert algorithm tutor evaluating a student's understanding of problem-solving approaches.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.3, // Lower temperature for more consistent scoring
			MaxTokens:   200,
		},
	)

	if err != nil {
		return 0, "", fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return 0, "", fmt.Errorf("no response from OpenAI")
	}

	response := resp.Choices[0].Message.Content
	score, feedback := l.parseResponse(response)

	return score, feedback, nil
}

func (l *LLMService) buildScoringPrompt(questionTitle, questionDescription, correctApproach, userAnswer string) string {
	return fmt.Sprintf(`You are evaluating a student's understanding of algorithm problem-solving logic.

**Problem:** %s

**Problem Description:**
%s

**Correct Approach:**
%s

**Student's Explanation:**
%s

---

**Your Task:**
Score the student's explanation from 0-5 based on how well they understand the correct approach:

- **5**: Perfect understanding. The student correctly identifies the key algorithmic pattern/data structure and explains the core logic accurately.
- **4**: Strong understanding. Minor details missing but core approach is correct.
- **3**: Acceptable understanding. Gets the main idea but has some conceptual gaps or inefficiencies.
- **2**: Weak understanding. Identifies some relevant concepts but approach is flawed.
- **1**: Poor understanding. Approach is incorrect but shows some problem-solving attempt.
- **0**: No understanding. Answer is completely wrong or irrelevant.

**Important:** Focus on the *algorithmic approach* and *logic*, NOT code syntax. The student should explain the strategy (e.g., "use a hashmap to store complements", "two pointers from both ends", "BFS traversal").

**Output Format (strictly follow this):**
SCORE: <number 0-5>
FEEDBACK: <one concise sentence explaining the score>

Example:
SCORE: 4
FEEDBACK: Correct identification of the two-pointer approach, but didn't mention the array must be sorted first.`,
		questionTitle,
		questionDescription,
		correctApproach,
		userAnswer,
	)
}

func (l *LLMService) parseResponse(response string) (int, string) {
	lines := strings.Split(strings.TrimSpace(response), "\n")

	score := 0
	feedback := "Unable to parse feedback from AI."

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "SCORE:") {
			scoreStr := strings.TrimSpace(strings.TrimPrefix(line, "SCORE:"))
			if s, err := strconv.Atoi(scoreStr); err == nil {
				score = s
			}
		}

		if strings.HasPrefix(line, "FEEDBACK:") {
			feedback = strings.TrimSpace(strings.TrimPrefix(line, "FEEDBACK:"))
		}
	}

	// Clamp score to valid range
	if score < 0 {
		score = 0
	}
	if score > 5 {
		score = 5
	}

	return score, feedback
}
