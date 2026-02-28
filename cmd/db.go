package cmd

import (
	"fmt"
	"os"
	"encoding/json"

	"paas-cli/internal/api"
	"paas-cli/internal/config"

	"github.com/spf13/cobra"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Manage databases",
}

var dbCreateType string

var dbCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a managed database",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		if cfg.Token == "" {
			fmt.Println("❌ Please login first: espacetech login")
			return
		}

		client := api.NewClient(cfg)
		db, err := client.CreateDatabase(args[0], dbCreateType)
		if err != nil {
			fmt.Printf("❌ Failed to create database: %v\n", err)
			return
		}

		fmt.Printf("✅ Created %s database '%s'\n", db.Type, db.Name)
		fmt.Printf("   ID:      %s\n", db.ID)
		fmt.Printf("   Host:    %s\n", db.Host)
		fmt.Printf("   Port:    %d\n", db.Port)
		fmt.Printf("   Status:  %s\n", db.Status)
		fmt.Println("\n📋 Next: link to a project with:")
		fmt.Printf("   espacetech db link %s --project <project-name>\n", args[0])
	},
}

var dbListCmd = &cobra.Command{
	Use:   "list",
	Short: "List your databases",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		if cfg.Token == "" {
			fmt.Println("❌ Please login first: espacetech login")
			return
		}

		client := api.NewClient(cfg)
		databases, err := client.ListDatabases()
		if err != nil {
			fmt.Printf("❌ Failed to list databases: %v\n", err)
			return
		}

		if len(databases) == 0 {
			fmt.Println("No databases found. Create one with: espacetech db create <name> --type postgres")
			return
		}

		fmt.Printf("🗄️  Your databases (%d):\n\n", len(databases))
		for _, db := range databases {
			linked := "unlinked"
			if db.ProjectID != "" {
				linked = "linked"
			}
			fmt.Printf("   %-15s  %-10s  %-10s  %s\n", db.Name, db.Type, db.Status, linked)
		}
	},
}

var dbInfoCmd = &cobra.Command{
	Use:   "info [name]",
	Short: "Show database details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		if cfg.Token == "" {
			fmt.Println("❌ Please login first: espacetech login")
			return
		}

		client := api.NewClient(cfg)
		db, err := findDatabaseByName(client, args[0])
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}

		fmt.Printf("🗄️  Database: %s\n\n", db.Name)
		fmt.Printf("   Type:       %s %s\n", db.Type, db.Version)
		fmt.Printf("   Status:     %s\n", db.Status)
		fmt.Printf("   Host:       %s\n", db.Host)
		fmt.Printf("   Port:       %d\n", db.Port)
		if db.DBName != "" {
			fmt.Printf("   Database:   %s\n", db.DBName)
			fmt.Printf("   Username:   %s\n", db.Username)
		}
		fmt.Printf("   Storage:    %d MB\n", db.StorageMB)
		fmt.Printf("   CPU:        %s\n", db.CPULimit)
		fmt.Printf("   Memory:     %s\n", db.MemoryLimit)
		if db.ProjectID != "" {
			fmt.Printf("   Linked to:  %s\n", db.ProjectID)
		} else {
			fmt.Printf("   Linked to:  (none)\n")
		}
	},
}

var dbLinkProject string

var dbLinkCmd = &cobra.Command{
	Use:   "link [db-name]",
	Short: "Link a database to a project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		if cfg.Token == "" {
			fmt.Println("❌ Please login first: espacetech login")
			return
		}

		client := api.NewClient(cfg)

		// Find database by name
		db, err := findDatabaseByName(client, args[0])
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}

		// Find project by name/slug or from .espacetech.json
		projectID := ""
		if dbLinkProject != "" {
			projects, err := client.ListProjects()
			if err != nil {
				fmt.Printf("❌ Failed to list projects: %v\n", err)
				return
			}
			for _, p := range projects {
				if p.Name == dbLinkProject || p.Slug == dbLinkProject {
					projectID = p.ID
					break
				}
			}
			if projectID == "" {
				fmt.Printf("❌ Project '%s' not found\n", dbLinkProject)
				return
			}
		} else {
			data, err := os.ReadFile(".espacetech.json")
			if err != nil {
				fmt.Println("❌ No --project flag and no .espacetech.json found")
				return
			}
			var projectCfg struct {
				ProjectID string `json:"project_id"`
			}
			json.Unmarshal(data, &projectCfg)
			projectID = projectCfg.ProjectID
		}

		err = client.LinkDatabase(db.ID, projectID)
		if err != nil {
			fmt.Printf("❌ Failed to link: %v\n", err)
			return
		}

		fmt.Printf("✅ Database '%s' linked to project. Connection string injected as env var.\n", args[0])
	},
}

var dbUnlinkCmd = &cobra.Command{
	Use:   "unlink [db-name]",
	Short: "Unlink a database from its project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		if cfg.Token == "" {
			fmt.Println("❌ Please login first: espacetech login")
			return
		}

		client := api.NewClient(cfg)
		db, err := findDatabaseByName(client, args[0])
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}

		err = client.UnlinkDatabase(db.ID)
		if err != nil {
			fmt.Printf("❌ Failed to unlink: %v\n", err)
			return
		}

		fmt.Printf("✅ Database '%s' unlinked. Env var removed from project.\n", args[0])
	},
}

var dbDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a database and all its data",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		if cfg.Token == "" {
			fmt.Println("❌ Please login first: espacetech login")
			return
		}

		client := api.NewClient(cfg)
		db, err := findDatabaseByName(client, args[0])
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}

		fmt.Printf("⚠️  This will permanently delete '%s' and all its data. Type the database name to confirm: ", args[0])
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != args[0] {
			fmt.Println("❌ Cancelled.")
			return
		}

		err = client.DeleteDatabase(db.ID)
		if err != nil {
			fmt.Printf("❌ Failed to delete: %v\n", err)
			return
		}

		fmt.Printf("✅ Database '%s' deleted.\n", args[0])
	},
}

// findDatabaseByName looks up a database by name from the user's list
func findDatabaseByName(client *api.Client, name string) (*api.DatabaseInfo, error) {
	databases, err := client.ListDatabases()
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %v", err)
	}
	for _, db := range databases {
		if db.Name == name {
			return &db, nil
		}
	}
	return nil, fmt.Errorf("database '%s' not found", name)
}

var dbExposeCmd = &cobra.Command{
	Use:   "expose [name]",
	Short: "Enable external access to a database",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		if cfg.Token == "" {
			fmt.Println("❌ Please login first: espacetech login")
			return
		}

		client := api.NewClient(cfg)
		db, err := findDatabaseByName(client, args[0])
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}

		result, err := client.ExposeDatabase(db.ID)
		if err != nil {
			fmt.Printf("❌ Failed to expose: %v\n", err)
			return
		}

		fmt.Printf("✅ Database '%s' exposed externally\n", args[0])
		fmt.Printf("   Host: %v\n", result["external_host"])
		fmt.Printf("   Port: %v\n", result["external_port"])
		fmt.Printf("   Connection: %v\n", result["connection"])
		fmt.Println("\n📋 Get full credentials with:")
		fmt.Printf("   espacetech db credentials %s\n", args[0])
	},
}

var dbUnexposeCmd = &cobra.Command{
	Use:   "unexpose [name]",
	Short: "Disable external access to a database",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		if cfg.Token == "" {
			fmt.Println("❌ Please login first: espacetech login")
			return
		}

		client := api.NewClient(cfg)
		db, err := findDatabaseByName(client, args[0])
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}

		err = client.UnexposeDatabase(db.ID)
		if err != nil {
			fmt.Printf("❌ Failed to unexpose: %v\n", err)
			return
		}

		fmt.Printf("✅ External access disabled for '%s'\n", args[0])
	},
}

var dbCredentialsCmd = &cobra.Command{
	Use:   "credentials [name]",
	Short: "Show database connection credentials",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		if cfg.Token == "" {
			fmt.Println("❌ Please login first: espacetech login")
			return
		}

		client := api.NewClient(cfg)
		db, err := findDatabaseByName(client, args[0])
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}

		creds, err := client.GetDatabaseCredentials(db.ID)
		if err != nil {
			fmt.Printf("❌ Failed to get credentials: %v\n", err)
			return
		}

		fmt.Printf("🔑 Credentials for '%s' (%s)\n\n", args[0], creds["type"])
		fmt.Printf("   Host:     %v\n", creds["host"])
		fmt.Printf("   Port:     %v\n", creds["port"])
		if creds["username"] != nil && creds["username"] != "" {
			fmt.Printf("   Username: %v\n", creds["username"])
		}
		if creds["database"] != nil && creds["database"] != "" {
			fmt.Printf("   Database: %v\n", creds["database"])
		}
		fmt.Printf("   Password: %v\n", creds["password"])
		fmt.Printf("\n   Internal URL: %v\n", creds["internal_url"])

		if creds["external_access"] == true {
			fmt.Printf("   External URL: %v\n", creds["external_url"])
		} else {
			fmt.Println("\n   ℹ️  External access is off. Enable with:")
			fmt.Printf("      espacetech db expose %s\n", args[0])
		}
	},
}

var dbStopCmd = &cobra.Command{
	Use:   "stop [name]",
	Short: "Stop a database (preserves data)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		if cfg.Token == "" {
			fmt.Println("❌ Please login first: espacetech login")
			return
		}

		client := api.NewClient(cfg)
		db, err := findDatabaseByName(client, args[0])
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}

		err = client.StopDatabase(db.ID)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}

		fmt.Printf("✅ Database '%s' stopped. Data is preserved.\n", args[0])
		fmt.Printf("   Restart with: espacetech db start %s\n", args[0])
	},
}

var dbStartCmd = &cobra.Command{
	Use:   "start [name]",
	Short: "Start a stopped database",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		if cfg.Token == "" {
			fmt.Println("❌ Please login first: espacetech login")
			return
		}

		client := api.NewClient(cfg)
		db, err := findDatabaseByName(client, args[0])
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}

		err = client.StartDatabase(db.ID)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}

		fmt.Printf("✅ Database '%s' started.\n", args[0])
	},
}

var dbRotateCmd = &cobra.Command{
	Use:   "rotate [name]",
	Short: "Rotate database password",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		if cfg.Token == "" {
			fmt.Println("❌ Please login first: espacetech login")
			return
		}

		client := api.NewClient(cfg)
		db, err := findDatabaseByName(client, args[0])
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}

		fmt.Printf("⚠️  This will change the password for '%s'. Connected clients will need to reconnect. Continue? (y/n): ", args[0])
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "y" && confirm != "Y" {
			fmt.Println("❌ Cancelled.")
			return
		}

		result, err := client.RotatePassword(db.ID)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}

		fmt.Printf("✅ Password rotated for '%s'\n", args[0])
		fmt.Printf("   New password: %v\n", result["new_password"])
		fmt.Println("\n   ⚠️  Save this password now — it won't be shown again.")
		fmt.Printf("   Get full connection string: espacetech db credentials %s\n", args[0])
	},
}

func init() {
	dbCreateCmd.Flags().StringVarP(&dbCreateType, "type", "t", "postgres", "Database type: postgres, redis, mongodb")
	dbLinkCmd.Flags().StringVarP(&dbLinkProject, "project", "p", "", "Project name or slug (defaults to current directory's project)")

	dbCmd.AddCommand(dbCreateCmd)
	dbCmd.AddCommand(dbListCmd)
	dbCmd.AddCommand(dbInfoCmd)
	dbCmd.AddCommand(dbLinkCmd)
	dbCmd.AddCommand(dbUnlinkCmd)
	dbCmd.AddCommand(dbDeleteCmd)
	rootCmd.AddCommand(dbCmd)
	dbCmd.AddCommand(dbExposeCmd)
	dbCmd.AddCommand(dbUnexposeCmd)
	dbCmd.AddCommand(dbCredentialsCmd)
	dbCmd.AddCommand(dbStopCmd)
	dbCmd.AddCommand(dbStartCmd)
	dbCmd.AddCommand(dbRotateCmd)
}