package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"paas-cli/internal/api"
	"paas-cli/internal/config"

	"github.com/spf13/cobra"
)

var deployProd bool

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy the current project",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		if cfg.Token == "" {
			fmt.Println("❌ Please login first: espacetech login")
			return
		}

		// Read project config
		data, err := os.ReadFile(".espacetech.json")
		if err != nil {
			fmt.Println("❌ No project config found. Run 'espacetech init' first.")
			return
		}

		var projectCfg struct {
			ProjectID string `json:"project_id"`
			Name      string `json:"name"`
			Slug      string `json:"slug"`
		}
		json.Unmarshal(data, &projectCfg)

		fmt.Printf("🚀 Deploying %s...\n", projectCfg.Name)

		client := api.NewClient(cfg)

		// Get current directory as source
		cwd, _ := os.Getwd()

		resp, err := client.Deploy(projectCfg.ProjectID, cwd, "CLI deploy", deployProd)
		if err != nil {
			fmt.Printf("❌ Deploy failed: %v\n", err)
			return
		}

		fmt.Printf("📦 Build queued (deployment: %s)\n", resp.DeploymentID)
		fmt.Println("⏳ Waiting for build...")

		// Poll for status
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
