package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type LeverScraper struct {
	client *http.Client
}

func NewLeverScraper() *LeverScraper {
	return &LeverScraper{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (l *LeverScraper) Name() string {
	return "lever"
}

// Lever API response structure
type leverPosting struct {
	ID					string			`json:"id"`
	Text				string			`json:"text"`
	HostedURL			string			`json:"hostedURL"`
	Categories 			leverCategories `json:"categories"`
	Description			string			`json:"description"`
	DescriptionPlain	string			`json:"descriptionPlain"`
	Lists 				[]leverList		`json:"lists"`
	CreatedAt			int64			`json:"createdAt"`
}

type leverCategories struct {
	Location	string	`json:"location"`
	Department	string	`json:"department"`
	Team		string	`json:"team"`
	Commitment	string	`json:"commitment"`
	Level		string	`json:"level"`
}

type leverList struct {
	Text	string	`json:"text"`
	Content string	`json:"content"`
}

func (l *LeverScraper) FetchJobs(ctx context.Context, slug string) ([]RawJob, error) {
	url := fmt.Sprintf("https://api.lever.co/v0/postings/%s?mode=json", slug)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := l.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching lever postings: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lever API returned status %d", resp.StatusCode)
	}

	var postings []leverPosting
	if err := json.NewDecoder(resp.Body).Decode(&postings); err != nil {
		return nil, fmt.Errorf("decoding lever response: %w", err)
	}

	jobs := make([]RawJob, 0, len(postings))
	for _, p := range postings {
		description := p.DescriptionPlain
		for _, list := range p.Lists {
			description += "\n\n" + list.Text + "\n" + list.Content
		}

		var postedAt *time.Time
		if p.CreatedAt > 0 {
			t := time.UnixMilli(p.CreatedAt)
			postedAt = &t
		}

		remote := strings.Contains(strings.ToLower(p.Categories.Location), "remote")

		jobs = append(jobs, RawJob{
			ExternalID: p.ID,
			Title:      p.Text,
			Description: description,
			Location:    p.Categories.Location,
			Remote:      remote,
			Department:  p.Categories.Department,
			URL:         p.HostedURL,
			PostedAt:    postedAt,
		})
	}

	return jobs, nil
}