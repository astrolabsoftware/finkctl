package cmd

import (
	"context"
	"os"
	"path"

	"github.com/astrolabsoftware/finkctl/resources"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	kubeConfigEnvName         = "KUBECONFIG"
	kubeConfigDefaultFilename = ".kube/config"
	configMapNameKafkaJaas    = "fink-kafka-jaas"
	configMapNamePathJaas     = "fink-kafka-jaas"
)

type kubeVars struct {
	ConfigMapNameKafkaJaas string
	ConfigMapPathKafkaJaas string
}

func getKubeVars() kubeVars {
	kubeVarsInstance := kubeVars{
		ConfigMapNameKafkaJaas: configMapNameKafkaJaas,
		ConfigMapPathKafkaJaas: "/etc/fink/" + configMapNameKafkaJaas,
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

	cm := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: configMapNameKafkaJaas,
		},
		Data: files,
	}
	_, err := clientSet.CoreV1().ConfigMaps("").Create(context.TODO(), &cm, metav1.CreateOptions{})
	if err != nil {
		panic(err.Error())
	}

}
