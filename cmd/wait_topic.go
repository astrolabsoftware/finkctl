/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var expected int

// topicCmd represents the topic command
var waitTopicCmd = &cobra.Command{
	Use:     "topic",
	Aliases: []string{"to", "topics"},
	Short:   "Wait for kafka topics produced by the fink-broker",
	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()
		slog.Info("Wait for kafka topics produced by the fink-broker")

		// Channel to signal when the condition is met
		allTopicsFound := make(chan bool, 1)

		// Start a goroutine to check the condition
		go func() {
			for {
				topics, err := getFinkTopics()
				cobra.CheckErr(err)
				t := time.Now()
				elapsed := t.Sub(start)
				slog.Info("Found topics", "elapsedTime", elapsed.Round(1), "topics", topics, "topicsCount", len(topics), "expected", expected)
				if len(topics) == expected {
					allTopicsFound <- true
					return
				}
				time.Sleep(time.Second * 10) // Adjust the sleep duration as needed
			}
		}()

		select {
		case <-allTopicsFound:
			fmt.Printf("Condition met: Found %d fink topics\n", expected)
		case <-time.After(timeout):
			fmt.Fprintf(os.Stderr, "error: timed out waiting for the condition on topics\n")
			os.Exit(1)
		}

	},
}

func init() {
	waitCmd.AddCommand(waitTopicCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// topicCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	waitTopicCmd.Flags().IntVarP(&expected, "expected", "e", 10, "Number of expected topics")
}
