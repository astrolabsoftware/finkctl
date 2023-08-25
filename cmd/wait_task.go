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
		logger.Info("List kafka topics produced by the fink-broker")
		clientSet, _ := setKubeClient()
		WaitForPodBySelectorRunning(clientSet, getCurrentNamespace(), "spark-role=executor", timeout)

	},
}

func init() {
	waitCmd.AddCommand(waitTaskCmd)
}
