package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"paas-cli/internal/api"
	"paas-cli/internal/config"

	"github.com/spf13/cobra"
)

var deployProd bool

type projectConfig struct {
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	SiteID    string `json:"site_id,omitempty"`
	SiteName  string `json:"site_name,omitempty"`
	SiteSlug  string `json:"site_slug,omitempty"`
}

type appChoice struct {
	Config projectConfig
	Dir    string // absolute path to app directory
	RelDir string // relative path from monorepo root
}

// findMonorepoRoot walks up from dir looking for turbo.json
func findMonorepoRoot(dir string) string {
	current := dir
	for {
		if _, err := os.Stat(filepath.Join(current, "turbo.json")); err == nil {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			return ""
		}
		current = parent
	}
}

// findInitializedApps scans a monorepo for apps that have .espacetech.json
func findInitializedApps(root string) []appChoice {
	var apps []appChoice
	skipDirs := map[string]bool{
		"node_modules": true, ".git": true, ".next": true,
		".turbo": true, "dist": true, ".cache": true,
	}

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && skipDirs[info.Name()] {
			return filepath.SkipDir
		}
		// Limit search depth
		relPath, _ := filepath.Rel(root, path)
		if info.IsDir() && len(strings.Split(relPath, string(filepath.Separator))) > 4 {
			return filepath.SkipDir
		}
		if info.Name() == ".espacetech.json" {
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return nil
			}
			var cfg projectConfig
			json.Unmarshal(data, &cfg)

			appDir := filepath.Dir(path)
			rel, _ := filepath.Rel(root, appDir)

			apps = append(apps, appChoice{
				Config: cfg,
				Dir:    appDir,
				RelDir: rel,
			})
		}
		return nil
	})
	return apps
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy the current project",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		if cfg.Token == "" {
			fmt.Println("❌ Please login first: espacetech login")
			return
		}

		cwd, _ := os.Getwd()
		var projCfg projectConfig
		var appDir string

		// Try to read .espacetech.json from CWD
		data, err := os.ReadFile(filepath.Join(cwd, ".espacetech.json"))
		if err != nil {
			// No config in CWD - check if inside a monorepo
			monorepoRoot := findMonorepoRoot(cwd)
			if monorepoRoot == "" {
				fmt.Println("❌ No project config found. Run 'espacetech init' first.")
				return
			}

			apps := findInitializedApps(monorepoRoot)
			if len(apps) == 0 {
				fmt.Println("❌ No initialized apps found in this monorepo.")
				fmt.Println("   Navigate to your app directory and run 'espacetech init' first.")
				return
			}

			var selected appChoice
			if len(apps) == 1 {
				selected = apps[0]
				fmt.Printf("📦 Found app: %s (%s)\n", selected.Config.Name, selected.RelDir)
			} else {
				fmt.Println("📦 Multiple apps found in this monorepo:")
				for i, app := range apps {
					fmt.Printf("  %d) %s (%s)\n", i+1, app.Config.Name, app.RelDir)
				}
				fmt.Print("\nSelect app to deploy: ")
				reader := bufio.NewReader(os.Stdin)
				input, _ := reader.ReadString('\n')
				input = strings.TrimSpace(input)
				choice, parseErr := strconv.Atoi(input)
				if parseErr != nil || choice < 1 || choice > len(apps) {
					fmt.Println("❌ Invalid selection")
					return
				}
				selected = apps[choice-1]
			}

			projCfg = selected.Config
			appDir = selected.Dir
		} else {
			json.Unmarshal(data, &projCfg)
			appDir = cwd
		}

		// Determine source directory and root_directory
		sourceDir := appDir
		rootDirectory := ""

		monorepoRoot := findMonorepoRoot(appDir)
		if monorepoRoot != "" && monorepoRoot != appDir {
			relPath, _ := filepath.Rel(monorepoRoot, appDir)
			rootDirectory = relPath
			sourceDir = monorepoRoot
			fmt.Printf("🚀 Deploying %s (monorepo: %s)...\n", projCfg.Name, rootDirectory)
		} else {
			siteLabel := ""
			if projCfg.SiteName != "" && projCfg.SiteName != "main" {
				siteLabel = fmt.Sprintf(" [site: %s]", projCfg.SiteName)
			}
			fmt.Printf("🚀 Deploying %s%s...\n", projCfg.Name, siteLabel)
		}

		client := api.NewClient(cfg)

		resp, err := client.Deploy(projCfg.ProjectID, projCfg.SiteID, sourceDir, "CLI deploy", deployProd, rootDirectory)
		if err != nil {
			fmt.Printf("❌ Deploy failed: %v\n", err)
			return
		}

		fmt.Printf("📦 Build queued (deployment: %s)\n", resp.DeploymentID)
		fmt.Println("⏳ Waiting for build...")

		for i := 0; i < 120; i++ {
			time.Sleep(3 * time.Second)

			deployment, err := client.GetDeployment(resp.DeploymentID)
			if err != nil {
				continue
			}

			switch deployment.Status {
			case "live":
				fmt.Println("\n✅ Deployed successfully!")
				if len(deployment.Domains) > 0 {
					fmt.Println("🌐 Your app is live at:")
					for _, d := range deployment.Domains {
						fmt.Printf("   https://%s\n", d)
					}
				} else {
					fmt.Println("🌐 Your app is live")
				}
				return
			case "failed":
				fmt.Println("\n❌ Deployment failed!")
				logs, _ := client.GetDeploymentLogs(resp.DeploymentID)
				if logs != "" {
					fmt.Println("\n📋 Build logs:")
					fmt.Println(logs)
				}
				return
			case "building":
				fmt.Print(".")
			case "deploying":
				fmt.Print("🔄")
			}
		}

		fmt.Println("\n⚠️  Deploy timed out. Check status with: espacetech status")
	},
}

func init() {
	deployCmd.Flags().BoolVarP(&deployProd, "prod", "p", false, "Deploy to production")
	rootCmd.AddCommand(deployCmd)
}
