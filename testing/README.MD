# Usage
This is for building and testing argo rollout integrations with Consul. For simplicity, everything uses the `make` command. The current version of these manual steps requires a custom build image of the argo controller included in the `install.yaml` file in this directory. Once required changes are made to the argo controller and released, these steps will be updated.

# Prerequisites
1. Kind is installed on the machine
2. Helm is installed on the machine
3. Kubectl is installed on the machine 5
4. Argo kubectl extension is installed on the machine
5. The plugin binary is in a directory on the machine 
6. Update any of the necessary variables in the `Makefile` to match your environment

# Verify v1 to v2 rollout
1. Run `make setup` to setup the system with the static-server/client, consul and argo. 
   - Everything is installed to the `default` namespace except for Argo gets installed in the `argo` namespace.
2. Run `make check-service-splitter` and `make check-service-resolver`
   - service-splitter shows 100% of traffic directed to stable
   - service-resolver only has a stable filter (`Service.Meta.version=1`)
3. Run `make rollout-watch` to watch the deployments
4. Run `make splitting-watch` to witness the traffic splitting between deployments
   - You should see 100% of traffic directed to v1
5. Run `make deploy-canary-v2` to deploy a canary rollout. 
   - splitting-watch: You should see traffic begin directing to V2 but most of the traffic is still directed to V1
   - rollout-watch: You should see the rollout now includes a canary deployment for v2
   - service-splitter: shows 80% of traffic directed to stable and 20% directed to canary
   - service-resolver: includes a canary filter (`Service.Meta.version=2`)
6. Run `make promote` to promote the canary deployment and watch it succeed. 
   - splitting-watch: You should see the traffic slowly shift to V2 until all traffic is directed to V2 and none to V1
   - rollout-watch: You should see more v2 deployments until there are 5 v1 and 5 v2 deployments. After some time, you should see the v1 deployments scale down to 0
   - service-splitter: slowly changes the percentages until canary is getting 100% of traffic. Finally, when finished shows 100% of traffic directed to stable
   - service-resolver: when finished, only has a stable filter (`Service.Meta.version=2`)

# Verify Abort Behavior
1. Run `Verify v1 to v2 rollout` steps 1-5
2. Run `make abort` to abort the rollout
   - splitting-watch: You should see traffic revert to entirely to v1
   - rollout-watch: You should see the v2 image still exists in a bad state
3. Run `make retry` to retry the rollout
   - splitting-watch: You should see traffic begin directing to V2 but most of the traffic is still directed to V1
   - rollout-watch: You should see the rollout now includes a canary deployment for v2
4. Run `make promote` to promote the canary deployment and watch it succeed.
    - splitting-watch: You should see the traffic slowly shift to V2 until all traffic is directed to V2 and none to V1
    - rollout-watch: You should see more v2 deployments until there are 5 v1 and 5 v2 deployments. After some time, you should see the v1 deployments scale down to 0

# Verify Undo Behavior
1. Run all steps for `Verify v1 to v2 rollouts`
2. Run `make undo`. This will begin a rollback to the previous version (v1) 
   - splitting-watch: You should see traffic begin directing to V2 but most of the traffic is still directed to V1
   - rollout-watch: You should see the rollout now includes a canary deployment for v1
3. Run `make promote` to promote the canary deployment and watch it succeed.
   - splitting-watch: You should see the traffic slowly shift to V1 until all traffic is directed to V1 and none to V2
   - rollout-watch: You should see more v1 deployments until there are 5 v1 and 5 v2 deployments. After some time, you should see the v2 deployments scale down to 0


