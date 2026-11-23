package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func writeExpectedFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "expected-topics.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("cannot write %s: %s", path, err)
	}
	return path
}

func TestLoadExpectedTopics(t *testing.T) {
	path := writeExpectedFile(t, "topics: [a, b]\ncds: [c]\n")
	groups, err := loadExpectedTopics(path)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	want := map[string][]string{"topics": {"a", "b"}, "cds": {"c"}}
	if !reflect.DeepEqual(groups, want) {
		t.Errorf("got %v, want %v", groups, want)
	}
}

func TestLoadExpectedTopicsWithoutTopicsGroup(t *testing.T) {
	path := writeExpectedFile(t, "cds: [c]\n")
	_, err := loadExpectedTopics(path)
	if err == nil || !strings.Contains(err.Error(), `"topics"`) {
		t.Errorf("expected an error naming the \"topics\" group, got %v", err)
	}
}

func TestLoadExpectedTopicsMissingFile(t *testing.T) {
	_, err := loadExpectedTopics(filepath.Join(t.TempDir(), "absent.yaml"))
	if err == nil {
		t.Error("expected an error for a missing file")
	}
}

func TestLoadExpectedTopicsInvalidYAML(t *testing.T) {
	path := writeExpectedFile(t, "topics: [a, b\n")
	_, err := loadExpectedTopics(path)
	if err == nil {
		t.Error("expected an error for invalid YAML")
	}
}

func TestRequiredTopics(t *testing.T) {
	groups := map[string][]string{"topics": {"a", "b"}, "cds": {"c"}}

	all, err := requiredTopics(groups, nil)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if want := []string{"a", "b", "c"}; !reflect.DeepEqual(all, want) {
		t.Errorf("without nothing: got %v, want %v", all, want)
	}

	withoutCds, err := requiredTopics(groups, []string{"cds"})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if want := []string{"a", "b"}; !reflect.DeepEqual(withoutCds, want) {
		t.Errorf("without cds: got %v, want %v", withoutCds, want)
	}
}

func TestRequiredTopicsUnknownGroup(t *testing.T) {
	groups := map[string][]string{"topics": {"a"}, "cds": {"c"}}
	_, err := requiredTopics(groups, []string{"CDS"})
	if err == nil {
		t.Fatal("expected an error for an unknown group")
	}
	for _, name := range []string{"CDS", "cds", "topics"} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error %q does not name %q", err, name)
		}
	}
}

func TestRequiredTopicsCannotDropTopics(t *testing.T) {
	groups := map[string][]string{"topics": {"a"}, "cds": {"c"}}
	if _, err := requiredTopics(groups, []string{"topics"}); err == nil {
		t.Error("expected an error when excluding the topics group")
	}
}

func TestCompareTopics(t *testing.T) {
	missing, extra := compareTopics([]string{"a", "b", "c"}, []string{"b", "x", "a"})
	if want := []string{"c"}; !reflect.DeepEqual(missing, want) {
		t.Errorf("missing: got %v, want %v", missing, want)
	}
	if want := []string{"x"}; !reflect.DeepEqual(extra, want) {
		t.Errorf("extra: got %v, want %v", extra, want)
	}
}
