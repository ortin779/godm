package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version number of godm",
	Long:  `The current installed version of the godm software`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("v0.1")
	},
}
