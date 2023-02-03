/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"log"

	"github.com/spf13/cobra"
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
		runSpark()
	},
}

func runSpark() {

	_, config := setKubeClient()

	api_server_url := config.Host

	bin := "stream2raw.py"

	var err_out error
	cmd_tpl := `spark-submit --master "k8s://%v" \
    --deploy-mode cluster \
    --conf spark.executor.instances=1 \
    --conf spark.kubernetes.authenticate.driver.serviceAccountName=spark \
    --conf spark.kubernetes.container.image="$IMAGE" \
    --conf spark.driver.extraJavaOptions="-Divy.cache.dir=/home/fink -Divy.home=/home/fink" \
    $ci_opt \
    local:///home/fink/fink-broker/bin/%v \
    -producer "${PRODUCER}" \
    -servers "${KAFKA_SOCKET}" -topic "${KAFKA_TOPIC}" \
    -schema "${FINK_ALERT_SCHEMA}" -startingoffsets_stream "${KAFKA_STARTING_OFFSET}" \
    -online_data_prefix "${ONLINE_DATA_PREFIX}" \
    -tinterval "${FINK_TRIGGER_UPDATE}" -log_level "${LOG_LEVEL}"`

	cmd := fmt.Sprintf(cmd_tpl, api_server_url, bin)

	out, errout, err := Shellout(cmd)
	if err != nil {
		err_msg := fmt.Sprintf("error creating kind cluster: %v\n", err)
		err_out = errors.New(err_msg)
	}

	outmsg := OutMsg{
		cmd:    cmd,
		err:    err_out,
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
