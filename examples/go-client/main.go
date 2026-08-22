package main

import (
	"context"
	"fmt"
	"log"

	"github.com/vikukumar/tarak/pkg/client"
)

func main() {
	fmt.Println("⚡ Tarak Go Client SDK Example")

	// 1. Initialize client from local kubeconfig (~/.tarak/config)
	c, err := client.NewClientFromKubeconfig("")
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	fmt.Printf("Connected to Tarak Server: %s\n\n", c.ServerURL())

	// 2. List all namespaces
	namespaces, err := c.Namespaces().List(context.Background())
	if err != nil {
		log.Fatalf("Failed to list namespaces: %v", err)
	}
	fmt.Printf("📦 Namespaces (%d):\n", len(namespaces))
	for _, ns := range namespaces {
		fmt.Printf("  - %s\n", ns)
	}

	// 3. List pods in default namespace
	pods, err := c.Pods("default").List(context.Background())
	if err != nil {
		log.Fatalf("Failed to list pods: %v", err)
	}
	fmt.Printf("\n🚀 Pods in 'default' (%d):\n", len(pods))
	for _, p := range pods {
		fmt.Printf("  - %s | Phase: %s | Node: %s | IP: %s\n", p.Name, p.Phase, p.Node, p.IP)
	}
}
