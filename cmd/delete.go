/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete all spark pods in current namespace",
	Long:  `Delete all spark pods in current namespace, using the label 'spark-app-selector'.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Delete all spark pods in current namespace")
		clientSet, _ := setKubeClient()

		pods, err := clientSet.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
			LabelSelector: "spark-app-selector",
		})
		if err != nil {
			log.Fatal(err)
		}

		for _, p := range pods.Items {
			err := clientSet.CoreV1().Pods(p.Namespace).Delete(context.TODO(), p.Name, metav1.DeleteOptions{})
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Delete pod %s", p.Name)
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deleteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deleteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
