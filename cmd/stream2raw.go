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

// stream2rawCmd represents the stream2raw command
var stream2rawCmd = &cobra.Command{
	Use:   "stream2raw",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("stream2raw called")

		for option := range sparkArgs {
			sparkArgs[option] = viper.GetString(option)
		}

		log.Printf("Args %v", sparkArgs)

		// TODO check error
		runSpark(sparkArgs)
	},
}

func runSpark(sparkArgs map[string]interface{}) {

	_, config := setKubeClient()

	apiServerUrl := config.Host

	sparkArgs["api_server_url"] = apiServerUrl
	sparkArgs["bin"] = "stream2raw.py"

	cmdTpl := `spark-submit --master "k8s://{{ .api_server_url }}" \
    --deploy-mode cluster \
    --conf spark.executor.instances=1 \
    --conf spark.kubernetes.authenticate.driver.serviceAccountName=spark \
    --conf spark.kubernetes.container.image="{{ .image }}" \
    --conf spark.driver.extraJavaOptions="-Divy.cache.dir=/home/fink -Divy.home=/home/fink" \
    $ci_opt \
    local:///home/fink/fink-broker/bin/{{ .bin }} \
    -producer "{{ .producer }}" \
    -servers "{{ .kafka_socket }}" -topic "{{ .kafka_topic }}" \
    -schema "{{ .fink_alert_schema }}" -startingoffsets_stream "{{ .kafka_starting_offset }}" \
    -online_data_prefix "{{ .online_data_prefix }}" \
    -tinterval "{{ .fink_trigger_update }}" -log_level "{{ .log_level }}"`

	cmd := format(cmdTpl, sparkArgs)

	out, errout := ExecCmd(cmd)

	outmsg := OutMsg{
		cmd:    cmd,
		out:    out,
		errout: errout}

	log.Printf("message: %v\n", outmsg)
}

func init() {
	sparkCmd.AddCommand(stream2rawCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// stream2rawCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// stream2rawCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
