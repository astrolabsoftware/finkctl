/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var timeout time.Duration

// waitCmd represents the wait command
var waitCmd = &cobra.Command{
	Use:   "wait",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(os.Stderr, "You must specify the type of resource to get."+
			"Use \"finkctl get --help\" for a complete list of supported resource")
		os.Exit(1)
	},
}

func init() {
	rootCmd.AddCommand(waitCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	waitCmd.PersistentFlags().DurationVar(&timeout, "timeout", 10*time.Second, "Timeout in seconds")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// waitCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
