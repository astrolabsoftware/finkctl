/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/

package cmd

import (
	"os"

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
	DistributionServers string     `mapstructure:"distribution_servers"`
	SubstreamPrefix     string     `mapstructure:"substream_prefix"`
	DistributionSchema  string     `mapstructure:"distribution_schema"`
	Night               string     `mapstructure:"night"`
	KafkaCreds          KafkaCreds `mapstructure:"kafka"`
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
		logger.Info(startMsg)

		sparkCmd, sc := generateSparkCmd(DISTRIBUTION)

		cmdTpl := sparkCmd + `-distribution_servers "{{ .DistributionServers }}" \
    -distribution_schema "{{ .DistributionSchema }}" \
    -substream_prefix "{{ .SubstreamPrefix }}" \
    -night "{{ .Night }}"`

		createExecutorPodTemplate(sc.PodTemplateFile)

		c := getDistributionConfig()

		createKafkaJaasConfigMap(&c)

		sparkCmd = format(cmdTpl, &c)

		ExecCmd(sparkCmd)
	},
}

func createExecutorPodTemplate(filename string) {
	c := getKubeVars()
	kafkaJaasConf := format(resources.ExecutorPodTemplate, &c)
	logger.Debugf("Writing PodTemplate to file: %s", filename)
	executorPodTemplateFile, err := os.Create(filename)
	if err != nil {
		logger.Fatal(err)
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

func getDistributionConfig() DistributionConfig {
	var c DistributionConfig

	if err := viper.UnmarshalKey(DISTRIBUTION, &c); err != nil {
		logger.Fatalf("Error while getting %s configuration: %v", DISTRIBUTION, err)
	}
	if c.DistributionServers == "" {
		c.DistributionServers = viper.GetString("stream2raw.kafka_socket")
	}
	if c.Night == "" {
		c.Night = viper.GetString("raw2science.night")
	}
	if c.KafkaCreds.Password == "" {
		c.KafkaCreds.Password = getKafkaPasswordFromSecret()
	}

	return c
}
