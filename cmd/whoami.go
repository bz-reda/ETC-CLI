package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"paas-cli/internal/config"

	"github.com/spf13/cobra"
)

var whoamiJSON bool

var whoamiCmd = &cobra.Command{
	Use:     "whoami",
	Aliases: []string{"account"},
	Short:   "Show the current CLI identity",
	Long:    "Prints the email, user ID, API endpoint and CLI version for the currently logged-in session. Tokens are never printed.",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		loggedIn := cfg.Token != "" || cfg.APIToken != ""

		if whoamiJSON {
			out := map[string]interface{}{
				"logged_in":    loggedIn,
				"email":        cfg.Email,
				"user_id":      cfg.UserID,
				"api_host":     cfg.APIHost,
				"cli_version":  version,
			}
			b, _ := json.MarshalIndent(out, "", "  ")
			fmt.Println(string(b))
			if !loggedIn {
				os.Exit(1)
			}
			return
		}

		if !loggedIn {
			fmt.Println("✗ Not signed in. Run: espacetech login")
			os.Exit(1)
		}

		fmt.Printf("✓ Signed in as: %s\n", cfg.Email)
		if cfg.UserID != "" {
			fmt.Printf("  User ID:      %s\n", cfg.UserID)
		}
		fmt.Printf("  API endpoint: %s\n", cfg.APIHost)
		fmt.Printf("  CLI version:  %s\n", version)
	},
}

func init() {
	whoamiCmd.Flags().BoolVar(&whoamiJSON, "json", false, "output as JSON (tokens are never included)")
	rootCmd.AddCommand(whoamiCmd)
}
