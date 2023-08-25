/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
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
	UseSSL          bool   `mapstructure:"use_ssl"`
	BucketName      string `mapstructure:"bucket"`
}

// s3Cmd represents the s3 command
var s3Cmd = &cobra.Command{
	Use:   "s3",
	Short: "Display S3 storage parameters",
	Long:  `Display all S3 storage parameters for running fink-broker on Spark over Kubernetes`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
	},
	Run: func(cmd *cobra.Command, args []string) {
		logConfiguration()
	},
}

func init() {

	rootCmd.AddCommand(s3Cmd)

	s3Cmd.PersistentFlags().String("endpoint", "", "S3 service URL")
	viper.BindPFlag("s3.endpoint", s3Cmd.PersistentFlags().Lookup("endpoint"))

}

func getS3Config() S3Config {
	var c S3Config

	if err := viper.UnmarshalKey(S3, &c); err != nil {
		logger.Fatalf("Error while getting %s configuration: %v", S3, err)
	}

	// FIXME UnmarshalKey() does not seems to support correctly nested key management
	c.Endpoint = viper.GetString("s3.endpoint")

	return c
}
