package matcher

import (
	"os"
	"testing"

	"github.com/Trungsherlock/jobgocli/internal/database"
)

func TestLLMMatcher(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping LLM matcher test")
	}

	m := NewLLMMatcher(apiKey)

	profile := database.Profile{
		Name:               "Trung",
		Skills:             `["Go","PostgreSQL","Docker","Kubernetes"]`,
		PreferredRoles:     `["backend engineer"]`,
		PreferredLocations: `["remote"]`,
		ExperienceYears:    3,
	}

	job := database.Job{
		Title:       "Senior Backend Engineer",
		Description: strPtr("We are looking for a backend engineer with experience in Go, PostgreSQL, and Docker. You will build and maintain microservices deployed on Kubernetes."),
		Location:    strPtr("Remote"),
		Remote:      true,
		Department:  strPtr("Engineering"),
	}

	result := m.Match(job, profile)

	t.Logf("LLM Score: %.1f", result.Score)
	t.Logf("LLM Reason: %s", result.Reason)

	if result.Score < 50 {
		t.Errorf("expected high match score for a perfect match, got %.1f", result.Score)
	}
	if result.Reason == "" {
		t.Error("expected a reason from LLM")
	}
}

func TestHybridMatcher(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping hybrid matcher test")
	}

	m := NewHybridMatcher(apiKey, 30.0)

	profile := database.Profile{
		Name:               "Trung",
		Skills:             `["Go","PostgreSQL","Docker","Kubernetes"]`,
		PreferredRoles:     `["backend engineer"]`,
		PreferredLocations: `["remote"]`,
		ExperienceYears:    3,
	}

	// This job should pass keyword threshold and get LLM-scored
	goodJob := database.Job{
		Title:       "Backend Engineer - Platform",
		Description: strPtr("Build APIs in Go with PostgreSQL. Deploy with Docker and Kubernetes."),
		Location:    strPtr("Remote"),
		Remote:      true,
	}

	result := m.Match(goodJob, profile)
	t.Logf("Hybrid (good job) Score: %.1f — %s", result.Score, result.Reason)

	if result.Score < 50 {
		t.Errorf("expected high score, got %.1f", result.Score)
	}

	// This job should NOT pass keyword threshold — stays keyword-scored
	badJob := database.Job{
		Title:       "Marketing Manager",
		Description: strPtr("Lead brand campaigns and growth marketing initiatives"),
		Location:    strPtr("London"),
	}

	result2 := m.Match(badJob, profile)
	t.Logf("Hybrid (bad job) Score: %.1f — %s", result2.Score, result2.Reason)

	if result2.Score > 30 {
		t.Errorf("expected low score for unrelated job, got %.1f", result2.Score)
	}
}
