package scraper

import (
	"context"
	"time"
)

type RawJob struct {
	ExternalID	string
	Title		string
	Description	string
	Location	string
	Remote		bool
	Department	string
	URL			string
	PostedAt	*time.Time
}

type Scraper interface {
	Name() string
	FetchJobs(ctx context.Context, slug string) ([]RawJob, error)
}