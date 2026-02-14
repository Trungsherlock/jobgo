CREATE TABLE IF NOT EXISTS companies (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    platform TEXT NOT NULL,
    slug TEXT NOT NULL,
    career_url TEXT,
    enabled BOOLEAN DEFAULT 1,
    last_scraped_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS jobs (
    id TEXT PRIMARY KEY,
    company_id TEXT NOT NULL REFERENCES companies(id),
    external_id TEXT,
    title TEXT NOT NULL,
    description TEXT,
    location TEXT,
    remote BOOLEAN,
    department TEXT,
    skills TEXT,
    url TEXT NOT NULL,
    posted_at DATETIME,
    scraped_at DATETIME,
    match_score REAL,
    match_reason TEXT,
    status TEXT DEFAULT "new",
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(company_id, external_id)
);

CREATE TABLE IF NOT EXISTS profile (
    id INTEGER PRIMARY KEY,
    name TEXT,
    email TEXT,
    skills TEXT,
    experience_years INTEGER,
    preferred_roles TEXT,
    preferred_locations TEXT,
    min_match_score REAL DEFAULT 50.0,
    resume_raw TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);