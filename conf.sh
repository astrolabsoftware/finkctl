# Build parameters
# ----------------
# Repository address
REPO="gitlab-registry.in2p3.fr/astrolabsoftware/fink"
IMAGE_TAG="2.7.1-20-gbd9d92b"
IMAGE="$REPO/fink-broker:$IMAGE_TAG"


# Kafka cluster parameters
# ------------------------
# Name for Kafka cluster
KAFKA_NS="kafka"
KAFKA_CLUSTER="kafka-cluster"


# Spark job 'stream2raw' parameters
# ---------------------------------
# Default values are the ones set in fink-alert-simulator CI environment
KAFKA_SOCKET=${KAFKA_SOCKET:-"kafka-cluster-kafka-external-bootstrap.kafka:9094"}
KAFKA_TOPIC=${KAFKA_TOPIC:-"ztf-stream-sim"}

FINK_ALERT_SIMULATOR_DIR="/tmp/fink-alert-simulator"

# submit the job in cluster mode - 1 driver + 1 executor
PRODUCER="sims"
FINK_ALERT_SCHEMA="/home/fink/fink-alert-schemas/ztf/ztf_public_20190903.schema.avro"
KAFKA_STARTING_OFFSET="earliest"
ONLINE_DATA_PREFIX="/home/fink/fink-broker/online"
FINK_TRIGGER_UPDATE=2
LOG_LEVEL="INFO"

NIGHT="20190903"
