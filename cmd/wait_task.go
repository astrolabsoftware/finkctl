/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

// topicCmd represents the topic command
var waitTaskCmd = &cobra.Command{
	Use:     "task",
	Aliases: []string{"tsk", "tasks"},
	Short:   "Wait for fink-broker pods to be launched",
	Long:    `Wait for fink-broker pods to be launched, timeout is applyed for each set of pod (driver and executor)`,
	Run: func(cmd *cobra.Command, args []string) {
		labelValues := []string{"driver", "executor"}
		expected_pods := 3
		clientSet, _ := setKubeClient()
		for _, value := range labelValues {
			selector := "spark-role=" + value
			slog.Info("Wait for fink-broker pods to be launched", "selector", selector)
			err := waitForPodExistsBySelector(clientSet, getCurrentNamespace(), selector, timeout, expected_pods)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: timed out waiting for %s pods to be created, reason: %s\n", value, err)
				os.Exit(1)
			}
			slog.Info("Wait for fink-broker pods to be ready", "selector", selector)
			err = waitForPodReadyBySelector(clientSet, getCurrentNamespace(), selector, timeout)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: timed out waiting for %s pods to be ready, reason: %s\n", value, err)
				os.Exit(1)
			}
			cmd.Printf("Pods with label '%s' are ready\n", selector)
		}
	},
}

func init() {
	waitCmd.AddCommand(waitTaskCmd)
}
