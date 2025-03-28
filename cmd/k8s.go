package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"
	"time"

	"github.com/astrolabsoftware/finkctl/v3/resources"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	applyv1 "k8s.io/client-go/applyconfigurations/core/v1"
	applymetav1 "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	podv1 "k8s.io/kubernetes/pkg/api/v1/pod"
)

const (
	kubeConfigEnvName         = "KUBECONFIG"
	kubeConfigDefaultFilename = ".kube/config"
	secretNameKafkaJaas       = "fink-kafka-jaas"
	secretPathKafkaJaas       = "/etc/fink-broker"
)

type kubeVars struct {
	SecretNameKafkaJaas string
	SecretPathKafkaJaas string
}

func getKubeVars() kubeVars {
	kubeVarsInstance := kubeVars{
		SecretNameKafkaJaas: secretNameKafkaJaas,
		SecretPathKafkaJaas: secretPathKafkaJaas,
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

func createKafkaJaasSecret(c *DistributionConfig, configMap bool) {
	kafkaJaasConf := format(resources.KafkaJaasConf, &c)

	clientSet, _ := setKubeClient()

	files := make(map[string]string)

	files[resources.KafkaJaasConfFile] = kafkaJaasConf

	var kind string
	version := "v1"
	name := secretNameKafkaJaas

	var err error
	if configMap {
		kind = "ConfigMap"
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
		_, err = clientSet.CoreV1().ConfigMaps(getCurrentNamespace()).Apply(
			context.TODO(), &cm,
			metav1.ApplyOptions{FieldManager: "application/apply-patch"})
	} else {
		kind = "Secret"
		secret := applyv1.SecretApplyConfiguration{
			TypeMetaApplyConfiguration: applymetav1.TypeMetaApplyConfiguration{
				Kind:       &kind,
				APIVersion: &version,
			},
			ObjectMetaApplyConfiguration: &applymetav1.ObjectMetaApplyConfiguration{
				Name: &name,
			},
			Data: convertToBytes(files),
		}
		_, err = clientSet.CoreV1().Secrets(getCurrentNamespace()).Apply(
			context.TODO(), &secret,
			metav1.ApplyOptions{FieldManager: "application/apply-patch"})
		slog.Debug("Secret created", "name", name, "secret", secret)
	}
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
func getKafkaTopics() ([]string, error) {

	clientSet, config := setKubeClient()
	pod := "kafka-cluster-dual-role-0"
	container := "kafka"

	// Command to list Kafka topics
	command := []string{
		"bin/kafka-topics.sh",
		"--bootstrap-server", "kafka-cluster-kafka-bootstrap.kafka:9092",
		"--list",
	}

	// Execute the command in the specified pod and container
	req := clientSet.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(pod).
		Namespace(kafkaNamespace).
		SubResource("exec").
		Param("container", container).
		Param("stdin", "false").
		Param("stdout", "true").
		Param("stderr", "true").
		Param("tty", "false")

	for _, arg := range command {
		req.Param("command", arg)
	}

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		fmt.Printf("Error creating executor: %v\n", err)
		return nil, err
	}

	// Capture the output
	output := &strings.Builder{}
	err = exec.StreamWithContext(context.TODO(), remotecommand.StreamOptions{
		Stdout: output,
		Stderr: output,
	})
	if err != nil {
		fmt.Printf("Error executing command: %v\n", err)
		fmt.Println(output.String())
		return nil, err
	}

	// Filter the output for topics containing the filter string
	topicNames := make([]string, 0)
	scanner := bufio.NewScanner(strings.NewReader(output.String()))
	fmt.Println("Filtered Kafka topics:")
	for scanner.Scan() {
		line := scanner.Text()
		topicNames = append(topicNames, line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading output: %v\n", err)
	}

	return topicNames, nil
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
	slog.Info("waiting for pod to be running", "podName", pod.Name)
	return wait.PollUntilContextTimeout(ctx, 5*time.Second, timeout, true, func(context context.Context) (bool, error) {
		pod, err := clientset.CoreV1().Pods(pod.Namespace).Get(ctx, pod.Name, metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("failed to detect pod %s. %+v", pod.Name, err)
		}

		if podv1.IsPodTerminal(pod) {
			return false, fmt.Errorf("job %s failed", pod.Name)
		}
		if podv1.IsPodReady(pod) {
			return true, nil
		}
		slog.Debug("pod is still initializing")
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
			slog.Debug("Found pods with label", "podCount", podCount, "selector", selector)
			// FIXME - check expected executor count
			if podCount >= expected {
				allPodsExists <- true
				return
			}
			time.Sleep(time.Millisecond * 500)
		}
	}()

	select {
	case <-allPodsExists:
		slog.Debug("Condition met: Found pods with label", "podCount", expected, "selector", selector)
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("error: timed out waiting for pods with label %s", selector)
	}
}
