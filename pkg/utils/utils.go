// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	pluginTypes "github.com/argoproj/argo-rollouts/utils/plugin/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewKubeConfig() (*rest.Config, error) {
	config, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		return nil, pluginTypes.RpcError{ErrorString: err.Error()}
	}
	return config, nil
}
