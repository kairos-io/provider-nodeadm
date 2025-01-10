package stages

import (
	"fmt"

	"github.com/aws/eks-hybrid/api/v1alpha1"
	yip "github.com/mudler/yip/pkg/schema"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kyaml "sigs.k8s.io/yaml"

	"github.com/kairos-io/provider-nodeadm/pkg/domain"
)

const (
	nodeConfigFile     = "node-config.yaml"
	nodeConfigTemplate = "node-config.tmpl"
)

// InitYipStages returns the stages required to run 'nodeadm init'.
func InitYipStages(nc domain.NodeadmConfig, proxyArgs string, handleDependencies bool) []yip.Stage {
	return []yip.Stage{
		initConfigStage(nc),
		initStage(proxyArgs, handleDependencies),
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

func initStage(proxyArgs string, handleDependencies bool) yip.Stage {
	stage := yip.Stage{
		Name: "Run nodeadm init",
		Commands: []string{
			fmt.Sprintf("bash %s %s %s %t %s", initScript, nodeConfigPath, runtimeRoot, len(proxyArgs) > 0, proxyArgs),
		},
	}
	if handleDependencies {
		stage.If = fmt.Sprintf("[ ! -f %s/nodeadm.init ]", runtimeRoot)
		stage.Commands = append(stage.Commands, fmt.Sprintf("touch %s/nodeadm.init", runtimeRoot))
	}
	return stage
}
