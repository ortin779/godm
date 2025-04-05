package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all downloads",
	RunE: func(cmd *cobra.Command, args []string) error {

		fmt.Println("Downloads:")
		fmt.Println("------------------------------------------------")
		fmt.Printf("%-20s %-12s %-10s %s\n", "Name", "Status", "Progress", "URL")
		fmt.Println("------------------------------------------------")

		return nil
	},
}
