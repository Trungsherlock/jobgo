package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:	"profile",
	Short:	"Manage user's profile",
}

var profileImportCmd = &cobra.Command{
	Use:	"import",
	Short:	"import user resume",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO: import profile", args[0])
	},
}

var profileShowCmd = &cobra.Command{
	Use:	"show",
	Short:	"Display current profile",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO: show profile",)
	},
}

var profileSetCmd = &cobra.Command{
	Use:	"set",
	Short:	"Update user profile",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO: update profile")
	},
}

func init() {
	rootCmd.AddCommand(profileCmd)
	profileCmd.AddCommand(profileImportCmd)
	profileCmd.AddCommand(profileShowCmd)
	profileCmd.AddCommand(profileSetCmd)

	profileSetCmd.Flags().String("skills", "", "Skills")
	profileSetCmd.Flags().String("roles", "", "Roles")
	profileSetCmd.Flags().String("locations", "", "Locations")
}