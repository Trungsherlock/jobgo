package database

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (d *DB) CreateJob(companyID, externalID, title, description, location, department, skills, url string, remote bool, postedAt *time.Time) (*Job, error) {
	id := uuid.New().String()
	_, err := d.Exec(
		`INSERT OR IGNORE INTO jobs (id, company_id, external_id, title, description, location, remote, department, skills, url, posted_at, scraped_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		id, companyID, externalID, title, description, location, remote, department, skills, url, postedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("inserting job: %w", err)
	}
	return d.GetJob(id)
}

func (d *DB) GetJob(id string) (*Job, error) {
	j := &Job{}
	err := d.QueryRow(
		`SELECT id, company_id, external_id, title, description, location, remote, department, skills, url, posted_at, scraped_at, match_score, match_reason, status, created_at
		 FROM jobs WHERE id = ?`, id,
	).Scan(&j.ID, &j.CompanyID, &j.ExternalID, &j.Title, &j.Description, &j.Location, &j.Remote, &j.Department, &j.Skills, &j.URL, &j.PostedAt, &j.ScrapedAt, &j.MatchScore, &j.MatchReason, &j.Status, &j.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("getting job: %w", err)
	}
	return j, nil
}

func (d *DB) ListJobs(minScore float64, companyID string, onlyNew bool, onlyRemote bool) ([]Job, error) {
	query := `SELECT id, company_id, external_id, title, description, location, remote, department, skills, url, posted_at, scraped_at, match_score, match_reason, status, created_at FROM jobs WHERE 1=1`
	var args []interface{}

	if minScore > 0 {
		query += " AND match_score >= ?"
		args = append(args, minScore)
	}
	if companyID != "" {
		query += " AND company_id = ?"
		args = append(args, companyID)
	}
	if onlyNew {
		query += " AND status = 'new'"
	}
	if onlyRemote {
		query += " AND remote = 1"
	}

	query += " ORDER BY match_score DESC, created_at DESC"

	rows, err := d.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing jobs: %w", err)
	}
	defer rows.Close()

	var jobs []Job
	for rows.Next() {
		var j Job
		if err := rows.Scan(&j.ID, &j.CompanyID, &j.ExternalID, &j.Title, &j.Description, &j.Location, &j.Remote, &j.Department, &j.Skills, &j.URL, &j.PostedAt, &j.ScrapedAt, &j.MatchScore, &j.MatchReason, &j.Status, &j.CreatedAt); err != nil {
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