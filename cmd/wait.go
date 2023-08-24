/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
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
		fmt.Println("wait called")
	},
}

func init() {
	rootCmd.AddCommand(waitCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	waitCmd.PersistentFlags().DurationVar(&timeout, "timeout", 10, "Timeout in seconds")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// waitCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
