package cli

import (
	"fmt"
	"os"
	"text/tabwriter"
	"encoding/json"

	"github.com/spf13/cobra"
)

var companyCmd = &cobra.Command{
	Use:	"company",
	Short:	"Manage target companies",
}

var companyAddCmd = &cobra.Command{
	Use:	"add",
	Short: 	"Add a company to track",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		platform, _ := cmd.Flags().GetString("platform")
		slug, _ := cmd.Flags().GetString("slug")

		if name == "" || platform == "" || slug == "" {
			return fmt.Errorf("--name, --platform, and --slug are required")
		}

		company, err := db.CreateCompany(name, platform, slug, "")
		if err != nil {
			return fmt.Errorf("adding company: %w", err)
		}
		fmt.Printf("Added company %s (id: %s)\n", company.Name, company.ID)
		return nil
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
	RunE: func(cmd *cobra.Command, args []string) error {
		companies, err := db.ListCompanies()
		if err != nil {
			return fmt.Errorf("listing companies: %w", err)
		}

		if len(companies) == 0 {
			fmt.Println("No companies tracked yet. Add one with: jobgo company add")
			return nil
		}

		output, _ := cmd.Flags().GetString("output")
		if output == "json" {
			data, _ := json.MarshalIndent(companies, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "ID\tNAME\tPLATFORM\tSLUG\tLAST SCRAPED")
		for _, c := range companies {
			lastScraped := "never"
			if c.LastScrapedAt != nil {
				lastScraped = c.LastScrapedAt.Format("2006-01-02 15:04")
			}
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", c.ID[:8], c.Name, c.Platform, c.Slug, lastScraped)
		}
		_ = w.Flush()
		return nil
	},
}

var companyRemoveCmd = &cobra.Command{
	Use:	"remove",
	Short:	"Remove a tracked company",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := db.DeleteCompany(args[0]); err != nil {
			return fmt.Errorf("removing company: %w", err)
		}
		fmt.Println("Company removed.")
		return nil
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