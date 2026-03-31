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

// localConfig reads .espacetech.json and returns projectID + siteID (siteID may be empty for old configs).
func localConfig() (projectID, siteID, name string, err error) {
	data, readErr := os.ReadFile(".espacetech.json")
	if readErr != nil {
		return "", "", "", fmt.Errorf("no project config found — run 'espacetech init' first")
	}
	var cfg struct {
		ProjectID string `json:"project_id"`
		Name      string `json:"name"`
		SiteID    string `json:"site_id"`
	}
	json.Unmarshal(data, &cfg)
	return cfg.ProjectID, cfg.SiteID, cfg.Name, nil
}

// getEnvVars fetches env vars using site-scoped endpoint when siteID is known,
// falling back to the legacy project endpoint for old .espacetech.json files.
func getEnvVars(client *api.Client, projectID, siteID string) (map[string]string, error) {
	if siteID != "" {
		return client.GetEnvVarsBySite(projectID, siteID)
	}
	return client.GetEnvVars(projectID)
}

// setEnvVars saves env vars using site-scoped endpoint when siteID is known.
func setEnvVars(client *api.Client, projectID, siteID string, vars map[string]string) error {
	if siteID != "" {
		return client.SetEnvVarsBySite(projectID, siteID, vars)
	}
	return client.SetEnvVars(projectID, vars)
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

		projectID, siteID, _, err := localConfig()
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}

		client := api.NewClient(cfg)

		current, err := getEnvVars(client, projectID, siteID)
		if err != nil {
			current = make(map[string]string)
		}

		for _, arg := range args {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) != 2 {
				fmt.Printf("❌ Invalid format: %s (use KEY=VALUE)\n", arg)
				return
			}
			current[parts[0]] = parts[1]
		}

		if err := setEnvVars(client, projectID, siteID, current); err != nil {
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

		projectID, siteID, name, err := localConfig()
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}

		client := api.NewClient(cfg)
		envVars, err := getEnvVars(client, projectID, siteID)
		if err != nil {
			fmt.Printf("❌ Failed to get env vars: %v\n", err)
			return
		}

		if len(envVars) == 0 {
			fmt.Println("No environment variables set.")
			return
		}

		fmt.Printf("🔧 Environment variables for %s:\n\n", name)
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

		projectID, siteID, _, err := localConfig()
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}

		client := api.NewClient(cfg)
		current, err := getEnvVars(client, projectID, siteID)
		if err != nil {
			fmt.Printf("❌ Failed to get env vars: %v\n", err)
			return
		}

		for _, key := range args {
			delete(current, key)
			fmt.Printf("   Removed: %s\n", key)
		}

		if err := setEnvVars(client, projectID, siteID, current); err != nil {
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
