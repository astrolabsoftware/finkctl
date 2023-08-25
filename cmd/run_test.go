package cmd

import (
	"log"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestGetSparkCmd(t *testing.T) {
	sc := SparkConfig{
		Image:             "gitlab-registry.in2p3.fr/astrolabsoftware/fink/fink-broker:testtag",
		Producer:          "sims",
		OnlineDataPrefix:  "/home/fink/fink-broker/online",
		FinkTriggerUpdate: "2",
		LogLevel:          "INFO",
	}

	sc.Binary = "changeme.py"
	sparkCmd := applyTemplate(sc, DISTRIBUTION)

	log.Printf("CMD %v", sparkCmd)
}

func TestViperReadConfigFile(t *testing.T) {
	type config struct {
		Name string
	}

	name := "foobar"
	if err := os.Setenv("VIPERTEST_NAME", name); err != nil {
		t.Fatal(err)
	}

	v := viper.New()
	v.SetEnvPrefix("VIPERTEST")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	for _, key := range viper.AllKeys() {
		viper.BindEnv(key, strings.ReplaceAll(key, ".", "_"))
	}

	// if you uncomment this line, the test will pass, otherwise it'll fail.
	// I would not expect to have to call bind if I'm using AutomaticEnv.
	v.BindEnv("name")

	c := config{}
	if err := v.Unmarshal(&c); err != nil {
		t.Fatal(err)
	}

	if v.GetString("name") != c.Name {
		t.Fatalf("expected name to be %q but got %q", v.GetString("name"), c.Name)
	}
}

func TestViperUnmarshalAutoEnv(t *testing.T) {
	type config struct {
		Name string
	}

	name := "foobar"
	if err := os.Setenv("VIPERTEST_NAME", name); err != nil {
		t.Fatal(err)
	}

	v := viper.New()
	v.SetEnvPrefix("VIPERTEST")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	for _, key := range viper.AllKeys() {
		viper.BindEnv(key, strings.ReplaceAll(key, ".", "_"))
	}

	// if you uncomment this line, the test will pass, otherwise it'll fail.
	// I would not expect to have to call bind if I'm using AutomaticEnv.
	v.BindEnv("name")

	c := config{}
	if err := v.Unmarshal(&c); err != nil {
		t.Fatal(err)
	}

	if v.GetString("name") != c.Name {
		t.Fatalf("expected name to be %q but got %q", v.GetString("name"), c.Name)
	}
}
