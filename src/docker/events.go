package docker

import (
	"context"
	
	"github.com/azukaar/cosmos-server/src/utils" 

	"github.com/docker/docker/api/types"
)

func DockerListenEvents() error {
	errD := connect()
	if errD != nil {
		return errD
	}
	
	go func() {
		msgs, errs := DockerClient.Events(context.Background(), types.EventsOptions{})

		for {
			select {
				case err := <-errs:
					utils.Error("Docker Events", err)
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