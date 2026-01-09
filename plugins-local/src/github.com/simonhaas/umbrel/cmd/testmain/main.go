package main

import (
	"context"
	"fmt"
	"os"
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

	for _, ci := range infos {
		fmt.Printf("ID: %s\n", ci.ID)
		fmt.Printf("Names: %v\n", ci.Names)
		fmt.Printf("Networks: %v\n", ci.Networks)
		fmt.Printf("PublishedPorts: %v\n", ci.PublishedPorts)
		fmt.Println("-----")
	}
}
