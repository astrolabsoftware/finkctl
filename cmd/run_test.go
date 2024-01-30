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
func TestGetSparkConfig(t *testing.T) {

	viper.Set(RUN+".fink_trigger_update", "2")
	viper.Set(RUN+".cpu", "4")
	viper.Set(DISTRIBUTION+".memory", "8GB")
	viper.Set(RUN+".image", "test-image")

	t.Logf("Viper config: %v", viper.AllSettings())

	sc := getSparkConfig(DISTRIBUTION)

	if sc.Cpu != "4" {
		t.Errorf("Expected CPU to be '4', but got '%s'", sc.Cpu)
	}

	if sc.Memory != "8GB" {
		t.Errorf("Expected Memory to be '8GB', but got '%s'", sc.Memory)
	}

	if sc.Binary != "distribute.py" {
		t.Errorf("Expected Binary to be 'distribute.py', but got '%s'", sc.Binary)
	}

	// if sc.ApiServerUrl != "test-api-server-url" {
	// 	t.Errorf("Expected ApiServerUrl to be 'test-api-server-url', but got '%s'", sc.ApiServerUrl)
	// }

	if sc.StorageClass != s3 {
		t.Errorf("Expected StorageClass to be %v, but got '%v'", s3, sc.StorageClass)
	}

	if sc.Image != "test-image" {
		t.Errorf("Expected Image to be 'test-image', but got '%s'", sc.Image)
	}
}
