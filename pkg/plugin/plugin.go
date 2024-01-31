package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	rolloutsPlugin "github.com/argoproj/argo-rollouts/rollout/trafficrouting/plugin/rpc"
	pluginTypes "github.com/argoproj/argo-rollouts/utils/plugin/types"
	consulv1aplha1 "github.com/hashicorp/consul-k8s/control-plane/api/v1alpha1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/wilkermichael/rollouts-plugin-trafficrouter-consul/pkg/utils"
)

const (
	serviceMetaVersionAnnotation     = "consul.hashicorp.com/service-meta-version"
	filterServiceMetaVersionTemplate = "Service.Meta.version == %s"
)

type ConsulTrafficRouting struct {
	ServiceName      string `json:"serviceName" protobuf:"bytes,1,opt,name=serviceName"`
	CanarySubsetName string `json:"canarySubsetName" protobuf:"bytes,2,opt,name=canarySubsetName"`
	StableSubsetName string `json:"stableSubsetName" protobuf:"bytes,3,opt,name=stableSubsetName"`
}

type RpcPlugin struct {
	K8SClient client.Client
	LogCtx    *logrus.Entry
	IsTest    bool
}

var _ rolloutsPlugin.TrafficRouterPlugin = (*RpcPlugin)(nil)

func (r *RpcPlugin) InitPlugin() pluginTypes.RpcError {
	if r.IsTest {
		return pluginTypes.RpcError{}
	}

	cfg, err := utils.NewKubeConfig()
	if err != nil {
		return pluginTypes.RpcError{ErrorString: err.Error()}
	}
	s := runtime.NewScheme()
	if err := consulv1aplha1.AddToScheme(s); err != nil {
		return pluginTypes.RpcError{ErrorString: err.Error()}
	}
	r.K8SClient, err = client.New(cfg, client.Options{Scheme: s})
	if err != nil {
		return pluginTypes.RpcError{ErrorString: err.Error()}
	}

	r.LogCtx = logrus.NewEntry(logrus.New())

	return pluginTypes.RpcError{}
}

func (r *RpcPlugin) UpdateHash(_ *v1alpha1.Rollout, _, _ string, _ []v1alpha1.WeightDestination) pluginTypes.RpcError {
	return pluginTypes.RpcError{}
}

func (r *RpcPlugin) SetWeight(rollout *v1alpha1.Rollout, desiredWeight int32, _ []v1alpha1.WeightDestination) pluginTypes.RpcError {
	ctx := context.TODO()
	consulConfig, err := getPluginConfig(rollout)
	if err != nil {
		return pluginTypes.RpcError{ErrorString: err.Error()}
	}

	serviceName := consulConfig.ServiceName
	canarySubsetName := consulConfig.CanarySubsetName
	stableSubsetName := consulConfig.StableSubsetName
	serviceMetaVersion := rollout.Spec.Template.GetObjectMeta().GetAnnotations()[serviceMetaVersionAnnotation]

	// This checks that we are performing a canary rollout, it is not
	// an error if this is empty. This will be empty on the initial rollout
	if rollout.Status.Canary == (v1alpha1.CanaryStatus{}) {
		r.LogCtx.Debug("Rollout does not have a CanaryStatus yet", "desiredWeight", desiredWeight)
		return pluginTypes.RpcError{}
	}

	// Get the service resolver
	serviceResolver := consulv1aplha1.ServiceResolver{}
	if err := r.K8SClient.Get(ctx, types.NamespacedName{Name: serviceName, Namespace: rollout.GetNamespace()}, &serviceResolver, &client.GetOptions{}); err != nil {
		return pluginTypes.RpcError{ErrorString: err.Error()}
	}

	// If the rollout is successful (not aborted) then modify the resolver
	if rolloutAborted(rollout) {
		r.LogCtx.Debug("Updating ServiceResolver for aborted rollout", "canarySubsetName", canarySubsetName, "serviceResolver", serviceResolver)
		err := r.updateResolverForAbortedRollout(ctx, canarySubsetName, serviceResolver)
		if err != nil {
			return pluginTypes.RpcError{ErrorString: err.Error()}
		}
	} else {
		// Check if the pods have completely rolled over, and we are finished, now set the resolver to the stable version
		if rolloutComplete(rollout) {
			r.LogCtx.Debug("Updating ServiceResolver for completion", "stableSubsetName", stableSubsetName, "canarySubsetName", canarySubsetName, "serviceMetaVersion", serviceMetaVersion, "serviceResolver", serviceResolver)
			err := r.updateResolverAfterCompletion(ctx, stableSubsetName, canarySubsetName, serviceMetaVersion, serviceResolver)
			if err != nil {
				return pluginTypes.RpcError{ErrorString: err.Error()}
			}
		} else {
			r.LogCtx.Debug("Updating ServiceResolver for rollout", "canarySubsetName", canarySubsetName, "serviceMetaVersion", serviceMetaVersion, "serviceResolver", serviceResolver)
			err := r.updateResolverForRollouts(ctx, canarySubsetName, serviceMetaVersion, serviceResolver)
			if err != nil {
				return pluginTypes.RpcError{ErrorString: err.Error()}
			}
		}
	}

	serviceSplitter := consulv1aplha1.ServiceSplitter{}
	if err := r.K8SClient.Get(ctx, types.NamespacedName{Name: serviceName, Namespace: rollout.GetNamespace()}, &serviceSplitter, &client.GetOptions{}); err != nil {
		return pluginTypes.RpcError{ErrorString: err.Error()}
	}
	if len(serviceSplitter.Spec.Splits) == 0 {
		return pluginTypes.RpcError{ErrorString: "spec.splits was not found in consul service splitter"}
	}

	for i, split := range serviceSplitter.Spec.Splits {
		switch split.ServiceSubset {
		case canarySubsetName:
			serviceSplitter.Spec.Splits[i].Weight = float32(desiredWeight)
		case stableSubsetName:
			serviceSplitter.Spec.Splits[i].Weight = float32(100 - desiredWeight)
		default:
			return pluginTypes.RpcError{ErrorString: "unexpected service split"}
		}
	}

	// Persist changes to the ServiceSplitter
	r.LogCtx.Debug("Updating ServiceSplitter", "serviceSplitter", serviceSplitter)
	if err := r.K8SClient.Update(ctx, &serviceSplitter, &client.UpdateOptions{}); err != nil {
		return pluginTypes.RpcError{ErrorString: err.Error()}
	}
	return pluginTypes.RpcError{}
}

func (r *RpcPlugin) SetHeaderRoute(_ *v1alpha1.Rollout, _ *v1alpha1.SetHeaderRoute) pluginTypes.RpcError {
	return pluginTypes.RpcError{}
}

func (r *RpcPlugin) VerifyWeight(_ *v1alpha1.Rollout, _ int32, _ []v1alpha1.WeightDestination) (pluginTypes.RpcVerified, pluginTypes.RpcError) {
	return pluginTypes.NotImplemented, pluginTypes.RpcError{}
}

func (r *RpcPlugin) Type() string {
	return Type
}

func (r *RpcPlugin) SetMirrorRoute(_ *v1alpha1.Rollout, _ *v1alpha1.SetMirrorRoute) pluginTypes.RpcError {
	return pluginTypes.RpcError{}
}

func (r *RpcPlugin) RemoveManagedRoutes(ro *v1alpha1.Rollout) pluginTypes.RpcError {
	return pluginTypes.RpcError{}
}

func (r *RpcPlugin) updateResolverAfterCompletion(ctx context.Context, stableSubsetName, canarySubsetName, serviceMetaVersion string, sr consulv1aplha1.ServiceResolver) error {
	if _, ok := sr.Spec.Subsets[canarySubsetName]; !ok {
		return errors.New(fmt.Sprintf("spec.subsets.%s.filter was not found in consul service resolver: %v", canarySubsetName, sr))
	}
	canarySubset := sr.Spec.Subsets[canarySubsetName]
	canarySubset.Filter = ""
	sr.Spec.Subsets[canarySubsetName] = canarySubset

	if _, ok := sr.Spec.Subsets[stableSubsetName]; !ok {
		return errors.New(fmt.Sprintf("spec.subsets.%s.filter was not found in consul service resolver: %v", canarySubsetName, sr))
	}
	stableSubset := sr.Spec.Subsets[stableSubsetName]
	stableSubset.Filter = fmt.Sprintf(filterServiceMetaVersionTemplate, serviceMetaVersion)
	sr.Spec.Subsets[stableSubsetName] = stableSubset

	// Persist changes to the ServiceResolver
	if err := r.K8SClient.Update(ctx, &sr, &client.UpdateOptions{}); err != nil {
		return pluginTypes.RpcError{ErrorString: err.Error()}
	}
	return nil
}

func (r *RpcPlugin) updateResolverForRollouts(ctx context.Context, canarySubsetName, serviceMetaVersion string, sr consulv1aplha1.ServiceResolver) error {
	if _, ok := sr.Spec.Subsets[canarySubsetName]; !ok {
		return errors.New(fmt.Sprintf("spec.subsets.%s.filter was not found in consul service resolver: %v", canarySubsetName, sr))
	}
	canarySubset := sr.Spec.Subsets[canarySubsetName]
	canarySubset.Filter = fmt.Sprintf(filterServiceMetaVersionTemplate, serviceMetaVersion)
	sr.Spec.Subsets[canarySubsetName] = canarySubset

	// Persist changes to the ServiceResolver
	if err := r.K8SClient.Update(ctx, &sr, &client.UpdateOptions{}); err != nil {
		return pluginTypes.RpcError{ErrorString: err.Error()}
	}
	return nil
}

func (r *RpcPlugin) updateResolverForAbortedRollout(ctx context.Context, canarySubsetName string, sr consulv1aplha1.ServiceResolver) error {
	if _, ok := sr.Spec.Subsets[canarySubsetName]; !ok {
		return errors.New(fmt.Sprintf("spec.subsets.%s.filter was not found in consul service resolver: %v", canarySubsetName, sr))
	}
	canarySubset := sr.Spec.Subsets[canarySubsetName]
	canarySubset.Filter = ""
	sr.Spec.Subsets[canarySubsetName] = canarySubset

	// Persist changes to the ServiceResolver
	if err := r.K8SClient.Update(ctx, &sr, &client.UpdateOptions{}); err != nil {
		return pluginTypes.RpcError{ErrorString: err.Error()}
	}
	return nil
}

func rolloutComplete(rollout *v1alpha1.Rollout) bool {
	rolloutCondition, err := completeCondition(rollout)
	if err != nil {
		return false
	}
	return strconv.FormatInt(rollout.GetObjectMeta().GetGeneration(), 10) == rollout.Status.ObservedGeneration &&
		rolloutCondition.Status == corev1.ConditionTrue
}

func completeCondition(rollout *v1alpha1.Rollout) (v1alpha1.RolloutCondition, error) {
	for i, condition := range rollout.Status.Conditions {
		if condition.Type == v1alpha1.RolloutCompleted {
			return rollout.Status.Conditions[i], nil
		}
	}
	return v1alpha1.RolloutCondition{}, errors.New("condition RolloutCompleted not found")
}

func rolloutAborted(rollout *v1alpha1.Rollout) bool {
	return rollout.Status.Abort
}

func getPluginConfig(rollout *v1alpha1.Rollout) (*ConsulTrafficRouting, error) {
	consulConfig := ConsulTrafficRouting{}
	if err := json.Unmarshal(rollout.Spec.Strategy.Canary.TrafficRouting.Plugins[ConfigKey], &consulConfig); err != nil {
		return nil, err
	}
	return &consulConfig, nil
}
