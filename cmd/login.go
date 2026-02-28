package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"paas-cli/internal/api"
	"paas-cli/internal/config"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Espace-Tech Cloud",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		client := api.NewClient(cfg)

		useEmail, _ := cmd.Flags().GetBool("email")

		if useEmail {
			emailPrompt := promptui.Prompt{Label: "Email"}
			email, _ := emailPrompt.Run()

			passPrompt := promptui.Prompt{Label: "Password", Mask: '*'}
			password, _ := passPrompt.Run()

			resp, err := client.Login(email, password)
			if err != nil {
				fmt.Printf("❌ Login failed: %v\n", err)
				return
			}

			cfg.Token = resp.Token
			cfg.APIToken = resp.APIToken
			cfg.UserID = resp.User.ID
			cfg.Email = resp.User.Email
			cfg.Save()

			fmt.Printf("✅ Logged in as %s (%s)\n", resp.User.Name, resp.User.Email)
			return
		}

		// Default: browser login
		browserLogin(cfg, client)
	},
}

func browserLogin(cfg *config.Config, client *api.Client) {
	fmt.Println("🔐 Opening browser for login...")

	resp, err := http.Post(cfg.APIHost+"/api/v1/auth/cli/session", "application/json", nil)
	if err != nil {
		fmt.Printf("❌ Failed to create login session: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var session struct {
		Code     string `json:"code"`
		LoginURL string `json:"login_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		fmt.Printf("❌ Failed to parse response: %v\n", err)
		return
	}

	openBrowser(session.LoginURL)
	fmt.Printf("\n   If browser didn't open, visit:\n   %s\n\n", session.LoginURL)
	fmt.Println("⏳ Waiting for approval in browser... (Ctrl+C to cancel)")

	for i := 0; i < 60; i++ {
		time.Sleep(5 * time.Second)

		pollResp, err := http.Get(cfg.APIHost + "/api/v1/auth/cli/session/" + session.Code)
		if err != nil {
			continue
		}

		var result struct {
			Confirmed bool   `json:"confirmed"`
			Token     string `json:"token"`
			Email     string `json:"email"`
			Error     string `json:"error"`
		}
		json.NewDecoder(pollResp.Body).Decode(&result)
		pollResp.Body.Close()

		if pollResp.StatusCode == http.StatusGone || pollResp.StatusCode == http.StatusNotFound {
			fmt.Println("❌ Login session expired. Please try again.")
			return
		}

		if result.Confirmed && result.Token != "" {
			cfg.Token = result.Token
			cfg.Email = result.Email

			me, err := client.GetMe(result.Token)
			if err == nil {
				cfg.UserID = me.ID
				cfg.Email = me.Email
			}

			cfg.Save()
			fmt.Printf("✅ Logged in as %s\n", cfg.Email)
			return
		}
	}

	fmt.Println("❌ Login timed out. Please try again.")
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	}
	if cmd != nil {
		cmd.Start()
	}
}

func init() {
	loginCmd.Flags().Bool("email", false, "Login with email/password instead of browser")
	rootCmd.AddCommand(loginCmd)
}
