/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// raw2scienceCmd represents the raw2science command
var raw2scienceCmd = &cobra.Command{
	Use:   "raw2science",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("raw2science called")

		for option := range sparkArgs {
			sparkArgs[option] = viper.GetString(option)
		}

		log.Printf("Args %v", sparkArgs)

		// TODO check error
		sparkArgs["bin"] = "raw2science.py"
		runSpark(sparkArgs)
	},
}

func init() {
	sparkCmd.AddCommand(raw2scienceCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// raw2scienceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// raw2scienceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
