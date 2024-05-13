/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

var c DistributionConfig

// createSecretCmd represents the createsecrets command
var createSecretCmd = &cobra.Command{
	Use:     "createsecrets",
	Aliases: []string{"cs"},
	Short:   "Create/update secrets used by Fink broker",
	Run: func(cmd *cobra.Command, args []string) {

		if c.KafkaCreds.Password == "" {
			c.KafkaCreds.Password = getKafkaPasswordFromSecret()
		}
		createKafkaJaasSecret(&c, false)
	},
}

func init() {
	rootCmd.AddCommand(createSecretCmd)
	createSecretCmd.PersistentFlags().StringVarP(&c.KafkaCreds.Username, "kafka-username", "k", "fink-producer", "Kafka username used by the Fink broker")
	createSecretCmd.PersistentFlags().StringVarP(&c.KafkaCreds.Password, "kafka-password", "p", "", "Kafka password used by the Fink broker, if not provided, it will be fetched from the kafka secret related to the kafkauser in the kafka namespace")
}
