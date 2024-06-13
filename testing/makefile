KUBERNETES_VERSION = v1.25.11
CONSUL_K8S_CHART_VERSION = 1.3.1
PLUGIN_DIR= $(shell realpath ../)
IMAGE=hashicorp/rollouts-plugin-trafficrouter-consul:0.0.1-alpha1

#Do Not change
VALUES_INIT_HELM=values_rollout_init.yaml

### TEST VERIFICATION
.PHONY: setup
setup: build-plugin create-kind-config consul-setup argo-setup

# Command for checking how the service is being split by running curl from inside a client pod
.PHONY: splitting-watch
splitting-watch:
	./scripts/splitting-watch.sh

.PHONY: rollout-watch
rollout-watch:
	kubectl argo rollouts get rollout static-server --watch

# Command used to deploy the canary deployment, will need to be promoted to continue
.PHONY: deploy-canary-v2
deploy-canary-v2:
	kubectl apply -f resources/canary_rollout_v2.yaml

# Command used to promote the canary deployment stopped with pause{}
.PHONY: promote
promote:
	kubectl argo rollouts promote static-server

.PHONY: undo
undo:
	kubectl argo rollouts undo static-server

.PHONY: abort
abort:
	kubectl argo rollouts abort static-server

.PHONY: kretryind
retry:
	kubectl argo rollouts retry rollout static-server

.PHONY: check-service-splitter
check-service-splitter:
	kubectl describe servicesplitters.consul.hashicorp.com

.PHONY: check-service-resolver
check-service-resolver:
	kubectl describe serviceresolver.consul.hashicorp.com

### SETUP KIND CLUSTER WITH CONSUL
# consul-setup sets up the kind cluster and deploys the consul helm chart
.PHONY: consul-setup
consul-setup: kind deploy
	./scripts/verify-pod-ready.sh consul-server 300; \
	./scripts/verify-pod-ready.sh consul-connect-injector 300; \

.PHONY: kind
kind: kind-delete
	kind create cluster --image kindest/node:$(KUBERNETES_VERSION) --name=dc1 --config ./resources/kind_config.yaml

.PHONY: add-helm-repo
add-helm-repo:
	helm repo add hashicorp https://helm.releases.hashicorp.com

.PHONY: kind-delete
kind-delete:
	kind delete cluster --name=dc1

.PHONY: deploy
deploy:
	helm repo update; helm install consul hashicorp/consul --version $(CONSUL_K8S_CHART_VERSION) -f values.yaml

# Install Argo
.PHONY: argo-setup
argo-setup: deploy-argo apply-crds

.PHONY: deploy-argo
deploy-argo:
	kubectl create namespace argo-rollouts; \
	helm install argo-rollouts argo/argo-rollouts -f values_rollout.yaml -n argo-rollouts; \
	kubectl apply -f $(PLUGIN_DIR)/yaml/rbac.yaml

.PHONY: apply-crds
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

.PHONY: create-kind-config
create-kind-config:
	./scripts/create-kind-config.sh ./resources/kind_config.yaml $(PLUGIN_DIR)/testing

.PHONY: build-plugin
build-plugin:
	@cd $(PLUGIN_DIR); CGO_ENABLED=0 GOOS=linux GOARCH=$(TARGETARCH) go build -v -o testing/rollouts-plugin-trafficrouter-consul ./

### RELEASE VERIFICATION
# IMAGE
.PHONY: setup-image
setup-image: consul-setup argo-setup-image

.PHONY: argo-setup-image
argo-setup-image: argo-image-values deploy-argo-image apply-crds

.PHONY: argo-image-values
argo-image-values:
	./scripts/create_values_rollout_init_helm.sh ./$(VALUES_INIT_HELM) $(IMAGE)

.PHONY: deploy-argo-image
deploy-argo-image:
	kubectl create namespace argo-rollouts; \
	helm install argo-rollouts argo/argo-rollouts -f $(VALUES_INIT_HELM) -n argo-rollouts; \
	kubectl apply -f $(PLUGIN_DIR)/yaml/rbac.yaml

## EXTRAS
.PHONY: install-required
install-required:
	brew install argoproj/tap/kubectl-argo-rollouts; \
    brew install helm; \
    brew install kind; \
	brew install yq; \
