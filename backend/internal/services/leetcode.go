package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	LeetCodeGraphQLEndpoint = "https://leetcode.com/graphql"
	UserAgent               = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"
)

type LeetCodeService struct {
	client *http.Client
}

func NewLeetCodeService() *LeetCodeService {
	return &LeetCodeService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// LeetCodeProblem represents a problem from the API
type LeetCodeProblem struct {
	QuestionID     string   `json:"questionId"`
	Title          string   `json:"title"`
	TitleSlug      string   `json:"titleSlug"`
	Difficulty     string   `json:"difficulty"`
	Content        string   `json:"content"`
	TopicTags      []Topic  `json:"topicTags"`
	Hints          []string `json:"hints"`
	SampleTestCase string   `json:"sampleTestCase"`
}

type Topic struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// GraphQL request/response structures
type graphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

// FetchRandomProblems fetches random problems by difficulty
func (l *LeetCodeService) FetchRandomProblems(easyCount, mediumCount, hardCount int) ([]*LeetCodeProblem, error) {
	problems := []*LeetCodeProblem{}

	difficulties := map[string]int{
		"EASY":   easyCount,
		"MEDIUM": mediumCount,
		"HARD":   hardCount,
	}

	current := 0

	for difficulty, count := range difficulties {
		for i := 0; i < count; i++ {
			current++

			slug, err := l.fetchRandomProblemSlug(difficulty)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch random %s problem: %w", difficulty, err)
			}

			problem, err := l.FetchProblemDetail(slug)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch problem details for %s: %w", slug, err)
			}

			problems = append(problems, problem)

			// Rate limiting
			time.Sleep(1 * time.Second)
		}
	}

	return problems, nil
}

// fetchRandomProblemSlug gets a random problem slug by difficulty
func (l *LeetCodeService) fetchRandomProblemSlug(difficulty string) (string, error) {
	query := `
		query randomQuestion($categorySlug: String, $filters: QuestionListFilterInput) {
			randomQuestion(categorySlug: $categorySlug, filters: $filters) {
				titleSlug
			}
		}
	`

	variables := map[string]interface{}{
		"categorySlug": "algorithms",
		"filters": map[string]interface{}{
			"difficulty": difficulty,
		},
	}

	respData, err := l.executeQuery(query, variables)
	if err != nil {
		return "", err
	}

	var result struct {
		RandomQuestion struct {
			TitleSlug string `json:"titleSlug"`
		} `json:"randomQuestion"`
	}

	if err := json.Unmarshal(respData, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if result.RandomQuestion.TitleSlug == "" {
		return "", fmt.Errorf("no random question returned")
	}

	return result.RandomQuestion.TitleSlug, nil
}

// FetchProblemDetail fetches full details for a specific problem
func (l *LeetCodeService) FetchProblemDetail(titleSlug string) (*LeetCodeProblem, error) {
	query := `
		query questionData($titleSlug: String!) {
			question(titleSlug: $titleSlug) {
				questionId
				title
				titleSlug
				difficulty
				content
				topicTags {
					name
					slug
				}
				hints
				sampleTestCase
			}
		}
	`

	variables := map[string]interface{}{
		"titleSlug": titleSlug,
	}

	respData, err := l.executeQuery(query, variables)
	if err != nil {
		return nil, err
	}

	var result struct {
		Question *LeetCodeProblem `json:"question"`
	}

	if err := json.Unmarshal(respData, &result); err != nil {
		return nil, fmt.Errorf("failed to parse problem details: %w", err)
	}

	if result.Question == nil {
		return nil, fmt.Errorf("problem not found: %s", titleSlug)
	}

	return result.Question, nil
}

// executeQuery sends a GraphQL request to LeetCode
func (l *LeetCodeService) executeQuery(query string, variables map[string]interface{}) (json.RawMessage, error) {
	reqBody := graphQLRequest{
		Query:     query,
		Variables: variables,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", LeetCodeGraphQLEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Referer", "https://leetcode.com/problemset/")

	// Execute request
	resp, err := l.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("LeetCode API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var gqlResp graphQLResponse
	if err := json.Unmarshal(body, &gqlResp); err != nil {
		return nil, fmt.Errorf("failed to parse GraphQL response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL error: %s", gqlResp.Errors[0].Message)
	}

	return gqlResp.Data, nil
}

// StripHTMLTags removes HTML tags from content (simple version)
func StripHTMLTags(content string) string {
	// Remove HTML tags (basic implementation)
	content = strings.ReplaceAll(content, "<p>", "\n")
	content = strings.ReplaceAll(content, "</p>", "\n")
	content = strings.ReplaceAll(content, "<strong>", "**")
	content = strings.ReplaceAll(content, "</strong>", "**")
	content = strings.ReplaceAll(content, "<code>", "`")
	content = strings.ReplaceAll(content, "</code>", "`")
	content = strings.ReplaceAll(content, "<pre>", "```\n")
	content = strings.ReplaceAll(content, "</pre>", "\n```")
	content = strings.ReplaceAll(content, "<ul>", "\n")
	content = strings.ReplaceAll(content, "</ul>", "\n")
	content = strings.ReplaceAll(content, "<li>", "- ")
	content = strings.ReplaceAll(content, "</li>", "\n")
	content = strings.ReplaceAll(content, "<em>", "*")
	content = strings.ReplaceAll(content, "</em>", "*")
	content = strings.ReplaceAll(content, "<br>", "\n")
	content = strings.ReplaceAll(content, "<br/>", "\n")

	// Remove remaining tags (simple regex alternative)
	for strings.Contains(content, "<") && strings.Contains(content, ">") {
		start := strings.Index(content, "<")
		end := strings.Index(content, ">")
		if start < end {
			content = content[:start] + content[end+1:]
		} else {
			break
		}
	}

	return strings.TrimSpace(content)
}

// GenerateApproachHint creates a hint from the problem's hints or topics
func GenerateApproachHint(problem *LeetCodeProblem) string {
	if len(problem.Hints) > 0 {
		// Use first hint as approach
		return problem.Hints[0]
	}

	// Fallback: generate from topics
	if len(problem.TopicTags) > 0 {
		topics := make([]string, len(problem.TopicTags))
		for i, tag := range problem.TopicTags {
			topics[i] = tag.Name
		}
		return fmt.Sprintf("Consider using: %s", strings.Join(topics, ", "))
	}

	return "Think about the optimal data structure and algorithmic approach for this problem."
}
