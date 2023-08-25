/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
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
		logger.Info("Wait for fink-broker pods to be launched %s", labelValues)
		clientSet, _ := setKubeClient()
		for _, value := range labelValues {
			err := waitForPodExistsBySelector(clientSet, getCurrentNamespace(), "spark-role="+value, timeout, expected_pods)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: timed out waiting for %s pods to be created, reason: %s", value, err)
				os.Exit(1)
			}
			err = waitForPodReadyBySelector(clientSet, getCurrentNamespace(), "spark-role="+value, timeout)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: timed out waiting for %s pods to be ready, reason: %s", value, err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	waitCmd.AddCommand(waitTaskCmd)
}
