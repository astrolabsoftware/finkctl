package resources

import (
	_ "embed"
)

const (
	KafkaJaasConfFile       = "kafka-jaas.conf"
	ExecutorPodTemplateFile = "executor-pod-template.yaml"
)

//go:embed kafka-jaas.conf
var KafkaJaasConf string

//go:embed executor-pod-template.yaml
var ExecutorPodTemplate string
