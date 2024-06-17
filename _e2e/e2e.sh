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

cp $DIR/finkctl.yaml.orig $DIR/finkctl.yaml

ink "Check -N parameter parsing"
if finkctl run raw2science --image=param_image -N 2020111101011 --dry-run
then
    ink -r "Expected to fail with -N parameter"
    exit 1
fi

ink "Check stream2raw dry-run"
finkctl run stream2raw --image=param_image -N 20000101 --dry-run > /tmp/stream2raw.out
diff /tmp/stream2raw.out $DIR/stream2raw.out.expected

ink "Check raw2science dry-run"
finkctl run raw2science --image=param_image -N 20000101 --dry-run > /tmp/raw2science.out
diff /tmp/raw2science.out $DIR/raw2science.out.expected

ink "Check distribution dry-run"
finkctl run distribution --image=param_image -N 20000101 --dry-run > /tmp/distribution.out
diff -I '^    --conf spark.kubernetes.executor.podTemplateFile=/tmp/fink-broker-[0-9]\+/executor-pod-template.yaml \\$' /tmp/distribution.out $DIR/distribution.out.expected

ink "Check stream2raw dry-run with 0 cpu"
sed "s/  cpu: 1/  cpu: 0/" $DIR/finkctl.yaml.orig > $DIR/finkctl.yaml
finkctl run stream2raw --image=param_image -N 20000101 --dry-run > /tmp/stream2raw-zerocpu.out
diff /tmp/stream2raw-zerocpu.out $DIR/stream2raw-zerocpu.out.expected
