package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ortin779/godm/internal/config"
	"github.com/ortin779/godm/internal/download"
	"github.com/spf13/cobra"
)

var (
	downloadDir   string
	downloadName  string
	downloadParts int
)

var addCmd = &cobra.Command{
	Use:   "add [url]",
	Short: "Add a new download",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]

		cfg, err := config.GetConfig()
		if err != nil {
			return err
		}

		fmt.Printf("Download added: %s\n", url)
		if downloadName == "" {
			downloadName = extractFileName(url)
		}
		if downloadDir == "" {
			downloadDir = cfg.DownloadDir
		}

		if downloadParts == 0 {
			downloadParts = cfg.DefaultParts
		}

		filePath := filepath.Join(downloadDir, downloadName)

		fmt.Println(filePath)

		downloadTask, err := download.NewTask(url, filePath, downloadParts)

		if err != nil {
			return err
		}

		return downloadTask.Start()
	},
}

func extractFileName(url string) string {
	parts := strings.Split(url, "/")
	filenamePart := parts[len(parts)-1]
	return filenamePart
}
