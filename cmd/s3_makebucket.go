/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

// makeBucketCmd represents the makeBucket command
var makeBucketCmd = &cobra.Command{
	Use:     "makebucket",
	Aliases: []string{"mb"},
	Short:   "Create a S3 bucket used by Fink broker",
	Run: func(cmd *cobra.Command, args []string) {
		logConfiguration()
		fmt.Println("Create S3 bucket for Fink broker")
		rc := getRunConfig("")
		c := getS3Config(rc.Night)
		slog.Debug("S3", "endpoint", c.Endpoint, "bucketName", c.BucketName)
		mc := setMinioClient(c)
		listBucket(mc)
		if !bucketExists(mc, c.BucketName) {
			makeBucket(mc, c.BucketName)
		} else {
			slog.Warn("Bucket exists", "bucketName", c.BucketName)
		}
	},
}

func init() {
	s3Cmd.AddCommand(makeBucketCmd)
}
