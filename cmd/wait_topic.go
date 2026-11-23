/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
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
		// Written by the poll, read on timeout while a poll may still run
		var mu sync.Mutex
		missing := required
		err = pollUntil(timeout, 10*time.Second, func() (bool, error) {
			topics, err := getFinkTopics()
			if err != nil {
				return false, err
			}
			stillMissing, extra := compareTopics(required, topics)
			mu.Lock()
			missing = stillMissing
			mu.Unlock()
			slog.Info("Found topics", "elapsedTime", time.Since(start).Round(time.Second),
				"found", len(topics), "missing", stillMissing)
			if len(extra) > 0 {
				slog.Info("Topics found but not expected", "topics", extra)
			}
			return len(stillMissing) == 0, nil
		})
		if errors.Is(err, errPollTimeout) {
			mu.Lock()
			fmt.Fprintf(os.Stderr, "error: timed out waiting for topics, missing: %s\n",
				strings.Join(missing, ", "))
			os.Exit(1)
		}
		cobra.CheckErr(err)
		fmt.Printf("Condition met: found the %d expected fink topics\n", len(required))
	},
}

func init() {
	waitCmd.AddCommand(waitTopicCmd)

	waitTopicCmd.Flags().StringVarP(&expectedFile, "expected-file", "f", "", "YAML file of the expected topics, by group")
	cobra.CheckErr(waitTopicCmd.MarkFlagRequired("expected-file"))
	waitTopicCmd.Flags().StringArrayVar(&withoutGroups, "without", nil, "Group of the expected-topics file not required (repeatable)")
}
