#!/bin/bash

set -euxo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

export FINKCONFIG="$DIR"

mkdir -p $HOME/.kube
cat >> $HOME/.kube/config << EOF
apiVersion: v1
clusters:
- cluster:
    server: https://127.0.0.1:34729
  name: kind-kind
contexts:
- context:
    cluster: kind-kind
    user: kind-kind
  name: kind-kind
current-context: kind-kind
kind: Config
preferences: {}
users:
- name: kind-kind
  user:
EOF

finkctl run raw2science --image=param_image --dry-run > /tmp/raw2science.out

diff /tmp/raw2science.out $DIR/raw2science.out.expected
