package scraper

import (
	"context"
	"testing"
	"time"
)

func TestLeverFetchJobs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	scraper := NewLeverScraper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	jobs, err := scraper.FetchJobs(ctx, "spotify")
	if err != nil {
		t.Fatalf("FetchJobs error: %v", err)
	}

	t.Logf("Found %d jobs from Lever/spotify", len(jobs))

	if len(jobs) == 0 {
		t.Log("Warning: no jobs returned - company may have no open positions")
		return
	}

	j := jobs[0]
	if j.ExternalID == "" {
		t.Error("Expected ExternalID to be set")
	}
	if j.Title == "" {
		t.Error("Expected Title to be set")
	}
	if j.URL == "" {
		t.Error("Expected URL to be set")
	}

	t.Logf("Sample job: %s - %s (%s)", j.Title, j.Location, j.URL)
}