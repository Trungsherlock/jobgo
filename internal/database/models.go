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
	H1bSponsorID	*string
	SponsorsH1b		bool
	H1bApprovalRate	*float64
	H1bTotalFiled	*int
}

type Job struct {
	ID				string
	CompanyID		string
	ExternalID		*string
	Title			string
	Description		*string
	Location		*string
	Remote			bool
	Department		*string
	Skills			*string
	URL				string
	PostedAt		*time.Time
	ScrapedAt		*time.Time
	MatchScore		*float64
	MatchReason		*string
	Status			string
	CreatedAt		time.Time
	ExperienceLevel	*string
	VisaMentioned	bool
	VisaSentiment	*string
	IsNewGrad		bool
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