/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// uploadsourceCmd represents the uploadsource command
var makeBucketCmd = &cobra.Command{
	Use:     "makebucket",
	Aliases: []string{"mb"},
	Short:   "Create a S3 bucket used by Fink broker",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Create S3 bucket for Fink broker")
		mc := setMinioClient()
		listBucket(mc)
		makeBucket(mc)

	},
}

func init() {
	s3Cmd.AddCommand(makeBucketCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// uploadsourceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// uploadsourceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
