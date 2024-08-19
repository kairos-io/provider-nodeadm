package provider

import (
	"os"
	"testing"

	"github.com/kairos-io/kairos-sdk/clusterplugin"
	"sigs.k8s.io/yaml"
)

func TestNodeadmProvider(t *testing.T) {
	opts, _ := os.ReadFile("testdata/options.yaml")
	expected, _ := os.ReadFile("testdata/expected.yaml")

	cluster := clusterplugin.Cluster{
		Options: string(opts),
	}
	schema := NodeadmProvider(cluster)
	got, _ := yaml.Marshal(schema)

	if string(got) != string(expected) {
		_ = os.WriteFile("testdata/got.yaml", got, 0644)
		t.Errorf("Expected %s, got %s", string(expected), string(got))
	}
}
