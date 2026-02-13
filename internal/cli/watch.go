package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:	"watch",
	Short:	"Run in watch mode, polling for new jobs on an interval",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO: watch")
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)
	watchCmd.Flags().Duration("interval", 30*time.Minute, "Polling interval between scrapes")
}