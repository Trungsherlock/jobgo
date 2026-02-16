package matcher

import (
	"github.com/Trungsherlock/jobgocli/internal/database"
)

type HybridMatcher struct {
	keyword   *KeywordMatcher
	llm       *LLMMatcher
	threshold float64
}

func NewHybridMatcher(apiKey string, threshold float64) *HybridMatcher {
	return &HybridMatcher{
		keyword:   NewKeywordMatcher(),
		llm:       NewLLMMatcher(apiKey),
		threshold: threshold,
	}
}

func (h *HybridMatcher) Match(job database.Job, profile database.Profile) MatchResult {
	keywordResult := h.keyword.Match(job, profile)

	if keywordResult.Score < h.threshold {
		return keywordResult
	}

	llmResult := h.llm.Match(job, profile)

	llmResult.Reason = "[LLM] " + llmResult.Reason
	return llmResult
}
