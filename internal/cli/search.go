package cli

import (
	"context"
	"fmt"
	"time"
	"os"

	"github.com/Trungsherlock/jobgocli/internal/database"
	"github.com/Trungsherlock/jobgocli/internal/matcher"
	"github.com/Trungsherlock/jobgocli/internal/scraper"
	"github.com/Trungsherlock/jobgocli/internal/worker"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var searchCmd = &cobra.Command{
	Use:	"search",
	Short:	"Triggers a full scrape of all enabled companies, stores new jobs, deduplicates against existing jobs",
	RunE: func(cmd *cobra.Command, args []string) error {
		timeout, _ := cmd.Flags().GetDuration("timeout")
		platformFilter, _ := cmd.Flags().GetString("platform")
		companyFilter, _ := cmd.Flags().GetString("company")

		// Get companies to scrape
		companies, err := db.ListCompanies()
		if err != nil {
			return fmt.Errorf("listing companies: %w", err)
		}

		if len(companies) == 0 {
			fmt.Println("No companies to scrape. Add some with: jobgo company add")
			return nil
		}

		// Apply filters
		var filtered []database.Company
		for _, c := range companies {
			if !c.Enabled {
				continue
			}
			if platformFilter != "" && c.Platform != platformFilter {
				continue
			}
			if companyFilter != "" && c.Name != companyFilter {
				continue
			}
			filtered = append(filtered, c)
		}

		if len(filtered) == 0 {
			fmt.Println("No companies match the specified filters.")
			return nil
		}

		fmt.Printf("Scraping %d companies... \n", len(filtered))

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// Run worker pool
		registry := scraper.NewRegistry()
		pool := worker.NewPool(registry, db, 5)
		results := pool.Run(ctx, filtered)

		profile, err := db.GetProfile()
		if err != nil {
			return fmt.Errorf("getting profile: %w", err)
		}

		if profile != nil {
			// Get all jobs without a match score
			unscoredJobs, err := db.ListUnscoredJobs()
			if err != nil {
				return fmt.Errorf("listing unscored jobs: %w", err)
			}

			if len(unscoredJobs) > 0 {
				fmt.Printf("\nScoring %d new jobs against profile...\n", len(unscoredJobs))
				var m matcher.Matcher
				apiKey := viper.GetString("anthropic_api_key")
				if apiKey == "" {
					apiKey = os.Getenv("ANTHROPIC_API_KEY")
				}

				matcherType := viper.GetString("matcher")
				switch matcherType {
				case "llm":
					if apiKey == "" {
						fmt.Println("No API key found for LLM matcher. Set ANTHROPIC_API_KEY environment variable or use --matcher=keyword")
						m = matcher.NewKeywordMatcher()
					} else {
						m = matcher.NewLLMMatcher(apiKey)
					}
				case "hybrid":
					if apiKey == "" {
						fmt.Println("No API key found for Hybrid matcher. Set ANTHROPIC_API_KEY environment variable or use --matcher=keyword")
						m = matcher.NewKeywordMatcher()
					} else {
						threshold := viper.GetFloat64("hybrid_threshold")
						if threshold == 0 {
							threshold = 50.0
						}
						m = matcher.NewHybridMatcher(apiKey, threshold)
					}
				default:
					m = matcher.NewKeywordMatcher()
				}
				scored := 0
				for _, job := range unscoredJobs {
					result := m.Match(job, *profile)
					_ = db.UpdateJobMatch(job.ID, result.Score, result.Reason)
					scored++
				}
				fmt.Printf("Scored %d jobs.\n", scored)
			}
		}

		var totalNew, failures int
		for _, r := range results {
			if r.Err != nil {
				failures++
				fmt.Printf("  FAIL  %s: %v\n", r.Company.Name, r.Err)
			} else {
				fmt.Printf("  OK    %s: %d new jobs\n", r.Company.Name, r.JobCount)
				totalNew += r.JobCount
			}
		}

		fmt.Printf("\nDone. Found %d new jobs from %d companies. %d failed.\n",
			totalNew, len(filtered)-failures, failures)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	
	searchCmd.Flags().String("company", "", "Company name")
	searchCmd.Flags().String("platform", "", "ATS platform (lever, greenhouse)")
	searchCmd.Flags().Duration("timeout", 30*time.Second, "Per-company scrape timeout")
}