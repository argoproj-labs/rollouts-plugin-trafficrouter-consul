// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"

	"github.com/argoproj-labs/rollouts-plugin-trafficrouter-consul/pkg/plugin"

	"github.com/argoproj-labs/rollouts-plugin-trafficrouter-consul/pkg/version"
	rolloutsPlugin "github.com/argoproj/argo-rollouts/rollout/trafficrouting/plugin/rpc"
	goPlugin "github.com/hashicorp/go-plugin"
	log "github.com/sirupsen/logrus"
)

// handshakeConfigs are used to just do a basic handshake between
// a plugin and host. If the handshake fails, a user friendly error is shown.
// This prevents users from executing bad plugins or executing a plugin
// directory. It is a UX feature, not a security feature.
var handshakeConfig = goPlugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "ARGO_ROLLOUTS_RPC_PLUGIN",
	MagicCookieValue: "trafficrouter",
}

func main() {
	// Create a flag to print the version of the plugin
	// This is useful for debugging and support
	versionFlag := flag.Bool("version", false, "Print the version of the plugin")
	flag.Parse()
	if *versionFlag {
		fmt.Println(version.GetHumanVersion())
		return
	}

	logCtx := log.WithFields(log.Fields{"plugin": "trafficrouter"})
	log.SetLevel(log.InfoLevel)

	rpcPluginImp := &plugin.RpcPlugin{
		LogCtx: logCtx,
	}

	var pluginMap = map[string]goPlugin.Plugin{
		"RpcTrafficRouterPlugin": &rolloutsPlugin.RpcTrafficRouterPlugin{Impl: rpcPluginImp},
	}

	goPlugin.Serve(&goPlugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
