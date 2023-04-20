/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const STREAM2RAW string = "stream2raw"

type Stream2RawConfig struct {
	KafkaSocket         string `mapstructure:"kafka_socket"`
	KafkaTopic          string `mapstructure:"kafka_topic"`
	FinkAlertSchema     string `mapstructure:"fink_alert_schema"`
	KafkaStartingOffset string `mapstructure:"kafka_starting_offset"`
}

// stream2rawCmd represents the stream2raw command
var stream2rawCmd = &cobra.Command{
	Use:     STREAM2RAW,
	Aliases: []string{"s2r"},
	Short:   "Launch Stream2raw service on Spark over Kubernetes",
	Long: `Launch Stream2raw service on Spark over Kubernetes. Stream2raw retrieves data from a Kafka stream
and writes it to a shared file system for further processing and analysis.`,
	Example: `  # Lauch stream2raw service on Spark, over Kubernetes
  finkctl spark stream2raw
  # Lauch stream2raw using a custom image
  finkctl spark stream2raw --image=gitlab-registry.in2p3.fr/astrolabsoftware/fink/fink-broker:2076184`,
	Run: func(cmd *cobra.Command, args []string) {

		startMsg := "Launch stream2raw service"
		logConfiguration()
		cmd.Printf(startMsg)
		logger.Info(startMsg)
		sparkCmd := generateSparkCmd(STREAM2RAW)

		cmdTpl := sparkCmd + `-servers "{{ .KafkaSocket }}" \
    -schema "{{ .FinkAlertSchema }}" \
    -startingoffsets_stream "{{ .KafkaStartingOffset }}" \
    -topic "{{ .KafkaTopic }}"`
		c := getStream2RawConfig()
		sparkCmd = format(cmdTpl, &c)

		ExecCmd(sparkCmd)

	},
}

func init() {
	sparkCmd.AddCommand(stream2rawCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// stream2rawCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// stream2rawCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func getStream2RawConfig() Stream2RawConfig {
	var c Stream2RawConfig
	if err := viper.UnmarshalKey(STREAM2RAW, &c); err != nil {
		log.Fatalf("Error while getting %s configuration: %v", STREAM2RAW, err)
	}

	return c
}
