package database

import "time"

type Company struct {
	ID				string		`json:"id"`
	Name			string		`json:"name"`
	Platform		string		`json:"platform"`
	Slug			string		`json:"slug"`
	CareerURL		string		`json:"career_url"`
	Enabled			bool		`json:"enabled"`
	LastScrapedAt	*time.Time	`json:"last_scraped_at"`
	CreatedAt		time.Time	`json:"created_at"`
	H1bSponsorID	*string		`json:"h1b_sponsor_id"`
	SponsorsH1b		bool		`json:"sponsors_h1b"`
	H1bApprovalRate	*float64	`json:"h1b_approval_rate"`
	H1bTotalFiled	*int		`json:"h1b_total_filed"`
	InCart			bool		`json:"in_cart"`
	CartAddedAt		*time.Time	`json:"cart_added_at"`
	LastNotifiedAt	*time.Time	`json:"last_notified_at"`
}

type Job struct {
	ID				string		`json:"id"`
	CompanyID		string		`json:"company_id"`
	CompanyName		string		`json:"company_name"`
	ExternalID		*string		`json:"external_id"`
	Title			string		`json:"title"`
	Description		*string		`json:"description"`
	Location		*string		`json:"location"`
	Remote			bool		`json:"remote"`
	Department		*string		`json:"department"`
	Skills			*string		`json:"skills"`
	URL				string		`json:"url"`
	PostedAt		*time.Time	`json:"posted_at"`
	ScrapedAt		*time.Time	`json:"scraped_at"`
	MatchScore		*float64	`json:"match_score"`
	MatchReason		*string		`json:"match_reason"`
	Status			string		`json:"status"`
	CreatedAt		time.Time	`json:"created_at"`
	ExperienceLevel	*string		`json:"experience_level"`
	VisaMentioned	bool		`json:"visa_mentioned"`
	VisaSentiment	*string		`json:"visa_sentiment"`
	IsNewGrad		bool		`json:"is_new_grad"`
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
	VisaRequired		bool
	ExperienceLevel		*string
}

type Application struct {
	ID			string
	JobID		string
	AppliedAt	time.Time
	Status		string
	Notes		string
	UpdatedAt	time.Time
}

type H1bSponsors struct {
	ID					string
	CompanyName			string
	NormalizedName		string
	City				string
	State				string
	NaicsCode			string
	FiscalYear			int
	InitialApprovals	int
	InitialDenials		int
	ContinuingApprovals	int
	ContinuingDenials	int
	ApprovalRate		float64
	TotalPetitions		int
}

type H1bLcas struct {
	ID				string
	EmployerName	string
	JobTitle		string
	SocCode			string
	WageFrom		float64
	WageTo			float64
	WageUnit		string
	WorksiteCity	string
	WorksiteState	string
	SubmitDate		time.Time
	DecisionDate	time.Time
	Status			string
}