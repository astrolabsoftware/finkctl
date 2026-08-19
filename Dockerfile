#
# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# Image running finkctl in the cluster, e.g. the fink-broker report CronJob.
# finkctl talks to the Kubernetes API only, so a static binary is enough.

FROM golang:1.21 AS builder

WORKDIR /src

# Dependencies first, so they stay cached across source-only changes
COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY cmd/ ./cmd/
COPY resources/ ./resources/

RUN CGO_ENABLED=0 go build -trimpath -o /finkctl .

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /finkctl /finkctl

USER nonroot:nonroot
ENTRYPOINT ["/finkctl"]
