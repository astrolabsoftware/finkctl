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
	"go.uber.org/zap"
)

var (
	dryRun    bool
	logger    *zap.SugaredLogger
	verbosity int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "finkctl",
	Short: "Command-line tool for managing and interacting with the Fink broker and its components on Spark over Kubernetes",
	Long: `finkctl is a command-line tool for managing and interacting with the Fink broker and its components.

	finkctl configuration directory is:
	1. directory referenced by FINKCONFIG environment variable
	2. current working directory
	3. $HOME/.finkctl
	Example of configuration files are available here:
	- https://github.com/astrolabsoftware/fink-broker/blob/master/itest/finkctl.yaml
	- https://github.com/astrolabsoftware/fink-broker/blob/master/itest/finkctl.secret.yaml
`,
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
	rootCmd.PersistentFlags().IntVarP(&verbosity, "verbosity", "v", 0, "Verbosity level (-v0 for minimal, -v2 for maximum)")
	cobra.OnInitialize(initLogger)

	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Only print the command")
}

// setUpLogs set the log output ans the log level
func initLogger() {
	var loglevelStr string
	if verbosity == 0 {
		loglevelStr = "error"
	} else if verbosity == 1 {
		loglevelStr = "info"
	} else {
		loglevelStr = "debug"
	}

	rawJSON := []byte(`{
		"level": "` + loglevelStr + `",
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
