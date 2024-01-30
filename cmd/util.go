package cmd

import (
	"os"
	"path"
	"strings"

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
		logger.Debugf("Use config file: %s", viper.ConfigFileUsed())
	} else {
		logger.Fatalf("Fail reading configuration files in $FINKCONFIG, $HOME/.fink, then $CWD: ", err, viper.ConfigFileUsed())
	}

	viper.SetConfigName("finkctl.secret.yaml")

	if err := viper.MergeInConfig(); err == nil {
		logger.Debugf("Use secret file: %s", viper.ConfigFileUsed())
	} else {
		logger.Fatalf("Fail reading secret files in $FINKCONFIG, $HOME/.fink, then $CWD: ", err, viper.ConfigFileUsed())
	}
}

func logConfiguration() {
	c := viper.AllSettings()
	bs, err := yaml.Marshal(c)
	if err != nil {
		logger.Fatalf("unable to marshal finkctl configuration to YAML: %v", err)
	}
	logger.Infof("Current finkctl configuration:\n%s", bs)
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
