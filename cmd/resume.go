package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var resumeCmd = &cobra.Command{
	Use:   "resume [filename]",
	Short: "Resume a paused download",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filename := args[0]

		fmt.Printf("Download resumed: %s\n", filename)
		return nil
	},
}
