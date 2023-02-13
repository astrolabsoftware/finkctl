# Build parameters
# ----------------
# Repository address
REPO="gitlab-registry.in2p3.fr/astrolabsoftware/fink"
IMAGE_TAG="2.7.1-26-gbb3ec8c"
export IMAGE="$REPO/fink-broker:$IMAGE_TAG"


# Kafka cluster parameters
# ------------------------
# Name for Kafka cluster
export KAFKA_NS="kafka"
export KAFKA_CLUSTER="kafka-cluster"


# Spark job 'stream2raw' parameters
# ---------------------------------
# Default values are the ones set in fink-alert-simulator CI environment
# TODO manage varaible below through environment variable if needed
# export KAFKA_SOCKET=${KAFKA_SOCKET:-"kafka-cluster-kafka-external-bootstrap.kafka:9094"}

export FINK_ALERT_SIMULATOR_DIR="/tmp/fink-alert-simulator"
