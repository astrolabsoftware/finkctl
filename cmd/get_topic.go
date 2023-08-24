/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// topicCmd represents the topic command
var getTopicCmd = &cobra.Command{
	Use:     "topic",
	Aliases: []string{"to", "topics"},
	Short:   "List kafka topics produced by the fink-broker",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("List kafka topics produced by the fink-broker")
		topics := getFinkTopics()
		if len(topics) == 0 {
			logger.Info("No fink topics found")
		} else {
			fmt.Println(strings.Join(topics, ", "))
		}
	},
}

func init() {
	getCmd.AddCommand(getTopicCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// topicCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// topicCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
