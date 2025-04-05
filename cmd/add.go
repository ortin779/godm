package cmd

import (
	"fmt"

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

		fmt.Printf("Download added: %s\n", url)
		if downloadName != "" {
			fmt.Printf("Name: %s\n", downloadName)
		}
		return nil
	},
}
