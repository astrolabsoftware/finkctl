package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

// requiredGroup is the group of an expected-topics file which is always
// required: it cannot be excluded with --without.
const requiredGroup = "topics"

// loadExpectedTopics reads an expected-topics file: named groups of topic
// names, e.g.
//
//	topics: [fink_sn_candidates_ztf, ...]
//	cds: [fink_simbad_grav_candidates_ztf, ...]
//
// The "topics" group is mandatory.
func loadExpectedTopics(path string) (map[string][]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read expected topics: %w", err)
	}
	groups := map[string][]string{}
	if err := yaml.UnmarshalStrict(content, &groups); err != nil {
		return nil, fmt.Errorf("invalid expected topics file %s: %w", path, err)
	}
	if len(groups[requiredGroup]) == 0 {
		return nil, fmt.Errorf("expected topics file %s has no %q group", path, requiredGroup)
	}
	return groups, nil
}

// requiredTopics returns the sorted topics of every group, except the groups
// in without. Excluding an unknown group, or the "topics" group, is an error:
// a typo must not silently change what is required.
func requiredTopics(groups map[string][]string, without []string) ([]string, error) {
	known := make([]string, 0, len(groups))
	for name := range groups {
		known = append(known, name)
	}
	sort.Strings(known)

	excluded := map[string]bool{}
	for _, name := range without {
		if name == requiredGroup {
			return nil, fmt.Errorf("group %q cannot be excluded", requiredGroup)
		}
		if _, ok := groups[name]; !ok {
			return nil, fmt.Errorf("unknown group %q, known groups: %s", name, strings.Join(known, ", "))
		}
		excluded[name] = true
	}

	required := []string{}
	for _, name := range known {
		if !excluded[name] {
			required = append(required, groups[name]...)
		}
	}
	sort.Strings(required)
	return required, nil
}

// compareTopics returns the required topics not found, and the found topics
// not required, both sorted.
func compareTopics(required, found []string) (missing, extra []string) {
	foundSet := map[string]bool{}
	for _, topic := range found {
		foundSet[topic] = true
	}
	requiredSet := map[string]bool{}
	for _, topic := range required {
		requiredSet[topic] = true
		if !foundSet[topic] {
			missing = append(missing, topic)
		}
	}
	for _, topic := range found {
		if !requiredSet[topic] {
			extra = append(extra, topic)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)
	return missing, extra
}
