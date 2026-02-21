package cli

import (
	"fmt"
	"strings"

	"github.com/Trungsherlock/jobgocli/internal/database"
	"github.com/Trungsherlock/jobgocli/internal/skills"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:	"profile",
	Short:	"Manage user's profile",
}

// var profileImportCmd = &cobra.Command{
// 	Use:	"import",
// 	Short:	"import user resume",
// 	Args: cobra.ExactArgs(1),
// 	Run: func(cmd *cobra.Command, args []string) {
// 		fmt.Println("TODO: import profile", args[0])
// 	},
// }

var profileShowCmd = &cobra.Command{
	Use:	"show",
	Short:	"Display current profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := db.GetProfile()
		if err != nil {
			return fmt.Errorf("getting profile: %w", err)
		}
		if p == nil {
			fmt.Println("No profile set. Create one with: jobgo profile set --skills \"Go,Docker\" --roles \"Backend,DevOps\" --locations \"Remote,NYC\"")
			return nil
		}

		fmt.Printf("Name:               %s\n", p.Name)
		fmt.Printf("Email:              %s\n", p.Email)
		fmt.Printf("Skills:             %s\n", p.Skills)
		fmt.Printf("Experience (years): %d\n", p.ExperienceYears)
		fmt.Printf("Preferred Roles:    %s\n", p.PreferredRoles)
		fmt.Printf("Preferred Locations:%s\n", p.PreferredLocations)
		fmt.Printf("Min Match Score:    %.0f\n", p.MinMatchScore)
		fmt.Printf("Visa Required:      %v\n", p.VisaRequired)
		return nil
	},
}

var profileSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set profile fields",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get existing profile or start fresh
		p, err := db.GetProfile()
		if err != nil {
			return err
		}
		if p == nil {
			p = &database.Profile{MinMatchScore: 50.0}
		}

		// Update only the fields that were explicitly passed
		if cmd.Flags().Changed("name") {
			p.Name, _ = cmd.Flags().GetString("name")
		}
		if cmd.Flags().Changed("email") {
			p.Email, _ = cmd.Flags().GetString("email")
		}
		if cmd.Flags().Changed("skills") {
			raw, _ := cmd.Flags().GetString("skills")
			p.Skills = normalizeSkillsCSV(raw)
		}
		if cmd.Flags().Changed("roles") {
			roles, _ := cmd.Flags().GetString("roles")
			p.PreferredRoles = toJSONArray(roles)
		}
		if cmd.Flags().Changed("locations") {
			locations, _ := cmd.Flags().GetString("locations")
			p.PreferredLocations = toJSONArray(locations)
		}
		if cmd.Flags().Changed("experience") {
			p.ExperienceYears, _ = cmd.Flags().GetInt("experience")
		}
		if cmd.Flags().Changed("min-match") {
			p.MinMatchScore, _ = cmd.Flags().GetFloat64("min-match")
		}

		if cmd.Flags().Changed("visa") {
    		p.VisaRequired, _ = cmd.Flags().GetBool("visa")
		}

		if err := db.UpsertProfile(p); err != nil {
			return fmt.Errorf("saving profile: %w", err)
		}

		fmt.Println("Profile updated.")
		return nil
	},
}

// toJSONArray converts "Go,Docker,K8s" to `["Go","Docker","K8s"]`
func toJSONArray(csv string) string {
	parts := strings.Split(csv, ",")
	quoted := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			quoted = append(quoted, `"`+p+`"`)
		}
	}
	return "[" + strings.Join(quoted, ",") + "]"
}

func normalizeSkillsCSV(csv string) string {
	parts := strings.Split(csv, ",")
	normalized := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		canonical := skills.Normalize(p)
		normalized = append(normalized, `"`+canonical+`"`)
	}
	return "[" + strings.Join(normalized, ",") + "]"
}

func init() {
	rootCmd.AddCommand(profileCmd)
	// profileCmd.AddCommand(profileImportCmd)
	profileCmd.AddCommand(profileShowCmd)
	profileCmd.AddCommand(profileSetCmd)

	profileSetCmd.Flags().String("name", "", "Your name")
	profileSetCmd.Flags().String("email", "", "Your email")
	profileSetCmd.Flags().String("skills", "", "Comma-separated skills (Go,Docker,K8s)")
	profileSetCmd.Flags().String("roles", "", "Comma-separated preferred roles")
	profileSetCmd.Flags().String("locations", "", "Comma-separated preferred locations")
	profileSetCmd.Flags().Int("experience", 0, "Years of experience")
	profileSetCmd.Flags().Float64("min-match", 50.0, "Minimum match score for notifications")
	profileSetCmd.Flags().Bool("visa", false, "Require H1B visa sponsorship")
}