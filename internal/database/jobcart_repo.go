package database

import "time"

func (d *DB) AddToCart(companyID string) error {
	_, err := d.Exec(
		`UPDATE companies SET in_cart = 1, cart_added_at = CURRENT_TIMESTAMP WHERE id = ?`, companyID,
	)
	return err
}

func (d *DB) RemoveFromCart(companyID string) error {
	_, err := d.Exec(
		`UPDATE companies SET in_cart = 0, cart_added_at = NULL WHERE id = ?`,
		companyID,
	)
	return err
}

func (d *DB) ListCartCompanies() ([]Company, error) {
	return d.listCompaniesWhere("COALESCE(in_cart, 0) = 1 ORDER BY cart_added_at DESC")
}

func (d *DB) UpdateLastNotified(companyID string, t time.Time) error {
	_, err := d.Exec(
		`UPDATE companies SET last_notified_at = ? WHERE id = ?`,
		t, companyID,
	)
	return err
}