#!/bin/bash

# Check if the correct number of arguments are passed
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 helm_file image"
    exit 1
fi

# New host path
helm_file=$1
image=$2

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
  initContainers:
    - name: copy-consul-plugin
      image: $image
      command: ["/bin/sh", "-c"]
      args:
        # Copy the binary from the image to the rollout container
        - cp /bin/rollouts-plugin-trafficrouter-consul /plugin-bin/hashicorp
      volumeMounts:
        - name: consul-plugin
          mountPath: /plugin-bin/hashicorp
  trafficRouterPlugins:
    trafficRouterPlugins: |-
      - name: "hashicorp/consul"
        location: "file:///plugin-bin/hashicorp/rollouts-plugin-trafficrouter-consul"
  volumes:
    - name: consul-plugin
      emptyDir: {}
  volumeMounts:
    - name: consul-plugin
      mountPath: /plugin-bin/hashicorp
EOF