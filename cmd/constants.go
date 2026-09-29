package cmd

const (
	finkPrefix      = "fink_"
	kafkaNamespace  = "kafka"
	kafkaSecretName = "fink-producer"

	// Kafka cluster access: the CLI tools are run inside a broker pod, whose
	// name is resolved from the Strimzi labels and falls back to the
	// well-known name of the single-node KRaft pool.
	kafkaBrokerSelector  = "strimzi.io/cluster=kafka-cluster,strimzi.io/broker-role=true"
	kafkaPodFallback     = "kafka-cluster-dual-role-0"
	kafkaContainer       = "kafka"
	kafkaBootstrapServer = "kafka-cluster-kafka-bootstrap.kafka:9092"
	kafkaTopicsBin       = "bin/kafka-topics.sh"
	kafkaOffsetsBin      = "bin/kafka-get-offsets.sh"
)
