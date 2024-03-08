
# finkctl
CLI tool for managing fink on Kubernetes.

To access documentation, run `finkctl -h`.

## Installation

go 1.21.+ is required.

`go install github.com/astrolabsoftware/finkctl/v3@<release_tag>`

## Configuration
To use finkctl, you need a configuration file. Set the FINKCONFIG environment variable to the directory containing `finkctl.yaml` and `finkctl.secret.yaml`. By default, it uses `$HOME/.fink`.

Example configuration files:
- [finkctl.yaml](../_e2e_/finkctl.yaml)
- [finkctl.secret.yaml](../_e2e_/finkctl.secret.yaml)
