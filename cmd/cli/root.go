package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "socle",
	Short: "Socle is a Go meta framework installer for building applications.",
	Long: `Socle CLI helps you scaffold and manage projects built on the Socle framework.

			Available Commands:
				new         				   - Create a new Socle project 

			Examples:
			socle new myapp 
			socle version
`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	color.Cyan("🚀 Ready to build with Socle!")
	cobra.OnInitialize(initConfig)

	rootCmd.AddCommand(newCmd)
}

func initConfig() {
	// future: config loading if needed
}

func exitGracefully(err error, msg ...string) {
	message := ""
	if len(msg) > 0 {
		message = msg[0]
	}

	if err != nil {
		color.Red("Error: %v\n", err)
	}

	if len(message) > 0 {
		color.Yellow(message)
	} else {
		color.Green("Finished!")
	}

	os.Exit(0)
}
