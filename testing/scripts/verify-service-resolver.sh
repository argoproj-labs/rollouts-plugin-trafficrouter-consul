#!/bin/bash

# Check if at least one argument is passed
if [ "$#" -lt 1 ]; then
    echo "Usage: $0 stable_filter [canary_filter]"
    exit 1
fi

# Run the kubectl command and save the output
output=$(kubectl get serviceresolvers.consul.hashicorp.com -o yaml)

# Extract the filters
stable_filter=$(echo "$output" | yq e '.items[].spec.subsets.stable.filter' -)
canary_filter=$(echo "$output" | yq e '.items[].spec.subsets.canary.filter' -)

# Verify the filters
if [ "$stable_filter" == "$1" ]; then
    if [ -z "$2" ]; then
        if [ -z "$canary_filter" ]; then
            echo "Filters are as expected"
        else
            echo "Unexpected canary filter. Expected: None, Actual: $canary_filter"
        fi
    else
        if [ "$canary_filter" == "$2" ]; then
            echo "Filters are as expected"
        else
            echo "Canary filter does not match. Expected: $2, Actual: $canary_filter"
        fi
    fi
else
    echo "Stable filter does not match. Expected: $1, Actual: $stable_filter"
fi