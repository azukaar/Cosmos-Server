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
	
	"go.uber.org/zap"
	"os"
	"log"
)

func main() {
	log.SetOutput(os.Stderr)
	logger := zap.NewExample()
	defer logger.Sync()

	utils.Log("Starting...")
	
	// utils.ReBootstrapContainer = docker.BootstrapContainerFromTags
	utils.PushShieldMetrics = metrics.PushShieldMetrics
	utils.GetContainerIPByName = docker.GetContainerIPByName
	utils.CheckDockerNetworkMode = docker.CheckDockerNetworkMode

	rand.Seed(time.Now().UnixNano())

	LoadConfig()
	
	utils.CheckHostNetwork()

	config := utils.GetMainConfig()
	if !config.NewInstall {
		MigratePre014()
	}

	utils.InitDBBuffers()
	
	go CRON()

	docker.ExportDocker()

	docker.DockerListenEvents()

	// docker.BootstrapAllContainersFromTags()

	docker.RemoveSelfUpdater()

	go func() {
		time.Sleep(180 * time.Second)
		docker.CheckUpdatesAvailable()
	}()

	version, err := docker.DockerClient.ServerVersion(context.Background())
	if err == nil {
		utils.Log("Docker API version: " + version.APIVersion)
	}

	config = utils.GetMainConfig()
	if !config.NewInstall {
		MigratePre013()

		utils.Log("Starting monitoring services...")

		metrics.Init()

		utils.Log("Starting market services...")

		market.Init()
		
		utils.Log("Starting OpenID services...")

		authorizationserver.Init()

		utils.Log("Starting constellation services...")

		constellation.InitDNS()
		
		constellation.Init()

		utils.Log("Starting server...")

	}

	StartServer()
}
