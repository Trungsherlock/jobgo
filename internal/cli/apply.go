package cli

import (
	"fmt"
	"os"
	"text/tabwriter"
	"encoding/json"

	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply [job-id]",
	Short: "Mark a job as applied",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		notes, _ := cmd.Flags().GetString("notes")

		job, err := db.GetJob(args[0])
		if err != nil {
			return fmt.Errorf("finding job: %w", err)
		}

		app, err := db.CreateApplication(job.ID, notes)
		if err != nil {
			return fmt.Errorf("creating application: %w", err)
		}

		companyName := job.CompanyID[:8]
		c, err := db.GetCompany(job.CompanyID)
		if err == nil {
			companyName = c.Name
		}

		fmt.Printf("Marked as applied: %s @ %s (app id: %s)\n", job.Title, companyName, app.ID[:8])
		return nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show application pipeline summary",
	RunE: func(cmd *cobra.Command, args []string) error {
		summaries, err := db.GetApplicationSummary()
		if err != nil {
			return err
		}

		if len(summaries) == 0 {
			fmt.Println("No applications yet. Apply with: jobgo apply <job-id>")
			return nil
		}

		output, _ := cmd.Flags().GetString("output")
		if output == "json" {
			data, _ := json.MarshalIndent(summaries, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "STATUS\tCOUNT")
		total := 0
		for _, s := range summaries {
			_, _ = fmt.Fprintf(w, "%s\t%d\n", s.Status, s.Count)
			total += s.Count
		}
		_, _ = fmt.Fprintf(w, "---\t---\n")
		_, _ = fmt.Fprintf(w, "total\t%d\n", total)
		_ = w.Flush()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(statusCmd)

	applyCmd.Flags().String("notes", "", "Notes about the application")
}
