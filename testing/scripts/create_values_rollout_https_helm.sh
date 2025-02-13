#!/bin/bash

# Check if the correct number of arguments are passed
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 helm_file binary"
    exit 1
fi

# New host path
helm_file=$1
binary=$2
registry=$3
repository=$4
tag=$5

# Create the YAML structure and write it to kind_config_file
cat << EOF > "$helm_file"
controller:
  image:
    # -- Registry to use
    registry: $registry
    # -- Repository to use
    repository: $repository
    # -- Overrides the image tag (default is the chart appVersion)
    tag: $tag
    # -- Image pull policy
    pullPolicy: IfNotPresent
  trafficRouterPlugins:
    - name: "hashicorp/consul"
      location: $binary # supports http(s):// urls and file://
  volumes:
    - name: consul-plugin
      emptyDir: {}
  volumeMounts:
    - name: consul-plugin
      mountPath: /plugin-bin/hashicorp
EOF