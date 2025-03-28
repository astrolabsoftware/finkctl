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
	"github.com/spf13/viper"
)

var noscience bool
var image string
var night string
var tonight bool

type storageClass int

const (
	s3 storageClass = iota
	hdfs
)

const (
	RUN             string = "run"
	tmp_path_prefix string = "fink-broker-"
)

// runCmd represents the spark command
var runCmd = &cobra.Command{
	Use:   RUN,
	Short: "Display Fink-broker parameters, for running it on Spark over Kubernetes",
	Long:  `Display all spark-submit parameters for running fink-broker on Spark over Kubernetes`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
	},
	Run: func(cmd *cobra.Command, args []string) {
		logConfiguration()
	},
}

type RunConfig struct {
	ApiServerUrl      string
	Binary            string
	Image             string `mapstructure:"image"`
	Night             string `mapstructure:"night"`
	OnlineDataPrefix  string `mapstructure:"online_data_prefix"`
	Packages          string
	LocalTmpDirectory string
	LogLevel          string `mapstructure:"log_level"`
	StorageClass      storageClass
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.PersistentFlags().StringVarP(&night, "night", "N", "", "Night to process, format YYYYMMDD, default is empty string, used in finkctl.yaml as {{.Night}} template")
	runCmd.PersistentFlags().StringVarP(&image, "image", "i", "", "fink-broker image name, used in finkctl.yaml as {{.Image}} template")
	runCmd.PersistentFlags().BoolVarP(&noscience, "noscience", "n", false, "Disable execution of science modules, can be overridden by exporting environment variable NOSCIENCE=true")
	runCmd.PersistentFlags().BoolVarP(&tonight, "tonight", "t", false, "Use tonight's date as night, format YYYYMMDD, used in finkctl.yaml as {{.Night}} template, supersed night flag")

	// FIXME validate support for env variable fo noscience?
	viper.BindPFlag("noscience", runCmd.PersistentFlags().Lookup("noscience"))
}

func getRunConfig(task string) RunConfig {

	var rc RunConfig

	if err := viper.UnmarshalKey(RUN, &rc); err != nil {
		slog.Error("Error while getting spark configuration", "task", task, "error", err)
	}

	_, config := setKubeClient()
	apiServerUrl := config.Host
	rc.ApiServerUrl = apiServerUrl

	if rc.OnlineDataPrefix == "" {
		rc.StorageClass = s3
	}

	if image != "" {
		rc.Image = image
	}

	YYYYMMDD := "20060102"
	if tonight {
		now := time.Now().UTC()
		tonight := now.Format(YYYYMMDD)
		rc.Night = tonight
	} else if night != "" {
		rc.Night = night
	}
	if rc.Night == "" {
		err := fmt.Errorf("night is empty, please provide a night")
		slog.Error("Error while getting spark configuration", "task", task, "error", err)
		os.Exit(1)
	}
	_, err1 := time.Parse(YYYYMMDD, rc.Night)
	if err1 != nil {
		err := fmt.Errorf("night has not the right format, please provide a night in the format YYYYMMDD")
		slog.Error("Error while getting spark configuration", "task", task, "error", err)
		os.Exit(1)
	}
	return rc
}
