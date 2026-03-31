package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out and clear saved credentials",
	Run: func(cmd *cobra.Command, args []string) {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("❌ Could not locate config: %v\n", err)
			return
		}

		configFile := filepath.Join(home, ".paas-cli.json")

		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			fmt.Println("ℹ️  Not logged in.")
			return
		}

		if err := os.Remove(configFile); err != nil {
			fmt.Printf("❌ Failed to remove credentials: %v\n", err)
			return
		}

		fmt.Println("👋 Logged out successfully.")
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
