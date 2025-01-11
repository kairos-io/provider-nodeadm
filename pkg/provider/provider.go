// Package provider contains the Kairos provider for EKS hybrid edge nodes.
package provider

import (
	"encoding/json"
	"fmt"

	"github.com/kairos-io/kairos-sdk/bus"
	"github.com/kairos-io/kairos-sdk/clusterplugin"
	"github.com/mudler/go-pluggable"
	yip "github.com/mudler/yip/pkg/schema"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	kyaml "sigs.k8s.io/yaml"

	"github.com/kairos-io/provider-nodeadm/pkg/domain"
	"github.com/kairos-io/provider-nodeadm/pkg/stages"
)

// NodeadmProvider is the kairos cluster provider for EKS hybrid edge nodes.
func NodeadmProvider(cluster clusterplugin.Cluster) yip.YipConfig {
	var nc domain.NodeadmConfig

	// Parse containerd and kubelet configuration overrides from the user-provided
	// cluster options, AKA nodeadm pack values
	if cluster.Options != "" {
		userOptions, _ := kyaml.YAMLToJSON([]byte(cluster.Options))
		_ = json.Unmarshal(userOptions, &nc.NodeConfiguration.UserConfig)
	}

	// Map provider options into the NodeadmConfig struct

	// Credential provider
	credentialProvider, ok := cluster.ProviderOptions[domain.CredentialProviderKey]
	if !ok {
		logrus.Fatalf("missing mandatory provider option %s", domain.CredentialProviderKey)
	}
	nc.CredentialProvider = domain.CredentialProvider(credentialProvider)

	// Kubernetes version
	nc.KubernetesVersion, ok = cluster.ProviderOptions[domain.KubernetesVersionKey]
	if !ok {
		logrus.Fatalf("missing mandatory provider option %s", domain.KubernetesVersionKey)
	}

	// Network configuration
	netConfig, ok := cluster.ProviderOptions[domain.NetworkConfigurationKey]
	if !ok {
		logrus.Fatalf("missing mandatory provider option %s", domain.NetworkConfigurationKey)
	}
	if err := json.Unmarshal([]byte(netConfig), &nc.NetworkConfiguration); err != nil {
		logrus.Fatalf("failed to unmarshal network configuration %+v: %v", netConfig, err)
	}

	// Node configuration
	nodeConfig, ok := cluster.ProviderOptions[domain.NodeConfigurationKey]
	if !ok {
		logrus.Fatalf("missing mandatory provider option %s", domain.NodeConfigurationKey)
	}
	if err := json.Unmarshal([]byte(nodeConfig), &nc.NodeConfiguration); err != nil {
		logrus.Fatalf("failed to unmarshal node configuration %+v: %v", nodeConfig, err)
	}

	// Generate yip stages
	stages.InitPaths(cluster)

	var proxyArgs string
	if stages.IsProxyConfigured(cluster.Env) {
		proxyArgs = fmt.Sprintf("%t %s %s %s",
			true,
			cluster.Env["HTTP_PROXY"],
			cluster.Env["HTTPS_PROXY"],
			stages.GetNoProxyConfig(cluster.Env, nc),
		)
	}

	bootBeforeStages := stages.PreInstallYipStages(cluster.Env, nc)
	bootBeforeStages = append(bootBeforeStages, stages.InitYipStages(nc, proxyArgs)...)

	cfg := yip.YipConfig{
		Name: "Kairos Provider Nodeadm",
		Stages: map[string][]yip.Stage{
			"boot.before": bootBeforeStages,
		},
	}

	return cfg
}

// Reset handles cluster reset events.
func Reset(event *pluggable.Event) pluggable.EventResponse {
	var (
		payload  bus.EventPayload
		config   clusterplugin.Config
		response pluggable.EventResponse
	)

	// parse the boot payload
	if err := json.Unmarshal([]byte(event.Data), &payload); err != nil {
		response.Error = fmt.Sprintf("failed to parse boot event: %v", err)
		return response
	}

	// parse config from boot payload
	if err := yaml.Unmarshal([]byte(payload.Config), &config); err != nil {
		response.Error = fmt.Sprintf("failed to parse config from boot event: %v", err)
		return response
	}

	if config.Cluster == nil {
		return response
	}
	stages.InitPaths(*config.Cluster)

	output, err := stages.ResetCmd().CombinedOutput()
	if err != nil {
		response.Error = fmt.Sprintf("failed to reset cluster: %s", string(output))
	}

	return response
}
