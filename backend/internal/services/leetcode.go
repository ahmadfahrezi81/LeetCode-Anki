package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
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

// Fixed version of the LeetCode service methods

// fetchTitleSlugByID gets the problem slug from its numeric LeetCode ID
// This uses questionList to find the problem by its frontendQuestionId
func (l *LeetCodeService) fetchTitleSlugByID(questionID int) (string, error) {
	query := `
        query problemsetQuestionList($categorySlug: String, $skip: Int, $limit: Int, $filters: QuestionListFilterInput) {
            problemsetQuestionList: questionList(
                categorySlug: $categorySlug
                skip: $skip
                limit: $limit
                filters: $filters
            ) {
                questions: data {
                    titleSlug
                    frontendQuestionId: questionFrontendId
                }
            }
        }
    `

	variables := map[string]interface{}{
		"categorySlug": "",
		"skip":         questionID - 1, // Skip to approximately the question
		"limit":        100,            // Fetch a range around the ID
		"filters":      map[string]interface{}{},
	}

	respData, err := l.executeQuery(query, variables)
	if err != nil {
		return "", err
	}

	var result struct {
		ProblemsetQuestionList struct {
			Questions []struct {
				TitleSlug          string `json:"titleSlug"`
				FrontendQuestionId string `json:"frontendQuestionId"`
			} `json:"questions"`
		} `json:"problemsetQuestionList"`
	}

	if err := json.Unmarshal(respData, &result); err != nil {
		return "", fmt.Errorf("failed to parse slug response for ID %d: %w", questionID, err)
	}

	// Find the question with matching frontend ID
	targetID := strconv.Itoa(questionID)
	for _, q := range result.ProblemsetQuestionList.Questions {
		if q.FrontendQuestionId == targetID {
			return q.TitleSlug, nil
		}
	}

	return "", fmt.Errorf("problem slug not found for ID: %d", questionID)
}

// Alternative approach: Use a mapping file or fetch all problems once
// This is more efficient for batch operations like LeetCode 150

func (l *LeetCodeService) FetchAllProblems() (map[int]string, error) {
	// This fetches all problems and creates an ID -> slug mapping
	allProblems := make(map[int]string)
	skip := 0
	limit := 100

	for {
		query := `
            query problemsetQuestionList($categorySlug: String, $skip: Int, $limit: Int, $filters: QuestionListFilterInput) {
                problemsetQuestionList: questionList(
                    categorySlug: $categorySlug
                    skip: $skip
                    limit: $limit
                    filters: $filters
                ) {
                    total: totalNum
                    questions: data {
                        titleSlug
                        frontendQuestionId: questionFrontendId
                    }
                }
            }
        `

		variables := map[string]interface{}{
			"categorySlug": "",
			"skip":         skip,
			"limit":        limit,
			"filters":      map[string]interface{}{},
		}

		respData, err := l.executeQuery(query, variables)
		if err != nil {
			return nil, err
		}

		var result struct {
			ProblemsetQuestionList struct {
				Total     int `json:"total"`
				Questions []struct {
					TitleSlug          string `json:"titleSlug"`
					FrontendQuestionId string `json:"frontendQuestionId"`
				} `json:"questions"`
			} `json:"problemsetQuestionList"`
		}

		if err := json.Unmarshal(respData, &result); err != nil {
			return nil, fmt.Errorf("failed to parse problems list: %w", err)
		}

		// Add to map
		for _, q := range result.ProblemsetQuestionList.Questions {
			id, err := strconv.Atoi(q.FrontendQuestionId)
			if err == nil {
				allProblems[id] = q.TitleSlug
			}
		}

		// Check if we've fetched all problems
		if len(result.ProblemsetQuestionList.Questions) < limit {
			break
		}

		skip += limit

		// Rate limiting
		time.Sleep(500 * time.Millisecond)
	}

	return allProblems, nil
}

// Optimized approach for seed_150/main.go
func (l *LeetCodeService) FetchProblemsByIDs(problemIDs []int) ([]*LeetCodeProblem, error) {
	// First, fetch all problems to build ID -> slug mapping
	fmt.Println("Building problem ID to slug mapping...")
	mapping, err := l.FetchAllProblems()
	if err != nil {
		return nil, fmt.Errorf("failed to build problem mapping: %w", err)
	}

	problems := make([]*LeetCodeProblem, 0, len(problemIDs))

	for _, id := range problemIDs {
		slug, ok := mapping[id]
		if !ok {
			return nil, fmt.Errorf("problem ID %d not found in mapping", id)
		}

		problem, err := l.FetchProblemDetail(slug)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch problem %d (%s): %w", id, slug, err)
		}

		problems = append(problems, problem)

		// Rate limiting
		time.Sleep(500 * time.Millisecond)
	}

	return problems, nil
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

// StripHTMLTags converts HTML content to Markdown using a robust library
func StripHTMLTags(content string) string {
	// Clean up HTML before conversion
	// Remove backticks inside code tags which cause double escaping
	content = strings.ReplaceAll(content, "<code>`", "<code>")
	content = strings.ReplaceAll(content, "`</code>", "</code>")

	converter := md.NewConverter("", true, nil)

	markdown, err := converter.ConvertString(content)
	if err != nil {
		// Fallback to simple stripping if conversion fails
		return strings.TrimSpace(content)
	}

	return strings.TrimSpace(markdown)
}
