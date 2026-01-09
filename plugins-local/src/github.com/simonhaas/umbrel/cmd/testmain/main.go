package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/simonhaas/umbrel"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Optionally set DOCKER_SOCKET env var to point to the docker socket path
	// e.g. export DOCKER_SOCKET=/var/run/docker.sock
	unixPath := os.Getenv("DOCKER_SOCKET")

	infos, err := umbrel.GetContainersInfo(ctx, unixPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "GetContainersInfo error: %v\n", err)
		os.Exit(1)
	}

	// only include containers attached to this network with names ending in _app_proxy_1
	networkFilter := "umbrel_main_network"
	nameSuffix := "_app_proxy_1"
	for _, ci := range infos {
		if ci.Networks == nil {
			continue
		}
		if _, ok := ci.Networks[networkFilter]; !ok {
			continue
		}

		match := false
		for _, n := range ci.Names {
			if strings.HasSuffix(n, nameSuffix) {
				match = true
				break
			}
		}
		if !match {
			continue
		}

		fmt.Printf("ID: %s\n", ci.ID)
		fmt.Printf("Names: %v\n", ci.Names)
		fmt.Printf("Networks: %v\n", ci.Networks)
		fmt.Printf("PublishedPorts: %v\n", ci.PublishedPorts)
		fmt.Println("-----")
	}
}
