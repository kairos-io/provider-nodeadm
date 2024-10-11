// Package provider contains the Kairos provider for EKS hybrid edge nodes.
package provider

import (
	"encoding/json"
	"fmt"

	"github.com/kairos-io/kairos-sdk/clusterplugin"
	yip "github.com/mudler/yip/pkg/schema"
	"github.com/sirupsen/logrus"

	"github.com/spectrocloud-labs/provider-nodeadm/pkg/domain"
	"github.com/spectrocloud-labs/provider-nodeadm/pkg/stages"
)

// NodeadmProvider is the kairos cluster provider for EKS hybrid edge nodes.
func NodeadmProvider(cluster clusterplugin.Cluster) yip.YipConfig {
	var nc domain.NodeadmConfig

	// As of now, we don't have any options to unmarshal from the cluster.Options field.
	// Eventually they may come from the k8s layer of the edge cluster profile if/when
	// nodeadm supports advanced configuration options for the kubelet, containerd, etc.
	//
	// if cluster.Options != "" {
	// 	userOptions, _ := kyaml.YAMLToJSON([]byte(cluster.Options))
	// 	_ = json.Unmarshal(userOptions, &nc)
	// }

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
	bootBeforeStages = append(bootBeforeStages, stages.InstallYipStages(nc, proxyArgs)...)
	bootBeforeStages = append(bootBeforeStages, stages.InitYipStages(nc, proxyArgs)...)

	cfg := yip.YipConfig{
		Name: "Kairos Provider Nodeadm",
		Stages: map[string][]yip.Stage{
			"boot.before": bootBeforeStages,
		},
	}

	return cfg
}
