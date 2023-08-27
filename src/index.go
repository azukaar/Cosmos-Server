package main

import (
	"math/rand"
	"time"
	"context"

	"github.com/azukaar/cosmos-server/src/docker"
	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/azukaar/cosmos-server/src/authorizationserver"
	"github.com/azukaar/cosmos-server/src/market"
	"github.com/azukaar/cosmos-server/src/constellation"
)

func main() {
	utils.Log("Starting...")
	
	utils.ReBootstrapContainer = docker.BootstrapContainerFromTags

	rand.Seed(time.Now().UnixNano())

	LoadConfig()

	go CRON()

	docker.Test()

	docker.DockerListenEvents()

	docker.BootstrapAllContainersFromTags()

	docker.RemoveSelfUpdater()

	go func() {
		time.Sleep(180 * time.Second)
		docker.CheckUpdatesAvailable()
	}()

	version, err := docker.DockerClient.ServerVersion(context.Background())
	if err == nil {
		utils.Log("Docker API version: " + version.APIVersion)
	}

	market.Init()
	
	authorizationserver.Init()

	constellation.InitDNS()
	
	constellation.Init()

	StartServer()
}
