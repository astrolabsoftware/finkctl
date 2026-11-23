package cmd

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// simTopicPattern matches the topic the fink-alert-simulator writes the alerts
// of a night to. Its name cannot be taken for a real ZTF topic (ztf_<night>).
var simTopicPattern = regexp.MustCompile(`^sim_ztf_public_([0-9]{8})$`)

// errNoSimTopic means the simulator has not written its topic (yet).
var errNoSimTopic = errors.New("no simulator topic")

var simNightTimeout time.Duration

// parseSimNight returns the night of the only simulator topic among topics.
// None is errNoSimTopic; several is an error naming them, with no guess.
func parseSimNight(topics []string) (string, error) {
	matches := []string{}
	night := ""
	for _, topic := range topics {
		if m := simTopicPattern.FindStringSubmatch(topic); m != nil {
			matches = append(matches, topic)
			night = m[1]
		}
	}
	switch len(matches) {
	case 0:
		return "", errNoSimTopic
	case 1:
		return night, nil
	default:
		return "", fmt.Errorf("several simulator topics, expected one: %s", strings.Join(matches, ", "))
	}
}

var getSimNightCmd = &cobra.Command{
	Use:   "sim-night",
	Short: "Print the night of the alerts written by the fink-alert-simulator",
	Long: `Print, alone on stdout, the night of the only sim_ztf_public_<YYYYMMDD>
Kafka topic. Wait for it until --timeout if the simulator has not written it
yet. Fail at once if there are several (e.g. a reused cluster).`,
	Run: func(cmd *cobra.Command, args []string) {
		deadline := time.After(simNightTimeout)
		for {
			topics, err := getKafkaTopics()
			cobra.CheckErr(err)
			night, err := parseSimNight(topics)
			if err == nil {
				fmt.Println(night)
				return
			}
			if !errors.Is(err, errNoSimTopic) {
				fmt.Fprintf(os.Stderr, "error: %s\n", err)
				os.Exit(1)
			}
			select {
			case <-deadline:
				fmt.Fprintf(os.Stderr, "error: no sim_ztf_public_<night> topic after %s\n", simNightTimeout)
				os.Exit(1)
			case <-time.After(10 * time.Second):
			}
		}
	},
}

func init() {
	getCmd.AddCommand(getSimNightCmd)
	getSimNightCmd.Flags().DurationVar(&simNightTimeout, "timeout", 5*time.Minute, "How long to wait for the simulator topic")
}
