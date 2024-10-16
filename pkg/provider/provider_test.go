package provider

import (
	"os"
	"testing"

	"github.com/kairos-io/kairos-sdk/clusterplugin"
	kyaml "sigs.k8s.io/yaml"

	"github.com/spectrocloud/provider-nodeadm/pkg/domain"
	"github.com/spectrocloud/provider-nodeadm/pkg/stages"
)

func TestNodeadmProvider(t *testing.T) {
	netConfig, _ := os.ReadFile("testdata/network-configuration.json")
	nodeConfig, _ := os.ReadFile("testdata/node-configuration.json")
	expected, _ := os.ReadFile("testdata/expected.yaml")

	cluster := clusterplugin.Cluster{
		ProviderOptions: map[string]string{
			domain.CredentialProviderKey:   "iam-ra",
			domain.KubernetesVersionKey:    "1.30.0",
			domain.NetworkConfigurationKey: string(netConfig),
			domain.NodeConfigurationKey:    string(nodeConfig),
		},
	}
	stages.InitPaths(cluster)
	schema := NodeadmProvider(cluster)
	got, _ := kyaml.Marshal(schema)

	if string(got) != string(expected) {
		_ = os.WriteFile("testdata/got.yaml", got, 0644)
		t.Errorf("Expected %s, got %s", string(expected), string(got))
	}
}
