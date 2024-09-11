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
	"github.com/azukaar/cosmos-server/src/metrics"
	"github.com/azukaar/cosmos-server/src/storage"
	"github.com/azukaar/cosmos-server/src/cron"
)

func main() {
	utils.Log("------------------------------------------")
	utils.Log("Starting Cosmos-Server version " + GetCosmosVersion())
	utils.Log("------------------------------------------")
	
	// utils.ReBootstrapContainer = docker.BootstrapContainerFromTags
	utils.PushShieldMetrics = metrics.PushShieldMetrics
	utils.GetContainerIPByName = docker.GetContainerIPByName
	utils.DoesContainerExist = docker.DoesContainerExist
	utils.CheckDockerNetworkMode = docker.CheckDockerNetworkMode

	rand.Seed(time.Now().UnixNano())

	LoadConfig()

	utils.CheckHostNetwork()
	
	go CRON()

	docker.ExportDocker()

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

	config := utils.GetMainConfig()
	
	if !config.NewInstall {
		MigratePre013()
		MigratePre014()

		docker.isInsideContainer()

		docker.CheckPuppetDB()

		utils.InitDBBuffers()

		utils.Log("Starting monitoring services...")

		metrics.Init()

		utils.Log("Starting market services...")

		market.Init()
		
		utils.Log("Starting OpenID services...")

		authorizationserver.Init()

		utils.Log("Starting constellation services...")

		utils.InitFBL()

		constellation.Init()

		storage.InitSnapRAIDConfig()
		
		// Has to be done last, so scheduler does not re-init
		cron.Init()

		utils.Log("Starting server...")
	}

	StartServer()
}
