/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// topicCmd represents the topic command
var waitTaskCmd = &cobra.Command{
	Use:     "task",
	Aliases: []string{"tsk", "tasks"},
	Short:   "Wait for fink-broker pods to be launched",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Wait for fink-broker pods to be launched")
		clientSet, _ := setKubeClient()
		WaitForPodReadyBySelector(clientSet, getCurrentNamespace(), "spark-role=driver", timeout)
		WaitForPodReadyBySelector(clientSet, getCurrentNamespace(), "spark-role=executor", timeout)
	},
}

func init() {
	waitCmd.AddCommand(waitTaskCmd)
}
