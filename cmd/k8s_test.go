package cmd

import (
	"os"
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func skipk8s(t *testing.T) {
	// TODO Check against the kubectl cli equivalent?
	if os.Getenv("FINKCTL_UTEST_K8S") == "" {
		t.Skip("Skipping Kubernetes tests")
	}
}

func TestGetKafkaTopics(t *testing.T) {
	skipk8s(t)
	topics, err := getKafkaTopics()
	if err != nil {
		t.Errorf("Error getting Kafka topics: %s", err)
	}

	t.Logf("Kafka topics for %s namespace: %s", kafkaNamespace, topics)
	// TODO Check against the kubectl cli equivalent
}

func TestListPods(t *testing.T) {
	skipk8s(t)
	clientSet, _ := setKubeClient()
	namespace := "kube-system"
	selector := "component=etcd"
	t.Logf("Wait for fink-broker pods with label '%s' to be created", selector)
	podList, _ := listPods(clientSet, namespace, selector)
	assert.Equal(t, len(podList.Items), 1, "Number of pods found should be 1")
	assert.Equal(t, podList.Items[0].Name, "etcd-kind-control-plane", "First pod found should be etcd-kind-control-plane")
}
