# rollouts-plugin-trafficrouter-consul
A plugin that allows argo-rollouts to work with Consul's service mesh for traffic shaping patterns.

# Testing
To run unit tests use `go test ./...`. For end-to-end verification follow the steps in `./testing/README.md`.

## How to integrate Consul with Argo Rollouts

### Install Rollouts Using Helm (Init Container)
This is the preferred method for installing this plugin.

Add the following code to your `values.yaml` file, then install the argo-rollouts by helm:

```yaml
  initContainers:
    - name: copy-consul-plugin
      image: <TODO Insert Release Image>
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
```
```bash
helm install argo-rollouts argo/argo-rollouts -f values.yaml -n argo-rollouts
```

### Install Rollouts Using Helm (Binary)

1. Build this plugin (i.e. `make build`).
2. Mount the built plugin on to the `argo-rollouts` container
3. Add the following code to your `values.yaml` file, then install the argo-rollouts by helm:

```yaml
  trafficRouterPlugins:
    trafficRouterPlugins: |-
      - name: "hashicorp/consul"
        location: "file:///plugin-bin/hashicorp/rollouts-plugin-trafficrouter-consul"
  volumes:
    - name: consul-route-plugin
      hostPath:
        # The path being mounted to, change this depending on your mount path
        path: /rollouts-plugin-trafficrouter-consul
        type: DirectoryOrCreate
  volumeMounts:
    - name: consul-route-plugin
      mountPath: /plugin-bin/hashicorp
```
```bash
helm install argo-rollouts argo/argo-rollouts -f values.yaml -n argo-rollouts
```

### Install the RBAC

After either mounting the binary or using an init container apply the RBAC using the provided `yaml/rbac.yaml`
```bash
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/rollouts-plugin-trafficrouter-consul/main/yaml/rbac.yaml
```

## Usage
