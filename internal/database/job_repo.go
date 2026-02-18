package database

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (d *DB) CreateJob(companyID, externalID, title, description, location, department, skills, url string, remote bool, postedAt *time.Time) (bool, error) {
	id := uuid.New().String()
	result, err := d.Exec(
		`INSERT OR IGNORE INTO jobs (id, company_id, external_id, title, description, location, remote, department, skills, url, posted_at, scraped_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		id, companyID, externalID, title, description, location, remote, department, skills, url, postedAt,
	)
	if err != nil {
		return false, fmt.Errorf("inserting job: %w", err)
	}
	n, _ := result.RowsAffected()
	return n > 0, nil
}

func (d *DB) GetJob(id string) (*Job, error) {
	j := &Job{}
	err := d.QueryRow(
		`SELECT id, company_id, external_id, title, description, location, remote, department, skills, url, posted_at, scraped_at, match_score, match_reason, status, created_at, experience_level, visa_mentioned, visa_sentiment, is_new_grad
		 FROM jobs WHERE id = ?`, id,
	).Scan(&j.ID, &j.CompanyID, &j.ExternalID, &j.Title, &j.Description, &j.Location, &j.Remote, &j.Department, &j.Skills, &j.URL, &j.PostedAt, &j.ScrapedAt, &j.MatchScore, &j.MatchReason, &j.Status, &j.CreatedAt, &j.ExperienceLevel, &j.VisaMentioned, &j.VisaSentiment, &j.IsNewGrad)
	if err != nil {
		return nil, fmt.Errorf("getting job: %w", err)
	}
	return j, nil
}

func (d *DB) ListJobs(minScore float64, companyID string, onlyNew bool, onlyRemote bool, onlyVisaFriendly bool, onlyNewGrad bool) ([]Job, error) {
	where := "1=1"
	var args []interface{}

	if minScore > 0 {
		where += " AND match_score >= ?"
		args = append(args, minScore)
	}
	if companyID != "" {
		where += " AND company_id = ?"
		args = append(args, companyID)
	}
	if onlyNew {
		where += " AND status = 'new'"
	}
	if onlyRemote {
		where += " AND remote = 1"
	}
	if onlyVisaFriendly {
		where += " AND (visa_sentiment = 'positive' OR visa_sentiment IS NULL OR visa_sentiment != 'negative')"
		where += " AND company_id IN (SELECT id FROM companies WHERE sponsors_h1b = 1)"
	}
	if onlyNewGrad {
		where += " AND is_new_grad = 1"
	}

	return d.listJobsWhere(where, args...)
}


func (d *DB) ListUnscoredJobs() ([]Job, error) {
	return d.listJobsWhere("match_score IS NULL")
}

func (d *DB) listJobsWhere(where string, args ...interface{}) ([]Job, error) {
	query := `SELECT id, company_id, external_id, title, description, location, remote, department, skills, url, posted_at, scraped_at, match_score, match_reason, status, created_at, experience_level, visa_mentioned, visa_sentiment, is_new_grad FROM jobs WHERE ` + where + ` ORDER BY created_at DESC`

	rows, err := d.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing jobs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var jobs []Job
	for rows.Next() {
		var j Job
		if err := rows.Scan(&j.ID, &j.CompanyID, &j.ExternalID, &j.Title, &j.Description, &j.Location, &j.Remote, &j.Department, &j.Skills, &j.URL, &j.PostedAt, &j.ScrapedAt, &j.MatchScore, &j.MatchReason, &j.Status, &j.CreatedAt, &j.ExperienceLevel, &j.VisaMentioned, &j.VisaSentiment, &j.IsNewGrad); err != nil {
			return nil, fmt.Errorf("scanning job: %w", err)
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

func (d *DB) UpdateJobStatus(id, status string) error {
	_, err := d.Exec(`UPDATE jobs SET status = ? WHERE id = ?`, status, id)
	return err
}

func (d *DB) UpdateJobMatch(id string, score float64, reason string) error {
	_, err := d.Exec(`UPDATE jobs SET match_score = ?, match_reason = ? WHERE id = ?`, score, reason, id)
	return err
}

func (d *DB) UpdateJobClassification(id string, experienceLevel string, isNewGrad bool, visaMentioned bool, visaSentiment string) error {
	_, err := d.Exec(
		`UPDATE jobs SET experience_level = ?, is_new_grad = ?, visa_mentioned = ?, visa_sentiment = ? WHERE id = ?`,
		experienceLevel, isNewGrad, visaMentioned, visaSentiment, id,
	)
	return err
}

func (d *DB) ListUnclassifiedJobs() ([]Job, error) {
	return d.listJobsWhere("experience_level IS NULL")
}
