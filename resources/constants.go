package resources

import (
	_ "embed"
)

//go:embed kafka-jaas.conf
var KafkaJaasConf string
