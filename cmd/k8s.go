package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	kafka "github.com/RedHatInsights/strimzi-client-go/apis/kafka.strimzi.io/v1beta2"
	"github.com/astrolabsoftware/finkctl/resources"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	applyv1 "k8s.io/client-go/applyconfigurations/core/v1"
	applymetav1 "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	podv1 "k8s.io/kubernetes/pkg/api/v1/pod"
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

	cm := applyv1.ConfigMapApplyConfiguration{
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
	url := fmt.Sprintf("/apis/kafka.strimzi.io/v1beta2/namespaces/%s/kafkatopics", kafkaNamespace)
	d, err := clientSet.RESTClient().Get().AbsPath(url).DoRaw(context.TODO())
	if err != nil {
		panic(err.Error())
	}
	if err := json.Unmarshal(d, &topics); err != nil {
		panic(err.Error())
	}

	topicNames := make([]string, len(topics.Items))
	for _, topic := range topics.Items {
		if topic.Namespace == kafkaNamespace {
			topicNames = append(topicNames, topic.Name)
		}
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

func waitForPodReady(ctx context.Context, clientset *kubernetes.Clientset, pod *v1.Pod, timeout time.Duration) error {
	logger.Infof("waiting for pod %s to be running...", pod.Name)
	return wait.PollUntilContextTimeout(ctx, 5*time.Second, timeout, true, func(context context.Context) (bool, error) {
		job, err := clientset.CoreV1().Pods(pod.Namespace).Get(ctx, pod.Name, metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("failed to detect pod %s. %+v", pod.Name, err)
		}

		if podv1.IsPodTerminal(pod) {
			return false, fmt.Errorf("job %s failed", job.Name)
		}
		if podv1.IsPodReady(pod) {
			return true, nil
		}
		logger.Debugf("pod is still initializing")
		return false, nil
	})
}

// Returns the list of currently scheduled or running pods in `namespace` with the given selector
func listPods(c *kubernetes.Clientset, namespace, selector string) (*v1.PodList, error) {
	listOptions := metav1.ListOptions{LabelSelector: selector}
	podList, err := c.CoreV1().Pods(namespace).List(context.TODO(), listOptions)
	if err != nil {
		return nil, err
	}
	return podList, nil
}

// Wait up to timeout seconds for all pods in 'namespace' with given 'selector' to enter running state.
// Returns an error if no pods are found or not all discovered pods enter running state.
func waitForPodReadyBySelector(c *kubernetes.Clientset, namespace, selector string, timeout time.Duration) error {
	podList, err := listPods(c, namespace, selector)
	if err != nil {
		return err
	}
	if len(podList.Items) == 0 {
		return fmt.Errorf("no pods in %s with selector %s", namespace, selector)
	}

	for _, pod := range podList.Items {
		if err := waitForPodReady(context.TODO(), c, &pod, timeout); err != nil {
			return err
		}
	}
	return nil
}

func waitForPodExistsBySelector(c *kubernetes.Clientset, namespace, selector string, timeout time.Duration, expected int) error {
	allPodsExists := make(chan bool, 1)

	go func() {
		for {
			podList, _ := listPods(c, namespace, selector)
			podCount := len(podList.Items)
			logger.Debugf("Found %d pods with label %s", podCount, selector)
			if podCount == expected {
				allPodsExists <- true
				return
			}
			time.Sleep(time.Millisecond * 500)
		}
	}()

	select {
	case <-allPodsExists:
		fmt.Printf("Condition met: Found %d pods with label %s", expected, selector)
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("error: timed out waiting for pods with label %s", selector)
	}
}
