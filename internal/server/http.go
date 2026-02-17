package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Trungsherlock/jobgocli/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	db     *database.DB
	router chi.Router
}

func New(db *database.DB) *Server {
	s := &Server{db: db}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	r.Route("/api", func(r chi.Router) {
		r.Get("/jobs", s.listJobs)
		r.Get("/jobs/{id}", s.getJob)
		r.Get("/companies", s.listCompanies)
		r.Post("/companies", s.addCompany)
		r.Delete("/companies/{id}", s.deleteCompany)
		r.Get("/profile", s.getProfile)
		r.Get("/stats", s.getStats)
	})

	s.router = r
}

func (s *Server) ListenAndServe(port int) error {
	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("API server listening on http://localhost%s\n", addr)
	return http.ListenAndServe(addr, s.router)
}

// --- Handlers ---

func (s *Server) listJobs(w http.ResponseWriter, r *http.Request) {
	minScore, _ := strconv.ParseFloat(r.URL.Query().Get("min_score"), 64)
	companyID := r.URL.Query().Get("company_id")
	onlyNew := r.URL.Query().Get("new") == "true"
	onlyRemote := r.URL.Query().Get("remote") == "true"

	jobs, err := s.db.ListJobs(minScore, companyID, onlyNew, onlyRemote)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, jobs)
}

func (s *Server) getJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	job, err := s.db.GetJob(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "job not found")
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (s *Server) listCompanies(w http.ResponseWriter, r *http.Request) {
	companies, err := s.db.ListCompanies()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, companies)
}

func (s *Server) addCompany(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Platform string `json:"platform"`
		Slug     string `json:"slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.Name == "" || req.Platform == "" || req.Slug == "" {
		writeError(w, http.StatusBadRequest, "name, platform, and slug are required")
		return
	}

	company, err := s.db.CreateCompany(req.Name, req.Platform, req.Slug, "")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, company)
}

func (s *Server) deleteCompany(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.db.DeleteCompany(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) getProfile(w http.ResponseWriter, r *http.Request) {
	profile, err := s.db.GetProfile()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if profile == nil {
		writeError(w, http.StatusNotFound, "no profile configured")
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

func (s *Server) getStats(w http.ResponseWriter, r *http.Request) {
	summaries, err := s.db.GetApplicationSummary()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, summaries)
}

// --- Helpers ---

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
