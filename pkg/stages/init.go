package stages

import (
	"fmt"

	yip "github.com/mudler/yip/pkg/schema"
	"github.com/sirupsen/logrus"

	"github.com/spectrocloud/provider-nodeadm/pkg/domain"
	"github.com/spectrocloud/provider-nodeadm/pkg/embed"
)

const (
	nodeConfigFile     = "node-config.yaml"
	nodeConfigTemplate = "node-config.tmpl"
)

// InitYipStages returns the stages required to run 'nodeadm init'.
func InitYipStages(nc domain.NodeadmConfig, proxyArgs string) []yip.Stage {
	return []yip.Stage{
		initConfigStage(nc),
		initStage(proxyArgs),
	}
}

func initConfigStage(nc domain.NodeadmConfig) yip.Stage {
	bs, err := embed.EFS.RenderTemplateBytes(nc.NodeConfiguration, "", nodeConfigTemplate)
	if err != nil {
		logrus.Fatal(err)
	}

	initConfigStage := yip.Stage{
		Name: "Generate Nodeadm Init Config File",
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

func initStage(proxyArgs string) yip.Stage {
	return yip.Stage{
		Name: "Run Nodeadm Init",
		If:   fmt.Sprintf("[ ! -f %s/nodeadm.init ]", runtimeRoot),
		Commands: []string{
			fmt.Sprintf("bash %s %s %s %t %s", initScript, nodeConfigPath, runtimeRoot, len(proxyArgs) > 0, proxyArgs),
			fmt.Sprintf("touch %s/nodeadm.init", runtimeRoot),
		},
	}
}
