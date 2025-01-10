package provider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kairos-io/kairos-sdk/clusterplugin"
	kyaml "sigs.k8s.io/yaml"

	"github.com/kairos-io/provider-nodeadm/pkg/domain"
	"github.com/kairos-io/provider-nodeadm/pkg/stages"
)

type testCase struct {
	name        string
	clusterFunc func(tc testCase) clusterplugin.Cluster
}

func (t testCase) configs() (string, string) {
	netConfig, _ := os.ReadFile(filepath.Join("testdata", t.name, "network-configuration.json"))
	nodeConfig, _ := os.ReadFile(filepath.Join("testdata", t.name, "node-configuration.json"))
	return string(netConfig), string(nodeConfig)
}

func (t testCase) expected() string {
	expected, _ := os.ReadFile(filepath.Join("testdata", t.name, "/expected.yaml"))
	return string(expected)
}

func TestNodeadmProvider(t *testing.T) {
	tests := []testCase{
		{
			name: "iam-ra-agent",
			clusterFunc: func(tc testCase) clusterplugin.Cluster {
				netConfig, nodeConfig := tc.configs()
				return clusterplugin.Cluster{
					ProviderOptions: map[string]string{
						domain.CredentialProviderKey:   string(domain.CredentialProviderIAMRolesAnywhere),
						domain.KubernetesVersionKey:    "1.30.0",
						domain.NetworkConfigurationKey: netConfig,
						domain.NodeConfigurationKey:    nodeConfig,
						domain.HandleDependenciesKey:   "true",
					},
				}
			},
		},
		{
			name: "iam-ra-appliance",
			clusterFunc: func(tc testCase) clusterplugin.Cluster {
				netConfig, nodeConfig := tc.configs()
				return clusterplugin.Cluster{
					ProviderOptions: map[string]string{
						domain.CredentialProviderKey:   string(domain.CredentialProviderIAMRolesAnywhere),
						domain.KubernetesVersionKey:    "1.30.0",
						domain.NetworkConfigurationKey: netConfig,
						domain.NodeConfigurationKey:    nodeConfig,
						domain.HandleDependenciesKey:   "false",
					},
				}
			},
		},
		{
			name: "ssm-agent",
			clusterFunc: func(tc testCase) clusterplugin.Cluster {
				netConfig, nodeConfig := tc.configs()
				return clusterplugin.Cluster{
					ProviderOptions: map[string]string{
						domain.CredentialProviderKey:   string(domain.CredentialProviderSystemsManager),
						domain.KubernetesVersionKey:    "1.30.0",
						domain.NetworkConfigurationKey: netConfig,
						domain.NodeConfigurationKey:    nodeConfig,
						domain.HandleDependenciesKey:   "true",
					},
				}
			},
		},
		{
			name: "ssm-custom-agent",
			clusterFunc: func(tc testCase) clusterplugin.Cluster {
				netConfig, nodeConfig := tc.configs()
				return clusterplugin.Cluster{
					Options: `
containerd:
  config: |
    [plugins."io.containerd.grpc.v1.cri".containerd]
    discard_unpacked_layers = false
kubelet:
  config:
    shutdownGracePeriod: 30s
  flags:
  - --node-labels=abc.company.com/test-label=true`,
					ProviderOptions: map[string]string{
						domain.CredentialProviderKey:   string(domain.CredentialProviderSystemsManager),
						domain.KubernetesVersionKey:    "1.30.0",
						domain.NetworkConfigurationKey: netConfig,
						domain.NodeConfigurationKey:    nodeConfig,
						domain.HandleDependenciesKey:   "true",
					},
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cluster := tt.clusterFunc(tt)
			expected := tt.expected()

			stages.InitPaths(cluster)
			schema := NodeadmProvider(cluster)
			got, _ := kyaml.Marshal(schema)

			if string(got) != expected {
				_ = os.WriteFile(filepath.Join("testdata", tt.name, "got.yaml"), got, 0644)
				t.Errorf("Expected %s, got %s", expected, string(got))
			}
		})
	}
}
