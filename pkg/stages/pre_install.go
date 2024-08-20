// Package stages contains helpers to generate yip stages for nodeadm.
package stages

import (
	"fmt"
	"path/filepath"
	"strings"

	yip "github.com/mudler/yip/pkg/schema"

	"github.com/spectrocloud-labs/provider-nodeadm/pkg/domain"
)

const (
	envPrefix        = "Environment="
	helperScriptPath = "/opt/nodeadm/scripts"
	k8sNoProxy       = ".svc,.svc.cluster,.svc.cluster.local"
)

var (
	initScript    = filepath.Join(helperScriptPath, "nodeadm-init.sh")
	installScript = filepath.Join(helperScriptPath, "nodeadm-install.sh")
	upgradeScript = filepath.Join(helperScriptPath, "nodeadm-upgrade.sh")
)

// PreInstallYipStages returns the setup stages required prior to running 'nodeadm install'.
func PreInstallYipStages(env map[string]string, nc domain.NodeadmConfig) []yip.Stage {
	return []yip.Stage{
		proxyStage(nc, env),
		commandsStage(),
		storeVersionStage(nc.KubernetesVersion),
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

func storeVersionStage(version string) yip.Stage {
	return yip.Stage{
		If:   "[ ! -f /opt/nodeadm/sentinel_kubernetes_version ]",
		Name: "Create kubernetes version sentinel file",
		Commands: []string{
			fmt.Sprintf("echo %s > /opt/nodeadm/sentinel_kubernetes_version", version),
		},
	}
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
