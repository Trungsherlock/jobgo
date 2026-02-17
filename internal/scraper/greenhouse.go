package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"strings"
)

type GreenhouseScraper struct {
	client *http.Client
}

func NewGreenhouseScraper() *GreenhouseScraper {
	return &GreenhouseScraper{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (g *GreenhouseScraper) Name() string {
	return "greenhouse"
}

// Greenhouse API response structure
type greenhouseJobList struct {
	JobList []greenhouseJob `json:"jobs"`
}

type greenhouseJob struct {
	ID			int64					`json:"id"`
	Title		string					`json:"title"`
	URL			string					`json:"absolute_url"`
	Location	greenhouseLocation		`json:"location"`
	UpdatedAt	string					`json:"updated_at"`
	Content		string					`json:"content"`
	Department 	[]greenhouseDept  		`json:"departments"`	
}

type greenhouseLocation struct {
	Name string `json:"name"`
}

type greenhouseDept struct {
	ID  	int64  `json:"id"`
	Name 	string `json:"name"`
}

func (g * GreenhouseScraper) FetchJobs(ctx context.Context, slug string) ([]RawJob, error) {
	url := fmt.Sprintf("https://boards-api.greenhouse.io/v1/boards/%s/jobs", slug)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching greenhouse jobs: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var jobList greenhouseJobList
	if err := json.NewDecoder(resp.Body).Decode(&jobList); err != nil {
		return nil, fmt.Errorf("decoding greenhouse response: %w", err)
	}

	url = fmt.Sprintf("https://boards-api.greenhouse.io/v1/boards/%s/jobs?content=true", slug)
	req, err = http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request for job content: %w", err)
	}
	resp, err = g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching greenhouse jobs with content: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusOK {
		_ = json.NewDecoder(resp.Body).Decode(&jobList)
	}

	jobs := make([]RawJob, 0, len(jobList.JobList))
	for _, j := range jobList.JobList {
		var postedAt *time.Time
		if j.UpdatedAt != "" {
			if t, err := time.Parse("2006-01-02T15:04:05-07:00", j.UpdatedAt); err == nil {
				postedAt = &t
			}
		}
		deptNames := make([]string, len(j.Department))
		for i, d := range j.Department {
			deptNames[i] = d.Name
		}
		jobs = append(jobs, RawJob{
			ExternalID: fmt.Sprintf("%d", j.ID),
			Title:      j.Title,
			Description: j.Content,
			Location:    j.Location.Name,
			Remote:      strings.Contains(strings.ToLower(j.Location.Name), "remote"),
			Department:  strings.Join(deptNames, ", "),
			URL:         j.URL,
			PostedAt:    postedAt,
		})
	}
	return jobs, nil
}