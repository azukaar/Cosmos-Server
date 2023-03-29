package docker

import (
	"context"
	
	"github.com/azukaar/cosmos-server/src/utils" 

	"github.com/docker/docker/api/types"
)

func DockerListenEvents() error {
	errD := Connect()
	if errD != nil {
		utils.Error("Docker did not connect. Not listening", errD)
		return errD
	}
	
	go func() {
		msgs, errs := DockerClient.Events(context.Background(), types.EventsOptions{})

		for {
			select {
				case err := <-errs:
					if err == nil {
						return
					}
					utils.Error("Docker Event Error", err)
				case msg := <-msgs:
					utils.Debug("Docker Event: " + msg.Type + " " + msg.Action + " " + msg.Actor.ID)
					if msg.Type == "container" && msg.Action == "start" {
						onDockerCreated(msg.Actor.ID)
					}
			}
		}
	}()

	return nil
}

func onDockerCreated(containerID string) {
	utils.Debug("onDockerCreated: " + containerID)
	BootstrapContainerFromTags(containerID)
}