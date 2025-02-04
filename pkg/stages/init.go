package stages

import (
	"fmt"

	yip "github.com/mudler/yip/pkg/schema"
)

// InitBootBeforeStages returns the boot.before stages required to run 'nodeadm init'.
func InitBootBeforeStages(proxyArgs string, handleDependencies bool) []yip.Stage {
	return []yip.Stage{
		initStage(proxyArgs, handleDependencies),
	}
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

// InitFSAfterStages returns the fs.after stages required to run 'nodeadm init'.
// These stages are only expected to run in appliance mode.
func InitFSAfterStages() []yip.Stage {
	return []yip.Stage{
		{
			If:   "[ -f /usr/bin/aws-iam-authenticator ] && [ ! -f /usr/local/bin/aws-iam-authenticator ]",
			Name: "Symlink aws-iam-authenticator",
			Commands: []string{
				"ln -s /usr/bin/aws-iam-authenticator /usr/local/bin/aws-iam-authenticator",
			},
		},
		{
			If:   "[ -f /usr/bin/aws_signing_helper ] && [ ! -f /usr/local/bin/aws_signing_helper ]",
			Name: "Symlink aws_signing_helper",
			Commands: []string{
				"ln -s /usr/bin/aws_signing_helper /usr/local/bin/aws_signing_helper",
			},
		},
	}
}
