package database

import (
	"os"
	"path/filepath"
	"testing"
)

// helper: creates an in-memory DB with migrations applied
func setupTestDB(t *testing.T) *DB {
	t.Helper()

	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}

	// Find migrations directory (relative to test file)
	migrationsDir := filepath.Join("..", "..", "migrations")
	if _, err := os.Stat(migrationsDir); err != nil {
		t.Fatalf("migrations directory not found: %v", err)
	}

	if err := db.Migrate(migrationsDir); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestCompanyCRUD(t *testing.T) {
	db := setupTestDB(t)

	// Create
	c, err := db.CreateCompany("Stripe", "lever", "stripe", "")
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}
	if c.Name != "Stripe" || c.Platform != "lever" {
		t.Errorf("got name=%s platform=%s, want Stripe/lever", c.Name, c.Platform)
	}

	// List
	companies, err := db.ListCompanies()
	if err != nil {
		t.Fatalf("ListCompanies: %v", err)
	}
	if len(companies) != 1 {
		t.Errorf("got %d companies, want 1", len(companies))
	}

	// Delete
	if err := db.DeleteCompany(c.ID); err != nil {
		t.Fatalf("DeleteCompany: %v", err)
	}
	companies, _ = db.ListCompanies()
	if len(companies) != 0 {
		t.Errorf("got %d companies after delete, want 0", len(companies))
	}
}

func TestJobCRUD(t *testing.T) {
	db := setupTestDB(t)

	c, _ := db.CreateCompany("Test Co", "lever", "testco", "")

	// Create
	created, err := db.CreateJob(c.ID, "ext-123", "Backend Engineer", "Go experience required", "Remote", "Engineering", `["Go"]`, "https://example.com/apply", true, nil)
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}
	if !created {
		t.Error("expected job to be created")
	}

	// List with filters
	jobs, err := db.ListJobs(0, "", false, true, false, false)
	if err != nil {
		t.Fatalf("ListJobs: %v", err)
	}
	if len(jobs) != 1 {
		t.Errorf("got %d jobs, want 1", len(jobs))
	}
	if jobs[0].Title != "Backend Engineer" {
		t.Errorf("got title=%s, want Backend Engineer", jobs[0].Title)
	}


	// Update status
	if err := db.UpdateJobStatus(jobs[0].ID, "applied"); err != nil {
		t.Fatalf("UpdateJobStatus: %v", err)
	}

	// Verify new jobs filter excludes it
	newJobs, _ := db.ListJobs(0, "", true, false, false, false)
	if len(newJobs) != 0 {
		t.Errorf("got %d new jobs after status change, want 0", len(newJobs))
	}
}

func TestProfileUpsert(t *testing.T) {
	db := setupTestDB(t)

	// No profile yet
	p, err := db.GetProfile()
	if err != nil {
		t.Fatalf("GetProfile: %v", err)
	}
	if p != nil {
		t.Error("expected nil profile initially")
	}

	// Create
	err = db.UpsertProfile(&Profile{
		Name:   "John",
		Skills: `["Go","Docker"]`,
	})
	if err != nil {
		t.Fatalf("UpsertProfile: %v", err)
	}

	p, _ = db.GetProfile()
	if p.Name != "John" {
		t.Errorf("got name=%s, want John", p.Name)
	}

	// Update (upsert)
	err = db.UpsertProfile(&Profile{
		Name:   "John Updated",
		Skills: `["Go","Docker","K8s"]`,
	})
	if err != nil {
		t.Fatalf("UpsertProfile update: %v", err)
	}

	p, _ = db.GetProfile()
	if p.Name != "John Updated" {
		t.Errorf("got name=%s, want John Updated", p.Name)
	}
}
