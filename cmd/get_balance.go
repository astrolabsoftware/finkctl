/*
Copyright © 2026 AstroLab Software
*/
package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

const (
	// Topics fed by the user-defined filters. They carry no night suffix, so
	// their high watermark is cumulative over every run: messages are
	// attributed to a night by bracketing the run window with
	// offset-by-timestamp lookups.
	filterTopicPattern = "fink_.*_ztf"

	// Per-night counting topic, fed by the whole science dataframe (see
	// bin/ztf/distribute.py in fink-broker), hence one message per science
	// alert.
	countingTopicPrefix = "fink_ztf_"

	// Reported when a value could not be determined, e.g. a dataset that does
	// not exist yet.
	unknownCount = int64(-1)
)

var (
	balancePrefix      string
	balanceNight       string
	balanceCron        string
	balanceNightOffset int
)

// topicCount is the number of messages a filter topic received during one run
// window. truncated marks a lower bound: the window starts at or before the
// oldest message Kafka still retains, so older segments were purged.
type topicCount struct {
	topic     string
	count     int64
	truncated bool
}

// nightBalance is what entered and what left the broker for one observing
// night. consumed is also the number of alerts in the raw dataset (stream2raw
// filters nothing) and distributed the number in the science dataset, so the
// effect of the quality cuts reads directly between the two.
type nightBalance struct {
	night       string
	consumed    int64
	rawFiles    int64
	rawBytes    int64
	sciFiles    int64
	sciBytes    int64
	distributed int64
	windowStart time.Time
	windowEnd   time.Time
	perTopic    []topicCount
}

// getBalanceCmd represents the "get balance" command
var getBalanceCmd = &cobra.Command{
	Use:     "balance",
	Aliases: []string{"bal"},
	Short:   "Report what entered and what left the fink-broker, per observing night",
	Long: `Report, for each observing night, the number of alerts consumed from the
survey Kafka cluster, the size of the raw and science datasets on HDFS, and the
number of messages pushed to each distribution topic.

Counts are read from the sources of truth, so no Spark job is needed:
  - consumed alerts come from the offsets recorded in the stream2raw checkpoint
  - distributed alerts come from the Kafka offsets of the output topics

Columns:
  NIGHT      observing night (YYYYMMDD), one row per night found under
             <prefix>/raw
  IN(kafka)  alerts stream2raw actually read from the survey Kafka cluster.
             Taken from the last *committed* batch of its checkpoint, not from
             the offsets directory: Spark records the offsets a batch is
             planned to reach before processing it, so a run dying on its first
             batch would otherwise look complete. The survey opens a fresh
             topic every night, so the run starts at offset 0 and that end
             offset is the count.
  RAW(f)     number of parquet files written by stream2raw
  RAW        total size of those files
  SCI(f)     number of parquet files written by raw2science
  SCI        total size of those files
  DISTRIB    messages pushed to the fink_* topics, from their Kafka end
             offsets. One alert reaching several filters is counted once per
             topic.

IN(kafka) and DISTRIB are the two ends of the broker and the only pair that
compares directly -- hence the TOTAL line summing just those two. The file
counts in between are micro-batch artefacts, not alert counts: raw2science
starts after stream2raw and flushes at its own pace, so fewer SCI files than
RAW files means nothing about how many alerts got through.`,
	Run: func(cmd *cobra.Command, args []string) {
		reporter, err := newBalanceReporter(balancePrefix, balanceCron, balanceNightOffset)
		cobra.CheckErr(err)

		nights := []string{balanceNight}
		if balanceNight == "" {
			nights, err = reporter.nights()
			cobra.CheckErr(err)
		}
		if len(nights) == 0 {
			return
		}

		balances := make([]nightBalance, 0, len(nights))
		for _, night := range nights {
			balance, err := reporter.balance(night)
			cobra.CheckErr(err)
			balances = append(balances, balance)
		}

		printReport(os.Stdout, balances, reporter.prefix)
	},
}

func init() {
	getCmd.AddCommand(getBalanceCmd)

	getBalanceCmd.Flags().StringVar(&balancePrefix, "prefix", "/user/185",
		"HDFS path prefix holding the raw and science datasets")
	getBalanceCmd.Flags().StringVar(&balanceNight, "night", "",
		"Observing night (YYYYMMDD). Default: every night found under <prefix>/raw")
	getBalanceCmd.Flags().StringVar(&balanceCron, "cron", "12:00",
		"UTC time at which a run starts, i.e. the chart value scheduled.schedule")
	getBalanceCmd.Flags().IntVar(&balanceNightOffset, "night-offset", 24,
		"Hours between a night and the run processing it, i.e. the chart value scheduled.nightOffsetHours")
}

// balanceReporter holds the pods the report is collected from. They are
// resolved once, as a report runs one command per night and per dataset.
type balanceReporter struct {
	prefix      string
	cron        string
	offsetHours int
	hdfsPod     string
	kafkaPod    string
}

func newBalanceReporter(prefix string, cron string, offsetHours int) (*balanceReporter, error) {
	if _, err := time.Parse("15:04", cron); err != nil {
		return nil, fmt.Errorf("invalid cron time %q, expected HH:MM: %w", cron, err)
	}
	if offsetHours < 0 {
		return nil, fmt.Errorf("invalid night offset %d, expected a positive number of hours", offsetHours)
	}

	return &balanceReporter{
		prefix:      prefix,
		cron:        cron,
		offsetHours: offsetHours,
		hdfsPod:     resolvePod(hdfsNamespace, hdfsNameNodeSelector, hdfsPodFallback),
		kafkaPod:    resolvePod(kafkaNamespace, kafkaBrokerSelector, kafkaPodFallback),
	}, nil
}

func (r *balanceReporter) hdfs(args ...string) (string, error) {
	command := append([]string{hdfsBin, "dfs"}, args...)
	return execInPod(hdfsNamespace, r.hdfsPod, hdfsContainer, command)
}

// kafkaOffsets returns the offsets of a topic selection at the given time,
// which is -1 (latest), -2 (earliest) or a number of milliseconds since epoch.
func (r *balanceReporter) kafkaOffsets(selector []string, timeSpec string) (map[string]int64, error) {
	command := append([]string{
		kafkaOffsetsBin,
		"--bootstrap-server", kafkaBootstrapServer,
		"--time", timeSpec,
	}, selector...)

	out, err := execInPod(kafkaNamespace, r.kafkaPod, kafkaContainer, command)
	if err != nil {
		return nil, err
	}
	return parseOffsets(out), nil
}

// nights returns every observing night present in the raw dataset.
func (r *balanceReporter) nights() ([]string, error) {
	out, err := r.hdfs("-ls", path.Join(r.prefix, "raw"))
	if err != nil {
		return nil, fmt.Errorf("unable to list nights under %s/raw: %w", r.prefix, err)
	}

	nights := make([]string, 0)
	for _, name := range parseListing(out) {
		if _, err := time.Parse("20060102", name); err == nil {
			nights = append(nights, name)
		}
	}
	sort.Strings(nights)
	return nights, nil
}

func (r *balanceReporter) balance(night string) (nightBalance, error) {
	start, end, err := runWindow(night, r.cron, r.offsetHours)
	if err != nil {
		return nightBalance{}, err
	}

	balance := nightBalance{
		night:       night,
		consumed:    r.consumed(night),
		distributed: r.distributed(night),
		windowStart: start,
		windowEnd:   end,
	}
	balance.rawFiles, balance.rawBytes = r.dataset(path.Join(r.prefix, "raw", night))
	balance.sciFiles, balance.sciBytes = r.dataset(path.Join(r.prefix, "science", night))

	balance.perTopic, err = r.filterCounts(start, end)
	if err != nil {
		return nightBalance{}, err
	}
	return balance, nil
}

// consumed returns the number of alerts stream2raw read from the survey Kafka
// cluster, taken from the last committed batch of its checkpoint. The survey
// opens a fresh topic every night, so the run starts at offset 0 and the end
// offset of that batch is the number of alerts consumed.
//
// The offsets directory is not enough on its own: Spark writes the offsets a
// batch is *planned* to reach before processing it, so a run that dies on its
// first batch still leaves a full-looking offsets file. Only a batch with a
// matching commit was actually processed.
func (r *balanceReporter) consumed(night string) int64 {
	checkpoint := path.Join(r.prefix, "raw_checkpoint", night)

	out, err := r.hdfs("-ls", path.Join(checkpoint, "commits"))
	if err != nil {
		slog.Debug("no stream2raw checkpoint", "night", night, "path", checkpoint)
		return unknownCount
	}

	last := lastBatch(parseListing(out))
	if last == "" {
		slog.Debug("no committed batch", "night", night, "path", checkpoint)
		return unknownCount
	}

	out, err = r.hdfs("-cat", path.Join(checkpoint, "offsets", last))
	if err != nil {
		slog.Debug("unable to read checkpoint batch", "night", night, "batch", last)
		return unknownCount
	}

	total, err := parseCheckpointOffsets(out)
	if err != nil {
		slog.Debug("unable to parse checkpoint batch", "night", night, "batch", last, "error", err)
		return unknownCount
	}
	return total
}

// dataset returns the number of files and the size in bytes of a dataset.
func (r *balanceReporter) dataset(dataPath string) (int64, int64) {
	out, err := r.hdfs("-count", dataPath)
	if err != nil {
		slog.Debug("no such dataset", "path", dataPath)
		return unknownCount, unknownCount
	}

	files, bytes, err := parseHdfsCount(out)
	if err != nil {
		slog.Debug("unable to parse hdfs count", "path", dataPath, "error", err)
		return unknownCount, unknownCount
	}
	return files, bytes
}

// distributed returns the number of science alerts pushed to the per-night
// counting topic.
func (r *balanceReporter) distributed(night string) int64 {
	offsets, err := r.kafkaOffsets([]string{"--topic", countingTopicPrefix + night}, "-1")
	if err != nil || len(offsets) == 0 {
		slog.Debug("no counting topic", "night", night)
		return unknownCount
	}

	total := int64(0)
	for _, offset := range offsets {
		if offset != unknownCount {
			total += offset
		}
	}
	return total
}

// filterCounts returns the number of messages each filter topic received during
// the run window.
func (r *balanceReporter) filterCounts(start time.Time, end time.Time) ([]topicCount, error) {
	selector := []string{"--topic-partitions", filterTopicPattern}

	hwm, err := r.kafkaOffsets(selector, "-1")
	if err != nil {
		return nil, fmt.Errorf("unable to read the offsets of %s: %w", filterTopicPattern, err)
	}
	if len(hwm) == 0 {
		return nil, nil
	}

	earliest, err := r.kafkaOffsets(selector, "-2")
	if err != nil {
		return nil, err
	}
	startOffsets, err := r.kafkaOffsets(selector, epochMillis(start))
	if err != nil {
		return nil, err
	}
	endOffsets, err := r.kafkaOffsets(selector, epochMillis(end))
	if err != nil {
		return nil, err
	}

	return computeDeltas(hwm, earliest, startOffsets, endOffsets), nil
}

func epochMillis(t time.Time) string {
	return strconv.FormatInt(t.UnixMilli(), 10)
}

// runWindow returns the window during which the run processing the given night
// produced its messages: from the cron firing that started it to the next one.
// concurrencyPolicy=Forbid guarantees the windows do not overlap.
func runWindow(night string, cron string, offsetHours int) (time.Time, time.Time, error) {
	day, err := time.Parse("20060102", night)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid night %q, expected YYYYMMDD: %w", night, err)
	}

	fire, err := time.Parse("15:04", cron)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid cron time %q, expected HH:MM: %w", cron, err)
	}

	start := day.AddDate(0, 0, offsetHours/24).
		Add(time.Duration(fire.Hour())*time.Hour + time.Duration(fire.Minute())*time.Minute)
	return start, start.Add(24 * time.Hour), nil
}

// parseListing returns the base names of the entries listed by "hdfs dfs -ls".
func parseListing(out string) []string {
	names := make([]string, 0)

	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		// permissions replicas owner group size date time path
		if len(fields) < 8 {
			continue
		}
		names = append(names, path.Base(fields[len(fields)-1]))
	}
	return names
}

// lastBatch returns the highest numbered batch of a checkpoint offsets
// directory, which holds the offsets reached by the query.
func lastBatch(names []string) string {
	last := ""
	highest := int64(-1)

	for _, name := range names {
		batch, err := strconv.ParseInt(name, 10, 64)
		if err != nil || batch <= highest {
			continue
		}
		highest = batch
		last = name
	}
	return last
}

// parseCheckpointOffsets sums the Kafka offsets recorded in a Spark structured
// streaming offsets file. The file holds one JSON object per line: only the one
// mapping topic -> {partition: offset} is a source offset, the others are
// metadata (watermark, configuration) and fail to unmarshal into that shape.
func parseCheckpointOffsets(content string) (int64, error) {
	total := unknownCount

	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "{") {
			continue
		}

		var sourceOffsets map[string]map[string]int64
		if err := json.Unmarshal([]byte(line), &sourceOffsets); err != nil {
			continue
		}

		for _, partitions := range sourceOffsets {
			for _, offset := range partitions {
				if total == unknownCount {
					total = 0
				}
				total += offset
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return unknownCount, err
	}
	return total, nil
}

// parseHdfsCount reads the output of "hdfs dfs -count <path>", made of the
// directory count, the file count, the content size and the path.
func parseHdfsCount(out string) (int64, int64, error) {
	fields := strings.Fields(out)
	if len(fields) < 3 {
		return unknownCount, unknownCount, fmt.Errorf("unexpected hdfs count output: %q", out)
	}

	files, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return unknownCount, unknownCount, fmt.Errorf("unexpected file count in %q: %w", out, err)
	}

	bytes, err := strconv.ParseInt(fields[2], 10, 64)
	if err != nil {
		return unknownCount, unknownCount, fmt.Errorf("unexpected content size in %q: %w", out, err)
	}
	return files, bytes, nil
}

// parseOffsets reads the output of kafka-get-offsets.sh, one
// "topic:partition:offset" per line, into a map keyed by "topic:partition".
// The offset is -1 or empty when no message carries a timestamp past the
// requested bound; both are reported as unknown.
func parseOffsets(out string) map[string]int64 {
	offsets := make(map[string]int64)

	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		fields := strings.Split(strings.TrimSpace(scanner.Text()), ":")
		if len(fields) != 3 {
			continue
		}

		key := fields[0] + ":" + fields[1]
		offset, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil || offset < 0 {
			offsets[key] = unknownCount
			continue
		}
		offsets[key] = offset
	}
	return offsets
}

// computeDeltas returns, per topic, the number of messages produced between the
// two bounds. A bound that is missing or unknown falls back to the high
// watermark, meaning everything was produced before it. A topic is flagged
// truncated when its start bound sits at or before the oldest retained message:
// the count is then a lower bound, Kafka having purged the older segments.
func computeDeltas(hwm map[string]int64, earliest map[string]int64,
	start map[string]int64, end map[string]int64) []topicCount {

	totals := make(map[string]int64)
	truncated := make(map[string]bool)

	for key, high := range hwm {
		separator := strings.LastIndex(key, ":")
		if separator < 0 {
			continue
		}
		topic := key[:separator]

		from := resolveOffset(start, key, high)
		to := resolveOffset(end, key, high)
		delta := to - from
		if delta < 0 {
			delta = 0
		}

		totals[topic] += delta
		if low, ok := earliest[key]; ok && low > 0 && from <= low {
			truncated[topic] = true
		}
	}

	counts := make([]topicCount, 0, len(totals))
	for topic, count := range totals {
		counts = append(counts, topicCount{topic: topic, count: count, truncated: truncated[topic]})
	}
	sort.Slice(counts, func(i, j int) bool { return counts[i].topic < counts[j].topic })
	return counts
}

func resolveOffset(offsets map[string]int64, key string, fallback int64) int64 {
	offset, found := offsets[key]
	if !found || offset == unknownCount {
		return fallback
	}
	return offset
}

func printReport(w io.Writer, balances []nightBalance, prefix string) {
	fmt.Fprintf(w, "Balance per observing night -- prefix %s\n\n", prefix)

	table := tabwriter.NewWriter(w, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(table, "NIGHT\tIN(kafka)\tRAW(f)\tRAW\tSCI(f)\tSCI\tDISTRIB\t")

	totalConsumed, totalDistributed := int64(0), int64(0)
	for _, balance := range balances {
		fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t\n",
			balance.night,
			formatCount(balance.consumed),
			formatCount(balance.rawFiles), formatBytes(balance.rawBytes),
			formatCount(balance.sciFiles), formatBytes(balance.sciBytes),
			formatCount(balance.distributed))

		if balance.consumed != unknownCount {
			totalConsumed += balance.consumed
		}
		if balance.distributed != unknownCount {
			totalDistributed += balance.distributed
		}
	}

	fmt.Fprintf(table, "TOTAL\t%d\t\t\t\t\t%d\t\n", totalConsumed, totalDistributed)
	table.Flush()

	fmt.Fprintln(w, "\nDistribution per filter topic, per run window")

	for _, balance := range balances {
		fmt.Fprintf(w, "\n  %s  [%s -> %s]\n", balance.night,
			balance.windowStart.Format("2006-01-02T15:04Z"),
			balance.windowEnd.Format("2006-01-02T15:04Z"))

		for _, count := range balance.perTopic {
			mark := ""
			if count.truncated {
				mark = "  ~"
			}
			fmt.Fprintf(w, "    %-50s %10d%s\n", count.topic, count.count, mark)
		}
	}

	fmt.Fprintln(w, "\n  ~ = window starts at or before the oldest retained message:")
	fmt.Fprintln(w, "      undercount, the segments were purged by the Kafka retention")
}

func formatCount(value int64) string {
	if value == unknownCount {
		return "n/a"
	}
	return strconv.FormatInt(value, 10)
}

func formatBytes(value int64) string {
	if value == unknownCount {
		return "n/a"
	}

	const unit = int64(1024)
	if value < unit {
		return fmt.Sprintf("%dB", value)
	}

	div, exp := unit, 0
	for size := value / unit; size >= unit; size /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%ciB", float64(value)/float64(div), "KMGTPE"[exp])
}
