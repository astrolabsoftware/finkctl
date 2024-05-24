/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/

package cmd

import (
	"log/slog"
	"os"
	"syscall"

	"github.com/astrolabsoftware/finkctl/v3/resources"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const DISTRIBUTION string = "distribution"
const DISTRIBUTION_BIN string = "distribute.py"

type KafkaCreds struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type DistributionConfig struct {
	Cpu                    string `mapstructure:"cpu"`
	DistributionServers    string `mapstructure:"distribution_servers"`
	Memory                 string `mapstructure:"memory"`
	SubstreamPrefix        string `mapstructure:"substream_prefix"`
	DistributionSchema     string `mapstructure:"distribution_schema"`
	KafkaBufferMemory      string `mapstructure:"kafka_buffer_memory"`
	KafkaDeliveryTimeoutMs string `mapstructure:"kafka_delivery_timeout_ms"`
	MmConfigPath           string `mapstructure:"mmconfigpath"`
	Night                  string
	KafkaCreds             KafkaCreds `mapstructure:"kafka"`
}

// distributionCmd represents the distribution command
var distributionCmd = &cobra.Command{
	Use:     DISTRIBUTION,
	Aliases: []string{"d", "dis"},
	Short:   "Launch Distribution service on Spark over Kubernetes",
	Long:    `Start fink-broker distribution service on Kubernetes`,
	Example: `  # Start fink-broker distribution service using image <image>
  finkctl spark --image=<image> distribution`,
	Run: func(cmd *cobra.Command, args []string) {
		startMsg := "Launch distribution service"
		slog.Info(startMsg)

		sparkCmd, rc := generateSparkCmd(DISTRIBUTION)
		c := getDistributionConfig(rc.Night)

		cmdTpl := sparkCmd + `-distribution_servers "{{ .DistributionServers }}" \
    -distribution_schema "{{ .DistributionSchema }}" \
    -substream_prefix "{{ .SubstreamPrefix }}" \
    -kafka_buffer_memory "{{ .KafkaBufferMemory }}" \
    -kafka_delivery_timeout_ms "{{ .KafkaDeliveryTimeoutMs }}" \
    -mmconfigpath "{{ .MmConfigPath }}" \
    -night "{{ .Night }}"`

		createExecutorPodTemplate(rc.PodTemplateFile)

		if dryRun {
			slog.Warn("Dry-run mode enabled, not creating KafkaJaasConfigMap")
		} else {
			createKafkaJaasConfigMap(&c)
		}
		sparkCmd = format(cmdTpl, &c)

		ExecCmd(sparkCmd)
	},
}

func createExecutorPodTemplate(filename string) {
	c := getKubeVars()
	kafkaJaasConf := format(resources.ExecutorPodTemplate, &c)
	slog.Debug("Writing PodTemplate", "destFile", filename)
	executorPodTemplateFile, err := os.Create(filename)
	if err != nil {
		slog.Error("Error while creating executor pod template file", "error", err)
		syscall.Exit(1)
	}
	defer executorPodTemplateFile.Close()

	// Write kafkaJaasConf to file
	executorPodTemplateFile.WriteString(kafkaJaasConf)
}

func init() {
	runCmd.AddCommand(distributionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// distributionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// distributionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func getDistributionConfig(night string) DistributionConfig {
	var c DistributionConfig

	if err := viper.UnmarshalKey(DISTRIBUTION, &c); err != nil {
		slog.Error("Error while getting configuration", "task", DISTRIBUTION, "error", err)
	}
	if c.DistributionServers == "" {
		c.DistributionServers = viper.GetString("stream2raw.kafka_socket")
	}
	c.Night = night
	if c.KafkaCreds.Password == "" {
		c.KafkaCreds.Password = getKafkaPasswordFromSecret()
	}

	return c
}
