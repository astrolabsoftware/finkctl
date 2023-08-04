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

SPARK_HOME="${SPARK_INSTALL_DIR}/${SPARK_NAME}"
PATH="$SPARK_HOME/bin:$PATH"

