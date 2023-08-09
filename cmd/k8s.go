package cmd

import (
	"context"
	"os"
	"path"

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

func getCurrentNamespace() string {
	// FIXME: this is a temporary solution
	// Watch https://stackoverflow.com/questions/76850938/how-to-get-current-namespace-of-an-out-cluster-go-kubernetes-client
	return "default"
}
