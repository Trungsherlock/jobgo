package database

import (
	"database/sql"
	"fmt"
)

func (d *DB) UpsertProfile(p *Profile) error {
	_, err := d.Exec(
		`INSERT INTO profile (id, name, email, skills, experience_years, preferred_roles, preferred_locations, min_match_score, resume_raw, visa_required, updated_at)
		 VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(id) DO UPDATE SET
		   name = excluded.name,
		   email = excluded.email,
		   skills = excluded.skills,
		   experience_years = excluded.experience_years,
		   preferred_roles = excluded.preferred_roles,
		   preferred_locations = excluded.preferred_locations,
		   min_match_score = excluded.min_match_score,
		   resume_raw = excluded.resume_raw,
		   visa_required = excluded.visa_required,
		   updated_at = CURRENT_TIMESTAMP`,
		p.Name, p.Email, p.Skills, p.ExperienceYears, p.PreferredRoles, p.PreferredLocations, p.MinMatchScore, p.ResumeRaw, p.VisaRequired,
	)
	return err
}

func (d *DB) GetProfile() (*Profile, error) {
	p := &Profile{}
	err := d.QueryRow(
		`SELECT id, name, email, skills, experience_years, preferred_roles, preferred_locations, min_match_score, resume_raw, created_at, updated_at, COALESCE(visa_required, 0), experience_level FROM profile WHERE id = 1`,
	).Scan(&p.ID, &p.Name, &p.Email, &p.Skills, &p.ExperienceYears, &p.PreferredRoles, &p.PreferredLocations, &p.MinMatchScore, &p.ResumeRaw, &p.CreatedAt, &p.UpdatedAt, &p.VisaRequired, &p.ExperienceLevel)
	if err == sql.ErrNoRows {
		return nil, nil // no profile yet
	}
	if err != nil {
		return nil, fmt.Errorf("getting profile: %w", err)
	}
	return p, nil
}
