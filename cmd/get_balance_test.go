package cmd

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/magiconair/properties/assert"
)

// A Spark structured streaming offsets file: a version marker, a metadata
// object, then the source offsets. Only the last line is a topic offset.
const checkpointBatch = `v1
{"batchWatermarkMs":0,"batchTimestampMs":1755511234567,"conf":{"spark.sql.streaming.join.stateFormatVersion":"2"}}
{"ztf_20260810_programid1":{"0":1841233,"1":1839902}}`

func TestParseCheckpointOffsets(t *testing.T) {
	total, err := parseCheckpointOffsets(checkpointBatch)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	assert.Equal(t, total, int64(3681135))
}

// A checkpoint holding no source offset must not be reported as zero alert.
func TestParseCheckpointOffsetsWithoutSource(t *testing.T) {
	total, err := parseCheckpointOffsets("v1\n{\"batchWatermarkMs\":0}")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	assert.Equal(t, total, unknownCount)
}

func TestParseHdfsCount(t *testing.T) {
	files, bytes, err := parseHdfsCount("           2            42          4823456789 /user/185/raw/20260810")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	assert.Equal(t, files, int64(42))
	assert.Equal(t, bytes, int64(4823456789))
}

func TestParseHdfsCountRejectsUnexpectedOutput(t *testing.T) {
	if _, _, err := parseHdfsCount("ls: `/user/185/raw/20260810': No such file or directory"); err == nil {
		t.Error("expected an error on an unexpected hdfs count output")
	}
}

func TestParseListingAndLastBatch(t *testing.T) {
	listing := `Found 3 items
-rw-r--r--   2 185 supergroup        432 2026-08-11 12:34 /user/185/raw_checkpoint/20260810/offsets/0
-rw-r--r--   2 185 supergroup        436 2026-08-11 18:02 /user/185/raw_checkpoint/20260810/offsets/9
-rw-r--r--   2 185 supergroup        436 2026-08-12 05:57 /user/185/raw_checkpoint/20260810/offsets/12`

	names := parseListing(listing)
	assert.Equal(t, names, []string{"0", "9", "12"})
	// Batches are numbered, not zero-padded: they must be ordered numerically.
	assert.Equal(t, lastBatch(names), "12")
}

// A run that died before committing anything leaves an empty commits
// directory: no batch was processed, which must not read as zero alert.
func TestLastBatchWithoutCommit(t *testing.T) {
	assert.Equal(t, lastBatch(parseListing("")), "")
}

func TestParseOffsets(t *testing.T) {
	offsets := parseOffsets(`fink_early_sn_candidates_ztf:0:1750
fink_sso_ztf_candidates_ztf:0:-1
fink_kn_candidates_ztf:0:
garbage line`)

	assert.Equal(t, len(offsets), 3)
	assert.Equal(t, offsets["fink_early_sn_candidates_ztf:0"], int64(1750))
	// An empty offset must be treated like -1, not like offset zero.
	assert.Equal(t, offsets["fink_sso_ztf_candidates_ztf:0"], unknownCount)
	assert.Equal(t, offsets["fink_kn_candidates_ztf:0"], unknownCount)
}

func TestComputeDeltas(t *testing.T) {
	hwm := map[string]int64{"fink_early_sn_candidates_ztf:0": 5000, "fink_sso_ztf_candidates_ztf:0": 120}
	earliest := map[string]int64{"fink_early_sn_candidates_ztf:0": 0, "fink_sso_ztf_candidates_ztf:0": 100}
	start := map[string]int64{"fink_early_sn_candidates_ztf:0": 1000, "fink_sso_ztf_candidates_ztf:0": 100}
	end := map[string]int64{"fink_early_sn_candidates_ztf:0": 1750, "fink_sso_ztf_candidates_ztf:0": 110}

	counts := computeDeltas(hwm, earliest, start, end)

	assert.Equal(t, counts, []topicCount{
		{topic: "fink_early_sn_candidates_ztf", count: 750, truncated: false},
		// Its window starts at the oldest retained message: lower bound.
		{topic: "fink_sso_ztf_candidates_ztf", count: 10, truncated: true},
	})
}

// The run of the current night has no end bound yet: no message carries a
// timestamp past it, so the count runs up to the high watermark.
func TestComputeDeltasWithoutEndBound(t *testing.T) {
	hwm := map[string]int64{"fink_early_sn_candidates_ztf:0": 5000}
	earliest := map[string]int64{"fink_early_sn_candidates_ztf:0": 0}
	start := map[string]int64{"fink_early_sn_candidates_ztf:0": 4200}
	end := map[string]int64{"fink_early_sn_candidates_ztf:0": unknownCount}

	counts := computeDeltas(hwm, earliest, start, end)

	assert.Equal(t, counts, []topicCount{
		{topic: "fink_early_sn_candidates_ztf", count: 800, truncated: false},
	})
}

func TestComputeDeltasSumsPartitions(t *testing.T) {
	hwm := map[string]int64{"fink_early_sn_candidates_ztf:0": 5000, "fink_early_sn_candidates_ztf:1": 3000}
	earliest := map[string]int64{"fink_early_sn_candidates_ztf:0": 0, "fink_early_sn_candidates_ztf:1": 0}
	start := map[string]int64{"fink_early_sn_candidates_ztf:0": 1000, "fink_early_sn_candidates_ztf:1": 500}
	end := map[string]int64{"fink_early_sn_candidates_ztf:0": 1750, "fink_early_sn_candidates_ztf:1": 900}

	counts := computeDeltas(hwm, earliest, start, end)

	assert.Equal(t, counts, []topicCount{
		{topic: "fink_early_sn_candidates_ztf", count: 1150, truncated: false},
	})
}

// A topic created after the window: every message it holds was produced later,
// so both bounds land on the same offset.
func TestComputeDeltasIgnoresLaterTopic(t *testing.T) {
	hwm := map[string]int64{"fink_new_filter_ztf:0": 40}
	earliest := map[string]int64{"fink_new_filter_ztf:0": 0}
	start := map[string]int64{"fink_new_filter_ztf:0": 0}
	end := map[string]int64{"fink_new_filter_ztf:0": 0}

	counts := computeDeltas(hwm, earliest, start, end)

	assert.Equal(t, counts, []topicCount{
		{topic: "fink_new_filter_ztf", count: 0, truncated: false},
	})
}

// With the default 24h offset, the run processing a night fires at the cron of
// the following day and ends at the next one.
func TestRunWindow(t *testing.T) {
	start, end, err := runWindow("20260810", "12:00", 24)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	assert.Equal(t, start, time.Date(2026, 8, 11, 12, 0, 0, 0, time.UTC))
	assert.Equal(t, end, time.Date(2026, 8, 12, 12, 0, 0, 0, time.UTC))
}

// With no offset the run processes the live night, starting the same day.
func TestRunWindowWithoutOffset(t *testing.T) {
	start, _, err := runWindow("20260810", "18:30", 0)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	assert.Equal(t, start, time.Date(2026, 8, 10, 18, 30, 0, 0, time.UTC))
}

func TestRunWindowRejectsInvalidNight(t *testing.T) {
	if _, _, err := runWindow("2026-08-10", "12:00", 24); err == nil {
		t.Error("expected an error on a night that is not YYYYMMDD")
	}
}

func TestFormatBytes(t *testing.T) {
	assert.Equal(t, formatBytes(unknownCount), "n/a")
	assert.Equal(t, formatBytes(512), "512B")
	assert.Equal(t, formatBytes(4823456789), "4.5GiB")
}

// fakeHdfs is an in-memory hdfsReader: dirs maps a directory to its entries,
// files a file to its content, sizes a dataset to its file count and size.
type fakeHdfs struct {
	dirs  map[string][]string
	files map[string]string
	sizes map[string][2]int64
}

func (f *fakeHdfs) list(dir string) ([]string, error) {
	if entries, ok := f.dirs[dir]; ok {
		return entries, nil
	}
	return nil, os.ErrNotExist
}

func (f *fakeHdfs) count(dataPath string) (int64, int64, error) {
	if size, ok := f.sizes[dataPath]; ok {
		return size[0], size[1], nil
	}
	return unknownCount, unknownCount, os.ErrNotExist
}

func (f *fakeHdfs) cat(filePath string) (string, error) {
	if content, ok := f.files[filePath]; ok {
		return content, nil
	}
	return "", os.ErrNotExist
}

func (f *fakeHdfs) close() {}

func TestNightsKeepsOnlyNightDirectories(t *testing.T) {
	reporter := &balanceReporter{prefix: "/user/185", hdfs: &fakeHdfs{dirs: map[string][]string{
		"/user/185/raw": {"20260811", "_spark_metadata", "20260809", "2026081"},
	}}}

	nights, err := reporter.nights()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	assert.Equal(t, nights, []string{"20260809", "20260811"})
}

func TestNightsFailsWithoutRawDataset(t *testing.T) {
	reporter := &balanceReporter{prefix: "/user/185", hdfs: &fakeHdfs{}}

	if _, err := reporter.nights(); err == nil {
		t.Error("expected an error when <prefix>/raw cannot be listed")
	}
}

// Only a committed batch counts: offsets/2 is planned but not committed.
func TestConsumedReadsLastCommittedBatch(t *testing.T) {
	checkpoint := "/user/185/raw_checkpoint/20260810"
	reporter := &balanceReporter{prefix: "/user/185", hdfs: &fakeHdfs{
		dirs: map[string][]string{checkpoint + "/commits": {"0", "1"}},
		files: map[string]string{
			checkpoint + "/offsets/1": checkpointBatch,
			checkpoint + "/offsets/2": strings.Replace(checkpointBatch, "1841233", "9999999", 1),
		},
	}}

	assert.Equal(t, reporter.consumed("20260810"), int64(3681135))
}

// A missing checkpoint, a run that committed nothing and an unreadable batch
// must all read as unknown, not as zero alert.
func TestConsumedUnknown(t *testing.T) {
	checkpoint := "/user/185/raw_checkpoint/20260810"

	for name, fake := range map[string]*fakeHdfs{
		"no checkpoint":  {},
		"no commit":      {dirs: map[string][]string{checkpoint + "/commits": {}}},
		"missing offset": {dirs: map[string][]string{checkpoint + "/commits": {"3"}}},
	} {
		reporter := &balanceReporter{prefix: "/user/185", hdfs: fake}
		if got := reporter.consumed("20260810"); got != unknownCount {
			t.Errorf("%s: consumed = %d, want unknown", name, got)
		}
	}
}

func TestDataset(t *testing.T) {
	reporter := &balanceReporter{prefix: "/user/185", hdfs: &fakeHdfs{sizes: map[string][2]int64{
		"/user/185/raw/20260810": {42, 4823456789},
	}}}

	files, bytes := reporter.dataset("/user/185/raw/20260810")
	assert.Equal(t, files, int64(42))
	assert.Equal(t, bytes, int64(4823456789))

	files, bytes = reporter.dataset("/user/185/science/20260810")
	assert.Equal(t, files, unknownCount)
	assert.Equal(t, bytes, unknownCount)
}

// An unreachable NameNode must fail fast with an error, not hang the report.
func TestNativeReaderUnreachableNamenode(t *testing.T) {
	start := time.Now()
	// TEST-NET-1 (RFC 5737): never routed, so the dial runs into its timeout
	// or is refused at once, depending on the host network.
	if _, err := newNativeReader([]string{"192.0.2.1:8020"}, "185"); err == nil {
		t.Fatal("expected an error on an unreachable NameNode")
	}
	if elapsed := time.Since(start); elapsed > hdfsDialTimeout+5*time.Second {
		t.Errorf("connection attempt took %s, beyond the dial timeout", elapsed)
	}
}
