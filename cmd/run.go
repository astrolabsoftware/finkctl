/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/astrolabsoftware/finkctl/v3/resources"
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
	Cpu               string `mapstructure:"cpu"`
	Image             string `mapstructure:"image"`
	Instances         string `mapstructure:"instances"`
	Night             string `mapstructure:"night"`
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

	if viper.GetString(task+".cpu") != "" {
		rc.Cpu = viper.GetString(task + ".cpu")
	}
	if viper.GetString(task+".memory") != "" {
		rc.Memory = viper.GetString(task + ".memory")
	}
	if viper.GetString(task+".instances") != "" {
		rc.Instances = viper.GetString(task + ".instances")
	}

	if task == DISTRIBUTION {
		rc.Binary = DISTRIBUTION_BIN
		var err error
		rc.LocalTmpDirectory, err = os.MkdirTemp(os.TempDir(), tmp_path_prefix)
		// Create a temporary file for kafka authentication
		if err != nil {
			log.Fatal(err)
		}
		rc.PodTemplateFile = path.Join(rc.LocalTmpDirectory, resources.ExecutorPodTemplateFile)
	} else {
		rc.Binary = fmt.Sprintf("%s.py", task)
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

func generateSparkCmd(task string) (string, RunConfig) {
	sc := getRunConfig(task)
	return applyTemplate(sc, task), sc
}

func applyTemplate(rc RunConfig, task string) string {

	// WARNING package are not deployed in spark-executor
	// see https://stackoverflow.com/a/67299668/2784039
	rc.Packages = `org.apache.spark:spark-streaming-kafka-0-10-assembly_2.12:3.2.3,\
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

	if rc.StorageClass == s3 {
		s3c := getS3Config(rc.Night)
		rc.OnlineDataPrefix = fmt.Sprintf("s3a://%s", s3c.BucketName)
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
    `, secretPathKafkaJaas, resources.KafkaJaasConfFile)
		cmdTpl += kafkaOptTpl
	}
	// TODO make it configurable at the task level using {{ .InstancesOption }}
	if rc.Instances != "" {
		cmdTpl += fmt.Sprintf(`--conf spark.executor.instances=%[1]s \
    `, rc.Instances)
	}
	// TODO make it configurable at the task level
	if rc.Memory != "" {
		cmdTpl += fmt.Sprintf(`--conf spark.driver.memory=%[1]s \
    --conf spark.executor.memory=%[1]s \
    `, rc.Memory)
	}
	// TODO make it configurable at the task level
	if rc.Cpu != "" {
		cmdTpl += fmt.Sprintf(`--conf spark.driver.cores=%[1]s \
    --conf spark.executor.cores=%[1]s \
    `, rc.Cpu)
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
	cmd := format(cmdTpl, &rc)
	return cmd
}
