package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var jobsCmd = &cobra.Command{
	Use:	"jobs",
	Short:	"Manage jobs",
}

var jobsListCmd = &cobra.Command{
	Use:	"list",
	Short:	"List jobs sorted by match score",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO: list jobs")
	},
}

var jobsShowCmd = &cobra.Command{
	Use:	"show",
	Short:	"Show full job description + match score + match reason",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO: show job", args[0])
	},
}

var jobsOpenCmd = &cobra.Command{
	Use:	"open",
	Short:	"Opens the job URL in the default browser",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO: open job", args[0])
	},
}

func init() {
	rootCmd.AddCommand(jobsCmd)
	jobsCmd.AddCommand(jobsListCmd)
	jobsCmd.AddCommand(jobsShowCmd)
	jobsCmd.AddCommand(jobsOpenCmd)

	jobsListCmd.Flags().Int("min-match", 50, "Minimum matching score")
	jobsListCmd.Flags().String("company", "", "Company name")
	jobsListCmd.Flags().Bool("new", false, "New Jobs (only unseen)")
	jobsListCmd.Flags().Bool("remote", false, "Only show remote jobs")
}