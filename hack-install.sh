#!/bin/bash

# This script is used to install the development finkctl command line tool
# to $PATH

set -euxo pipefail

DIR=$(cd "$(dirname "$0")"; pwd -P)

go install $DIR
