package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Trungsherlock/jobgocli/internal/database"
	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

type MCPServer struct {
	db     *database.DB
	server *mcpserver.MCPServer
}

func NewMCP(db *database.DB) *MCPServer {
	s := mcpserver.NewMCPServer(
		"JobGo",
		"1.0.0",
		mcpserver.WithToolCapabilities(true),
	)

	m := &MCPServer{db: db, server: s}
	m.registerTools()
	return m
}

func (m *MCPServer) registerTools() {
	// search_jobs tool
	m.server.AddTool(
		mcp.NewTool("search_jobs",
			mcp.WithDescription("Search for jobs matching criteria. Returns a list of job postings with match scores."),
			mcp.WithNumber("min_score", mcp.Description("Minimum match score (0-100)"), mcp.DefaultNumber(0)),
			mcp.WithBoolean("remote_only", mcp.Description("Only return remote jobs"), mcp.DefaultBool(false)),
			mcp.WithBoolean("new_only", mcp.Description("Only return unseen jobs"), mcp.DefaultBool(false)),
		),
		m.searchJobs,
	)

	// get_job_details tool
	m.server.AddTool(
		mcp.NewTool("get_job_details",
			mcp.WithDescription("Get full details of a specific job posting including description and match reason."),
			mcp.WithString("job_id", mcp.Required(), mcp.Description("The job ID")),
		),
		m.getJobDetails,
	)

	// list_companies tool
	m.server.AddTool(
		mcp.NewTool("list_companies",
			mcp.WithDescription("List all tracked companies and their scraping status."),
		),
		m.listCompanies,
	)

	// get_profile tool
	m.server.AddTool(
		mcp.NewTool("get_profile",
			mcp.WithDescription("Get the user's profile including skills, preferred roles, and locations."),
		),
		m.getProfile,
	)

	// get_stats tool
	m.server.AddTool(
		mcp.NewTool("get_stats",
			mcp.WithDescription("Get application pipeline statistics showing how many applications are in each stage."),
		),
		m.getStats,
	)
}

func (m *MCPServer) searchJobs(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, _ := req.Params.Arguments.(map[string]interface{})
	minScore, _ := args["min_score"].(float64)
	remoteOnly, _ := args["remote_only"].(bool)
	newOnly, _ := args["new_only"].(bool)

	jobs, err := m.db.ListJobs(minScore, "", newOnly, remoteOnly, false, false)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Build a concise summary for the AI
	type jobSummary struct {
		ID       string   `json:"id"`
		Title    string   `json:"title"`
		Company  string   `json:"company"`
		Location string   `json:"location"`
		Remote   bool     `json:"remote"`
		Score    *float64 `json:"score"`
		Status   string   `json:"status"`
		URL      string   `json:"url"`
	}

	summaries := make([]jobSummary, 0, len(jobs))
	for _, j := range jobs {
		companyName := j.CompanyID[:8]
		c, err := m.db.GetCompany(j.CompanyID)
		if err == nil {
			companyName = c.Name
		}
		location := ""
		if j.Location != nil {
			location = *j.Location
		}
		summaries = append(summaries, jobSummary{
			ID:       j.ID,
			Title:    j.Title,
			Company:  companyName,
			Location: location,
			Remote:   j.Remote,
			Score:    j.MatchScore,
			Status:   j.Status,
			URL:      j.URL,
		})
	}

	data, _ := json.MarshalIndent(summaries, "", "  ")
	return mcp.NewToolResultText(fmt.Sprintf("Found %d jobs:\n%s", len(summaries), string(data))), nil
}

func (m *MCPServer) getJobDetails(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, _ := req.Params.Arguments.(map[string]interface{})
	jobID, _ := args["job_id"].(string)

	job, err := m.db.GetJob(jobID)
	if err != nil {
		return mcp.NewToolResultError("Job not found: " + err.Error()), nil
	}

	companyName := job.CompanyID[:8]
	c, err := m.db.GetCompany(job.CompanyID)
	if err == nil {
		companyName = c.Name
	}

	details := map[string]interface{}{
		"id":      job.ID,
		"title":   job.Title,
		"company": companyName,
		"url":     job.URL,
		"remote":  job.Remote,
		"status":  job.Status,
	}
	if job.Location != nil {
		details["location"] = *job.Location
	}
	if job.Department != nil {
		details["department"] = *job.Department
	}
	if job.Description != nil {
		details["description"] = *job.Description
	}
	if job.MatchScore != nil {
		details["match_score"] = *job.MatchScore
	}
	if job.MatchReason != nil {
		details["match_reason"] = *job.MatchReason
	}

	data, _ := json.MarshalIndent(details, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (m *MCPServer) listCompanies(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	companies, err := m.db.ListCompanies()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	data, _ := json.MarshalIndent(companies, "", "  ")
	return mcp.NewToolResultText(fmt.Sprintf("%d companies:\n%s", len(companies), string(data))), nil
}

func (m *MCPServer) getProfile(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	profile, err := m.db.GetProfile()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if profile == nil {
		return mcp.NewToolResultText("No profile configured. Use 'jobgo profile set' to create one."), nil
	}
	data, _ := json.MarshalIndent(profile, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (m *MCPServer) getStats(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	summaries, err := m.db.GetApplicationSummary()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if len(summaries) == 0 {
		return mcp.NewToolResultText("No applications yet."), nil
	}
	data, _ := json.MarshalIndent(summaries, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (m *MCPServer) ServeStdio() error {
	return mcpserver.ServeStdio(m.server)
}

func (m *MCPServer) ServeSSE(port int) error {
	sseServer := mcpserver.NewSSEServer(m.server)
	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("MCP SSE server listening on http://localhost%s\n", addr)
	return sseServer.Start(addr)
}
