package cmd

import (
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

func logConfiguration() {
	c := viper.AllSettings()
	bs, err := yaml.Marshal(c)
	if err != nil {
		logger.Fatalf("unable to marshal finkctl configuration to YAML: %v", err)
	}
	logger.Infof("Current finkctl configuration:\n%s", bs)
}
