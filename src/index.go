package main

import (
	"github.com/azukaar/cosmos-server/src/utils"
	"time"
	"github.com/azukaar/cosmos-server/src/docker"
	"math/rand"
)

func main() {
	  utils.Log("Starting...")

		rand.Seed(time.Now().UnixNano())

		LoadConfig()
		
		docker.Test()

		docker.DockerListenEvents()

		docker.BootstrapAllContainersFromTags()
		
		StartServer()
}