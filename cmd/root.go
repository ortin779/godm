package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "godm",
	Short: "godm - A Go Download Manager with concurrent downloads",
	Long: `A flexible download manager built in Go with support for concurrent downloads,
pause and resume functionality, and a command-line interface.`,
}

func init() {
	addCmd.Flags().StringVarP(&downloadDir, "dir", "d", "", "Directory to save the download")
	addCmd.Flags().StringVarP(&downloadName, "name", "n", "", "Name for the download")
	addCmd.Flags().IntVarP(&downloadParts, "parts", "p", 0, "Number of parts for parallel download")

	configCmd.Flags().StringVarP(&configPath, "path", "p", "", "Default download path")
	configCmd.Flags().IntVarP(&configParts, "parts", "n", 0, "Default number of parts")
	configCmd.Flags().IntVarP(&maxDownloads, "max_downloads", "m", 0, "Default max no of downloads")

	// Add subcommands
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(pauseCmd)
	rootCmd.AddCommand(resumeCmd)
	rootCmd.AddCommand(versionCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
