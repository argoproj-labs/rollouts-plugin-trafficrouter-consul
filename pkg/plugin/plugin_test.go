package plugin

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	pluginTypes "github.com/argoproj/argo-rollouts/utils/plugin/types"
	consulv1aplha1 "github.com/hashicorp/consul-k8s/control-plane/api/v1alpha1"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestSetWeight(t *testing.T) {
	testCases := []struct {
		testName         string
		rollout          *v1alpha1.Rollout
		desiredWeight    int32
		inputResolver    *consulv1aplha1.ServiceResolver
		inputSplitter    *consulv1aplha1.ServiceSplitter
		expectedResolver *consulv1aplha1.ServiceResolver
		expectedSplitter *consulv1aplha1.ServiceSplitter
		expectedError    pluginTypes.RpcError
	}{
		{
			testName: "desired weight 50",
			rollout: &v1alpha1.Rollout{
				ObjectMeta: v1.ObjectMeta{
					Name:       "rollout",
					Namespace:  "default",
					Generation: 10,
				},
				Spec: v1alpha1.RolloutSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: v1.ObjectMeta{
							Annotations: map[string]string{
								"consul.hashicorp.com/service-meta-version": "2",
							},
						},
					},
					Strategy: v1alpha1.RolloutStrategy{
						Canary: &v1alpha1.CanaryStrategy{
							TrafficRouting: &v1alpha1.RolloutTrafficRouting{
								Plugins: map[string]json.RawMessage{
									ConfigKey: pluginJson(),
								},
							},
						},
					},
				},
				Status: v1alpha1.RolloutStatus{
					ObservedGeneration: "10",
					Conditions: []v1alpha1.RolloutCondition{
						{
							Type:   v1alpha1.RolloutCompleted,
							Status: corev1.ConditionFalse,
						},
					},
					Canary: v1alpha1.CanaryStatus{
						Weights: &v1alpha1.TrafficWeights{
							Canary: v1alpha1.WeightDestination{
								Weight: 50,
							},
							Stable: v1alpha1.WeightDestination{
								Weight: 50,
							},
						},
					},
				},
			},
			desiredWeight: 50,
			inputResolver: defaultResolver(),
			inputSplitter: defaultSplitter(),
			expectedResolver: &consulv1aplha1.ServiceResolver{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
				Spec: consulv1aplha1.ServiceResolverSpec{
					Subsets: map[string]consulv1aplha1.ServiceResolverSubset{
						"stable": {
							Filter: "Service.Meta.version == 1",
						},
						"canary": {
							Filter: "Service.Meta.version == 2",
						},
					},
				},
			},
			expectedSplitter: &consulv1aplha1.ServiceSplitter{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
				Spec: consulv1aplha1.ServiceSplitterSpec{
					Splits: []consulv1aplha1.ServiceSplit{
						{
							Weight:        50,
							ServiceSubset: "canary",
						},
						{
							Weight:        50,
							ServiceSubset: "stable",
						},
					},
				},
			},
			expectedError: pluginTypes.RpcError{},
		},
	}

	for _, testCase := range testCases {
		s := runtime.NewScheme()
		require.NoError(t, consulv1aplha1.AddToScheme(s))

		objs := []client.Object{}
		objs = append(objs, testCase.inputResolver, testCase.inputSplitter)
		namespacedName := types.NamespacedName{Name: "test-service", Namespace: "default"}

		k8sClient := fake.NewClientBuilder().WithScheme(s).WithObjects(objs...).Build()
		p := &RpcPlugin{
			K8SClient: k8sClient,
			IsTest:    true,
			LogCtx:    logrus.NewEntry(logrus.New()),
		}
		err := p.SetWeight(testCase.rollout, testCase.desiredWeight, []v1alpha1.WeightDestination{})
		require.Equal(t, testCase.expectedError, err, "errors should be equal")
		actualResolver := &consulv1aplha1.ServiceResolver{}
		actualSplitter := &consulv1aplha1.ServiceSplitter{}
		require.NoError(t, k8sClient.Get(context.TODO(), namespacedName, actualResolver, &client.GetOptions{}))
		require.NoError(t, k8sClient.Get(context.TODO(), namespacedName, actualSplitter, &client.GetOptions{}))
		require.ElementsMatch(t, testCase.expectedSplitter.Spec.Splits, actualSplitter.Spec.Splits)
		require.Equal(t, testCase.expectedResolver.Spec.Subsets["canary"], actualResolver.Spec.Subsets["canary"])
		require.Equal(t, testCase.expectedResolver.Spec.Subsets["stable"], actualResolver.Spec.Subsets["stable"])
	}
}

func pluginJson() []byte {
	config := ConsulTrafficRouting{
		ServiceName:      "test-service",
		CanarySubsetName: "canary",
		StableSubsetName: "stable",
	}
	jsonConfig, _ := json.Marshal(config)
	return jsonConfig
}

func defaultSplitter() *consulv1aplha1.ServiceSplitter {
	return &consulv1aplha1.ServiceSplitter{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-service",
			Namespace: "default",
		},
		Spec: consulv1aplha1.ServiceSplitterSpec{
			Splits: []consulv1aplha1.ServiceSplit{
				{
					Weight:        100,
					ServiceSubset: "stable",
				},
				{
					Weight:        0,
					ServiceSubset: "canary",
				},
			},
		},
	}
}

func defaultResolver() *consulv1aplha1.ServiceResolver {
	return &consulv1aplha1.ServiceResolver{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-service",
			Namespace: "default",
		},
		Spec: consulv1aplha1.ServiceResolverSpec{
			Subsets: map[string]consulv1aplha1.ServiceResolverSubset{
				"stable": {
					Filter: "Service.Meta.version == 1",
				},
				"canary": {
					Filter: "",
				},
			},
		},
	}
}
