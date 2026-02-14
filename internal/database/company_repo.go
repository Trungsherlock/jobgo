package database

import (
	"fmt"

	"github.com/google/uuid"
)

func (d *DB) CreateCompany(name, platform, slug, careerURL string) (*Company, error) {
	id := uuid.New().String()
	_, err := d.Exec(
		`INSERT INTO companies (id, name, platform, slug, career_url) VALUES (?, ?, ?, ?, ?)`,
		id, name, platform, slug, careerURL,
	)
	if err != nil {
		return nil, fmt.Errorf("inserting company: %w", err)
	}
	return d.GetCompany(id)
}

func (d *DB) GetCompany(id string) (*Company, error) {
	c := &Company{}
	err := d.QueryRow(
		`SELECT id, name, platform, slug, career_url, enabled, last_scraped_at, created_at FROM companies WHERE id = ?`, id,
	).Scan(&c.ID, &c.Name, &c.Platform, &c.Slug, &c.CareerURL, &c.Enabled, &c.LastScrapedAt, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("getting company: %w", err)
	}
	return c, nil
}

func (d *DB) ListCompanies() ([]Company, error) {
	rows, err := d.Query(`SELECT id, name, platform, slug, career_url, enabled, last_scraped_at, created_at FROM companies ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("listing companies: %w", err)
	}
	defer rows.Close()

	var companies []Company
	for rows.Next() {
		var c Company
		if err := rows.Scan(&c.ID, &c.Name, &c.Platform, &c.Slug, &c.CareerURL, &c.Enabled, &c.LastScrapedAt, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning company: %w", err)
		}
		companies = append(companies, c)
	}
	return companies, rows.Err()
}

func (d *DB) DeleteCompany(id string) error {
	result, err := d.Exec(`DELETE FROM companies WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("deleting company: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("company not found: %s", id)
	}
	return nil
}

func (d *DB) UpdateCompanyLastScraped(id string) error {
	_, err := d.Exec(`UPDATE companies SET last_scraped_at = CURRENT_TIMESTAMP WHERE id = ?`, id)
	return err
}