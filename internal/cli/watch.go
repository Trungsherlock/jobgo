package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Trungsherlock/jobgo/internal/database"
	"github.com/Trungsherlock/jobgo/internal/filter"
	"github.com/Trungsherlock/jobgo/internal/matcher"
	"github.com/Trungsherlock/jobgo/internal/notifier"
	"github.com/Trungsherlock/jobgo/internal/scraper"
	"github.com/Trungsherlock/jobgo/internal/worker"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Run in watch mode, polling for new jobs on an interval",
	RunE: func(cmd *cobra.Command, args []string) error {
		interval, _ := cmd.Flags().GetDuration("interval")
		minScore, _ := cmd.Flags().GetFloat64("min-score")
		cartOnly, _ := cmd.Flags().GetBool("cart")

		var notifiers []notifier.Notifier
		notifiers = append(notifiers, notifier.NewTerminalNotifier())

		notifyTypes := viper.GetStringSlice("notify")
		for _, n := range notifyTypes {
			switch n {
			case "desktop":
				notifiers = append(notifiers, notifier.NewDesktopNotifier())
			case "webhook":
				webhookURL := viper.GetString("webhook_url")
				if webhookURL != "" {
					notifiers = append(notifiers, notifier.NewWebhookNotifier(webhookURL))
				}
			}
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigCh
			fmt.Println("\nShutting down gracefully...")
			cancel()
		}()

		if cartOnly {
			fmt.Printf("Watching cart companies every %s (min score: %.0f). Press Ctrl+C to stop.\n\n", interval, minScore)
		} else {
			fmt.Printf("Watching for new jobs every %s (min score: %.0f). Press Ctrl+C to stop.\n\n", interval, minScore)
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		runCycle(ctx, minScore, notifiers, cartOnly)

		for {
			select {
			case <-ctx.Done():
				fmt.Println("Watch stopped.")
				return nil
			case <-ticker.C:
				runCycle(ctx, minScore, notifiers, cartOnly)
			}
		}
	},
}

func runCycle(ctx context.Context, minScore float64, notifiers []notifier.Notifier, cartOnly bool) {
	if ctx.Err() != nil {
		return
	}

	var companies []database.Company
	var err error
	if cartOnly {
		companies, err = db.ListCartCompanies()
	} else {
		companies, err = db.ListCompanies()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing companies: %v\n", err)
		return
	}

	var enabled []database.Company
	if cartOnly {
		// Cart companies are already filtered; just ensure they're enabled
		for _, c := range companies {
			if c.Enabled {
				enabled = append(enabled, c)
			}
		}
	} else {
		for _, c := range companies {
			if c.Enabled {
				enabled = append(enabled, c)
			}
		}
	}

	if len(enabled) == 0 {
		if cartOnly {
			fmt.Println("No companies in cart to watch.")
		} else {
			fmt.Println("No companies to watch.")
		}
		return
	}

	fmt.Printf("[%s] Scraping %d companies...\n", time.Now().Format("15:04:05"), len(enabled))

	registry := scraper.NewRegistry()
	pool := worker.NewPool(registry, db, 5)
	results := pool.Run(ctx, enabled)

	totalNew := 0
	for _, r := range results {
		if r.Err != nil {
			fmt.Printf("  FAIL  %s: %v\n", r.Company.Name, r.Err)
		} else {
			totalNew += r.JobCount
		}
	}

	// Score unscored jobs
	profile, _ := db.GetProfile()
	if profile != nil {
		unscoredJobs, _ := db.ListUnscoredJobs()
		pipeline := matcher.NewPipeline()
		for _, job := range unscoredJobs {
			result := pipeline.Score(job, *profile)
			_ = db.UpdateJobSkillScore(job.ID, result.Score, result.MatchedSkills, result.MissingSkills, result.Reason)
		}
	}

	// Print new high-match jobs
	if totalNew > 0 && profile != nil {
		highMatches, _ := db.ListJobs(minScore, "", true, false, false, false, false)

		params := filter.Params{}
		if profile.PreferredRoles != "" {
			params.Titles = parseJSONArray(profile.PreferredRoles)
		}
		if profile.PreferredLocations != "" {
			params.Locations = parseJSONArray(profile.PreferredLocations)
		}
		var sponsorIDs map[string]bool
		if profile.VisaRequired {
			allCompanies, _ := db.ListCompanies()
			sponsorIDs = make(map[string]bool)
			for _, c := range allCompanies {
				if c.SponsorsH1b {
					sponsorIDs[c.ID] = true
				}
			}
			params.H1BOnly = true
		}

		filtered := filter.Apply(highMatches, filter.Build(params, sponsorIDs))

		for _, j := range filtered {
			score := 0.0
			if j.SkillScore != nil {
				score = *j.SkillScore
			}
			for _, n := range notifiers {
				_ = n.Notify(j, j.CompanyName, score)
			}
		}
	}
}

func parseJSONArray(s string) []string {
	if s == "" {
		return nil
	}
	var arr []string
	_ = json.Unmarshal([]byte(s), &arr)
	return arr
}

func init() {
	rootCmd.AddCommand(watchCmd)
	watchCmd.Flags().Duration("interval", 30*time.Minute, "Polling interval between scrapes")
	watchCmd.Flags().Float64("min-score", 50.0, "Minimum score to highlight")
	watchCmd.Flags().Bool("cart", false, "Only watch companies in your job cart")
}
