package main

import (
	"math/rand"
	"time"

	"github.com/azukaar/cosmos-server/src/docker"
	"github.com/azukaar/cosmos-server/src/utils"
)

func main() {
	utils.Log("Starting...")
	// utils.Log("Smart Shield estimates the capacity at " + strconv.Itoa((int)(proxy.MaxUsers)) + " concurrent users")

	rand.Seed(time.Now().UnixNano())

	LoadConfig()

	checkVersion()

	go CRON()

	docker.Test()

	docker.DockerListenEvents()

	docker.BootstrapAllContainersFromTags()

	StartServer()
}
