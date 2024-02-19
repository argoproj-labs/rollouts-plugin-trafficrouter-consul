// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: Apache-2.0

package plugin

// Type holds this plugin type
const Type = "Consul"

// ConfigKey used to identify the plugin in argo-rollouts configmap.
// see https://argoproj.github.io/argo-rollouts/features/traffic-management/plugins/
const ConfigKey = "hashicorp/consul"
