package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var companyCmd = &cobra.Command{
	Use:	"company",
	Short:	"Manage target companies",
}

var companyAddCmd = &cobra.Command{
	Use:	"add",
	Short: "Add a company to track",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO: add company")
	},
}

var companyImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Bulk import companies from a YAML file",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO: import companies")
	},
}

var companyListCmd = &cobra.Command{
	Use:	"list",
	Short:	"List tracked companies",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO: list companies")
	},
}

var companyRemoveCmd = &cobra.Command{
	Use:	"remove",
	Short:	"Remove a tracked company",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO: remove company", args[0])
	},
}

func init() {
	rootCmd.AddCommand(companyCmd)
	companyCmd.AddCommand(companyAddCmd)
	companyCmd.AddCommand(companyImportCmd)
	companyCmd.AddCommand(companyRemoveCmd)
	companyCmd.AddCommand(companyListCmd)

	companyAddCmd.Flags().String("name", "", "Company name")
	companyAddCmd.Flags().String("platform", "", "ATS platform (lever, greenhouse)")
	companyAddCmd.Flags().String("slug", "", "Platform slug")
	companyImportCmd.Flags().String("file", "", "Path to companies YAML file")
}