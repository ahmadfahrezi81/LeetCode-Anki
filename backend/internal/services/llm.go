package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"leetcode-anki/backend/config"
	"leetcode-anki/backend/internal/models"
	"log"

	openai "github.com/sashabaranov/go-openai"
)

type LLMService struct {
	client *openai.Client
}

func NewLLMService() *LLMService {
	client := openai.NewClient(config.AppConfig.OpenAIKey)
	return &LLMService{client: client}
}

// LLMResponse matches the JSON structure from the LLM
type LLMResponse struct {
	Score     int           `json:"score"`
	SubScores SubScoresJSON `json:"sub_scores"`
	Feedback  string        `json:"feedback"`
	Solution  SolutionJSON  `json:"solution"`
}

type SubScoresJSON struct {
	PatternRecognition      int `json:"pattern_recognition"`
	AlgorithmicCorrectness  int `json:"algorithmic_correctness"`
	ComplexityUnderstanding int `json:"complexity_understanding"`
	EdgeCaseAwareness       int `json:"edge_case_awareness"`
}

type SolutionJSON struct {
	Pattern               string   `json:"pattern"`
	WhyThisPattern        string   `json:"why_this_pattern"`
	ApproachSteps         []string `json:"approach_steps"`
	Pseudocode            string   `json:"pseudocode"`
	TimeComplexity        string   `json:"time_complexity"`
	SpaceComplexity       string   `json:"space_complexity"`
	ComplexityExplanation string   `json:"complexity_explanation"`
	KeyInsights           []string `json:"key_insights"`
	CommonPitfalls        []string `json:"common_pitfalls"`
	CorrectApproach       string   `json:"correct_approach"`
}

// ScoreAnswer uses GPT to score the user's explanation with comprehensive feedback
func (l *LLMService) ScoreAnswer(ctx context.Context, questionTitle, questionDescription, userAnswer string) (int, string, string, *models.SubScores, *models.SolutionBreakdown, error) {
	prompt := l.buildScoringPrompt(questionTitle, questionDescription, userAnswer)

	resp, err := l.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are an expert algorithm tutor. You provide structured feedback in JSON format to help students master problem-solving patterns.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.1,
			MaxTokens:   2500,
		},
	)

	if err != nil {
		return 0, "", "", nil, nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return 0, "", "", nil, nil, fmt.Errorf("no response from OpenAI")
	}

	response := resp.Choices[0].Message.Content

	// Log the raw response for debugging
	log.Printf("🤖 Raw LLM Response:\n%s\n", response)

	// Parse the JSON response
	score, feedback, correctApproach, subScores, solutionBreakdown, err := l.parseJSONResponse(response)

	log.Printf("📊 Score: %d", score)
	log.Printf("📝 Feedback: %s", feedback)
	log.Printf("🎯 Correct Approach: %s", correctApproach)
	log.Printf("📈 SubScores: %+v", subScores)
	log.Printf("💡 Solution Breakdown: %+v", solutionBreakdown)

	if err != nil {
		log.Printf("❌ Failed to parse LLM response: %v", err)
		return 0, "", "", nil, nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	return score, feedback, correctApproach, subScores, solutionBreakdown, nil
}

func (l *LLMService) buildScoringPrompt(questionTitle, questionDescription, userAnswer string) string {
	return fmt.Sprintf(`You are evaluating a student's understanding of algorithm problem-solving.

**Problem:** %s

**Problem Description:**
%s

**Student's Explanation:**
%s

---

**Your Task:**
Evaluate the student's understanding and provide comprehensive, pedagogical feedback.

**Evaluation Criteria:**

1. **Overall Score (0-5):**
   - 5: Perfect understanding
   - 4: Strong understanding, minor gaps
   - 3: Acceptable, some conceptual gaps
   - 2: Weak, flawed approach
   - 1: Poor, incorrect approach
   - 0: No understanding

2. **Sub-Scores (each 0-5):**
   - Pattern Recognition
   - Algorithmic Correctness
   - Complexity Understanding
   - Edge Case Awareness

3. **Feedback:** 2-4 paragraphs covering what they got right, what they missed, and how to improve.

4. **Solution Breakdown:** Complete step-by-step explanation with pattern, approach, pseudocode, complexity, insights, and common pitfalls.

**CRITICAL: You must respond with ONLY valid JSON. No markdown, no backticks, no preamble. Just pure JSON.**

**Output Format:**
{
  "score": <0-5>,
  "sub_scores": {
    "pattern_recognition": <0-5>,
    "algorithmic_correctness": <0-5>,
    "complexity_understanding": <0-5>,
    "edge_case_awareness": <0-5>
  },
  "feedback": "<Multi-paragraph detailed feedback here>",
  "solution": {
    "pattern": "<Pattern name>",
    "why_this_pattern": "<Explanation>",
    "approach_steps": [
      "<Step 1>",
      "<Step 2>",
      "<Step 3>"
    ],
    "pseudocode": "<Clean pseudocode here>",
    "time_complexity": "<e.g., O(n)>",
    "space_complexity": "<e.g., O(1)>",
    "complexity_explanation": "<Why this complexity>",
    "key_insights": [
      "<Insight 1>",
      "<Insight 2>"
    ],
    "common_pitfalls": [
      "<Pitfall 1>",
      "<Pitfall 2>"
    ],
    "correct_approach": "<1-2 sentence summary>"
  }
}`, questionTitle, questionDescription, userAnswer)
}

func (l *LLMService) parseJSONResponse(response string) (int, string, string, *models.SubScores, *models.SolutionBreakdown, error) {
	// Clean up potential markdown formatting
	cleaned := cleanJSONResponse(response)

	var llmResp LLMResponse
	if err := json.Unmarshal([]byte(cleaned), &llmResp); err != nil {
		return 0, "", "", nil, nil, fmt.Errorf("JSON unmarshal error: %w", err)
	}

	// Validate and clamp scores
	score := clampScore(llmResp.Score)

	subScores := &models.SubScores{
		PatternRecognition:      clampScore(llmResp.SubScores.PatternRecognition),
		AlgorithmicCorrectness:  clampScore(llmResp.SubScores.AlgorithmicCorrectness),
		ComplexityUnderstanding: clampScore(llmResp.SubScores.ComplexityUnderstanding),
		EdgeCaseAwareness:       clampScore(llmResp.SubScores.EdgeCaseAwareness),
	}

	solutionBreakdown := &models.SolutionBreakdown{
		Pattern:               llmResp.Solution.Pattern,
		WhyThisPattern:        llmResp.Solution.WhyThisPattern,
		ApproachSteps:         llmResp.Solution.ApproachSteps,
		Pseudocode:            llmResp.Solution.Pseudocode,
		TimeComplexity:        llmResp.Solution.TimeComplexity,
		SpaceComplexity:       llmResp.Solution.SpaceComplexity,
		ComplexityExplanation: llmResp.Solution.ComplexityExplanation,
		KeyInsights:           llmResp.Solution.KeyInsights,
		CommonPitfalls:        llmResp.Solution.CommonPitfalls,
	}

	// Ensure arrays are not nil
	if solutionBreakdown.ApproachSteps == nil {
		solutionBreakdown.ApproachSteps = []string{}
	}
	if solutionBreakdown.KeyInsights == nil {
		solutionBreakdown.KeyInsights = []string{}
	}
	if solutionBreakdown.CommonPitfalls == nil {
		solutionBreakdown.CommonPitfalls = []string{}
	}

	return score, llmResp.Feedback, llmResp.Solution.CorrectApproach, subScores, solutionBreakdown, nil
}

// cleanJSONResponse removes markdown code fences and extra whitespace
func cleanJSONResponse(response string) string {
	// Remove markdown JSON code fences if present
	if len(response) > 7 && response[:7] == "```json" {
		response = response[7:]
	}
	if len(response) > 3 && response[len(response)-3:] == "```" {
		response = response[:len(response)-3]
	}

	// Trim whitespace
	return response
}

// clampScore ensures scores are within 0-5 range
func clampScore(score int) int {
	if score < 0 {
		return 0
	}
	if score > 5 {
		return 5
	}
	return score
}

// TranscribeAudio uses OpenAI Whisper API to transcribe audio to text
func (l *LLMService) TranscribeAudio(ctx context.Context, audioFile io.Reader, filename string) (string, error) {
	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: filename,
		Reader:   audioFile,
		Prompt:   "This is a technical explanation of an algorithm or data structure problem. The speaker may mention terms like hashmap, binary search, O(n), pseudocode, edge cases, etc.",
	}

	resp, err := l.client.CreateTranscription(ctx, req)
	if err != nil {
		return "", fmt.Errorf("whisper API error: %w", err)
	}

	log.Printf("🎤 Transcribed audio: %s", resp.Text)
	return resp.Text, nil
}

// EnhanceAnswer uses GPT-4o-mini to clean up transcription errors only
func (l *LLMService) EnhanceAnswer(ctx context.Context, rawTranscription string) (string, error) {
	prompt := fmt.Sprintf(`You are cleaning up a voice transcription of a student explaining their algorithm approach.

**Raw Transcription:**
%s

**Your Task:**
Fix ONLY speech-to-text errors and basic grammar. DO NOT add any content, explanations, or pseudocode that wasn't explicitly said.

**Fix:**
- Speech recognition errors (e.g., "hash map" → "hashmap", "oh of n" → "O(n)", "for loop" → "for loop")
- Technical term corrections (binary search, two pointers, sliding window, etc.)
- Basic grammar and punctuation
- Remove filler words (um, uh, like, you know)

**DO NOT:**
- Add explanations the user didn't say
- Generate pseudocode unless they dictated it line by line
- Add steps or details they didn't mention
- Restructure their explanation significantly

**Output the minimally cleaned transcription directly, preserving their exact words and approach.**`, rawTranscription)

	resp, err := l.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are an expert at cleaning up technical transcriptions and formatting algorithm explanations.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.3,
			MaxTokens:   500,
		},
	)

	if err != nil {
		return "", fmt.Errorf("GPT API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from GPT")
	}

	enhanced := resp.Choices[0].Message.Content
	log.Printf("✨ Enhanced answer: %s", enhanced)
	return enhanced, nil
}
