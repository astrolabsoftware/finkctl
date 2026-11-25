package cmd

import (
	"log/slog"
	"strings"
)

// getKafkaTopics returns the list of Kafka topics produced by fink-broker
func getFinkTopics() ([]string, error) {
	finkTopics := []string{}
	topics, err := getKafkaTopics()
	if err != nil {
		return nil, err
	}
	for _, t := range topics {
		if strings.HasPrefix(t, finkPrefix) {
			finkTopics = append(finkTopics, t)
		}
	}
	slog.Debug("List of fink topics", "topics", finkTopics)
	return finkTopics, nil
}
