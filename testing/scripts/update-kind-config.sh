#!/bin/bash

# Check if the correct number of arguments are passed
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 kind_config_file new_host_path"
    exit 1
fi

# New host path
kind_config_file=$1
new_host_path=$2

# Use yq to update the hostPath
yq e ".nodes[0].extraMounts[0].hostPath = \"$new_host_path\"" -i "$kind_config_file"