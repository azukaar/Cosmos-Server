package docker

import (
	"context"
	"sync"
	"time"

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
					// Check connection
					errD := Connect()
					if errD != nil {
						utils.Fatal("Docker connection died, couldn't recover... Restarting", errD)
					}
					msgs, errs = DockerClient.Events(context.Background(), types.EventsOptions{})

				case msg := <-msgs:
					utils.Debug("Docker Event: " + msg.Type + " " + msg.Action + " " + msg.Actor.ID)
					if msg.Type == "container" && msg.Action == "start" {
						onDockerStarted(msg.Actor.ID)
					}

					// on container destroy and network disconnect
					if msg.Type == "container" && msg.Action == "destroy" {
						onDockerDestroyed(msg.Actor.ID)
					}
					if msg.Type == "container" && msg.Action == "create" {
						onDockerCreated(msg.Actor.ID)
					}
					if msg.Type == "network" && msg.Action == "disconnect" {
						onNetworkDisconnect(msg.Actor.ID)
					}
					if msg.Type == "network" && msg.Action == "destroy" {
						onNetworkDestroy(msg.Actor.ID)
					}
					if msg.Type == "network" && msg.Action == "create" {
						onNetworkCreate(msg.Actor.ID)
					}
					if msg.Type == "network" && msg.Action == "connect" {
						onNetworkConnect(msg.Actor.ID)
					}
			}
		}
	}()

	return nil
}

var (
	timer    *time.Timer
	interval = 30000 * time.Millisecond
	mu       sync.Mutex
	cancelFunc      context.CancelFunc
)

func DebouncedExportDocker() {
	mu.Lock()
	defer mu.Unlock()

	if timer != nil {
		if cancelFunc != nil {
			cancelFunc() // cancel the previous context
		}
		timer.Stop()
	}

	// Create a new context and cancelFunc
	ctx, newCancelFunc := context.WithCancel(context.Background())
	cancelFunc = newCancelFunc

	timer = time.AfterFunc(interval, func() {
		select {
		case <-ctx.Done():
			// if the context was canceled, don't execute ExportDocker
			return
		default:
			ExportDocker()
		}
	})
}

func onDockerStarted(containerID string) {
	utils.Debug("onDockerStarted: " + containerID)
	BootstrapContainerFromTags(containerID)
	// DebouncedExportDocker()
}

func onDockerDestroyed(containerID string) {
	utils.Debug("onDockerDestroyed: " + containerID)
	// DebouncedExportDocker()
}

func onNetworkDisconnect(networkID string) {
	utils.Debug("onNetworkDisconnect: " + networkID)
	DebouncedNetworkCleanUp(networkID)
	// DebouncedExportDocker()
}

func onDockerCreated(containerID string) {
	utils.Debug("onDockerCreated: " + containerID)
	// DebouncedExportDocker()
}

func onNetworkDestroy(networkID string) {
	utils.Debug("onNetworkDestroy: " + networkID)
	// DebouncedExportDocker()
}

func onNetworkCreate(networkID string) {
	utils.Debug("onNetworkCreate: " + networkID)
	// DebouncedExportDocker()
}

func onNetworkConnect(networkID string) {
	utils.Debug("onNetworkConnect: " + networkID)
	// DebouncedExportDocker()
}
