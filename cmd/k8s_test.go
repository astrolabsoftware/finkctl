package cmd

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

// TestGetCurrentNamespace tests the getCurrentNamespace function
// NOTE: an acces to a kubernetes cluster is required to run this test
func TestGetCurrentNamespace(t *testing.T) {
	ns := getCurrentNamespace()
	assert.Equal(t, ns, "default")
}
