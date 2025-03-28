package resources

import (
	_ "embed"
)

const (
	KafkaJaasConfFile = "kafka-jaas.conf"
)

//go:embed kafka-jaas.conf
var KafkaJaasConf string
