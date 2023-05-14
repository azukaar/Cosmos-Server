package docker

import (
	"context"
	"errors"
	"time"
	"fmt"
	"github.com/azukaar/cosmos-server/src/utils" 

	"github.com/docker/docker/client"
	// natting "github.com/docker/go-connections/nat"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types"
)

var DockerClient *client.Client
var DockerContext context.Context
var DockerNetworkName = "cosmos-network"

func getIdFromName(name string) (string, error) {
	containers, err := DockerClient.ContainerList(DockerContext, types.ContainerListOptions{})
	if err != nil {
		utils.Error("Docker Container List", err)
		return "", err
	}

	for _, container := range containers {
		if container.Names[0] == name {
			utils.Warn(container.Names[0] + " == " + name + " == " + container.ID)
			return container.ID, nil
		}
	}

	return "", errors.New("Container not found")
}

var DockerIsConnected = false

func Connect() error {
	if DockerClient != nil {
		// check if connection is still alive
		ping, err := DockerClient.Ping(DockerContext)
		if ping.APIVersion != "" && err == nil {
			DockerIsConnected = true
			return nil
		} else {
			DockerIsConnected = false
			DockerClient = nil
			utils.Error("Docker Connection died, will try to connect again", err)
		}
	}
	if DockerClient == nil {
		ctx := context.Background()
		client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			DockerIsConnected = false
			return err
		}
		defer client.Close()

		DockerClient = client
		DockerContext = ctx

		ping, err := DockerClient.Ping(DockerContext)
		if ping.APIVersion != "" && err == nil {
			DockerIsConnected = true
			utils.Log("Docker Connected")
		} else {
			DockerIsConnected = false
			utils.Error("Docker Connection - Cannot ping Daemon. Is it running?", nil)
			return errors.New("Docker Connection - Cannot ping Daemon. Is it running?")
		}
		
		// if running in Docker, connect to main network
		// if os.Getenv("HOSTNAME") != "" {
		// 	ConnectToNetwork(os.Getenv("HOSTNAME"))
		// }
	}

	return nil
}

func EditContainer(oldContainerID string, newConfig types.ContainerJSON) (string, error) {
	
	utils.Debug("VOLUMES:" + fmt.Sprintf("%v", newConfig.HostConfig.Mounts))
	
	if(oldContainerID != "") {
		// no need to re-lock if we are reverting
		DockerNetworkLock <- true
		defer func() { 
			<-DockerNetworkLock 
			utils.Debug("Unlocking EDIT Container")
		}()

		errD := Connect()
		if errD != nil {
			return "", errD
		}
	}
	
	newName := newConfig.Name
	oldContainer := newConfig

	if(oldContainerID != "") {
		utils.Log("EditContainer - Container updating. Retriveing currently running " + oldContainerID)

		var err error

		// get container informations
		// https://godoc.org/github.com/docker/docker/api/types#ContainerJSON
		oldContainer, err = DockerClient.ContainerInspect(DockerContext, oldContainerID)

		utils.Debug("OLD VOLUMES:" + fmt.Sprintf("%v", oldContainer.HostConfig.Mounts))

		if err != nil {
			return "", err
		}

		// if no name, use the same one, that will force Docker to create a hostname if not set
		newName = oldContainer.Name
		newConfig.Config.Hostname = newName

		// stop and remove container
		stopError := DockerClient.ContainerStop(DockerContext, oldContainerID, container.StopOptions{})
		if stopError != nil {
			return "", stopError
		}

		removeError := DockerClient.ContainerRemove(DockerContext, oldContainerID, types.ContainerRemoveOptions{})
		if removeError != nil {
			return "", removeError
		}

		// wait for container to be destroyed
		//
		for {
			_, err := DockerClient.ContainerInspect(DockerContext, oldContainerID)
			if err != nil {
				break
			} else {
				utils.Log("EditContainer - Waiting for container to be destroyed")
				time.Sleep(1 * time.Second)
			}
		}

		utils.Log("EditContainer - Container stopped " + oldContainerID)
	} else {
		utils.Log("EditContainer - Revert started")
	}

	// recreate container with new informations
	createResponse, createError := DockerClient.ContainerCreate(
		DockerContext,
		newConfig.Config,
		newConfig.HostConfig,
		nil,
		nil,
		newName,
	)

	utils.Log("EditContainer - Container recreated. Re-connecting networks " + createResponse.ID)

	// is force secure
	isForceSecure := newConfig.Config.Labels["cosmos-force-network-secured"] == "true"
	
	// re-connect to networks
	for networkName, _ := range oldContainer.NetworkSettings.Networks {
		if(isForceSecure && networkName == "bridge") {
			utils.Log("EditContainer - Skipping network " + networkName + " (cosmos-force-network-secured is true)")
			continue
		}
		utils.Log("EditContainer - Connecting to network " + networkName)
		errNet := ConnectToNetworkSync(networkName, createResponse.ID)
		if errNet != nil {
			utils.Error("EditContainer - Failed to connect to network " + networkName, errNet)
		} else {
			utils.Debug("EditContainer - New Container connected to network " + networkName)
		}
	}
	
	utils.Log("EditContainer - Networks Connected. Starting new container " + createResponse.ID)

	runError := DockerClient.ContainerStart(DockerContext, createResponse.ID, types.ContainerStartOptions{})

	if createError != nil || runError != nil {
		if(oldContainerID == "") {
			if(createError == nil) {
				utils.Error("EditContainer - Failed to revert. Container is re-created but in broken state.", runError)
				return "", runError
			} else {
				utils.Error("EditContainer - Failed to revert. Giving up.", createError)
				return "", createError
			}
		}

		utils.Log("EditContainer - Failed to edit, attempting to revert changes")

		if(createError == nil) {
			utils.Log("EditContainer - Killing new broken container")
			DockerClient.ContainerKill(DockerContext, createResponse.ID, "")
		}

		utils.Log("EditContainer - Reverting...")
		// attempt to restore container
		restored, restoreError := EditContainer("", oldContainer)

		if restoreError != nil {
			utils.Error("EditContainer - Failed to restore container", restoreError)

			if createError != nil {
				utils.Error("EditContainer - re-create container ", createError)
				return "", createError
			} else {
				utils.Error("EditContainer - re-start container ", runError)
				return "", runError
			}
		} else {
			utils.Log("EditContainer - Container restored " + oldContainerID)
			errorWas := ""
			if createError != nil {
				errorWas = createError.Error()
			} else {
				errorWas = runError.Error()
			}
			return restored, errors.New("Failed to edit container, but restored to previous state. Error was: " + errorWas)
		}
	}

	utils.Log("EditContainer - Container started. All done! " + createResponse.ID)

	return createResponse.ID, nil
}

func ListContainers() ([]types.Container, error) {
	errD := Connect()
	if errD != nil {
		return nil, errD
	}

	containers, err := DockerClient.ContainerList(DockerContext, types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		return nil, err
	}

	return containers, nil
}

func AddLabels(containerConfig types.ContainerJSON, labels map[string]string) error {
	for key, value := range labels {
		containerConfig.Config.Labels[key] = value
	}

	return nil
}

func RemoveLabels(containerConfig types.ContainerJSON, labels []string) error {
	for _, label := range labels {
		delete(containerConfig.Config.Labels, label)
	}

	return nil
}

func IsLabel(containerConfig types.ContainerJSON, label string) bool {
	if containerConfig.Config.Labels[label] == "true" {
		return true
	}
	return false
}
func HasLabel(containerConfig types.ContainerJSON, label string) bool {
	if containerConfig.Config.Labels[label] != "" {
		return true
	}
	return false
}
func GetLabel(containerConfig types.ContainerJSON, label string) string {
	return containerConfig.Config.Labels[label]
}

func Test() error {

	// connect()

	// jellyfin, _ := DockerClient.ContainerInspect(DockerContext, "jellyfin")
	// ports := GetAllPorts(jellyfin)
	// fmt.Println(ports)

	// json jellyfin

	// fmt.Println(jellyfin.NetworkSettings)

	return nil
}
