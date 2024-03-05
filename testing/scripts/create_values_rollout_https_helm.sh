#!/bin/bash

# Check if the correct number of arguments are passed
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 helm_file binary"
    exit 1
fi

# New host path
helm_file=$1
binary=$2

# Create the YAML structure and write it to kind_config_file
cat << EOF > "$helm_file"
controller:
  image:
    # -- Registry to use
    registry: docker.io
    # -- Repository to use
    repository: wilko1989/argo-rollouts
    # -- Overrides the image tag (default is the chart appVersion)
    tag: latest
    # -- Image pull policy
    pullPolicy: Always
  trafficRouterPlugins:
    trafficRouterPlugins: |-
      - name: "hashicorp/consul"
        location: $binary # supports http(s):// urls and file://
  volumes:
    - name: consul-plugin
      emptyDir: {}
  volumeMounts:
    - name: consul-plugin
      mountPath: /plugin-bin/hashicorp
EOF