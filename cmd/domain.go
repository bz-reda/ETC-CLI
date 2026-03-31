package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"paas-cli/internal/api"
	"paas-cli/internal/config"

	"github.com/spf13/cobra"
)

var domainCmd = &cobra.Command{
	Use:   "domain",
	Short: "Manage project domains",
}

var domainAddCmd = &cobra.Command{
	Use:   "add [domain]",
	Short: "Add a custom domain to the current project",
	Args:  cobra.ExactArgs(1),
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

		var projectCfg ProjectConfig
		json.Unmarshal(data, &projectCfg)

		client := api.NewClient(cfg)
		domain := args[0]

		err = client.AddDomain(projectCfg.ProjectID, projectCfg.SiteID, domain)
		if err != nil {
			fmt.Printf("❌ Failed to add domain: %v\n", err)
			return
		}

		fmt.Printf("✅ Domain '%s' added to %s\n", domain, projectCfg.Name)
		fmt.Println("\n📋 Next steps:")
		fmt.Printf("   1. Add an A record in your DNS: %s → 65.109.68.181\n", domain)
		fmt.Printf("   2. Redeploy: espacetech deploy --prod\n")
		fmt.Printf("   3. SSL will be provisioned automatically\n")
	},
}

var domainListCmd = &cobra.Command{
	Use:   "list",
	Short: "List domains for the current project",
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

		var projectCfg ProjectConfig
		json.Unmarshal(data, &projectCfg)

		client := api.NewClient(cfg)
		domains, err := client.ListDomains(projectCfg.ProjectID)
		if err != nil {
			fmt.Printf("❌ Failed to list domains: %v\n", err)
			return
		}

		if len(domains) == 0 {
			fmt.Println("No domains configured.")
			return
		}

		fmt.Printf("🌐 Domains for %s:\n\n", projectCfg.Name)
		for _, d := range domains {
			fmt.Printf("   https://%s\n", d)
		}
	},
}

var domainRemoveCmd = &cobra.Command{
	Use:   "remove [domain]",
	Short: "Remove a domain from the current project",
	Args:  cobra.ExactArgs(1),
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
		}
		json.Unmarshal(data, &projectCfg)

		client := api.NewClient(cfg)
		err = client.RemoveDomain(projectCfg.ProjectID, args[0])
		if err != nil {
			fmt.Printf("❌ Failed to remove domain: %v\n", err)
			return
		}

		fmt.Printf("✅ Domain '%s' removed. Redeploy to apply changes.\n", args[0])
	},
}

func init() {
	domainCmd.AddCommand(domainAddCmd)
	domainCmd.AddCommand(domainListCmd)
	domainCmd.AddCommand(domainRemoveCmd)
	rootCmd.AddCommand(domainCmd)
}