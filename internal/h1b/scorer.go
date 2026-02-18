package h1b

import (
	"fmt"
	"strings"

	"github.com/Trungsherlock/jobgocli/internal/database"
)

type H1BAdjustment struct {
	Delta	float64
	Reason	string
}

func ScoreH1B(job database.Job, company database.Company, profile database.Profile) H1BAdjustment {
	if !profile.VisaRequired {
		return H1BAdjustment{}
	}

	var delta float64
	var reasons []string

	if company.SponsorsH1b {
		rate := 0.0
		if company.H1bApprovalRate != nil {
			rate = *company.H1bApprovalRate
		}

		if rate >= 90 {
			delta += 15
			reasons = append(reasons, fmt.Sprintf("H1B sponsor (%.0f%% approval)", rate))
		} else if rate >= 70 {
			delta += 10
			reasons = append(reasons, fmt.Sprintf("H1B sponsor (%.0f%% approval)", rate))
		} else {
			delta += 5
			reasons = append(reasons, fmt.Sprintf("H1B sponsor (%.0f%% approval)", rate))
		}
	}

	sentiment := ""
	if job.VisaSentiment != nil {
		sentiment = *job.VisaSentiment
	}

	switch sentiment {
	case "positive":
		delta += 10
		reasons = append(reasons, "Visa Sponsorship mentioned positively")
	case "negative":
		delta -= 20
		reasons = append(reasons, "No visa sponsorship")
	}
	if job.IsNewGrad && profile.ExperienceYears <= 2 {
		delta += 5
		reasons = append(reasons, "New Grad friendly")
	}

	reason := ""
	if len(reasons) > 0 {
		reason = strings.Join(reasons, " | ")
	}
	return H1BAdjustment {
		Delta: delta,
		Reason: reason,
	}
}