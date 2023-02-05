/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var image string

// sparkCmd represents the spark command
var sparkCmd = &cobra.Command{
	Use:   "spark",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("spark called")

	},
}

func init() {
	rootCmd.AddCommand(sparkCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// sparkCmd.PersistentFlags().String("foo", "", "A help for foo")

	viper.AutomaticEnv()
	option := "image"
	sparkCmd.PersistentFlags().String(option, "", "fink-broker image name")
	viper.BindPFlag(option, sparkCmd.PersistentFlags().Lookup(option))
	option = "producer"
	sparkCmd.PersistentFlags().String(option, "", "fink-broker image name")
	viper.BindPFlag(option, sparkCmd.PersistentFlags().Lookup(option))
	option = "kafka_socket"
	sparkCmd.PersistentFlags().String(option, "", "fink-broker image name")
	viper.BindPFlag(option, sparkCmd.PersistentFlags().Lookup(option))
	option = "kafka_topic"
	sparkCmd.PersistentFlags().String(option, "", "fink-broker image name")
	viper.BindPFlag(option, sparkCmd.PersistentFlags().Lookup(option))
	option = "fink_alert_schema"
	sparkCmd.PersistentFlags().String(option, "", "fink-broker image name")
	viper.BindPFlag(option, sparkCmd.PersistentFlags().Lookup(option))
	option = "kafka_starting_offset"
	sparkCmd.PersistentFlags().String(option, "", "fink-broker image name")
	viper.BindPFlag(option, sparkCmd.PersistentFlags().Lookup(option))
	option = "online_data_prefix"
	sparkCmd.PersistentFlags().String(option, "", "fink-broker image name")
	viper.BindPFlag(option, sparkCmd.PersistentFlags().Lookup(option))
	option = "fink_trigger_update"
	sparkCmd.PersistentFlags().String(option, "", "fink-broker image name")
	viper.BindPFlag(option, sparkCmd.PersistentFlags().Lookup(option))
	option = "log_level"
	sparkCmd.PersistentFlags().String(option, "", "fink-broker image name")
	viper.BindPFlag(option, sparkCmd.PersistentFlags().Lookup(option))

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// sparkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
