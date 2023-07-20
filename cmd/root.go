/*
Copyright © 2023 Fabrice Jammes fabrice.jammes@in2p3.fr

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var cfgFile string
var dryRun bool
var secretCfgFile string

var (
	logger   *zap.SugaredLogger
	logLevel int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "finkctl",
	Short: "Command-line tool for managing and interacting with the Fink broker and its components on Spark over Kubernetes",
	Long:  `finkctl is a command-line tool for managing and interacting with the Fink broker and its components.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}

}

func init() {
	cobra.OnInitialize(initLogger, initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $CWD/.finkctl then $HOME/.finkctl)")
	rootCmd.PersistentFlags().StringVar(&secretCfgFile, "secret", "", "config file with secret (default is $CWD/.finkctl.secret then $HOME/.finkctl.secret)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Only print the command")
	rootCmd.PersistentFlags().IntVarP(&logLevel, "log-level", "v", 0, "Set-up log level")
}

// setUpLogs set the log output ans the log level
func initLogger() {

	rawJSON := []byte(`{
		"level": "debug",
		"encoding": "console",
		"outputPaths": ["stdout", "/tmp/logs"],
		"errorOutputPaths": ["stderr"],
		"encoderConfig": {
		  "messageKey": "message",
		  "levelKey": "level",
		  "levelEncoder": "lowercase"
		}
	  }`)

	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}
	_logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	defer _logger.Sync()
	logger = _logger.Sugar()

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	// Find home directory.
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	cwd, err1 := os.Getwd()
	cobra.CheckErr(err1)

	viper.AddConfigPath(cwd)
	viper.AddConfigPath(home)
	viper.SetConfigType("yaml")

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName(".finkctl")
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
		viper.SetConfigName(".finkctl.secret")
	}

	if err := viper.MergeInConfig(); err == nil {
		logger.Debugf("Use secret file: %s", viper.ConfigFileUsed())
	} else {
		logger.Fatalf("Fail reading secret file: ", err, viper.ConfigFileUsed())
	}
}
