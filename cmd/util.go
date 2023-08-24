package cmd

import (
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	// Find home directory.
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	cwd, err1 := os.Getwd()
	cobra.CheckErr(err1)

	viper.AddConfigPath(cwd)
	viper.AddConfigPath(path.Join(home, ".finkctl"))
	viper.SetConfigType("yaml")

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("finkctl.yaml")
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		logger.Debugf("Use config file: %s", viper.ConfigFileUsed())
	} else {
		logger.Fatalf("Fail reading configuration file: ", err, viper.ConfigFileUsed())
	}

	if secretCfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(secretCfgFile)
	} else {
		viper.SetConfigName("finkctl.secret.yaml")
	}

	if err := viper.MergeInConfig(); err == nil {
		logger.Debugf("Use secret file: %s", viper.ConfigFileUsed())
	} else {
		logger.Fatalf("Fail reading secret file: ", err, viper.ConfigFileUsed())
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
func getFinkTopics() []string {
	topics := []string{}
	for _, t := range getKafkaTopics() {
		if strings.HasPrefix(t, finkPrefix) {
			topics = append(topics, t)
		}
	}
	return topics
}
