// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	consulv1aplha1 "github.com/hashicorp/consul-k8s/control-plane/api/v1alpha1"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		expectedError    string
	}{
		{
			testName: "in progress, desired weight 50",
			rollout: &v1alpha1.Rollout{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "rollout",
					Namespace:  "default",
					Generation: 10,
				},
				Spec: v1alpha1.RolloutSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
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
				ObjectMeta: metav1.ObjectMeta{
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
				ObjectMeta: metav1.ObjectMeta{
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
		},
		{
			testName: "in progress, desired weight 25",
			rollout: &v1alpha1.Rollout{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "rollout",
					Namespace:  "default",
					Generation: 10,
				},
				Spec: v1alpha1.RolloutSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
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
								Weight: 25,
							},
							Stable: v1alpha1.WeightDestination{
								Weight: 75,
							},
						},
					},
				},
			},
			desiredWeight: 25,
			inputResolver: defaultResolver(),
			inputSplitter: defaultSplitter(),
			expectedResolver: &consulv1aplha1.ServiceResolver{
				ObjectMeta: metav1.ObjectMeta{
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
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
				Spec: consulv1aplha1.ServiceSplitterSpec{
					Splits: []consulv1aplha1.ServiceSplit{
						{
							Weight:        25,
							ServiceSubset: "canary",
						},
						{
							Weight:        75,
							ServiceSubset: "stable",
						},
					},
				},
			},
		},
		{
			testName: "in progress, desired weight 75",
			rollout: &v1alpha1.Rollout{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "rollout",
					Namespace:  "default",
					Generation: 10,
				},
				Spec: v1alpha1.RolloutSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
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
								Weight: 75,
							},
							Stable: v1alpha1.WeightDestination{
								Weight: 25,
							},
						},
					},
				},
			},
			desiredWeight: 75,
			inputResolver: defaultResolver(),
			inputSplitter: defaultSplitter(),
			expectedResolver: &consulv1aplha1.ServiceResolver{
				ObjectMeta: metav1.ObjectMeta{
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
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
				Spec: consulv1aplha1.ServiceSplitterSpec{
					Splits: []consulv1aplha1.ServiceSplit{
						{
							Weight:        75,
							ServiceSubset: "canary",
						},
						{
							Weight:        25,
							ServiceSubset: "stable",
						},
					},
				},
			},
		},
		{
			testName: "in progress, desired weight 0",
			rollout: &v1alpha1.Rollout{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "rollout",
					Namespace:  "default",
					Generation: 10,
				},
				Spec: v1alpha1.RolloutSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
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
								Weight: 0,
							},
							Stable: v1alpha1.WeightDestination{
								Weight: 100,
							},
						},
					},
				},
			},
			desiredWeight: 0,
			inputResolver: defaultResolver(),
			inputSplitter: defaultSplitter(),
			expectedResolver: &consulv1aplha1.ServiceResolver{
				ObjectMeta: metav1.ObjectMeta{
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
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
				Spec: consulv1aplha1.ServiceSplitterSpec{
					Splits: []consulv1aplha1.ServiceSplit{
						{
							Weight:        0,
							ServiceSubset: "canary",
						},
						{
							Weight:        100,
							ServiceSubset: "stable",
						},
					},
				},
			},
		},
		{
			testName: "completed, desired weight 0",
			rollout: &v1alpha1.Rollout{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "rollout",
					Namespace:  "default",
					Generation: 10,
				},
				Spec: v1alpha1.RolloutSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
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
							Status: corev1.ConditionTrue,
						},
					},
					Canary: v1alpha1.CanaryStatus{
						Weights: &v1alpha1.TrafficWeights{
							Canary: v1alpha1.WeightDestination{
								Weight: 0,
							},
							Stable: v1alpha1.WeightDestination{
								Weight: 100,
							},
						},
					},
				},
			},
			desiredWeight: 0,
			inputResolver: defaultResolver(),
			inputSplitter: defaultSplitter(),
			expectedResolver: &consulv1aplha1.ServiceResolver{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
				Spec: consulv1aplha1.ServiceResolverSpec{
					Subsets: map[string]consulv1aplha1.ServiceResolverSubset{
						"stable": {
							Filter: "Service.Meta.version == 2",
						},
						"canary": {
							Filter: "",
						},
					},
				},
			},
			expectedSplitter: &consulv1aplha1.ServiceSplitter{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
				Spec: consulv1aplha1.ServiceSplitterSpec{
					Splits: []consulv1aplha1.ServiceSplit{
						{
							Weight:        0,
							ServiceSubset: "canary",
						},
						{
							Weight:        100,
							ServiceSubset: "stable",
						},
					},
				},
			},
		},
		{
			testName: "in progress, desired weight 100",
			rollout: &v1alpha1.Rollout{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "rollout",
					Namespace:  "default",
					Generation: 10,
				},
				Spec: v1alpha1.RolloutSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
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
								Weight: 100,
							},
							Stable: v1alpha1.WeightDestination{
								Weight: 0,
							},
						},
					},
				},
			},
			desiredWeight: 100,
			inputResolver: defaultResolver(),
			inputSplitter: defaultSplitter(),
			expectedResolver: &consulv1aplha1.ServiceResolver{
				ObjectMeta: metav1.ObjectMeta{
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
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
				Spec: consulv1aplha1.ServiceSplitterSpec{
					Splits: []consulv1aplha1.ServiceSplit{
						{
							Weight:        100,
							ServiceSubset: "canary",
						},
						{
							Weight:        0,
							ServiceSubset: "stable",
						},
					},
				},
			},
		},
		{
			testName: "aborted rollout",
			rollout: &v1alpha1.Rollout{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "rollout",
					Namespace:  "default",
					Generation: 10,
				},
				Spec: v1alpha1.RolloutSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
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
					Abort:              true,
					AbortedAt:          &metav1.Time{Time: time.Now()},
					Conditions: []v1alpha1.RolloutCondition{
						{
							Type:   v1alpha1.RolloutCompleted,
							Status: corev1.ConditionFalse,
						},
					},
					Canary: v1alpha1.CanaryStatus{
						Weights: &v1alpha1.TrafficWeights{
							Canary: v1alpha1.WeightDestination{
								Weight: 100,
							},
							Stable: v1alpha1.WeightDestination{
								Weight: 0,
							},
						},
					},
				},
			},
			desiredWeight: 0,
			inputResolver: defaultResolver(),
			inputSplitter: defaultSplitter(),
			expectedResolver: &consulv1aplha1.ServiceResolver{
				ObjectMeta: metav1.ObjectMeta{
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
			},
			expectedSplitter: &consulv1aplha1.ServiceSplitter{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
				Spec: consulv1aplha1.ServiceSplitterSpec{
					Splits: []consulv1aplha1.ServiceSplit{
						{
							Weight:        0,
							ServiceSubset: "canary",
						},
						{
							Weight:        100,
							ServiceSubset: "stable",
						},
					},
				},
			},
		},
		{
			testName: "in progress, desired weight 50, non-default-suffix",
			rollout: &v1alpha1.Rollout{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "rollout",
					Namespace:  "default",
					Generation: 10,
				},
				Spec: v1alpha1.RolloutSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								"consul.hashicorp.com/service-meta-number": "2",
							},
						},
					},
					Strategy: v1alpha1.RolloutStrategy{
						Canary: &v1alpha1.CanaryStrategy{
							TrafficRouting: &v1alpha1.RolloutTrafficRouting{
								Plugins: map[string]json.RawMessage{
									ConfigKey: pluginJsonWithSuffix("number"),
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
			inputResolver: &consulv1aplha1.ServiceResolver{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
				Spec: consulv1aplha1.ServiceResolverSpec{
					Subsets: map[string]consulv1aplha1.ServiceResolverSubset{
						"stable": {
							Filter: "Service.Meta.number == 1",
						},
						"canary": {
							Filter: "",
						},
					},
				},
			},
			inputSplitter: defaultSplitter(),
			expectedResolver: &consulv1aplha1.ServiceResolver{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
				Spec: consulv1aplha1.ServiceResolverSpec{
					Subsets: map[string]consulv1aplha1.ServiceResolverSubset{
						"stable": {
							Filter: "Service.Meta.number == 1",
						},
						"canary": {
							Filter: "Service.Meta.number == 2",
						},
					},
				},
			},
			expectedSplitter: &consulv1aplha1.ServiceSplitter{
				ObjectMeta: metav1.ObjectMeta{
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
		},
		{
			testName: "empty canary status returns immediately",
			rollout: &v1alpha1.Rollout{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "rollout",
					Namespace:  "default",
					Generation: 10,
				},
				Spec: v1alpha1.RolloutSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								"consul.hashicorp.com/service-meta-number": "2",
							},
						},
					},
					Strategy: v1alpha1.RolloutStrategy{
						Canary: &v1alpha1.CanaryStrategy{
							TrafficRouting: &v1alpha1.RolloutTrafficRouting{
								Plugins: map[string]json.RawMessage{
									ConfigKey: pluginJsonWithSuffix("number"),
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
					Canary: v1alpha1.CanaryStatus{},
				},
			},
			desiredWeight:    50,
			inputResolver:    defaultResolver(),
			inputSplitter:    defaultSplitter(),
			expectedResolver: defaultResolver(),
			expectedSplitter: defaultSplitter(),
		},
		{
			testName: "error invalid rollout config",
			rollout: &v1alpha1.Rollout{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "rollout",
					Namespace:  "default",
					Generation: 10,
				},
				Spec: v1alpha1.RolloutSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								"consul.hashicorp.com/service-meta-number": "2",
							},
						},
					},
					Strategy: v1alpha1.RolloutStrategy{
						Canary: &v1alpha1.CanaryStrategy{
							TrafficRouting: &v1alpha1.RolloutTrafficRouting{
								Plugins: map[string]json.RawMessage{
									ConfigKey: invalidPlugin(),
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
			desiredWeight:    50,
			inputResolver:    defaultResolver(),
			inputSplitter:    defaultSplitter(),
			expectedResolver: nil,
			expectedSplitter: nil,
			expectedError:    "invalid consul traffic routing configuration.",
		},
		{
			testName: "error missing resolver",
			rollout: &v1alpha1.Rollout{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "rollout",
					Namespace:  "default",
					Generation: 10,
				},
				Spec: v1alpha1.RolloutSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
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
			desiredWeight:    50,
			inputResolver:    nil,
			inputSplitter:    defaultSplitter(),
			expectedResolver: nil,
			expectedSplitter: nil,
			expectedError:    "serviceresolvers.consul.hashicorp.com \"test-service\" not found",
		},
		{
			testName: "error missing splitter",
			rollout: &v1alpha1.Rollout{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "rollout",
					Namespace:  "default",
					Generation: 10,
				},
				Spec: v1alpha1.RolloutSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
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
			desiredWeight:    50,
			inputResolver:    defaultResolver(),
			inputSplitter:    nil,
			expectedResolver: nil,
			expectedSplitter: nil,
			expectedError:    "servicesplitters.consul.hashicorp.com \"test-service\" not found",
		},
		{
			testName: "error in progress rollout invalid resolver",
			rollout: &v1alpha1.Rollout{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "rollout",
					Namespace:  "default",
					Generation: 10,
				},
				Spec: v1alpha1.RolloutSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
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
			inputResolver: &consulv1aplha1.ServiceResolver{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
				Spec: consulv1aplha1.ServiceResolverSpec{
					Subsets: map[string]consulv1aplha1.ServiceResolverSubset{
						"foo": {
							Filter: "Service.Meta.version == 1",
						},
						"bar": {
							Filter: "",
						},
					},
				},
			},
			inputSplitter: defaultSplitter(),
			expectedError: "spec.subsets.canary.filter was not found in consul service resolver",
		},
		{
			testName: "error aborted rollout invalid resolver",
			rollout: &v1alpha1.Rollout{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "rollout",
					Namespace:  "default",
					Generation: 10,
				},
				Spec: v1alpha1.RolloutSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
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
					Abort:              true,
					AbortedAt:          &metav1.Time{Time: time.Now()},
					Conditions: []v1alpha1.RolloutCondition{
						{
							Type:   v1alpha1.RolloutCompleted,
							Status: corev1.ConditionFalse,
						},
					},
					Canary: v1alpha1.CanaryStatus{
						Weights: &v1alpha1.TrafficWeights{
							Canary: v1alpha1.WeightDestination{
								Weight: 100,
							},
							Stable: v1alpha1.WeightDestination{
								Weight: 0,
							},
						},
					},
				},
			},
			desiredWeight: 0,
			inputResolver: &consulv1aplha1.ServiceResolver{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
				Spec: consulv1aplha1.ServiceResolverSpec{
					Subsets: map[string]consulv1aplha1.ServiceResolverSubset{
						"foo": {
							Filter: "Service.Meta.version == 1",
						},
						"bar": {
							Filter: "",
						},
					},
				},
			},
			inputSplitter: defaultSplitter(),
			expectedError: "spec.subsets.canary.filter was not found in consul service resolver",
		},
		{
			testName: "error completed rollout invalid resolver invalid canary subset",
			rollout: &v1alpha1.Rollout{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "rollout",
					Namespace:  "default",
					Generation: 10,
				},
				Spec: v1alpha1.RolloutSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
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
							Status: corev1.ConditionTrue,
						},
					},
					Canary: v1alpha1.CanaryStatus{
						Weights: &v1alpha1.TrafficWeights{
							Canary: v1alpha1.WeightDestination{
								Weight: 0,
							},
							Stable: v1alpha1.WeightDestination{
								Weight: 100,
							},
						},
					},
				},
			},
			desiredWeight: 0,
			inputResolver: &consulv1aplha1.ServiceResolver{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
				Spec: consulv1aplha1.ServiceResolverSpec{
					Subsets: map[string]consulv1aplha1.ServiceResolverSubset{
						"foo": {
							Filter: "Service.Meta.version == 1",
						},
						"stable": {
							Filter: "",
						},
					},
				},
			},
			inputSplitter: defaultSplitter(),
			expectedError: "spec.subsets.canary.filter was not found in consul service resolver",
		},
		{
			testName: "error completed rollout invalid resolver invalid stable subset",
			rollout: &v1alpha1.Rollout{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "rollout",
					Namespace:  "default",
					Generation: 10,
				},
				Spec: v1alpha1.RolloutSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
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
							Status: corev1.ConditionTrue,
						},
					},
					Canary: v1alpha1.CanaryStatus{
						Weights: &v1alpha1.TrafficWeights{
							Canary: v1alpha1.WeightDestination{
								Weight: 0,
							},
							Stable: v1alpha1.WeightDestination{
								Weight: 100,
							},
						},
					},
				},
			},
			desiredWeight: 0,
			inputResolver: &consulv1aplha1.ServiceResolver{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
				Spec: consulv1aplha1.ServiceResolverSpec{
					Subsets: map[string]consulv1aplha1.ServiceResolverSubset{
						"canary": {
							Filter: "Service.Meta.version == 1",
						},
						"bar": {
							Filter: "",
						},
					},
				},
			},
			inputSplitter: defaultSplitter(),
			expectedError: "spec.subsets.stable.filter was not found in consul service resolver",
		},
		{
			testName: "error missing splitter subsets",
			rollout: &v1alpha1.Rollout{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "rollout",
					Namespace:  "default",
					Generation: 10,
				},
				Spec: v1alpha1.RolloutSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
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
			inputSplitter: &consulv1aplha1.ServiceSplitter{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
				Spec: consulv1aplha1.ServiceSplitterSpec{
					Splits: []consulv1aplha1.ServiceSplit{},
				},
			},
			expectedError: "spec.splits was not found in consul service splitter",
		},
		{
			testName: "error invalid number of splitter subsets",
			rollout: &v1alpha1.Rollout{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "rollout",
					Namespace:  "default",
					Generation: 10,
				},
				Spec: v1alpha1.RolloutSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
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
			inputSplitter: &consulv1aplha1.ServiceSplitter{
				ObjectMeta: metav1.ObjectMeta{
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
						{
							Weight:        0,
							ServiceSubset: "other",
						},
					},
				},
			},
			expectedError: "unexpected number of service splits. Expected 2, found 3",
		},
		{
			testName: "error invalid splitter subset names",
			rollout: &v1alpha1.Rollout{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "rollout",
					Namespace:  "default",
					Generation: 10,
				},
				Spec: v1alpha1.RolloutSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
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
			inputSplitter: &consulv1aplha1.ServiceSplitter{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
				Spec: consulv1aplha1.ServiceSplitterSpec{
					Splits: []consulv1aplha1.ServiceSplit{
						{
							Weight:        100,
							ServiceSubset: "foo",
						},
						{
							Weight:        0,
							ServiceSubset: "bar",
						},
					},
				},
			},
			expectedError: "unexpected service split",
		},
		{
			testName: "error invalid resolver status not synced",
			rollout: &v1alpha1.Rollout{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "rollout",
					Namespace:  "default",
					Generation: 10,
				},
				Spec: v1alpha1.RolloutSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
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
			inputResolver: &consulv1aplha1.ServiceResolver{
				ObjectMeta: metav1.ObjectMeta{
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
				Status: consulv1aplha1.Status{
					Conditions: []consulv1aplha1.Condition{
						{
							Type:   consulv1aplha1.ConditionSynced,
							Status: corev1.ConditionFalse,
						},
					},
				},
			},
			inputSplitter: defaultSplitter(),
			expectedError: "service resolver has not synced with Consul. The service resolver needs to be up to date before rollout can continue",
		},
		{
			testName: "error invalid resolver last synced time mismatch",
			rollout: &v1alpha1.Rollout{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "rollout",
					Namespace:  "default",
					Generation: 10,
				},
				Spec: v1alpha1.RolloutSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
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
			inputResolver: &consulv1aplha1.ServiceResolver{
				ObjectMeta: metav1.ObjectMeta{
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
				Status: consulv1aplha1.Status{
					Conditions: []consulv1aplha1.Condition{
						{
							Type:               consulv1aplha1.ConditionSynced,
							Status:             corev1.ConditionTrue,
							LastTransitionTime: metav1.Time{Time: unknownSyncTime(t)},
						},
					},
					LastSyncedTime: &metav1.Time{Time: time.Now()},
				},
			},
			inputSplitter: defaultSplitter(),
			expectedError: "service resolver has not synced with Consul. The service resolver needs to be up to date before rollout can continue",
		},
		{
			testName: "error invalid splitter status not synced",
			rollout: &v1alpha1.Rollout{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "rollout",
					Namespace:  "default",
					Generation: 10,
				},
				Spec: v1alpha1.RolloutSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
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
			inputSplitter: &consulv1aplha1.ServiceSplitter{
				ObjectMeta: metav1.ObjectMeta{
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
				Status: consulv1aplha1.Status{
					Conditions: []consulv1aplha1.Condition{
						{
							Type:   consulv1aplha1.ConditionSynced,
							Status: corev1.ConditionFalse,
						},
					},
				},
			},
			expectedError: "service splitter has not synced with Consul. The service splitter needs to be up to date before rollout can continue",
		},
		{
			testName: "error invalid splitter last synced time mismatch",
			rollout: &v1alpha1.Rollout{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "rollout",
					Namespace:  "default",
					Generation: 10,
				},
				Spec: v1alpha1.RolloutSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
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
			inputSplitter: &consulv1aplha1.ServiceSplitter{
				ObjectMeta: metav1.ObjectMeta{
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
				Status: consulv1aplha1.Status{
					Conditions: []consulv1aplha1.Condition{
						{
							Type:               consulv1aplha1.ConditionSynced,
							Status:             corev1.ConditionTrue,
							LastTransitionTime: metav1.Time{Time: unknownSyncTime(t)},
						},
					},
					LastSyncedTime: &metav1.Time{Time: time.Now()},
				},
			},
			expectedError: "service splitter has not synced with Consul. The service splitter needs to be up to date before rollout can continue",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			s := runtime.NewScheme()
			require.NoError(t, consulv1aplha1.AddToScheme(s))

			objs := []client.Object{}

			if testCase.inputResolver != nil {
				objs = append(objs, testCase.inputResolver)
			}
			if testCase.inputSplitter != nil {
				objs = append(objs, testCase.inputSplitter)
			}

			namespacedName := types.NamespacedName{Name: "test-service", Namespace: "default"}

			k8sClient := fake.NewClientBuilder().WithScheme(s).WithObjects(objs...).Build()
			p := &RpcPlugin{
				K8SClient: k8sClient,
				IsTest:    true,
				LogCtx:    logrus.NewEntry(logrus.New()),
			}
			err := p.SetWeight(testCase.rollout, testCase.desiredWeight, []v1alpha1.WeightDestination{})
			if testCase.expectedError == "" {
				actualResolver := &consulv1aplha1.ServiceResolver{}
				actualSplitter := &consulv1aplha1.ServiceSplitter{}
				require.NoError(t, k8sClient.Get(context.TODO(), namespacedName, actualResolver, &client.GetOptions{}))
				require.NoError(t, k8sClient.Get(context.TODO(), namespacedName, actualSplitter, &client.GetOptions{}))
				require.ElementsMatch(t, testCase.expectedSplitter.Spec.Splits, actualSplitter.Spec.Splits)
				require.Equal(t, testCase.expectedResolver.Spec.Subsets["canary"], actualResolver.Spec.Subsets["canary"])
				require.Equal(t, testCase.expectedResolver.Spec.Subsets["stable"], actualResolver.Spec.Subsets["stable"])
			} else {
				require.Contains(t, err.ErrorString, testCase.expectedError)
			}
		})
	}
}

func TestRpcPluginType(t *testing.T) {
	p := &RpcPlugin{}
	require.Equal(t, Type, p.Type())
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

func pluginJsonWithSuffix(suffix string) []byte {
	config := ConsulTrafficRouting{
		ServiceName:                 "test-service",
		CanarySubsetName:            "canary",
		StableSubsetName:            "stable",
		ServiceMetaAnnotationSuffix: suffix,
	}
	jsonConfig, _ := json.Marshal(config)
	return jsonConfig
}

func invalidPlugin() []byte {
	config := struct{}{}
	jsonConfig, _ := json.Marshal(config)
	return jsonConfig
}

func defaultSplitter() *consulv1aplha1.ServiceSplitter {
	return &consulv1aplha1.ServiceSplitter{
		ObjectMeta: metav1.ObjectMeta{
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
		Status: consulv1aplha1.Status{
			Conditions: []consulv1aplha1.Condition{
				{
					Type:               consulv1aplha1.ConditionSynced,
					Status:             corev1.ConditionTrue,
					LastTransitionTime: metav1.Time{Time: time.Now()},
				},
			},
			LastSyncedTime: &metav1.Time{Time: time.Now()},
		},
	}
}

func defaultResolver() *consulv1aplha1.ServiceResolver {
	return &consulv1aplha1.ServiceResolver{
		ObjectMeta: metav1.ObjectMeta{
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
		Status: consulv1aplha1.Status{
			Conditions: []consulv1aplha1.Condition{
				{
					Type:               consulv1aplha1.ConditionSynced,
					Status:             corev1.ConditionTrue,
					LastTransitionTime: metav1.Time{Time: time.Now()},
				},
			},
			LastSyncedTime: &metav1.Time{Time: time.Now()},
		},
	}
}

func unknownSyncTime(t *testing.T) time.Time {
	const layout = "2006-01-02 15:04:05"
	timeString := "2023-03-06 08:30:00"
	parsedTime, err := time.Parse(layout, timeString)
	require.NoError(t, err)
	return parsedTime
}
