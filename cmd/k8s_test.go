package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestMain(m *testing.M) {
	verbosity = 3
	initLogger()
	code := m.Run()
	os.Exit(code)
}

// TestGetCurrentNamespace tests the getCurrentNamespace function
// NOTE: an acces to a kubernetes cluster is required to run this test
func TestGetCurrentNamespace(t *testing.T) {
	ns := getCurrentNamespace()
	// TODO Check againt the kubectl cli equivalent
	assert.Equal(t, ns, "default")
}

func TestGetKafkaPasswordFromSecret(t *testing.T) {
	secret := getKafkaPasswordFromSecret()
	// TODO Check against the kubectl cli equivalent
	t.Logf("XXXXX : %s", secret)
	assert.Equal(t, secret, "TODO")
}

func TestGetKafkaTopics(t *testing.T) {
	topics := getKafkaTopics()

	t.Logf("Kafka topics for %s namespace: %s", kafkaNamespace, topics)
	// TODO Check against the kubectl cli equivalent
}

func TestListPods(t *testing.T) {
	clientSet, _ := setKubeClient()
	namespace := "kube-system"
	selector := "component=etcd"
	logger.Infof("Wait for fink-broker pods with label '%s' to be created", selector)
	podList, _ := listPods(clientSet, namespace, selector)
	assert.Equal(t, len(podList.Items), 1, "Number of pods found should be 1")
	assert.Equal(t, podList.Items[0].Name, "etcd-kind-control-plane", "First pod found should be etcd-kind-control-plane")
}

func TestWaitForPodExistsBySelector(t *testing.T) {
	expected_pods := 1
	clientSet, _ := setKubeClient()
	namespace := "kube-system"
	selector := "component=etcd"
	logger.Infof("Wait for pods with label '%s' to be created", selector)
	err := waitForPodExistsBySelector(clientSet, namespace, selector, timeout, expected_pods)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: timed out waiting for %s pods to be created, reason: %s\n", selector, err)
		os.Exit(1)
	}
}

func TestWaitForPodReadyBySelector(t *testing.T) {
	clientSet, _ := setKubeClient()
	namespace := "kube-system"
	selector := "component=etcd"
	logger.Infof("Wait for pods with label '%s' to be created", selector)
	err := waitForPodReadyBySelector(clientSet, namespace, selector, timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: timed out waiting for %s pods to be created, reason: %s\n", selector, err)
		os.Exit(1)
	}
}
