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

// sparkCmd represents the spark command
var sparkCmd = &cobra.Command{
	Use:   "spark",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("spark called")

	},
}

// TODO split parameters for stream2raw and raw2science
var sparkArgs = map[string]interface{}{
	"image":                 "delicious",
	"producer":              "delicious",
	"kafka_socket":          "delicious",
	"kafka_topic":           "delicious",
	"fink_alert_schema":     "delicious",
	"kafka_starting_offset": "delicious",
	"online_data_prefix":    "delicious",
	"fink_trigger_update":   "delicious",
	"log_level":             "delicious",
	"night":                 "delicious",
}

func init() {
	rootCmd.AddCommand(sparkCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// sparkCmd.PersistentFlags().String("foo", "", "A help for foo")

	viper.AutomaticEnv()

	for option := range sparkArgs {
		sparkCmd.PersistentFlags().String(option, "", "fink-broker image name")
		viper.BindPFlag(option, sparkCmd.PersistentFlags().Lookup(option))
	}

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// sparkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func runSpark(sparkArgs map[string]interface{}) {

	_, config := setKubeClient()

	apiServerUrl := config.Host

	sparkArgs["api_server_url"] = apiServerUrl

	cmdTpl := `spark-submit --master "k8s://{{ .api_server_url }}" \
    --deploy-mode cluster \
    --conf spark.executor.instances=1 \
    --conf spark.kubernetes.authenticate.driver.serviceAccountName=spark \
    --conf spark.kubernetes.container.image="{{ .image }}" \
    --conf spark.driver.extraJavaOptions="-Divy.cache.dir=/home/fink -Divy.home=/home/fink" \
    $ci_opt \
    local:///home/fink/fink-broker/bin/{{ .bin }} \
    -log_level "{{ .log_level }}" \
    -online_data_prefix "{{ .online_data_prefix }}" \
    -producer "{{ .producer }}" \
    -tinterval "{{ .fink_trigger_update }}" \
    `

	switch sparkArgs["bin"] {
	case "stream2raw.py":
		cmdTpl += `-servers "{{ .kafka_socket }}"
    -schema "{{ .fink_alert_schema }}"
    -startingoffsets_stream "{{ .kafka_starting_offset }}" \
    -topic "{{ .kafka_topic }}"`

	case "raw2science.py":
		cmdTpl += `-night "{{ .night }}"`
	}

	cmd := format(cmdTpl, sparkArgs)

	out, errout := ExecCmd(cmd)

	outmsg := OutMsg{
		cmd:    cmd,
		out:    out,
		errout: errout}

	log.Printf("message: %v\n", outmsg)
}
