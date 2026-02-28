package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"paas-cli/internal/api"
	"paas-cli/internal/config"

	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage environment variables",
}

var envSetCmd = &cobra.Command{
	Use:   "set [KEY=VALUE...]",
	Short: "Set environment variables",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		if cfg.Token == "" {
			fmt.Println("❌ Please login first: espacetech login")
			return
		}

		data, err := os.ReadFile(".espacetech.json")
		if err != nil {
			fmt.Println("❌ No project config found. Run 'espacetech init' first.")
			return
		}

		var projectCfg struct {
			ProjectID string `json:"project_id"`
			Name      string `json:"name"`
		}
		json.Unmarshal(data, &projectCfg)

		client := api.NewClient(cfg)

		// Get current env vars
		current, err := client.GetEnvVars(projectCfg.ProjectID)
		if err != nil {
			current = make(map[string]string)
		}

		// Parse and merge new vars
		for _, arg := range args {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) != 2 {
				fmt.Printf("❌ Invalid format: %s (use KEY=VALUE)\n", arg)
				return
			}
			current[parts[0]] = parts[1]
		}

		err = client.SetEnvVars(projectCfg.ProjectID, current)
		if err != nil {
			fmt.Printf("❌ Failed to set env vars: %v\n", err)
			return
		}

		fmt.Println("✅ Environment variables updated:")
		for _, arg := range args {
			parts := strings.SplitN(arg, "=", 2)
			fmt.Printf("   %s=%s\n", parts[0], parts[1])
		}
		fmt.Println("\n🔄 Redeploy to apply: espacetech deploy --prod")
	},
}

var envListCmd = &cobra.Command{
	Use:   "list",
	Short: "List environment variables",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		if cfg.Token == "" {
			fmt.Println("❌ Please login first: espacetech login")
			return
		}

		data, err := os.ReadFile(".espacetech.json")
		if err != nil {
			fmt.Println("❌ No project config found. Run 'espacetech init' first.")
			return
		}

		var projectCfg struct {
			ProjectID string `json:"project_id"`
			Name      string `json:"name"`
		}
		json.Unmarshal(data, &projectCfg)

		client := api.NewClient(cfg)
		envVars, err := client.GetEnvVars(projectCfg.ProjectID)
		if err != nil {
			fmt.Printf("❌ Failed to get env vars: %v\n", err)
			return
		}

		if len(envVars) == 0 {
			fmt.Println("No environment variables set.")
			return
		}

		fmt.Printf("🔧 Environment variables for %s:\n\n", projectCfg.Name)
		for k, v := range envVars {
			fmt.Printf("   %s=%s\n", k, v)
		}
	},
}

var envRemoveCmd = &cobra.Command{
	Use:   "remove [KEY...]",
	Short: "Remove environment variables",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		if cfg.Token == "" {
			fmt.Println("❌ Please login first: espacetech login")
			return
		}

		data, err := os.ReadFile(".espacetech.json")
		if err != nil {
			fmt.Println("❌ No project config found. Run 'espacetech init' first.")
			return
		}

		var projectCfg struct {
			ProjectID string `json:"project_id"`
			Name      string `json:"name"`
		}
		json.Unmarshal(data, &projectCfg)

		client := api.NewClient(cfg)
		current, err := client.GetEnvVars(projectCfg.ProjectID)
		if err != nil {
			fmt.Printf("❌ Failed to get env vars: %v\n", err)
			return
		}

		for _, key := range args {
			delete(current, key)
			fmt.Printf("   Removed: %s\n", key)
		}

		err = client.SetEnvVars(projectCfg.ProjectID, current)
		if err != nil {
			fmt.Printf("❌ Failed to update env vars: %v\n", err)
			return
		}

		fmt.Println("✅ Environment variables updated")
		fmt.Println("🔄 Redeploy to apply: espacetech deploy --prod")
	},
}

func init() {
	envCmd.AddCommand(envSetCmd)
	envCmd.AddCommand(envListCmd)
	envCmd.AddCommand(envRemoveCmd)
	rootCmd.AddCommand(envCmd)
}