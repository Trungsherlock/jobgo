package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Trungsherlock/jobgocli/internal/database"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var db *database.DB

var rootCmd = &cobra.Command{
	Use:	"jobgo",
	Short:	"A CLI tool to find and track job opportunities",
	Long:	`jobgo crawls jobs from your target companies, matches them against your resume, and notifies you in real-time.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("getting home dir: %w", err)
		}

		dbPath := viper.GetString("database.path")
		if dbPath == "" {
			dbPath = filepath.Join(home, ".jobgo", "jobgo.db")
		}

		db, err = database.New(dbPath)
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}

		migrationsDir := viper.GetString("database.migrations")
		if migrationsDir == "" {
			candidates := []string{
				"migrations",
				filepath.Join(home, ".jobgo", "migrations"),
			}
			for _, c := range candidates {
				if _, err := os.Stat(c); err == nil {
					migrationsDir = c
					break
				}
			}
		}

		if migrationsDir != "" {
			if err := db.Migrate(migrationsDir); err != nil {
				return fmt.Errorf("running migration: %w", err)
			}
		}

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		if db != nil {
			return db.Close()
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().String("config", "", "config file (default is $HOME/.jobgo/config.yaml)")
	rootCmd.PersistentFlags().Bool("verbose", false, "enable verbose output")
	rootCmd.PersistentFlags().Bool("debug", false, "enable debug output")
	rootCmd.PersistentFlags().StringP("output", "o", "table", "Output format (table, json)")
}

func initConfig() {
	cfgFile, _ := rootCmd.Flags().GetString("config")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		viper.AddConfigPath(home + "/.jobgo")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		if v, _ := rootCmd.Flags().GetBool("verbose"); v {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}
	}
}