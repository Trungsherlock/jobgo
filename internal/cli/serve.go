package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:	"serve",
	Short:	"Start the API server",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO: serve")
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().Int("port", 8080, "Port to listen on")
	serveCmd.Flags().Bool("mcp", false, "Run as MCP server")
}