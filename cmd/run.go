/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/astrolabsoftware/finkctl/v3/resources"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var noscience bool
var image string

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

type SparkConfig struct {
	ApiServerUrl      string
	Binary            string
	Cpu               string `mapstructure:"cpu"`
	Image             string `mapstructure:"image"`
	Producer          string `mapstructure:"producer"`
	OnlineDataPrefix  string `mapstructure:"online_data_prefix"`
	Packages          string
	FinkTriggerUpdate string `mapstructure:"fink_trigger_update"`
	LocalTmpDirectory string
	LogLevel          string `mapstructure:"log_level"`
	Memory            string `mapstructure:"memory"`
	PodTemplateFile   string
	StorageClass      storageClass
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.PersistentFlags().StringVarP(&image, "image", "i", "", "fink-broker image name")
	runCmd.PersistentFlags().BoolVarP(&noscience, "noscience", "n", false, "Disable execution of science modules, can be overridden by exporting environment variable NOSCIENCE=true")

	// FIXME validate support for env variable fo noscience?
	viper.BindPFlag("noscience", runCmd.PersistentFlags().Lookup("noscience"))
	viper.AutomaticEnv()
}

func getSparkConfig(task string) SparkConfig {

	var c SparkConfig

	if err := viper.UnmarshalKey(RUN, &c); err != nil {
		logger.Fatalf("Error while getting spark configuration: %v", err)
	}

	if viper.GetString(task+".cpu") != "" {
		c.Cpu = viper.GetString(task + ".cpu")
	}
	if viper.GetString(task+".memory") != "" {
		c.Memory = viper.GetString(task + ".memory")
	}

	if task == DISTRIBUTION {
		c.Binary = DISTRIBUTION_BIN
		var err error
		c.LocalTmpDirectory, err = os.MkdirTemp(os.TempDir(), tmp_path_prefix)
		// Create a temporary file for kafka authentication
		if err != nil {
			log.Fatal(err)
		}
		c.PodTemplateFile = path.Join(c.LocalTmpDirectory, resources.ExecutorPodTemplateFile)
	} else {
		c.Binary = fmt.Sprintf("%s.py", task)
	}

	_, config := setKubeClient()
	apiServerUrl := config.Host
	c.ApiServerUrl = apiServerUrl

	if c.OnlineDataPrefix == "" {
		c.StorageClass = s3
	}

	if image != "" {
		c.Image = image
	}
	return c
}

func generateSparkCmd(task string) (string, SparkConfig) {
	sc := getSparkConfig(task)
	return applyTemplate(sc, task), sc
}

func applyTemplate(sc SparkConfig, task string) string {

	// WARNING package are not deployed in spark-executor
	// see https://stackoverflow.com/a/67299668/2784039
	sc.Packages = `org.apache.spark:spark-streaming-kafka-0-10-assembly_2.12:3.2.3,\
org.apache.spark:spark-sql-kafka-0-10_2.12:3.2.3,\
org.apache.spark:spark-avro_2.12:3.2.3,\
org.apache.spark:spark-token-provider-kafka-0-10_2.12:3.2.3,\
org.apache.hbase:hbase-shaded-mapreduce:2.2.7,\
com.amazonaws:aws-java-sdk-bundle:1.11.375,\
org.apache.hadoop:hadoop-aws:3.2.3`

	// TODO check https://docs.cloudera.com/cdp-private-cloud-base/7.1.8/ozone-storing-data/topics/ozone-config-spark-s3a.html
	cmdTpl := fmt.Sprintf(`spark-submit --master "k8s://{{ .ApiServerUrl }}" \
    --deploy-mode cluster \
    --conf spark.executor.instances=1 \
    --conf spark.kubernetes.namespace=%s \
    --conf spark.kubernetes.authenticate.driver.serviceAccountName=spark \
    --conf spark.kubernetes.container.image="{{ .Image }}" \
    --conf spark.driver.extraJavaOptions="-Divy.cache.dir=/tmp -Divy.home=/tmp" \
    `, getCurrentNamespace())

	if sc.StorageClass == s3 {
		s3c := getS3Config()
		sc.OnlineDataPrefix = fmt.Sprintf("s3a://%s", s3c.BucketName)
		s3OptTpl := `--conf spark.hadoop.fs.s3a.endpoint={{ .Endpoint }} \
    --conf spark.hadoop.fs.s3a.access.key="{{ .AccessKeyID }}" \
    --conf spark.hadoop.fs.s3a.secret.key="{{ .SecretAccessKey }}" \
    --conf spark.hadoop.mapreduce.fileoutputcommitter.algorithm.version=2 \
    --conf spark.hadoop.fs.s3a.connection.ssl.enabled={{ .UseSSL }} \
    --conf spark.hadoop.fs.s3a.fast.upload=true \
    --conf spark.hadoop.fs.s3a.path.style.access=true \
    --conf spark.hadoop.fs.s3a.aws.credentials.provider=org.apache.hadoop.fs.s3a.SimpleAWSCredentialsProvider \
    --conf spark.hadoop.fs.s3a.impl="org.apache.hadoop.fs.s3a.S3AFileSystem" \
    `
		cmdTpl += format(s3OptTpl, &s3c)
	}

	if task == DISTRIBUTION {
		kafkaOptTpl := fmt.Sprintf(`--conf spark.kubernetes.executor.podTemplateFile={{ .PodTemplateFile }} \
    --conf "spark.executor.extraJavaOptions=-Djava.security.auth.login.config=%s/%s" \
    `, configMapPathKafkaJaas, resources.KafkaJaasConfFile)
		cmdTpl += kafkaOptTpl
	}

	if sc.Memory != "" {
		cmdTpl += fmt.Sprintf(`--conf spark.driver.memory=%[1]s \
    --conf spark.executor.memory=%[1]s \
    `, sc.Memory)
	}

	if sc.Cpu != "" {
		cmdTpl += fmt.Sprintf(`--conf spark.kubernetes.driver.request.cores=%[1]s \
    --conf spark.kubernetes.executor.request.cores=%[1]s \
    `, sc.Cpu)
	}
	cmdTpl += `local:///home/fink/fink-broker/bin/{{ .Binary }} \
    -log_level "{{ .LogLevel }}" \
    -online_data_prefix "{{ .OnlineDataPrefix }}" \
    -producer "{{ .Producer }}" \
    -tinterval "{{ .FinkTriggerUpdate }}" \
    `
	if noscience {
		cmdTpl += `--noscience \
    `
	}
	cmd := format(cmdTpl, &sc)
	return cmd
}
