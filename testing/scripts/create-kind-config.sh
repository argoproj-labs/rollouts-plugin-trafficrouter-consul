#!/bin/bash

# Check if the correct number of arguments are passed
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 kind_config_file host_path"
    exit 1
fi

# New host path
kind_config_file=$1
host_path=$2

# Create the YAML structure and write it to kind_config_file
cat << EOF > "$kind_config_file"
apiVersion: kind.x-k8s.io/v1alpha4
kind: Cluster
nodes:
  - role: control-plane
    extraMounts:
      - hostPath: $host_path
        containerPath: /rollouts-plugin-trafficrouter-consul
EOF