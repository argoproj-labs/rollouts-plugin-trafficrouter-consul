#!/bin/bash

# Check if the correct number of arguments are passed
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 pod_name timeout"
    exit 1
fi

pod_name=$1
timeout=$2
start_time=$(date +%s)

echo "Waiting for $pod_name pod to be ready"
while true; do
    # Get the status of the pod
    output=$(kubectl get pods -o json)
    status=$(echo "$output" | jq -r ".items[] | select(.metadata.name | startswith(\"$pod_name\")).status.conditions[] | select(.type == \"Ready\").status")

    if [[ $status == "True" ]]; then
        echo "Pod $pod_name is ready"
        exit 0
    else
        current_time=$(date +%s)
        elapsed_time=$((current_time - start_time))

        if ((elapsed_time > timeout)); then
            echo "Timeout while waiting for pod to be ready"
            exit 1
        fi

        # Wait for a bit before checking again
        sleep 5
    fi
done