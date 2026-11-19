/*
Copyright © 2026 AstroLab Software
*/
package cmd

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/magiconair/properties/assert"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kerr"
)

// The run window of the tests, in milliseconds since epoch
const (
	windowStart = int64(1760616000000)
	windowEnd   = windowStart + 24*3600*1000
)

// fakeAdmin answers like a Kafka cluster holding two filter topics, the
// counting topic of one night, an unrelated topic and an internal one. The
// offsets match those fakeOffsetsShell prints, so that both readers can be
// compared.
type fakeAdmin struct {
	// listErr fails the offset lookups, like an unreachable broker
	listErr error
	// calls records the offset lookups, as "end", "start" or "milli"
	calls []string
}

func (f *fakeAdmin) ListTopicsWithInternal(ctx context.Context, topics ...string) (kadm.TopicDetails, error) {
	details := kadm.TopicDetails{}
	for topic, partitions := range map[string]int32{
		"fink_early_ztf":     2,
		"fink_late_ztf":      1,
		"fink_ztf_20261016":  1,
		"ztf_20261016_other": 1,
		"__consumer_offsets": 1,
	} {
		detail := kadm.TopicDetail{Topic: topic, Partitions: kadm.PartitionDetails{}, IsInternal: strings.HasPrefix(topic, "__")}
		for p := int32(0); p < partitions; p++ {
			detail.Partitions[p] = kadm.PartitionDetail{Topic: topic, Partition: p}
		}
		details[topic] = detail
	}
	// A topic the brokers failed to describe is listed with its error
	details["fink_broken_ztf"] = kadm.TopicDetail{Topic: "fink_broken_ztf", Err: kerr.LeaderNotAvailable}
	return details, nil
}

func (f *fakeAdmin) ListStartOffsets(ctx context.Context, topics ...string) (kadm.ListedOffsets, error) {
	f.calls = append(f.calls, "start")
	return f.listed(topics, map[string]int64{
		"fink_early_ztf:0": 10, "fink_early_ztf:1": 0, "fink_late_ztf:0": 0, "fink_ztf_20261016:0": 0,
	})
}

func (f *fakeAdmin) ListEndOffsets(ctx context.Context, topics ...string) (kadm.ListedOffsets, error) {
	f.calls = append(f.calls, "end")
	return f.listed(topics, map[string]int64{
		"fink_early_ztf:0": 100, "fink_early_ztf:1": 50, "fink_late_ztf:0": 30, "fink_ztf_20261016:0": 1234,
	})
}

// ListOffsetsAfterMilli answers, as kadm does, the end offset for a partition
// with no message past the timestamp.
func (f *fakeAdmin) ListOffsetsAfterMilli(ctx context.Context, millisecond int64, topics ...string) (kadm.ListedOffsets, error) {
	f.calls = append(f.calls, "milli")
	switch millisecond {
	case windowStart:
		return f.listed(topics, map[string]int64{"fink_early_ztf:0": 40, "fink_early_ztf:1": 20, "fink_late_ztf:0": 30})
	case windowEnd:
		return f.listed(topics, map[string]int64{"fink_early_ztf:0": 90, "fink_early_ztf:1": 50, "fink_late_ztf:0": 30})
	}
	return nil, errors.New("unexpected timestamp")
}

func (f *fakeAdmin) Close() {}

// listed returns the offsets of the requested topics only, as kadm does.
func (f *fakeAdmin) listed(topics []string, offsets map[string]int64) (kadm.ListedOffsets, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	listed := kadm.ListedOffsets{}
	for _, topic := range topics {
		listed[topic] = map[int32]kadm.ListedOffset{}
		for key, offset := range offsets {
			if strings.HasPrefix(key, topic+":") {
				partition := int32(key[len(key)-1] - '0')
				listed[topic][partition] = kadm.ListedOffset{Topic: topic, Partition: partition, Offset: offset}
			}
		}
	}
	return listed, nil
}

// fakeOffsetsShell prints what kafka-get-offsets.sh prints for the cluster of
// fakeAdmin. Like the real tool, it leaves out a partition without a message
// past the timestamp.
func fakeOffsetsShell(t *testing.T) podExecutor {
	return func(namespace string, pod string, container string, command []string) (string, error) {
		timeSpec, topics := argAfter(command, "--time"), argAfter(command, "--topic")
		if topics == regexp.QuoteMeta("fink_ztf_20261016") && timeSpec == "-1" {
			return "fink_ztf_20261016:0:1234\n", nil
		}
		if topics != filterTopicPattern {
			return "Error occurred: Could not match any topic-partitions with the specified filters\n",
				errors.New("command terminated with exit code 1")
		}
		switch timeSpec {
		case "-1":
			return "fink_early_ztf:0:100\nfink_early_ztf:1:50\nfink_late_ztf:0:30\n", nil
		case "-2":
			return "fink_early_ztf:0:10\nfink_early_ztf:1:0\nfink_late_ztf:0:0\n", nil
		case "1760616000000":
			return "fink_early_ztf:0:40\nfink_early_ztf:1:20\n", nil
		case "1760702400000":
			return "fink_early_ztf:0:90\n", nil
		}
		t.Errorf("unexpected command %v", command)
		return "", errors.New("unexpected command")
	}
}

func argAfter(command []string, flag string) string {
	for i := 0; i < len(command)-1; i++ {
		if command[i] == flag {
			return command[i+1]
		}
	}
	return ""
}

// Both readers must give the same report on the same cluster.
func TestFilterCountsSameInBothModes(t *testing.T) {
	start, end := time.UnixMilli(windowStart), time.UnixMilli(windowEnd)
	want := []topicCount{
		{topic: "fink_early_ztf", count: (90 - 40) + (50 - 20)},
		{topic: "fink_late_ztf", count: 0},
	}

	for name, kafka := range map[string]kafkaReader{
		"exec":   &execKafkaReader{exec: fakeOffsetsShell(t)},
		"native": &nativeKafkaReader{admin: &fakeAdmin{}, timeout: time.Second},
	} {
		reporter := &balanceReporter{kafka: kafka}
		counts, err := reporter.filterCounts(start, end)
		if err != nil {
			t.Fatalf("%s: unexpected error: %s", name, err)
		}
		assert.Equal(t, counts, want, name)
	}
}

func TestDistributedSameInBothModes(t *testing.T) {
	for name, kafka := range map[string]kafkaReader{
		"exec":   &execKafkaReader{exec: fakeOffsetsShell(t)},
		"native": &nativeKafkaReader{admin: &fakeAdmin{}, timeout: time.Second},
	} {
		reporter := &balanceReporter{kafka: kafka}
		assert.Equal(t, reporter.distributed("20261016"), int64(1234), name)
		assert.Equal(t, reporter.distributed("20261017"), unknownCount, name)
	}
}

// The topic pattern must match the whole name, as with kafka-get-offsets.sh,
// and select neither an unrelated topic nor one the brokers failed to describe.
func TestMatchingTopics(t *testing.T) {
	details, _ := (&fakeAdmin{}).ListTopicsWithInternal(context.Background())

	names := matchingTopics(details, regexp.MustCompile("^(?:"+filterTopicPattern+")$"))
	assert.Equal(t, names, []string{"fink_early_ztf", "fink_late_ztf"})

	names = matchingTopics(details, regexp.MustCompile("^(?:fink_ztf)$"))
	assert.Equal(t, names, []string{})
}

// A partition without an offset (-1), one that failed to load and the
// placeholder kadm adds for a missing topic are all left out.
func TestListedOffsets(t *testing.T) {
	listed := kadm.ListedOffsets{
		"fink_early_ztf": {
			0: {Topic: "fink_early_ztf", Partition: 0, Offset: 42},
			1: {Topic: "fink_early_ztf", Partition: 1, Offset: -1},
			2: {Topic: "fink_early_ztf", Partition: 2, Offset: 7, Err: kerr.NotLeaderForPartition},
		},
		"fink_gone_ztf": {
			-1: {Topic: "fink_gone_ztf", Partition: -1, Err: kerr.UnknownTopicOrPartition},
		},
	}

	assert.Equal(t, listedOffsets(listed), map[string]int64{"fink_early_ztf:0": 42})
}

// No matching topic is an empty result, and no offset lookup is made: kadm
// would list the offsets of every topic.
func TestNativeOffsetsWithoutMatchingTopic(t *testing.T) {
	admin := &fakeAdmin{}
	reader := &nativeKafkaReader{admin: admin, timeout: time.Second}

	offsets, err := reader.offsets("fink_.*_lsst", offsetLatest)

	assert.Equal(t, err, nil)
	assert.Equal(t, offsets, map[string]int64{})
	assert.Equal(t, len(admin.calls), 0)
}

func TestNativeOffsetsSelectsTheLookup(t *testing.T) {
	admin := &fakeAdmin{}
	reader := &nativeKafkaReader{admin: admin, timeout: time.Second}

	for _, at := range []int64{offsetLatest, offsetEarliest, windowStart} {
		if _, err := reader.offsets(filterTopicPattern, at); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
	}
	assert.Equal(t, admin.calls, []string{"end", "start", "milli"})
}

// A failed lookup fails the report.
func TestNativeOffsetsFailsOnLookupError(t *testing.T) {
	reader := &nativeKafkaReader{admin: &fakeAdmin{listErr: errors.New("connection refused")}, timeout: time.Second}

	_, err := reader.offsets(filterTopicPattern, offsetLatest)

	assert.Equal(t, err != nil, true)
}

func TestNativeOffsetsRejectsInvalidPattern(t *testing.T) {
	reader := &nativeKafkaReader{admin: &fakeAdmin{}, timeout: time.Second}

	_, err := reader.offsets("fink_(", offsetLatest)

	assert.Equal(t, err != nil, true)
}

// The exec reader runs kafka-get-offsets.sh with the time and the topic
// pattern it was given.
func TestExecOffsetsCommand(t *testing.T) {
	var got []string
	reader := &execKafkaReader{pod: "kafka-cluster-dual-role-0", exec: func(namespace string, pod string, container string, command []string) (string, error) {
		got = command
		return "fink_early_ztf:0:100\n", nil
	}}

	offsets, err := reader.offsets(filterTopicPattern, offsetEarliest)

	assert.Equal(t, err, nil)
	assert.Equal(t, offsets, map[string]int64{"fink_early_ztf:0": 100})
	assert.Equal(t, got, []string{kafkaOffsetsBin, "--bootstrap-server", kafkaBootstrapServer,
		"--time", "-2", "--topic", filterTopicPattern})
}

// An unreachable broker must fail the lookup within the timeout, not hang the
// report.
func TestNativeReaderUnreachableBroker(t *testing.T) {
	reader, err := newNativeKafkaReader([]string{"127.0.0.1:1"})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	defer reader.close()
	reader.timeout = 2 * time.Second

	start := time.Now()
	if _, err := reader.offsets(filterTopicPattern, offsetLatest); err == nil {
		t.Fatal("expected an error on an unreachable broker")
	}
	if elapsed := time.Since(start); elapsed > reader.timeout+5*time.Second {
		t.Errorf("lookup took %s, beyond the timeout", elapsed)
	}
}

// A bootstrap list with spaces after the commas, or a trailing comma, gives
// clean host:port entries.
func TestSplitHosts(t *testing.T) {
	got := splitHosts(" a:9092, b:9092 ,,c:9092,")
	want := []string{"a:9092", "b:9092", "c:9092"}
	if len(got) != len(want) {
		t.Fatalf("splitHosts = %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("splitHosts = %q, want %q", got, want)
		}
	}
}
