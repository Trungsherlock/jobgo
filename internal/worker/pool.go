package worker

import (
	"context"
	"fmt"
	"sync"

	"github.com/Trungsherlock/jobgocli/internal/database"
	"github.com/Trungsherlock/jobgocli/internal/scraper"
)

type Result struct {
	Company 	database.Company
	JobCount	int
	Err 		error
}

type Pool struct {
	registry 	*scraper.Registry
	db 			*database.DB
	workers 	int
}

func NewPool(registry *scraper.Registry, db *database.DB, workers int) *Pool {
	return &Pool{
		registry: registry,
		db: db,
		workers: workers,
	}
}

func (p *Pool) Run(ctx context.Context, companies []database.Company) []Result {
	var (
		wg 		sync.WaitGroup
		mu 		sync.Mutex
		results []Result
	)

	// Buffered channel as a worker queue
	jobs := make(chan database.Company, len(companies))
	for _, c := range companies {
		jobs <- c
	}
	close(jobs)

	// Spawn workers
	for i := 0; i < p.workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for company := range jobs {
				if ctx.Err() != nil {
					mu.Lock()
					results = append(results, Result{
						Company: company,
						Err: ctx.Err(),
					})
					mu.Unlock()
					continue
				}

				result := p.scrapeCompany(ctx, company)
				mu.Lock()
				results = append(results, result)
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()
	return results
}

func (p *Pool) scrapeCompany(ctx context.Context, company database.Company) Result {
	s, err := p.registry.Get(company.Platform)
	if err != nil {
		return Result{Company: company, Err: err}
	}

	rawJobs, err := s.FetchJobs(ctx, company.Slug)
	if err != nil {
		return Result{Company: company, Err: fmt.Errorf("scraping %s: %w", company.Name, err)}
	}

	newCount := 0
	for _, rj := range rawJobs {
		created, err := p.db.CreateJob(
			company.ID,
			rj.ExternalID,
			rj.Title,
			rj.Description,
			rj.Location,
			rj.Department,
			"",
			rj.URL,
			rj.Remote,
			rj.PostedAt,
		)
		if err != nil {
			continue
		}
		if created {
			newCount++
		}
	}

	_ = p.db.UpdateCompanyLastScraped(company.ID)
	return Result{
		Company: company,
		JobCount: newCount,
	}
}