// Package stages contains helpers to generate yip stages for nodeadm.
package stages

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/aws/eks-hybrid/api/v1alpha1"
	"github.com/kairos-io/kairos-sdk/clusterplugin"
	yip "github.com/mudler/yip/pkg/schema"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kyaml "sigs.k8s.io/yaml"

	"github.com/kairos-io/provider-nodeadm/pkg/domain"
)

const (
	envPrefix      = "Environment="
	k8sNoProxy     = ".svc,.svc.cluster,.svc.cluster.local"
	nodeConfigFile = "node-config.yaml"
	nodeadmRoot    = "/opt/nodeadmutil"
)

var (
	initScript, installScript, resetScript, upgradeScript string
	nodeConfigPath                                        string
	runtimeRoot                                           string
)

// InitPaths initializes all nodeadm paths relative to the cluster's root path.
func InitPaths(cluster clusterplugin.Cluster) {
	clusterRootPath := clusterRootPath(cluster)

	runtimeRoot = filepath.Join(clusterRootPath, nodeadmRoot)

	nodeConfigPath = filepath.Join(runtimeRoot, nodeConfigFile)

	initScript = filepath.Join(clusterRootPath, nodeadmRoot, "scripts", "nodeadm-init.sh")
	installScript = filepath.Join(clusterRootPath, nodeadmRoot, "scripts", "nodeadm-install.sh")
	resetScript = filepath.Join(clusterRootPath, nodeadmRoot, "scripts", "nodeadm-reset.sh")
	upgradeScript = filepath.Join(clusterRootPath, nodeadmRoot, "scripts", "nodeadm-upgrade.sh")
}

func clusterRootPath(cluster clusterplugin.Cluster) string {
	rootpath := cluster.ProviderOptions[domain.ClusterRootPathKey]
	if rootpath == "" {
		return "/"
	}
	return rootpath
}

// PreInstallBootBeforeStages returns the setup stages required during boot.before, prior to running 'nodeadm install'.
func PreInstallBootBeforeStages(env map[string]string, nc domain.NodeadmConfig) []yip.Stage {
	return []yip.Stage{
		proxyStage(nc, env),
		commandsStage(),
		initConfigStage(nc),
	}
}

func commandsStage() yip.Stage {
	return yip.Stage{
		Name: "Run Pre-installation Commands",
		Commands: []string{
			"mkdir -p /etc/iam/pki",
		},
	}
}

func initConfigStage(nc domain.NodeadmConfig) yip.Stage {
	bs, err := toHybridConfig(nc)
	if err != nil {
		logrus.Fatal(err)
	}

	initConfigStage := yip.Stage{
		Name: "Generate nodeadm config file",
		Files: []yip.File{
			{
				Path:        nodeConfigPath,
				Permissions: 0640,
				Content:     string(bs),
			},
		},
	}

	if nc.NodeConfiguration.IAMRolesAnywhere != nil {
		initConfigStage.Files = append(initConfigStage.Files, []yip.File{
			{
				Path:        "/etc/iam/pki/server.pem",
				Permissions: 0600,
				Content:     nc.NodeConfiguration.IAMRolesAnywhere.Certificate,
			},
			{
				Path:        "/etc/iam/pki/server.key",
				Permissions: 0400,
				Content:     nc.NodeConfiguration.IAMRolesAnywhere.PrivateKey,
			},
		}...)
	}

	return initConfigStage
}

func toHybridConfig(nc domain.NodeadmConfig) ([]byte, error) {
	nodeConfig := v1alpha1.NodeConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "node.eks.aws/v1alpha1",
			Kind:       "NodeConfig",
		},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: v1alpha1.NodeConfigSpec{
			Cluster: v1alpha1.ClusterDetails{
				Name:   nc.NodeConfiguration.ClusterName,
				Region: nc.NodeConfiguration.Region,
			},
			Hybrid: &v1alpha1.HybridOptions{
				EnableCredentialsFile: false,
			},
		},
	}
	if nc.NodeConfiguration.UserConfig != nil {
		nodeConfig.Spec.Containerd = nc.NodeConfiguration.UserConfig.Containerd
		nodeConfig.Spec.Kubelet = nc.NodeConfiguration.UserConfig.Kubelet
	}
	if nc.NodeConfiguration.IAMRolesAnywhere != nil && nc.NodeConfiguration.IAMRolesAnywhere.RoleARN != "" {
		nodeConfig.Spec.Hybrid.IAMRolesAnywhere = &v1alpha1.IAMRolesAnywhere{
			NodeName:       nc.NodeConfiguration.IAMRolesAnywhere.NodeName,
			TrustAnchorARN: nc.NodeConfiguration.IAMRolesAnywhere.TrustAnchorARN,
			ProfileARN:     nc.NodeConfiguration.IAMRolesAnywhere.ProfileARN,
			RoleARN:        nc.NodeConfiguration.IAMRolesAnywhere.RoleARN,
		}
	}
	if nc.NodeConfiguration.SSM != nil && nc.NodeConfiguration.SSM.ActivationCode != "" {
		nodeConfig.Spec.Hybrid.SSM = &v1alpha1.SSM{
			ActivationCode: nc.NodeConfiguration.SSM.ActivationCode,
			ActivationID:   nc.NodeConfiguration.SSM.ActivationID,
		}
	}
	return kyaml.Marshal(nodeConfig)
}

func proxyStage(nc domain.NodeadmConfig, env map[string]string) yip.Stage {
	daemonProxyEnv := daemonProxyEnv(env, nc.NetworkConfiguration)
	return yip.Stage{
		Name: "Set proxy env",
		Files: []yip.File{
			{
				Path:        filepath.Join("/etc/systemd/system/containerd.service.d", "http-proxy.conf"),
				Permissions: 0400,
				Content:     daemonProxyEnv,
			},
			{
				Path:        filepath.Join("/etc/systemd/system/kubelet.service.d", "http-proxy.conf"),
				Permissions: 0400,
				Content:     daemonProxyEnv,
			},
			{
				Path:        filepath.Join("/etc/apt", "apt.conf"),
				Permissions: 0400,
				Content:     aptProxyEnv(env),
			},
		},
	}
}

func daemonProxyEnv(env map[string]string, nc domain.NetworkConfig) string {
	var proxy []string

	httpProxy := env["HTTP_PROXY"]
	httpsProxy := env["HTTPS_PROXY"]
	userNoProxy := env["NO_PROXY"]

	if IsProxyConfigured(env) {
		proxy = append(proxy, "[Service]")

		if len(httpProxy) > 0 {
			proxy = append(proxy, fmt.Sprintf("%s\"HTTP_PROXY=%s\"", envPrefix, httpProxy))
		}
		if len(httpsProxy) > 0 {
			proxy = append(proxy, fmt.Sprintf("%s\"HTTPS_PROXY=%s\"", envPrefix, httpsProxy))
		}

		noProxy := defaultNoProxy(nc)
		if len(userNoProxy) > 0 {
			noProxy = noProxy + "," + userNoProxy
		}

		proxy = append(proxy, fmt.Sprintf("%s\"NO_PROXY=%s\"", envPrefix, noProxy))
	}

	return strings.Join(proxy, "\n")
}

func aptProxyEnv(env map[string]string) string {
	var proxy []string

	httpProxy := env["HTTP_PROXY"]
	httpsProxy := env["HTTPS_PROXY"]

	if IsProxyConfigured(env) {
		if len(httpProxy) > 0 {
			proxy = append(proxy, fmt.Sprintf("Acquire::http::Proxy \"%s\";", httpProxy))
		}
		if len(httpsProxy) > 0 {
			proxy = append(proxy, fmt.Sprintf("Acquire::https::Proxy \"%s\";", httpsProxy))
		}
	}

	return strings.Join(proxy, "\n")
}

// GetNoProxyConfig derives the NO_PROXY environment variable value from the cluster's
// network configuration and environment variables.
func GetNoProxyConfig(env map[string]string, nc domain.NodeadmConfig) string {
	defaultNoProxy := defaultNoProxy(nc.NetworkConfiguration)
	userNoProxy := env["NO_PROXY"]
	if len(userNoProxy) > 0 {
		return defaultNoProxy + "," + userNoProxy
	}
	return defaultNoProxy
}

// IsProxyConfigured checks if the HTTP_PROXY or HTTPS_PROXY environment variables are set.
func IsProxyConfigured(env map[string]string) bool {
	return len(env["HTTP_PROXY"]) > 0 || len(env["HTTPS_PROXY"]) > 0
}

func defaultNoProxy(nc domain.NetworkConfig) string {
	var noProxy string

	if len(nc.PodCIDR) > 0 {
		noProxy = nc.PodCIDR
	}
	if len(nc.ServiceCIDR) > 0 {
		noProxy = noProxy + "," + nc.ServiceCIDR
	}

	return noProxy + "," + k8sNoProxy
}
