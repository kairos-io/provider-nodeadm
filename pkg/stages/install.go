package stages

import (
	"fmt"

	yip "github.com/mudler/yip/pkg/schema"

	"github.com/kairos-io/provider-nodeadm/pkg/domain"
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
		If:   fmt.Sprintf("[ ! -f %s/nodeadm.install ]", runtimeRoot),
		Commands: []string{
			fmt.Sprintf(
				"bash %s %s %s %s %t %s",
				installScript, nc.KubernetesVersion, nc.CredentialProvider, runtimeRoot, len(proxyArgs) > 0, proxyArgs,
			),
			fmt.Sprintf("touch %s/nodeadm.install", runtimeRoot),
		},
	}
}

func upgradeStage(nc domain.NodeadmConfig, proxyArgs string) yip.Stage {
	return yip.Stage{
		Name: "Run Nodeadm Upgrade",
		Commands: []string{
			fmt.Sprintf(
				"bash %s %s %s %s %t %s",
				upgradeScript, nc.KubernetesVersion, nodeConfigPath, runtimeRoot, len(proxyArgs) > 0, proxyArgs,
			),
		},
	}
}
