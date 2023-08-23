package cmd

import (
	"context"
	"encoding/json"
	"os"
	"path"

	kafka "github.com/RedHatInsights/strimzi-client-go/apis/kafka.strimzi.io/v1beta2"
	"github.com/astrolabsoftware/finkctl/resources"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	applymetav1 "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	kubeConfigEnvName         = "KUBECONFIG"
	kubeConfigDefaultFilename = ".kube/config"
	configMapNameKafkaJaas    = "fink-kafka-jaas"
	configMapPathKafkaJaas    = "/etc/fink-broker"
)

type kubeVars struct {
	ConfigMapNameKafkaJaas string
	ConfigMapPathKafkaJaas string
}

func getKubeVars() kubeVars {
	kubeVarsInstance := kubeVars{
		ConfigMapNameKafkaJaas: configMapNameKafkaJaas,
		ConfigMapPathKafkaJaas: configMapPathKafkaJaas,
	}
	return kubeVarsInstance
}

func getKubeConfig() string {
	kubeConfigFilename := os.Getenv(kubeConfigEnvName)
	// Fallback to default kubeconfig file location if no env variable set
	if kubeConfigFilename == "" {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		kubeConfigFilename = path.Join(home, kubeConfigDefaultFilename)
	}
	return kubeConfigFilename
}

func setKubeClient() (*kubernetes.Clientset, *rest.Config) {

	kubeconfig := getKubeConfig()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset, config
}

func createKafkaJaasConfigMap(c *DistributionConfig) {
	kafkaJaasConf := format(resources.KafkaJaasConf, &c)

	clientSet, _ := setKubeClient()

	files := make(map[string]string)

	files[resources.KafkaJaasConfFile] = kafkaJaasConf

	kind := "ConfigMap"
	version := "v1"
	name := configMapNameKafkaJaas

	cm := v1.ConfigMapApplyConfiguration{
		TypeMetaApplyConfiguration: applymetav1.TypeMetaApplyConfiguration{
			Kind:       &kind,
			APIVersion: &version,
		},
		ObjectMetaApplyConfiguration: &applymetav1.ObjectMetaApplyConfiguration{
			Name: &name,
		},
		Data: files,
	}

	_, err := clientSet.CoreV1().ConfigMaps(getCurrentNamespace()).Apply(
		context.TODO(), &cm,
		metav1.ApplyOptions{FieldManager: "application/apply-patch"})
	if err != nil {
		panic(err.Error())
	}
}

// getKafkaPasswordFromSecret returns the kafka password
// equivalent to "kubectl get -n kafka secrets/fink-producer --template={{.data.password}} | base64 --decode"
func getKafkaPasswordFromSecret() string {

	clientSet, _ := setKubeClient()

	secret, err := clientSet.CoreV1().Secrets(kafkaNamespace).Get(
		context.TODO(), kafkaSecretName, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	return string(secret.Data["password"])
}

// getKafkaTopic returns the kafka topics
// equivalent to "kubectl get -n kafka kafkatopics.kafka.strimzi.io --template='{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}'"
func getKafkaTopics() []string {

	clientSet, _ := setKubeClient()

	topics := &kafka.KafkaTopicList{}
	d, err := clientSet.RESTClient().Get().Namespace(kafkaNamespace).AbsPath("/apis/kafka.strimzi.io/v1beta2/kafkatopics").DoRaw(context.TODO())
	if err != nil {
		panic(err.Error())
	}
	logger.Debugf("Unmarshalling kafka topics: %s", string(d))
	if err := json.Unmarshal(d, &topics); err != nil {
		panic(err.Error())
	}

	topicNames := make([]string, len(topics.Items))
	for _, topic := range topics.Items {
		topicNames = append(topicNames, topic.Name)
	}
	return topicNames
}

// getCurrentNamespace returns the current namespace
// for the current context of the kubeconfig file
func getCurrentNamespace() string {

	kubeconfig := getKubeConfig()

	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	ns := config.Contexts[config.CurrentContext].Namespace

	if len(ns) == 0 {
		ns = "default"
	}

	return ns
}
