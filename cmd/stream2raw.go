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
	Aliases: []string{"s2"},
	Short:   "Launch 'stream2raw' job on Spark",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println("stream2raw called")

		sparkCmd := generateSparkCmd(STREAM2RAW)

		cmdTpl := sparkCmd + `-servers "{{ .KafkaSocket }}" \
    -schema "{{ .FinkAlertSchema }}" \
    -startingoffsets_stream "{{ .KafkaStartingOffset }}" \
    -topic "{{ .KafkaTopic }}"`
		c := getStream2RawConfig()
		sparkCmd = format(cmdTpl, &c)

		out, errout := ExecCmd(sparkCmd)
		outmsg := OutMsg{
			cmd:    sparkCmd,
			out:    out,
			errout: errout}
		log.Printf("message: %v\n", outmsg)
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
