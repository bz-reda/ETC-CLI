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

var linkCmd = &cobra.Command{
	Use:   "link [project-slug]",
	Short: "Link the current directory to an existing project",
	Long: `Link the current directory to a project that already exists on Espace-Tech Cloud.

Use this instead of 'init' when you clone a repository on a new machine:
init creates a brand-new project, while link connects to one you already own.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		if cfg.Token == "" {
			fmt.Println("❌ Please login first: espacetech login")
			return
		}

		if _, err := os.Stat(".espacetech.json"); err == nil {
			fmt.Println("⚠️  This directory is already linked. Delete .espacetech.json to re-link.")
			return
		}

		client := api.NewClient(cfg)

		projects, err := client.ListProjects()
		if err != nil {
			fmt.Printf("❌ Failed to list projects: %v\n", err)
			return
		}
		if len(projects) == 0 {
			fmt.Println("❌ You don't have any projects yet. Create one with: espacetech init")
			return
		}

		var project *api.Project

		if len(args) == 1 {
			target := args[0]
			for i, p := range projects {
				if p.Slug == target || p.Name == target || p.ID == target {
					project = &projects[i]
					break
				}
			}
			if project == nil {
				fmt.Printf("❌ No project found matching '%s'\n", target)
				fmt.Println("   Run 'espacetech link' without arguments to pick from a list.")
				return
			}
		} else {
			labels := make([]string, len(projects))
			for i, p := range projects {
				labels[i] = fmt.Sprintf("%s  (slug: %s, framework: %s)", p.Name, p.Slug, p.Framework)
			}
			sel := promptui.Select{
				Label: "Select a project to link",
				Items: labels,
				Size:  10,
			}
			idx, _, err := sel.Run()
			if err != nil {
				fmt.Println("❌ Cancelled")
				return
			}
			project = &projects[idx]
		}

		sites, err := client.ListSites(project.ID)
		if err != nil {
			fmt.Printf("❌ Failed to list sites: %v\n", err)
			return
		}
		if len(sites) == 0 {
			fmt.Println("❌ This project has no sites. Create one with: espacetech site add <name>")
			return
		}

		var site *api.Site
		if len(sites) == 1 {
			site = &sites[0]
		} else {
			labels := make([]string, len(sites))
			for i, s := range sites {
				labels[i] = fmt.Sprintf("%s  (slug: %s, status: %s)", s.Name, s.Slug, s.Status)
			}
			sel := promptui.Select{
				Label: "Select a site",
				Items: labels,
				Size:  10,
			}
			idx, _, err := sel.Run()
			if err != nil {
				fmt.Println("❌ Cancelled")
				return
			}
			site = &sites[idx]
		}

		projectCfg := ProjectConfig{
			ProjectID: project.ID,
			Name:      project.Name,
			Slug:      project.Slug,
			Framework: project.Framework,
			SiteID:    site.ID,
			SiteName:  site.Name,
			SiteSlug:  site.Slug,
		}
		data, _ := json.MarshalIndent(projectCfg, "", "  ")
		if err := os.WriteFile(".espacetech.json", data, 0644); err != nil {
			fmt.Printf("❌ Failed to write .espacetech.json: %v\n", err)
			return
		}

		fmt.Printf("✅ Linked to project '%s' (slug: %s)\n", project.Name, project.Slug)
		fmt.Printf("   Site: %s (slug: %s)\n", site.Name, site.Slug)
		fmt.Println("📁 Config saved to .espacetech.json")
		fmt.Println("\nNext: run 'espacetech deploy --prod' to deploy")
	},
}

func init() {
	rootCmd.AddCommand(linkCmd)
}
