package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Trungsherlock/jobgocli/internal/database"
	"github.com/Trungsherlock/jobgocli/internal/matcher"
	"github.com/Trungsherlock/jobgocli/internal/scraper"
	"github.com/Trungsherlock/jobgocli/internal/worker"
	"github.com/Trungsherlock/jobgocli/internal/notifier"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Run in watch mode, polling for new jobs on an interval",
	RunE: func(cmd *cobra.Command, args []string) error {
		interval, _ := cmd.Flags().GetDuration("interval")
		minScore, _ := cmd.Flags().GetFloat64("min-score")

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

		fmt.Printf("Watching for new jobs every %s (min score: %.0f). Press Ctrl+C to stop.\n\n", interval, minScore)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		runCycle(ctx, minScore, notifiers)

		for {
			select {
			case <-ctx.Done():
				fmt.Println("Watch stopped.")
				return nil
			case <-ticker.C:
				runCycle(ctx, minScore, notifiers)
			}
		}
	},
}

func runCycle(ctx context.Context, minScore float64, notifiers []notifier.Notifier) {
	if ctx.Err() != nil {
		return
	}

	companies, err := db.ListCompanies()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing companies: %v\n", err)
		return
	}

	var enabled []database.Company
	for _, c := range companies {
		if c.Enabled {
			enabled = append(enabled, c)
		}
	}

	if len(enabled) == 0 {
		fmt.Println("No companies to watch.")
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
		m := matcher.NewKeywordMatcher()
		for _, job := range unscoredJobs {
			result := m.Match(job, *profile)
			_ = db.UpdateJobMatch(job.ID, result.Score, result.Reason)
		}
	}

	// Print new high-match jobs
	if totalNew > 0 && profile != nil {
        highMatches, _ := db.ListJobs(minScore, "", true, false, false, false)
        for _, j := range highMatches {
            score := 0.0
            if j.MatchScore != nil {
                score = *j.MatchScore
            }
            companyName := j.CompanyID[:8]
            c, err := db.GetCompany(j.CompanyID)
            if err == nil {
                companyName = c.Name
            }
            for _, n := range notifiers {
                _ = n.Notify(j, companyName, score)
            }
        }
    }
}

func init() {
	rootCmd.AddCommand(watchCmd)
	watchCmd.Flags().Duration("interval", 30*time.Minute, "Polling interval between scrapes")
	watchCmd.Flags().Float64("min-score", 50.0, "Minimum score to highlight")
}
