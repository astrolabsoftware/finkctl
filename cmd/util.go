package cmd

import (
	"log/slog"
	"os"
	"path"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// initConfig reads config files
func initConfig() {

	var finkConfigPath = os.Getenv("FINKCONFIG")
	if len(finkConfigPath) == 0 {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		cwd, err1 := os.Getwd()
		cobra.CheckErr(err1)
		viper.AddConfigPath(cwd)

		finkConfigPath = path.Join(home, ".finkctl")
	}

	viper.AddConfigPath(finkConfigPath)
	viper.SetConfigType("yaml")

	viper.SetConfigName("finkctl.yaml")

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		slog.Debug("Use config", "file", viper.ConfigFileUsed())
	} else {
		slog.Error("Fail reading configuration files in $FINKCONFIG, $HOME/.fink, then $CWD", "error", err, "config_file_used", viper.ConfigFileUsed())
		syscall.Exit(1)
	}

	viper.SetConfigName("finkctl.secret.yaml")

	if err := viper.MergeInConfig(); err == nil {
		slog.Debug("Use secret", "file", viper.ConfigFileUsed())
	} else {
		slog.Error("Fail reading secret files in $FINKCONFIG, $HOME/.fink, then $CWD", "error", err, "config_file_used", viper.ConfigFileUsed())
		syscall.Exit(1)
	}
}

func logConfiguration() {
	c := viper.AllSettings()
	bs, err := yaml.Marshal(c)
	if err != nil {
		slog.Error("unable to marshal finkctl configuration to YAML", "error", err)
		syscall.Exit(1)
	}
	slog.Info("Current finkctl configuration", "data", bs)
}

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
	return finkTopics, nil
}

// applyVarTemplate use a template to generate a configuration value
// and example template is: "xxxx-{{ .Night }}-yyyy"
func applyVarTemplate(template string, night string) string {
	type TmplData struct {
		Night string
	}
	return format(template, &TmplData{Night: night})
}

func convertToBytes(src map[string]string) map[string][]byte {
	if src == nil {
		return nil
	}
	dst := make(map[string][]byte, len(src))
	for k, v := range src {
		dst[k] = []byte(v)
	}
	return dst
}
