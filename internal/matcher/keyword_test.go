package matcher

import (
	"testing"

	"github.com/Trungsherlock/jobgocli/internal/database"
)

func strPtr(s string) *string { return &s }

func TestKeywordMatcher(t *testing.T) {
	m := NewKeywordMatcher()

	profile := database.Profile{
		Skills:             `["Go","PostgreSQL","Docker","Kubernetes"]`,
		PreferredRoles:     `["backend engineer"]`,
		PreferredLocations: `["remote"]`,
	}

	tests := []struct {
		name     string
		job      database.Job
		minScore float64
		maxScore float64
	}{
		{
			name: "perfect match",
			job: database.Job{
				Title:       "Senior Backend Engineer",
				Description: strPtr("We need someone with Go, PostgreSQL, Docker and Kubernetes experience"),
				Remote:      true,
			},
			minScore: 80,
			maxScore: 100,
		},
		{
			name: "partial skill match",
			job: database.Job{
				Title:       "Backend Engineer",
				Description: strPtr("Experience with Go and PostgreSQL required"),
				Location:    strPtr("New York"),
			},
			minScore: 40,
			maxScore: 70,
		},
		{
			name: "no match",
			job: database.Job{
				Title:       "Marketing Manager",
				Description: strPtr("Lead our marketing campaigns and brand strategy"),
				Location:    strPtr("London"),
			},
			minScore: 0,
			maxScore: 10,
		},
		{
			name: "title match only",
			job: database.Job{
				Title:       "Backend Engineer - Payments",
				Description: strPtr("Work on payment processing systems using Java and Spring"),
			},
			minScore: 25,
			maxScore: 45,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.Match(tt.job, profile)
			if result.Score < tt.minScore || result.Score > tt.maxScore {
				t.Errorf("score %.1f not in expected range [%.0f, %.0f]. Reason: %s",
					result.Score, tt.minScore, tt.maxScore, result.Reason)
			}
			t.Logf("Score: %.1f — %s", result.Score, result.Reason)
		})
	}
}
