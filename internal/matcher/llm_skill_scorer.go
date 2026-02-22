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

type LLMSkillScorer struct {
	apiKey string
	model  string
	client *http.Client
}

func NewLLMSkillScorer(apiKey string) *LLMSkillScorer {
	return &LLMSkillScorer{
		apiKey: apiKey,
		model: "claude-haiku-4-5-20251001",
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

type llmSkillResponse struct {
	Score			float64		`json:"score"`
	MatchedSkills	[]string	`json:"matched_skills"`
	MissingSkills	[]string	`json:"missing_skills"`
	Reason			string		`json:"reason"`
}

func (l *LLMSkillScorer) Score(job database.Job, profile database.Profile) (SkillScoreResult, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    description := ""
    if job.Description != nil {
        description = *job.Description
        if len(description) > 3000 {
            description = description[:3000] + "..."
        }
    }
    if description == "" {
        return SkillScoreResult{Score: 0, Reason: "No job description"}, nil
    }

    prompt := fmt.Sprintf(`You are a technical recruiter evaluating skill fit.

Candidate skills: %s

Job description:
%s

Rate the TECHNICAL SKILL FIT ONLY from 0-100.
Do NOT consider location, job title, experience level, or visa status.
Focus purely on: does this candidate have the technical skills this job needs?

Respond with ONLY valid JSON, no explanation outside the JSON:
{
  "score": <number 0-100>,
  "matched_skills": [<skills the candidate has that the job wants>],
  "missing_skills": [<skills the job wants that the candidate lacks>],
  "reason": "<2 sentence explanation of the skill fit>"
}`,
        profile.Skills,
        description,
    )

    reqBody := claudeRequest{
        Model:     l.model,
        MaxTokens: 300,
        Messages:  []claudeMessage{{Role: "user", Content: prompt}},
    }

    bodyBytes, err := json.Marshal(reqBody)
    if err != nil {
        return SkillScoreResult{}, fmt.Errorf("marshaling request: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(bodyBytes))
    if err != nil {
        return SkillScoreResult{}, fmt.Errorf("creating request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("x-api-key", l.apiKey)
    req.Header.Set("anthropic-version", "2023-06-01")

    resp, err := l.client.Do(req)
    if err != nil {
        return SkillScoreResult{}, fmt.Errorf("calling API: %w", err)
    }
    defer func() { _ = resp.Body.Close() }()

    var claudeResp claudeResponse
    if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
        return SkillScoreResult{}, fmt.Errorf("decoding response: %w", err)
    }
    if claudeResp.Error != nil {
        return SkillScoreResult{}, fmt.Errorf("API error: %s", claudeResp.Error.Message)
    }
    if len(claudeResp.Content) == 0 {
        return SkillScoreResult{}, fmt.Errorf("empty response")
    }

    text := strings.TrimSpace(claudeResp.Content[0].Text)

    text = strings.TrimPrefix(text, "```json")
    text = strings.TrimPrefix(text, "```")
    text = strings.TrimSuffix(text, "```")
    text = strings.TrimSpace(text)

    var llmResp llmSkillResponse
    if err := json.Unmarshal([]byte(text), &llmResp); err != nil {
        return SkillScoreResult{}, fmt.Errorf("parsing response: %w (raw: %s)", err, text)
    }

    if llmResp.Score < 0 {
        llmResp.Score = 0
    }
    if llmResp.Score > 100 {
        llmResp.Score = 100
    }

    return SkillScoreResult(llmResp), nil
}