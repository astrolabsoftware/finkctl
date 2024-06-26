/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const RAW2SCIENCE string = "raw2science"

type Raw2ScienceConfig struct {
	MmConfigPath string `mapstructure:"mmconfigpath"`
	Night        string
}

// raw2scienceCmd represents the raw2science command
var raw2scienceCmd = &cobra.Command{
	Use:     RAW2SCIENCE,
	Aliases: []string{"r"},
	Short:   "Launch Raw2science service on Spark over Kubernetes",
	Long: `Launch Raw2science service on Spark over Kubernetes. Raw2science processes and analyzes data from a
shared file system and send it to Kafka streams.`,
	Example: `  # Lauch raw2science service on Spark, over Kubernetes
	  finkctl spark raw2science
	  # Lauch raw2science using a custom image
	  finkctl spark raw2science --image=gitlab-registry.in2p3.fr/astrolabsoftware/fink/fink-broker:2076184`,
	Run: func(cmd *cobra.Command, args []string) {

		slog.Info("Launch raw2science service")

		sparkCmd, rc := generateSparkCmd(RAW2SCIENCE)

		cmdTpl := sparkCmd + `-night "{{ .Night }}" -mmconfigpath "{{ .MmConfigPath }}"`
		c := getRaw2ScienceConfig(rc.Night)
		sparkCmd = format(cmdTpl, &c)
		ExecCmd(sparkCmd)
	},
}

func init() {
	runCmd.AddCommand(raw2scienceCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// raw2scienceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// raw2scienceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func getRaw2ScienceConfig(night string) Raw2ScienceConfig {
	var c Raw2ScienceConfig
	if err := viper.UnmarshalKey(RAW2SCIENCE, &c); err != nil {
		slog.Error("Error while getting configuration", "task", RAW2SCIENCE, "error", err)
	}
	c.Night = night
	return c
}
