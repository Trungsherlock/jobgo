package database

import (
	"fmt"

	"github.com/google/uuid"
)

func (d *DB) CreateApplication(jobID, notes string) (*Application, error) {
	id := uuid.New().String()
	_, err := d.Exec(
		`INSERT INTO applications (id, job_id, notes) VALUES (?, ?, ?)`,
		id, jobID, notes,
	)
	if err != nil {
		return nil, fmt.Errorf("creating application: %w", err)
	}

	_, _ = d.Exec(`UPDATE jobs SET status = 'applied' WHERE id = ?`, jobID)

	return d.GetApplication(id)
}

func (d *DB) GetApplication(id string) (*Application, error) {
	a := &Application{}
	err := d.QueryRow(
		`SELECT id, job_id, applied_at, status, notes, updated_at FROM applications WHERE id = ?`, id,
	).Scan(&a.ID, &a.JobID, &a.AppliedAt, &a.Status, &a.Notes, &a.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("getting application: %w", err)
	}
	return a, nil
}

func (d *DB) UpdateApplication(jobID, status, notes string) error {
	_, err := d.Exec(
		`UPDATE applications SET status = ?, notes = ?, updated_at = CURRENT_TIMESTAMP WHERE job_id = ?`,
		status, notes, jobID,
	)
	if err != nil {
		return fmt.Errorf("updating application: %w", err)
	}
	_, err = d.Exec(`UPDATE jobs SET status = ? WHERE id = ?`, status, jobID)
	return err
}

type StatusSummary struct {
	Status string
	Count  int
}

func (d *DB) GetApplicationSummary() ([]StatusSummary, error) {
	rows, err := d.Query(`SELECT status, COUNT(*) FROM applications GROUP BY status ORDER BY COUNT(*) DESC`)
	if err != nil {
		return nil, fmt.Errorf("getting summary: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var summaries []StatusSummary
	for rows.Next() {
		var s StatusSummary
		if err := rows.Scan(&s.Status, &s.Count); err != nil {
			return nil, err
		}
		summaries = append(summaries, s)
	}
	return summaries, rows.Err()
}
