package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

const (
	kubeConfigEnvName         = "KUBECONFIG"
	kubeConfigDefaultFilename = ".kube/config"
)

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

// getKubeRestConfig returns the client configuration, taken from the service
// account when running inside a pod and from the kubeconfig file otherwise.
// finkctl runs both ways: interactively from a workstation, and as a CronJob
// in the cluster, where no kubeconfig is mounted.
func getKubeRestConfig() (*rest.Config, error) {
	if config, err := rest.InClusterConfig(); err == nil {
		slog.Debug("using in-cluster kubernetes configuration")
		return config, nil
	}

	kubeconfig := getKubeConfig()
	slog.Debug("using kubeconfig file", "path", kubeconfig)
	// use the current context in kubeconfig
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

func setKubeClient() (*kubernetes.Clientset, *rest.Config) {

	config, err := getKubeRestConfig()
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

// resolvePod returns the name of the first pod matching the selector in the
// namespace, falling back to the well-known name when the lookup returns
// nothing, so the command keeps working if the operator labels ever change.
func resolvePod(namespace string, selector string, fallback string) string {

	clientSet, _ := setKubeClient()

	podList, err := listPods(clientSet, namespace, selector)
	if err != nil || len(podList.Items) == 0 {
		slog.Debug("no pod matching selector, using well-known name",
			"namespace", namespace, "selector", selector, "pod", fallback)
		return fallback
	}
	return podList.Items[0].Name
}

// execInPod runs a command inside a container of a running pod and returns its
// combined output, the equivalent of
// "kubectl exec -n <namespace> <pod> -c <container> -- <command>"
func execInPod(namespace string, pod string, container string, command []string) (string, error) {

	clientSet, config := setKubeClient()

	slog.Debug("executing command in pod", "namespace", namespace, "pod", pod, "command", command)

	req := clientSet.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(pod).
		Namespace(namespace).
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
		return "", fmt.Errorf("unable to create executor for pod %s/%s: %w", namespace, pod, err)
	}

	// Capture the output
	output := &strings.Builder{}
	err = exec.StreamWithContext(context.TODO(), remotecommand.StreamOptions{
		Stdout: output,
		Stderr: output,
	})
	if err != nil {
		return output.String(), fmt.Errorf("command %v failed in pod %s/%s: %w", command, namespace, pod, err)
	}

	return output.String(), nil
}

// getKafkaTopic returns the kafka topics
// equivalent to "kubectl get -n kafka kafkatopics.kafka.strimzi.io --template='{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}'"
func getKafkaTopics() ([]string, error) {

	// Command to list Kafka topics
	command := []string{
		kafkaTopicsBin,
		"--bootstrap-server", kafkaBootstrapServer,
		"--list",
	}

	pod := resolvePod(kafkaNamespace, kafkaBrokerSelector, kafkaPodFallback)

	out, err := execInPod(kafkaNamespace, pod, kafkaContainer, command)
	if err != nil {
		return nil, fmt.Errorf("cannot list the kafka topics: %w, output: %s", err, strings.TrimSpace(out))
	}
	// Filter the output for topics containing the filter string
	topicNames := make([]string, 0)
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		topicNames = append(topicNames, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("cannot read the kafka topics: %w", err)
	}
	slog.Debug("Kafka topics", "topics", topicNames)

	return topicNames, nil
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
