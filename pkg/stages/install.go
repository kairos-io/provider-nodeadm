package stages

import (
	"fmt"

	yip "github.com/mudler/yip/pkg/schema"

	"github.com/spectrocloud-labs/provider-nodeadm/pkg/domain"
)

// InstallYipStages returns the stages required to run 'nodeadm install' and 'nodeadm upgrade'.
func InstallYipStages(nc domain.NodeadmConfig, proxyArgs string) []yip.Stage {
	return []yip.Stage{
		installStage(nc, proxyArgs),
		upgradeStage(nc, proxyArgs),
	}
}

func installStage(nc domain.NodeadmConfig, proxyArgs string) yip.Stage {
	return yip.Stage{
		Name: "Run Nodeadm Install",
		If:   "[ ! -f /opt/nodeadm/nodeadm.install ]",
		Commands: []string{
			fmt.Sprintf("bash %s %s %s %s", installScript, nc.KubernetesVersion, nc.CredentialProvider, proxyArgs),
			"touch /opt/nodeadm/nodeadm.install",
		},
	}
}

func upgradeStage(nc domain.NodeadmConfig, proxyArgs string) yip.Stage {
	return yip.Stage{
		Name: "Run Nodeadm Upgrade",
		Commands: []string{
			fmt.Sprintf("bash %s %s %s", upgradeScript, nc.KubernetesVersion, proxyArgs),
		},
	}
}
