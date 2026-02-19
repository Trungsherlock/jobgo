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
		`SELECT id, name, platform, slug, career_url, enabled, last_scraped_at, created_at,
		h1b_sponsor_id, sponsors_h1b, h1b_approval_rate, h1b_total_filed, COALESCE(in_cart, 0), cart_added_at, last_notified_at FROM companies WHERE id = ?`, id,
	).Scan(&c.ID, &c.Name, &c.Platform, &c.Slug, &c.CareerURL, &c.Enabled, &c.LastScrapedAt, &c.CreatedAt, &c.H1bSponsorID, &c.SponsorsH1b, &c.H1bApprovalRate, &c.H1bTotalFiled, &c.InCart, &c.CartAddedAt, &c.LastNotifiedAt)
	if err != nil {
		return nil, fmt.Errorf("getting company: %w", err)
	}
	return c, nil
}

func (d *DB) ListCompanies() ([]Company, error) {
	return d.listCompaniesWhere("1=1 ORDER BY name")
}

func (d *DB) DeleteCompany(idPrefix string) error {
	result, err := d.Exec(`DELETE FROM companies WHERE id LIKE ?`, idPrefix+"%")
	if err != nil {
		return fmt.Errorf("deleting company: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("company not found: %s", idPrefix)
	}
	if n > 1 {
		return fmt.Errorf("ambiguous prefix %s matched %d companies", idPrefix, n)
	}
	return nil
}

func (d *DB) UpdateCompanyLastScraped(id string) error {
	_, err := d.Exec(`UPDATE companies SET last_scraped_at = CURRENT_TIMESTAMP WHERE id = ?`, id)
	return err
}

func (d *DB) listCompaniesWhere(where string) ([]Company, error) {
	rows, err := d.Query(`SELECT id, name, platform, slug, career_url, enabled, last_scraped_at, created_at,
        h1b_sponsor_id, sponsors_h1b, h1b_approval_rate, h1b_total_filed,
        COALESCE(in_cart, 0), cart_added_at, last_notified_at
        FROM companies WHERE ` + where)
    if err != nil {
        return nil, err
    }
    defer func() { _ = rows.Close() }()

	var companies []Company
	for rows.Next() {
		var c Company
		if err := rows.Scan(&c.ID, &c.Name, &c.Platform, &c.Slug, &c.CareerURL, &c.Enabled,
		&c.LastScrapedAt, &c.CreatedAt, &c.H1bSponsorID, &c.SponsorsH1b,
		&c.H1bApprovalRate, &c.H1bTotalFiled,
		&c.InCart, &c.CartAddedAt, &c.LastNotifiedAt); err != nil {
			return nil, err
		}
		companies = append(companies, c)
	}
	return companies, rows.Err()
}