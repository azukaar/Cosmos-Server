package main

import (
	"math/rand"
	"time"
	"context"

	"github.com/azukaar/cosmos-server/src/docker"
	"github.com/azukaar/cosmos-server/src/utils"
)

func main() {
	utils.Log("Starting...")

	rand.Seed(time.Now().UnixNano())

	LoadConfig()

	go CRON()

	docker.Test()

	docker.DockerListenEvents()

	docker.BootstrapAllContainersFromTags()

	docker.RemoveSelfUpdater()

	go func() {
		time.Sleep(30 * time.Second)
		docker.CheckUpdatesAvailable()
	}()

	version, err := docker.DockerClient.ServerVersion(context.Background())
	if err == nil {
		utils.Log("Docker API version: " + version.APIVersion)
	}
	
	StartServer()
}
