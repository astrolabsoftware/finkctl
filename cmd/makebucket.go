/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// makeBucketCmd represents the makeBucket command
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
		c := getS3Config()
		logger.Debugf("S3 endpoint: %s, bucket name: %s", c.Endpoint, c.BucketName)
		mc := setMinioClient(c)
		listBucket(mc)
		if !bucketExists(mc, c.BucketName) {
			makeBucket(mc, c.BucketName)
		} else {
			logger.Warnf("Bucket exists: %s", c.BucketName)
		}
	},
}

func init() {
	s3Cmd.AddCommand(makeBucketCmd)
}
