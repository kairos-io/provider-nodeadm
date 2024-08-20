// Package provider contains the Kairos provider for EKS hybrid edge nodes.
package provider

import (
	"encoding/json"
	"fmt"

	"github.com/kairos-io/kairos-sdk/clusterplugin"
	yip "github.com/mudler/yip/pkg/schema"
	kyaml "sigs.k8s.io/yaml"

	"github.com/spectrocloud-labs/provider-nodeadm/pkg/domain"
	"github.com/spectrocloud-labs/provider-nodeadm/pkg/stages"
)

// NodeadmProvider is the kairos cluster provider for EKS hybrid edge nodes.
func NodeadmProvider(cluster clusterplugin.Cluster) yip.YipConfig {
	var nc domain.NodeadmConfig

	// Unmarshal user-provided options into the NodeadmConfig struct
	if cluster.Options != "" {
		userOptions, _ := kyaml.YAMLToJSON([]byte(cluster.Options))
		_ = json.Unmarshal(userOptions, &nc)
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
