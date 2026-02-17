package matcher

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Trungsherlock/jobgocli/internal/database"
)

type KeywordMatcher struct {}

func NewKeywordMatcher() *KeywordMatcher {
	return &KeywordMatcher{}
}

func (k *KeywordMatcher) Match(job database.Job, profile database.Profile) MatchResult {
	profileSkills := parseJSONArray(profile.Skills)
	profileRoles := parseJSONArray(profile.PreferredRoles)
	profileLocations := parseJSONArray(profile.PreferredLocations)

	if len(profileSkills) == 0 && len(profileRoles) == 0 {
		return MatchResult{
			Score: 0,
			Reason: "No skills or roles specified in profile",
		}
	}

	jobText := strings.ToLower(job.Title)
	if job.Description != nil {
		jobText += " " + strings.ToLower(*job.Description)
	}
	if job.Department != nil {
		jobText += " " + strings.ToLower(*job.Department)
	}

	var score float64
	var reasons []string

	if len(profileSkills) > 0 {
		matchedSkills := 0
		var matched []string
		for _, skill := range profileSkills {
			if strings.Contains(jobText, strings.ToLower(skill)) {
				matchedSkills++
				matched = append(matched, skill)
			}
		}

		skillScore := float64(matchedSkills) / float64(len(profileSkills)) * 50.0
		score += skillScore
		if len(matched) > 0 {
			reasons = append(reasons, fmt.Sprintf("Matched skills: %s (%.1f%%)", strings.Join(matched, ", "), skillScore))
		}
	}

	if len(profileRoles) > 0 {
		titleLower := strings.ToLower(job.Title)
		bestRoleMatch := 0.0
		var matchedRole string
		for _, role := range profileRoles {
			roleLower := strings.ToLower(role)
			if strings.Contains(titleLower, roleLower) {
				bestRoleMatch = 1.0
				matchedRole = role
				break
			}

			roleWords := strings.Fields(roleLower)
			titleWords := strings.Fields(titleLower)
			overlap := wordOverlap(roleWords, titleWords)
			if overlap > bestRoleMatch {
				bestRoleMatch = overlap
				matchedRole = role
			}
		}
		roleScore := bestRoleMatch * 30.0
		score += roleScore
		if matchedRole != "" && bestRoleMatch > 0.3 {
			reasons = append(reasons, fmt.Sprintf("Role: %s (%.0f%%)", matchedRole, bestRoleMatch*100))
		}
	}

	if len(profileLocations) > 0 {
		jobLocation := ""
		if job.Location != nil {
			jobLocation = strings.ToLower(*job.Location)
		}
		for _, loc := range profileLocations {
			locLower := strings.ToLower(loc)
			if locLower == "remote" && job.Remote {
				score += 20.0
				reasons = append(reasons, "Location: remote")
				break
			}
			if strings.Contains(jobLocation, locLower) {
				score += 20.0
				reasons = append(reasons, fmt.Sprintf("Location: %s", loc))
				break
			}
		}
	}

	reason := "No match"
	if len(reasons) > 0 {
		reason = strings.Join(reasons, " | ")
	}

	return MatchResult{
		Score:  score,
		Reason: reason,
	}
}

func wordOverlap(a, b []string) float64 {
	if len(a) == 0 {
		return 0
	}
	bSet := make(map[string]bool, len(b))
	for _, w := range b {
		bSet[w] = true
	}
	matches := 0
	for _, w := range a {
		if bSet[w] {
			matches++
		}
	}
	return float64(matches) / float64(len(a))
}

func parseJSONArray(s string) []string {
	var arr []string
	if s == "" {
		return nil
	}
	_ = json.Unmarshal([]byte(s), &arr)
	return arr
}