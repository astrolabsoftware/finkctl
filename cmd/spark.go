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

var minimal bool

// sparkCmd represents the spark command
var sparkCmd = &cobra.Command{
	Use:     "spark",
	Aliases: []string{"s"},
	Short:   "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Display Spark Configuration")
		var sc SparkConfig
		if err := viper.UnmarshalKey("spark", &sc); err != nil {
			fmt.Println(err)
		}
		fmt.Println(sc)
		var s2rc Stream2RawConfig
		if err := viper.UnmarshalKey("stream2raw", &s2rc); err != nil {
			fmt.Println(err)
		}
		fmt.Println(s2rc)
		var r2sc Raw2ScienceConfig
		if err := viper.UnmarshalKey("raw2science", &r2sc); err != nil {
			fmt.Println(err)
		}
		fmt.Println(r2sc)
	},
}

type SparkConfig struct {
	ApiServerUrl      string
	Binary            string
	Image             string `mapstructure:"image"`
	Producer          string `mapstructure:"producer"`
	OnlineDataPrefix  string `mapstructure:"online_data_prefix"`
	FinkTriggerUpdate string `mapstructure:"fink_trigger_update"`
	LogLevel          string `mapstructure:"log_level"`
}

func init() {
	rootCmd.AddCommand(sparkCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	sparkCmd.PersistentFlags().BoolVarP(&minimal, "minimal", "m", false, "Set minimal cpu/memory requests for spark pods")

	sparkCmd.PersistentFlags().String("image", "", "fink-broker image name")
	viper.BindPFlag("image", sparkCmd.PersistentFlags().Lookup("image"))

	// log.Printf("CONFIG::: %s\n", cfgFile)
	// if cfgFile != "" {
	// 	// Use config file from the flag.
	// 	viper.SetConfigFile(cfgFile)
	// }

	// If a config file is found, read it in.
	// if err := viper.ReadInConfig(); err == nil {
	// 	fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	// }
	// fmt.Fprintln(os.Stdout, "Using config file:", viper.ConfigFileUsed())

	// for option := range sparkArgs {
	// 	log.Printf("option %v", viper.GetString(option))
	// }

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// sparkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func getSparkConfig(task string) SparkConfig {
	var c SparkConfig
	if err := viper.UnmarshalKey("spark", &c); err != nil {
		log.Fatalf("Error while getting spark configuration: %v", err)
	}

	if task == DISTRIBUTION {
		c.Binary = DISTRIBUTION_BIN
	} else {
		c.Binary = fmt.Sprintf("%s.py", task)
	}

	_, config := setKubeClient()
	apiServerUrl := config.Host
	c.ApiServerUrl = apiServerUrl

	if c.Image == "" {
		c.Image = viper.GetString("image")
	}
	return c
}

func generateSparkCmd(task string) string {

	sc := getSparkConfig(task)
	return applyTemplate(sc)
}

func applyTemplate(sc SparkConfig) string {
	// TODO check https://docs.cloudera.com/cdp-private-cloud-base/7.1.8/ozone-storing-data/topics/ozone-config-spark-s3a.html
	cmdTpl := `spark-submit --master "k8s://{{ .ApiServerUrl }}" \
    --deploy-mode cluster \
    --packages "org.apache.spark:spark-core_2.12:3.2.1,com.amazonaws:aws-java-sdk:1.11.967,org.apache.hadoop:hadoop-aws:3.2.3,org.apache.hadoop:hadoop-common:3.2.1,org.apache.hadoop:hadoop-client:3.2.1" \
    --conf spark.executor.instances=1 \
    --conf spark.kubernetes.authenticate.driver.serviceAccountName=spark \
    --conf spark.kubernetes.container.image="{{ .Image }}" \
    --conf spark.driver.extraJavaOptions="-Divy.cache.dir=/home/fink -Divy.home=/home/fink" \
    --conf spark.hadoop.fs.s3a.endpoint=http://minio.minio-dev:9000 \
    --conf spark.hadoop.fs.s3a.access.key="minioadmin" \
    --conf spark.hadoop.fs.s3a.secret.key="minioadmin" \
    --conf spark.hadoop.mapreduce.fileoutputcommitter.algorithm.version=2 \
    --conf spark.hadoop.fs.s3a.connection.ssl.enabled=false \
    --conf spark.hadoop.fs.s3a.fast.upload=true \
    --conf spark.hadoop.fs.s3a.path.style.access=true \
    --conf spark.hadoop.fs.s3a.impl="org.apache.hadoop.fs.s3a.S3AFileSystem" \
    `
	if minimal {
		cmdTpl += `--conf spark.kubernetes.driver.request.cores=0 \
    --conf spark.kubernetes.executor.request.cores=0 \
    --conf spark.driver.memory=500m \
    --conf spark.executor.memory=500m \
    `
	}
	cmdTpl += `local:///home/fink/fink-broker/bin/{{ .Binary }} \
    -log_level "{{ .LogLevel }}" \
    -online_data_prefix "{{ .OnlineDataPrefix }}" \
    -producer "{{ .Producer }}" \
    -tinterval "{{ .FinkTriggerUpdate }}" \
    `
	cmd := format(cmdTpl, &sc)
	return cmd
}
