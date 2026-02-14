package database

import "time"

type Company struct {
	ID				string
	Name			string
	Platform		string
	Slug			string
	CareerURL		string
	Enabled			bool
	LastScrapedAt	*time.Time
	CreatedAt		time.Time
}

type Job struct {
	ID			string
	CompanyID	string
	ExternalID	*string
	Title		string
	Description	*string
	Location	*string
	Remote		bool
	Department	*string
	Skills		*string
	URL			string
	PostedAt	*time.Time
	ScrapedAt	*time.Time
	MatchScore	*float64
	MatchReason	*string
	Status		string
	CreatedAt	time.Time
}

type Profile struct {
	ID					int
	Name				string
	Email				string
	Skills				string
	ExperienceYears		int
	PreferredRoles		string
	PreferredLocations	string
	MinMatchScore		float64
	ResumeRaw			string
	CreatedAt			time.Time
	UpdatedAt			time.Time
}

type Application struct {
	ID			string
	JobID		string
	AppliedAt	time.Time
	Status		string
	Notes		string
	UpdatedAt	time.Time
}