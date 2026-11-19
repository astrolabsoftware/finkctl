/*
Copyright © 2026 AstroLab Software
*/
package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

// Special times of an offset lookup, as kafka-get-offsets.sh --time takes
// them. Any other value is a number of milliseconds since epoch.
const (
	offsetLatest   = int64(-1)
	offsetEarliest = int64(-2)
)

// kafkaRequestTimeout bounds each lookup of the native client, so that an
// unreachable Kafka fails the report instead of hanging it.
const kafkaRequestTimeout = 30 * time.Second

// kafkaReader reads the offsets of the topics whose whole name matches a
// regular expression. It lets get balance read Kafka either natively or
// through a broker pod, and be tested without Kafka.
type kafkaReader interface {
	// offsets returns, keyed by "topic:partition", the offset of every
	// partition of the matching topics at the given time: offsetLatest,
	// offsetEarliest or milliseconds since epoch. A partition without such
	// an offset is left out. No matching topic gives an empty map, not an
	// error, as before the broker distributes its first alert.
	offsets(topics string, at int64) (map[string]int64, error)
	close()
}

// noMatchingTopicMsg is printed by kafka-get-offsets.sh, which then exits 1,
// when no topic matches its selection
const noMatchingTopicMsg = "Could not match any topic-partitions"

// podExecutor runs a command in a container of a pod and returns its output
type podExecutor func(namespace string, pod string, container string, command []string) (string, error)

// execKafkaReader runs kafka-get-offsets.sh in a broker pod. It needs the
// pods/exec right in the kafka namespace and starts a JVM on every lookup, but
// works from outside the cluster, e.g. from the CI runner.
type execKafkaReader struct {
	pod string
	// exec runs a command in a pod, execInPod when nil; tests replace it
	exec podExecutor
}

func newExecKafkaReader() *execKafkaReader {
	return &execKafkaReader{pod: resolvePod(kafkaNamespace, kafkaBrokerSelector, kafkaPodFallback)}
}

func (r *execKafkaReader) offsets(topics string, at int64) (map[string]int64, error) {
	// --topic takes a regular expression matched against the whole topic
	// name, and selects every partition
	command := []string{
		kafkaOffsetsBin,
		"--bootstrap-server", kafkaBootstrapServer,
		"--time", strconv.FormatInt(at, 10),
		"--topic", topics,
	}

	exec := r.exec
	if exec == nil {
		exec = execInPod
	}
	out, err := exec(kafkaNamespace, r.pod, kafkaContainer, command)
	if err != nil && strings.Contains(out, noMatchingTopicMsg) {
		// No topic matches the selection yet, e.g. before the broker
		// distributes its first alert: there is no offset, not an error.
		slog.Info("no Kafka topic available yet", "topics", topics)
		return map[string]int64{}, nil
	}
	if err != nil {
		return nil, err
	}
	return parseOffsets(out), nil
}

func (r *execKafkaReader) close() {}

// kafkaAdmin is the part of the kadm client the native reader uses, so that
// tests can answer in its place.
type kafkaAdmin interface {
	ListTopicsWithInternal(ctx context.Context, topics ...string) (kadm.TopicDetails, error)
	ListStartOffsets(ctx context.Context, topics ...string) (kadm.ListedOffsets, error)
	ListEndOffsets(ctx context.Context, topics ...string) (kadm.ListedOffsets, error)
	ListOffsetsAfterMilli(ctx context.Context, millisecond int64, topics ...string) (kadm.ListedOffsets, error)
	Close()
}

// nativeKafkaReader talks to the Kafka brokers with the franz-go client. In
// the cluster, the plain listener needs no credentials.
type nativeKafkaReader struct {
	admin   kafkaAdmin
	timeout time.Duration
}

func newNativeKafkaReader(bootstrap []string) (*nativeKafkaReader, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(bootstrap...),
		kgo.DialTimeout(10*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create a Kafka client for %s: %w", strings.Join(bootstrap, ","), err)
	}
	return &nativeKafkaReader{admin: kadm.NewClient(client), timeout: kafkaRequestTimeout}, nil
}

func (r *nativeKafkaReader) offsets(topics string, at int64) (map[string]int64, error) {
	// Anchored, as kafka-get-offsets.sh matches the whole topic name
	pattern, err := regexp.Compile("^(?:" + topics + ")$")
	if err != nil {
		return nil, fmt.Errorf("invalid topic pattern %q: %w", topics, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	details, err := r.admin.ListTopicsWithInternal(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to list the Kafka topics: %w", err)
	}
	names := matchingTopics(details, pattern)
	if len(names) == 0 {
		// kadm lists every topic when given none, so stop here
		slog.Info("no Kafka topic available yet", "topics", topics)
		return map[string]int64{}, nil
	}

	var listed kadm.ListedOffsets
	switch at {
	case offsetLatest:
		listed, err = r.admin.ListEndOffsets(ctx, names...)
	case offsetEarliest:
		listed, err = r.admin.ListStartOffsets(ctx, names...)
	default:
		listed, err = r.admin.ListOffsetsAfterMilli(ctx, at, names...)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to read the offsets of %s: %w", topics, err)
	}
	return listedOffsets(listed), nil
}

func (r *nativeKafkaReader) close() {
	r.admin.Close()
}

// matchingTopics returns the topics whose name matches the pattern, leaving
// out those the brokers failed to describe.
func matchingTopics(details kadm.TopicDetails, pattern *regexp.Regexp) []string {
	names := make([]string, 0)
	for _, name := range details.Names() {
		detail := details[name]
		if detail.Err != nil || !pattern.MatchString(name) {
			continue
		}
		names = append(names, name)
	}
	return names
}

// listedOffsets turns a kadm listing into a map keyed by "topic:partition",
// the shape parseOffsets gives for kafka-get-offsets.sh. Like that tool, it
// skips a partition whose offset could not be read, with a warning.
// kafka-get-offsets.sh leaves out a partition with no message after the
// timestamp, while kadm returns its end offset instead; computeDeltas falls
// back to the end offset in that case, so both give the same count. The
// negative-offset check is only a guard.
func listedOffsets(listed kadm.ListedOffsets) map[string]int64 {
	offsets := make(map[string]int64)
	listed.Each(func(o kadm.ListedOffset) {
		if o.Err != nil {
			slog.Warn("skip the offset of a Kafka partition", "topic", o.Topic, "partition", o.Partition, "error", o.Err)
			return
		}
		if o.Offset < 0 {
			return
		}
		offsets[fmt.Sprintf("%s:%d", o.Topic, o.Partition)] = o.Offset
	})
	return offsets
}
