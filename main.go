// Package main is the entrypoint for provider-nodeadm.
package main

import (
	"os"

	"github.com/kairos-io/kairos-sdk/clusterplugin"
	"github.com/sirupsen/logrus"

	"github.com/spectrocloud-labs/provider-nodeadm/pkg/provider"
)

func main() {
	f, err := os.OpenFile("/var/log/provider-nodeadm.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	logrus.SetOutput(f)

	plugin := clusterplugin.ClusterPlugin{
		Provider: provider.NodeadmProvider,
	}

	if err := plugin.Run(); err != nil {
		logrus.Fatal(err)
	}
}
