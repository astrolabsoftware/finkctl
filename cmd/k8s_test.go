package cmd

import (
	"fmt"
	"testing"

	"github.com/magiconair/properties/assert"
)

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
	fmt.Printf("XXXXX : %s", secret)
	assert.Equal(t, secret, "TODO")
}

func TestGetKafkaTopics(t *testing.T) {
	topics := getKafkaTopics()

	t.Logf("Kafka topics: %s", topics)
	// TODO Check against the kubectl cli equivalent
}
