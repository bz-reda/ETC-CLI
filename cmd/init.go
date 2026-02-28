package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"paas-cli/internal/api"
	"paas-cli/internal/config"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

type ProjectConfig struct {
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Framework string `json:"framework"`
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new project in the current directory",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		if cfg.Token == "" {
			fmt.Println("❌ Please login first: espacetech login")
			return
		}

		// Check if already initialized
		if _, err := os.Stat(".espacetech.json"); err == nil {
			fmt.Println("⚠️  Project already initialized. Delete .espacetech.json to re-initialize.")
			return
		}

		// Project name
		namePrompt := promptui.Prompt{Label: "Project name"}
		name, _ := namePrompt.Run()

		// Framework
		frameworkSelect := promptui.Select{
			Label: "Framework",
			Items: []string{"nextjs"},
		}
		_, framework, _ := frameworkSelect.Run()

		client := api.NewClient(cfg)
		project, err := client.CreateProject(name, framework)
		if err != nil {
			fmt.Printf("❌ Failed to create project: %v\n", err)
			return
		}

		// Domain
		domainPrompt := promptui.Prompt{
			Label:   "Domain (e.g., mysite.com, leave empty to skip)",
			Default: "",
		}
		domain, _ := domainPrompt.Run()

		if domain != "" {
			if err := client.AddDomain(project.ID, domain); err != nil {
				fmt.Printf("⚠️  Failed to add domain: %v\n", err)
			} else {
				fmt.Printf("🌐 Domain %s added\n", domain)
			}
		}

		// Save project config
		projectCfg := ProjectConfig{
			ProjectID: project.ID,
			Name:      project.Name,
			Slug:      project.Slug,
			Framework: project.Framework,
		}
		data, _ := json.MarshalIndent(projectCfg, "", "  ")
		os.WriteFile(".espacetech.json", data, 0644)

		fmt.Printf("✅ Project '%s' created (slug: %s)\n", project.Name, project.Slug)
		fmt.Println("📁 Config saved to .espacetech.json")
		fmt.Println("\nNext: run 'espacetech deploy --prod' to deploy")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
