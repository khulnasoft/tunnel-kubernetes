package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/khulnasoft/tunnel-kubernetes/pkg/artifacts"
	"github.com/khulnasoft/tunnel-kubernetes/pkg/k8s"
	tk "github.com/khulnasoft/tunnel-kubernetes/pkg/tunnelk8s"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"

	"context"
)

func main() {
	ctx := context.Background()

	cluster, err := k8s.GetCluster(k8s.WithBurst(100), k8s.WithQPS(100))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Current namespace:", cluster.GetCurrentNamespace())

	tunnelk8s := tk.New(cluster, tk.WithExcludeOwned(true))
	fmt.Println("Scanning cluster")

	//tunnel k8s #cluster
	start := time.Now()
	artifacts, err := tunnelk8s.ListArtifacts(ctx)
	end := time.Now()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Scan took %v\n", end.Sub(start))
	printArtifacts(artifacts)

	fmt.Println("Scanning kind 'pods' with exclude-owned=true")
	artifacts, err = tunnelk8s.Resources("pod").AllNamespaces().ListArtifacts(ctx)
	if err != nil {
		log.Fatal(err)
	}
	printArtifacts(artifacts)

	fmt.Println("Scanning namespace 'default'")
	//tunnel k8s --namespace default
	artifacts, err = tunnelk8s.Namespace("default").ListArtifacts(ctx)
	if err != nil {
		log.Fatal(err)
	}
	printArtifacts(artifacts)
	fmt.Println("Scanning all namespaces ")
	artifacts, err = tunnelk8s.AllNamespaces().ListArtifacts(ctx)
	if err != nil {
		log.Fatal(err)
	}
	printArtifacts(artifacts)

	fmt.Println("Scanning 'deployments'")

	//tunnel k8s deployment
	artifacts, err = tunnelk8s.Namespace("default").Resources("deployment").ListArtifacts(ctx)
	if err != nil {
		log.Fatal(err)
	}
	printArtifacts(artifacts)

	fmt.Println("Scanning 'cm,pods'")
	//tunnel k8s clusterroles,pods
	artifacts, err = tunnelk8s.Namespace("default").Resources("cm,pods").ListArtifacts(ctx)
	if err != nil {
		log.Fatal(err)
	}
	printArtifacts(artifacts)

	tolerations := []corev1.Toleration{
		{
			Effect:   corev1.TaintEffectNoSchedule,
			Operator: corev1.TolerationOperator(corev1.NodeSelectorOpExists),
		},
		{
			Effect:   corev1.TaintEffectNoExecute,
			Operator: corev1.TolerationOperator(corev1.NodeSelectorOpExists),
		},
		{
			Effect:            corev1.TaintEffectNoExecute,
			Key:               "node.kubernetes.io/not-ready",
			Operator:          corev1.TolerationOperator(corev1.NodeSelectorOpExists),
			TolerationSeconds: ptr.To[int64](300),
		},
		{
			Effect:            corev1.TaintEffectNoExecute,
			Key:               "node.kubernetes.io/unreachable",
			Operator:          corev1.TolerationOperator(corev1.NodeSelectorOpExists),
			TolerationSeconds: ptr.To[int64](300),
		},
	}

	// collect node info
	ar, err := tunnelk8s.ListArtifactAndNodeInfo(ctx, []tk.NodeCollectorOption{
		tk.WithScanJobNamespace("tunnel-temp"),
		tk.WithIgnoreLabels(map[string]string{"chen": "test"}),
		tk.WithTolerations(tolerations),
	}...)
	if err != nil {
		log.Fatal(err)
	}
	for _, a := range ar {
		if a.Kind != "NodeInfo" {
			continue
		}
		fmt.Println(a.RawResource)
	}

	bi, err := tunnelk8s.ListClusterBomInfo(ctx)
	if err != nil {
		log.Fatal(err)
	}
	bb, err := json.Marshal(bi)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(string(bb))
}

func printArtifacts(artifacts []*artifacts.Artifact) {
	for _, artifact := range artifacts {
		printArtifact(artifact)
	}
}

func printArtifact(artifact *artifacts.Artifact) {
	fmt.Printf(
		"Name: %s, Kind: %s, Namespace: %s, Images: %v\n",
		artifact.Name,
		artifact.Kind,
		artifact.Namespace,
		artifact.Images,
	)
}
