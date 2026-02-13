package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:	"search",
	Short:	"Triggers a full scrape of all enabled companies, stores new jobs, deduplicates against existing jobs",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO: search")
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	
	searchCmd.Flags().String("company", "", "Company name")
	searchCmd.Flags().String("platform", "", "ATS platform (lever, greenhouse)")
	searchCmd.Flags().Duration("timeout", 30*time.Second, "Per-company scrape timeout")
}