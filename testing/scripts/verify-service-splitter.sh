#!/bin/bash

# Check if the correct number of arguments are passed
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 stable_weight canary_weight"
    exit 1
fi

# Run the kubectl command and save the output
output=$(kubectl get servicesplitters.consul.hashicorp.com -o yaml)

# Extract the weights
stable_weight=$(echo "$output" | yq e '.items[].spec.splits[] | select(.serviceSubset == "stable") | .weight' -)
canary_weight=$(echo "$output" | yq e '.items[].spec.splits[] | select(.serviceSubset == "canary") | .weight' -)

# Verify the weights
if [ "$stable_weight" -eq "$1" ] && [ "$canary_weight" -eq "$2" ]; then
    echo "Weights are as expected"
else
    echo "Weights are not as expected"
    echo "Expected weights {stable: $1, canary: $2}. Actual weights {stable: $stable_weight, canary: $canary_weight}"
fi