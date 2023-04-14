/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const S3 string = "s3"

// endpoint := "localhost:9000"
// accessKeyID := "minioadmin"
// secretAccessKey := "minioadmin"
// useSSL := false

type S3Config struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"id"`
	SecretAccessKey string `mapstructure:"secret"`
	UseSSL          string `mapstructure:"use_ssl"`
	BucketName      string `mapstructure:"bucket"`
}

// s3Cmd represents the s3 command
var s3Cmd = &cobra.Command{
	Use:   "s3",
	Short: "Display S3 storage parameters",
	Long:  `Display all S3 storage parameters for running fink-broker on Spark over Kubernetes`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("s3 called")
	},
}

func init() {
	rootCmd.AddCommand(s3Cmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// s3Cmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// s3Cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func getS3Config() S3Config {
	var c S3Config

	if err := viper.UnmarshalKey(S3, &c); err != nil {
		logger.Fatalf("Error while getting %s configuration: %v", S3, err)
	}

	return c
}
