/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "micro",
	Short: "micro is a light-weight micro service framework",
	Long: `micro is a light-weight micro service framework,
and provides you with code generation and service discover capabilities.
Now I am working on to support more features to micro!  
`,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

// micro gen -I ./idl idl/service/user.proto
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
