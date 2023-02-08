/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/

package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const DISTRIBUTION string = "distribution"
const DISTRIBUTION_BIN string = "distribute.py"

type DistributionConfig struct {
	DistributionServers string `mapstructure:"distribution_servers"`
	SubstreamPrefix     string `mapstructure:"substream_prefix"`
	DistributionSchema  string `mapstructure:"distribution_schema"`
	Night               string `mapstructure:"night"`
}

// distributionCmd represents the distribution command
var distributionCmd = &cobra.Command{
	Use:     DISTRIBUTION,
	Aliases: []string{"di"},
	Short:   "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("distribution called")

		sparkCmd := generateSparkCmd(DISTRIBUTION)

		cmdTpl := sparkCmd + `-distribution_servers "{{ .DistributionServers }}" \
    -distribution_schema "{{ .DistributionSchema }}" \
    -substream_prefix "{{ .SubstreamPrefix }}" \
    -night "{{ .Night }}"`
		c := getDistributionConfig()
		sparkCmd = format(cmdTpl, &c)

		out, errout := ExecCmd(sparkCmd)
		outmsg := OutMsg{
			cmd:    sparkCmd,
			out:    out,
			errout: errout}
		log.Printf("message: %v\n", outmsg)
	},

	/*
		elif [[ $service == "distribution" ]]; then
		  if [[ $ELASTICC == true ]]; then
		    SCRIPT=${FINK_HOME}/bin/distribute_elasticc.py
		  else
		    SCRIPT=${FINK_HOME}/bin/distribute.py
		  fi
		  # Read configuration for redistribution
		  source ${FINK_HOME}/conf/fink.conf.distribution
		  # Start the Spark Producer
		  spark-submit --master ${SPARK_MASTER} \
		  --files ${FINK_HOME}/conf/fink_kafka_producer_jaas.conf \
		  --driver-java-options "-Djava.security.auth.login.config=${FINK_PRODUCER_JAAS}" \
		  --conf "spark.driver.extraJavaOptions=-Djava.security.auth.login.config=${FINK_PRODUCER_JAAS}" \
		  --conf "spark.executor.extraJavaOptions=-Djava.security.auth.login.config=${FINK_PRODUCER_JAAS}" \
		  $SCRIPT ${HELP_ON_SERVICE} \
		  -producer ${PRODUCER} \
		  -online_data_prefix ${ONLINE_DATA_PREFIX} \
		  -distribution_servers ${DISTRIBUTION_SERVERS} \
		  -distribution_schema ${DISTRIBUTION_SCHEMA} \
		  -substream_prefix ${SUBSTREAM_PREFIX} \
		  -tinterval ${FINK_TRIGGER_UPDATE} \
		  -night ${NIGHT} \
		  -log_level ${LOG_LEVEL} ${EXIT_AFTER}
	*/

	/*
		 	files: /home/fink/fink-broker/conf/fink_kafka_producer_jaas.conf
			driver-java-options: "-Djava.security.auth.login.config=${FINK_PRODUCER_JAAS}"
			--conf spark.kubernetes.driver.secrets.spark-secret=/etc/secrets
			--conf spark.kubernetes.executor.secrets.spark-secret=/etc/secrets
			--conf "spark.driver.extraJavaOptions=-Djava.security.auth.login.config=${FINK_PRODUCER_JAAS}"
			--conf "spark.executor.extraJavaOptions=-Djava.security.auth.login.config=${FINK_PRODUCER_JAAS}"
	*/
}

func init() {
	sparkCmd.AddCommand(distributionCmd)

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
		log.Fatalf("Error while getting %s configuration: %v", DISTRIBUTION, err)
	}
	if c.DistributionServers == "" {
		c.DistributionServers = viper.GetString("stream2raw.kafka_socket")
	}
	if c.Night == "" {
		c.Night = viper.GetString("raw2science.night")
	}

	return c
}
