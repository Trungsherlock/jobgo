package matcher

import "github.com/Trungsherlock/jobgocli/internal/database"

type MatchResult struct {
	Score 	float64
	Reason string
}

type Matcher interface {
	Match(job database.Job, profile database.Profile) MatchResult
}


