package matcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Trungsherlock/jobgocli/internal/database"
)

type LLMMatcher struct {
	apiKey string
	model string
	client *http.Client
}

func NewLLMMatcher(apiKey string) *LLMMatcher {
	return &LLMMatcher{
		apiKey: apiKey,
		model: "claude-sonnet-4-5-20250929",
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

type claudeRequest struct {
	Model		string			`json:"model"`
	MaxTokens 	int				`json:"max_tokens"`
	Messages 	[]claudeMessage `json:"messages"`
}

type claudeMessage struct {
	Role 	string `json:"role"`
	Content string `json:"content"`
}

type claudeResponse struct {
	Content	[]claudeContent	`json:"content"`
	Error	*claudeError		`json:"error,omitempty"`
}

type claudeContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type claudeError struct {
	Type 	string `json:"type"`
	Message string `json:"message"`
}

type llmMatchResponse struct {
	Score 	float64 `json:"score"`
	Reason	string 	`json:"reason"`
}

func (l *LLMMatcher) Match(job database.Job, profile database.Profile) MatchResult {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := l.callAPI(ctx, job, profile)
	if err != nil {
		// Fallback to keyword matcher on API failure
		return NewKeywordMatcher().Match(job, profile)
	}

	return result
}

func (l *LLMMatcher) callAPI(ctx context.Context, job database.Job, profile database.Profile) (MatchResult, error) {
	// Build the prompt
	description := ""
	if job.Description != nil {
		description = *job.Description
		if len(description) > 3000 {
			description = description[:3000] + "..."
		}
	}

	location := ""
	if job.Location != nil {
		location = *job.Location
	}

	department := ""
	if job.Department != nil {
		department = *job.Department
	}

	prompt := fmt.Sprintf(`You are a job matching assistant. Rate how well this job matches the candidate's profile.

		## Candidate Profile
		- Name: %s
		- Skills: %s
		- Preferred Roles: %s
		- Preferred Locations: %s
		- Experience: %d years

		## Job Posting
		- Title: %s
		- Company Department: %s
		- Location: %s
		- Remote: %v
		- Description: %s

		## Instructions
		Rate the match from 0 to 100 and explain why in 1-2 sentences.
		Consider: skill overlap, role relevance, location fit, and seniority match.

		Respond with ONLY valid JSON in this exact format:
		{"score": <number>, "reason": "<string>"}`,
		profile.Name,
		profile.Skills,
		profile.PreferredRoles,
		profile.PreferredLocations,
		profile.ExperienceYears,
		job.Title,
		department,
		location,
		job.Remote,
		description,
	)

	reqBody := claudeRequest{
		Model:     l.model,
		MaxTokens: 150,
		Messages: []claudeMessage{
			{Role: "user", Content: prompt},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return MatchResult{}, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(bodyBytes))
	if err != nil {
		return MatchResult{}, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", l.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := l.client.Do(req)
	if err != nil {
		return MatchResult{}, fmt.Errorf("calling Claude API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var claudeResp claudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return MatchResult{}, fmt.Errorf("decoding response: %w", err)
	}

	if claudeResp.Error != nil {
		return MatchResult{}, fmt.Errorf("API error: %s", claudeResp.Error.Message)
	}

	if len(claudeResp.Content) == 0 {
		return MatchResult{}, fmt.Errorf("empty response from API")
	}

	// Parse the JSON response from Claude
	text := claudeResp.Content[0].Text
	text = strings.TrimSpace(text)

	var matchResp llmMatchResponse
	if err := json.Unmarshal([]byte(text), &matchResp); err != nil {
		return MatchResult{}, fmt.Errorf("parsing match response: %w (raw: %s)", err, text)
	}

	// Clamp score
	if matchResp.Score < 0 {
		matchResp.Score = 0
	}
	if matchResp.Score > 100 {
		matchResp.Score = 100
	}

	return MatchResult(matchResp), nil
}
