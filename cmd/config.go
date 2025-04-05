package cmd

import (
	"fmt"

	"github.com/ortin779/godm/internal/config"
	"github.com/spf13/cobra"
)

var (
	configPath   string
	configParts  int
	maxDownloads int
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View or update configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if updating config
		if configPath != "" || configParts > 0 || maxDownloads > 0 {
			err := config.UpdateConfig(configPath, maxDownloads, configParts)
			if err != nil {
				return fmt.Errorf("failed to update configuration: %v", err)
			}
			fmt.Println("Configuration updated successfully.")
		}

		// Display current config
		cfg, err := config.GetConfig()
		if err != nil {
			return fmt.Errorf("failed to get configuration: %v", err)
		}
		fmt.Println("Current Configuration:")
		fmt.Printf("Default download path: %s\n", cfg.DownloadDir)
		fmt.Printf("Default parts: %d\n", cfg.DefaultParts)
		fmt.Printf("Default Max downloads: %d\n", cfg.MaxDownloads)

		return nil
	},
}
