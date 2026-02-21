package matcher

import (
    "fmt"
    "strings"

    "github.com/Trungsherlock/jobgocli/internal/database"
    "github.com/Trungsherlock/jobgocli/internal/skills"
)

type SkillScoreResult struct {
    Score         float64  `json:"score"`
    MatchedSkills []string `json:"matched_skills"`
    MissingSkills []string `json:"missing_skills"`
    Reason        string   `json:"reason"`
}

type SkillScorer struct{}

func NewSkillScorer() *SkillScorer {
    return &SkillScorer{}
}

func (s *SkillScorer) Score(job database.Job, profile database.Profile) SkillScoreResult {
    userSkills := parseJSONArray(profile.Skills)
    if len(userSkills) == 0 {
        return SkillScoreResult{Score: 0, Reason: "No skills in profile"}
    }
    if job.Description == nil || *job.Description == "" {
        return SkillScoreResult{Score: 0, Reason: "No job description"}
    }

    jobSkills := skills.ExtractFromJob(*job.Description)

    userSet := make(map[string]bool, len(userSkills))
    for _, s := range userSkills {
        userSet[skills.Normalize(s)] = true
    }

    requiredMatched  := intersect(userSet, jobSkills.Required)
    preferredMatched := intersect(userSet, jobSkills.Preferred)
    mentionedMatched := intersect(userSet, jobSkills.Mentioned)

    var score float64
    if len(jobSkills.Required) > 0 {
        score += float64(len(requiredMatched)) / float64(len(jobSkills.Required)) * 70
    } else {
        score += 70
    }
    if len(jobSkills.Preferred) > 0 {
        score += float64(len(preferredMatched)) / float64(len(jobSkills.Preferred)) * 20
    } else {
        score += 20
    }
    if len(jobSkills.Mentioned) > 0 {
        score += float64(len(mentionedMatched)) / float64(len(jobSkills.Mentioned)) * 10
    } else {
        score += 10
    }

    matched := append(append(requiredMatched, preferredMatched...), mentionedMatched...)

    missing := difference(userSet, append(jobSkills.Required, jobSkills.Preferred...))

    reason := buildReason(requiredMatched, jobSkills.Required, missing)

    return SkillScoreResult{
        Score:         score,
        MatchedSkills: matched,
        MissingSkills: missing,
        Reason:        reason,
    }
}

func intersect(userSet map[string]bool, jobSkills []string) []string {
    var result []string
    for _, s := range jobSkills {
        if userSet[skills.Normalize(s)] {
            result = append(result, s)
        }
    }
    return result
}

func difference(userSet map[string]bool, jobSkills []string) []string {
    var result []string
    for _, s := range jobSkills {
        if !userSet[skills.Normalize(s)] {
            result = append(result, s)
        }
    }
    return result
}

func buildReason(requiredMatched, required, missing []string) string {
    if len(required) == 0 {
        return "No required skills listed in job description"
    }

    coverage := fmt.Sprintf("%d/%d required skills matched", len(requiredMatched), len(required))

    if len(missing) == 0 {
        return coverage + ". Full match on all required and preferred skills."
    }

    missingRequired := []string{}
    for _, m := range missing {
        for _, r := range required {
            if strings.EqualFold(m, r) {
                missingRequired = append(missingRequired, m)
                break
            }
        }
    }

    if len(missingRequired) > 0 {
        return fmt.Sprintf("%s. Missing required: %s.", coverage, strings.Join(missingRequired, ", "))
    }
    return fmt.Sprintf("%s. Missing preferred: %s.", coverage, strings.Join(missing, ", "))
}
