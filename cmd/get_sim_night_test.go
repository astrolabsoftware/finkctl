package cmd

import (
	"errors"
	"strings"
	"testing"
)

func TestParseSimNight(t *testing.T) {
	night, err := parseSimNight([]string{"fink_sn_candidates_ztf", "sim_ztf_public_20200101", "__consumer_offsets"})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if night != "20200101" {
		t.Errorf("got %q, want 20200101", night)
	}
}

// Look-alikes are not simulator topics: a real ZTF topic, a short night, a
// suffix.
func TestParseSimNightNone(t *testing.T) {
	_, err := parseSimNight([]string{"ztf_20200101", "sim_ztf_public_2020010", "sim_ztf_public_20200101_x"})
	if !errors.Is(err, errNoSimTopic) {
		t.Errorf("expected errNoSimTopic, got %v", err)
	}
}

func TestParseSimNightSeveral(t *testing.T) {
	_, err := parseSimNight([]string{"sim_ztf_public_20200101", "sim_ztf_public_20200102"})
	if err == nil || errors.Is(err, errNoSimTopic) {
		t.Fatalf("expected an error other than errNoSimTopic, got %v", err)
	}
	for _, topic := range []string{"sim_ztf_public_20200101", "sim_ztf_public_20200102"} {
		if !strings.Contains(err.Error(), topic) {
			t.Errorf("error %q does not name %s", err, topic)
		}
	}
}
