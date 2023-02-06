# Build parameters
# ----------------
# Repository address
REPO="gitlab-registry.in2p3.fr/astrolabsoftware/fink"
# Tag to apply to the built image, or to identify the image to be pushed
GIT_HASH="$(git -C $DIR describe --dirty --always)"
IMAGE_TAG="$GIT_HASH"
# WARNING "spark-py" is hard-coded in spark build script
export IMAGE="$REPO/fink-broker:$IMAGE_TAG"


# Spark parameters
# ----------------
# Assuming Scala 2.11

# Spark image tag
# Spark image is built here: https://github.com/astrolabsoftware/k8s-spark-py/
SPARK_IMAGE_TAG="k8s-3.1.3"

# Spark version
SPARK_VERSION="3.1.3"

# Name for the Spark archive
SPARK_NAME="spark-${SPARK_VERSION}-bin-hadoop3.2"

# Spark install location
SPARK_INSTALL_DIR="${HOME}/fink-k8s-tmp"

export SPARK_HOME="${SPARK_INSTALL_DIR}/${SPARK_NAME}"
export PATH="$SPARK_HOME/bin:$PATH"

# Kafka cluster parameters
# ------------------------
# Name for Kafka cluster
export KAFKA_NS="kafka"
export KAFKA_CLUSTER="kafka-cluster"


# Spark job 'stream2raw' parameters
# ---------------------------------
# Default values are the ones set in fink-alert-simulator CI environment
export KAFKA_SOCKET=${KAFKA_SOCKET:-"kafka-cluster-kafka-external-bootstrap.kafka:9094"}
export KAFKA_TOPIC=${KAFKA_TOPIC:-"ztf-stream-sim"}

FINK_ALERT_SIMULATOR_DIR="/tmp/fink-alert-simulator"

# submit the job in cluster mode - 1 driver + 1 executor
export PRODUCER="sims"
export FINK_ALERT_SCHEMA="/home/fink/fink-alert-schemas/ztf/ztf_public_20190903.schema.avro"
export KAFKA_STARTING_OFFSET="earliest"
export ONLINE_DATA_PREFIX="/home/fink/fink-broker/online"
export FINK_TRIGGER_UPDATE=2
export LOG_LEVEL="INFO"
ci=${CI:-false}

export NIGHT="20190903"
