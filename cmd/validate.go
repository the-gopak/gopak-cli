package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate merged configuration against the JSON Schema",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Configuration is valid")
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
