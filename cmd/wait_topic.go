/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	expectedFile  string
	withoutGroups []string
)

// topicCmd represents the topic command
var waitTopicCmd = &cobra.Command{
	Use:     "topic",
	Aliases: []string{"to", "topics"},
	Short:   "Wait for the kafka topics produced by the fink-broker, by name",
	Long: `Wait until every topic of an expected-topics file exists.

The file holds named groups of topic names. The "topics" group is always
required; any other group is required unless excluded with --without, e.g.
--without cds while the CDS xmatch service is down. On timeout, the missing
topics are listed. Topics found but not expected are logged, not an error.`,
	Run: func(cmd *cobra.Command, args []string) {
		groups, err := loadExpectedTopics(expectedFile)
		cobra.CheckErr(err)
		required, err := requiredTopics(groups, withoutGroups)
		cobra.CheckErr(err)

		start := time.Now()
		slog.Info("Wait for kafka topics produced by the fink-broker", "expected", len(required))
		deadline := time.After(timeout)
		for {
			topics, err := getFinkTopics()
			cobra.CheckErr(err)
			missing, extra := compareTopics(required, topics)
			slog.Info("Found topics", "elapsedTime", time.Since(start).Round(time.Second),
				"found", len(topics), "missing", missing)
			if len(extra) > 0 {
				slog.Info("Topics found but not expected", "topics", extra)
			}
			if len(missing) == 0 {
				fmt.Printf("Condition met: found the %d expected fink topics\n", len(required))
				return
			}
			select {
			case <-deadline:
				fmt.Fprintf(os.Stderr, "error: timed out waiting for topics, missing: %s\n",
					strings.Join(missing, ", "))
				os.Exit(1)
			case <-time.After(10 * time.Second):
			}
		}
	},
}

func init() {
	waitCmd.AddCommand(waitTopicCmd)

	waitTopicCmd.Flags().StringVarP(&expectedFile, "expected-file", "f", "", "YAML file of the expected topics, by group")
	cobra.CheckErr(waitTopicCmd.MarkFlagRequired("expected-file"))
	waitTopicCmd.Flags().StringArrayVar(&withoutGroups, "without", nil, "Group of the expected-topics file not required (repeatable)")
}
