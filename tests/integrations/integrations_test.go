package e2e

import (
	"context"
	"encoding/json"
	tk "github.com/khulnasoft/tunnel-kubernetes/pkg/tunnelk8s"
	"github.com/stretchr/testify/assert"
	"os"
	"reflect"
	"testing"

	"github.com/khulnasoft/tunnel-kubernetes/pkg/k8s"
)

func TestNodeInfo(T *testing.T) {
	ctx := context.Background()
	if testing.Short() {
		T.Skip("skipping end-to-end test")
	}
	cluster, err := k8s.GetCluster(k8s.WithBurst(100))
	if err != nil {
		panic(err)
	}
	tunnelk8s := tk.New(cluster, tk.WithExcludeOwned(true))
	// collect node info
	ar, err := tunnelk8s.ListArtifactAndNodeInfo(ctx, []tk.NodeCollectorOption{
		tk.WithScanJobNamespace("tunnel-temp"),
		tk.WithCommandPaths([]string{"./testdata"}),
		tk.WithScanJobImageRef("ghcr.io/khulnasoft/node-collector:0.3.0"),
	}...)
	assert.NoError(T, err)
	for _, a := range ar {
		if a.Kind != "NodeInfo" {
			continue
		}
		var expectedNodeIbfo map[string]interface{}
		b, err := os.ReadFile("./testdata/expected_node_info.json")
		assert.NoError(T, err)
		err = json.Unmarshal(b, &expectedNodeIbfo)
		assert.NoError(T, err)
		assert.True(T, reflect.DeepEqual(expectedNodeIbfo["info"], a.RawResource["info"]))
	}
}

func TestKBOM(T *testing.T) {
	ctx := context.Background()
	if testing.Short() {
		T.Skip("skipping end-to-end test")
	}
	cluster, err := k8s.GetCluster(k8s.WithBurst(100))
	if err != nil {
		panic(err)
	}
	tunnelk8s := tk.New(cluster, tk.WithExcludeOwned(true))
	// collect bom info
	gotBom, err := tunnelk8s.ListClusterBomInfo(ctx)
	assert.NoError(T, err)
	assert.True(T, len(gotBom) == 10)
}
