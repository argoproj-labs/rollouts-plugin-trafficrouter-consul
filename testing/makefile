KUBERNETES_VERSION = v1.25.11
CONSUL_K8S_CHART_VERSION = 1.3.1
PLUGIN_DIR= $(shell realpath ../)

### TEST VERIFICATION
setup: build-plugin create-kind-config consul-setup argo-setup

# Command for checking how the service is being split by running curl from inside a client pod
splitting-watch:
	./scripts/splitting-watch.sh

rollout-watch:
	kubectl argo rollouts get rollout static-server --watch

# Command used to deploy the canary deployment, will need to be promoted to continue
deploy-canary-v2:
	kubectl apply -f resources/canary_rollout_v2.yaml

# Command used to promote the canary deployment stopped with pause{}
promote:
	kubectl argo rollouts promote static-server

undo:
	kubectl argo rollouts undo static-server

abort:
	kubectl argo rollouts abort static-server

retry:
	kubectl argo rollouts retry rollout static-server

check-service-splitter:
	kubectl describe servicesplitters.consul.hashicorp.com

check-service-resolver:
	kubectl describe serviceresolver.consul.hashicorp.com

### SETUP KIND CLUSTER WITH CONSUL
# consul-setup sets up the kind cluster and deploys the consul helm chart
consul-setup: kind deploy
	./scripts/verify-pod-ready.sh consul-server 300; \
	./scripts/verify-pod-ready.sh consul-connect-injector 300; \

kind: kind-delete
	kind create cluster --image kindest/node:$(KUBERNETES_VERSION) --name=dc1 --config ./resources/kind_config.yaml

add-helm-repo:
	helm repo add hashicorp https://helm.releases.hashicorp.com

kind-delete:
	kind delete cluster --name=dc1

deploy:
	helm repo update; helm install consul hashicorp/consul --version $(CONSUL_K8S_CHART_VERSION) -f values.yaml

# Install Argo
argo-setup: deploy-argo apply-crds

deploy-argo:
	kubectl create namespace argo-rollouts; \
	kubectl apply -n argo-rollouts -f install.yaml
	kubectl apply -f $(PLUGIN_DIR)/yaml/rbac.yaml

apply-crds:
	kubectl apply -f resources/proxy_defaults.yaml \
	-f resources/serviceaccount.yaml \
	-f resources/service.yaml \
	-f resources/serviceaccount_client.yaml \
	-f resources/service_client.yaml \
	-f resources/deployment_client.yaml \
	-f resources/service_resolver.yaml \
	-f resources/service_splitter.yaml \
	-f resources/canary_rollout.yaml

create-kind-config:
	./scripts/create-kind-config.sh ./resources/kind_config.yaml $(PLUGIN_DIR)

build-plugin:
	make -C $(PLUGIN_DIR) build

## EXTRAS
install-required:
	brew install argoproj/tap/kubectl-argo-rollouts; \
    brew install helm; \
    brew install kind; \
	brew install yq; \
