package docker

import (
	"os"

	"github.com/azukaar/cosmos-server/src/utils" 
	
	"github.com/docker/docker/api/types"
	network "github.com/docker/docker/api/types/network"
	natting "github.com/docker/go-connections/nat"
)

func CreateCosmosNetwork() (string, error) {
	// check if network exists already
	networks, err := DockerClient.NetworkList(DockerContext, types.NetworkListOptions{})
	if err != nil {
		utils.Error("Docker Network List", err)
		return "", err
	}

	newNeworkName := ""
	for {
		newNeworkName = "cosmos-network-" + utils.GenerateRandomString(9)
		exists := false
		for _, network := range networks {
			if network.Name == newNeworkName {
				exists = true
			}
		}
		if !exists {
			break
		}
	}

	utils.Log("Creating new secure network: " + newNeworkName)

	// create network
	_, err = DockerClient.NetworkCreate(DockerContext, newNeworkName, types.NetworkCreate{
		Attachable: true,
	})

	if err != nil {
		utils.Error("Docker Network Create", err)
		return "", err
	}

	//if running in Docker, connect to main network
	if os.Getenv("HOSTNAME") != "" {
		err := DockerClient.NetworkConnect(DockerContext, newNeworkName, os.Getenv("HOSTNAME"), &network.EndpointSettings{})
	
		if err != nil {
			utils.Error("Docker Network Connect", err)
			return "", err
		}
	}

	return newNeworkName, nil
}

func ConnectToSecureNetwork(containerConfig types.ContainerJSON) (bool, error) {
	errD := connect()
	if errD != nil {
		utils.Error("Docker Connect", errD)
		return false, errD
	}

	if(!HasLabel(containerConfig, "cosmos-network-name")) {
		newNetwork, errNC := CreateCosmosNetwork()
		if errNC != nil {
			utils.Error("Docker Network Create", errNC)
			return false, errNC
		}

		utils.Log("Connected to new secure network for container: " + newNetwork)

		AddLabels(containerConfig, map[string]string{
			"cosmos-network-name": newNetwork,
		})

		return true, nil
	}
	
	netName := GetLabel(containerConfig, "cosmos-network-name")

	connected, err := IsConnectedToNetwork(containerConfig, netName)
	if err != nil {
		utils.Error("IsConnectedToNetwork", err)
		return false, err
	}
	if(connected) {
		return false, nil
	}

	errCo := DockerClient.NetworkConnect(DockerContext, netName, containerConfig.ID, &network.EndpointSettings{})

	if errCo != nil {
		utils.Error("Docker Network Connect", errCo)
		return false, errCo
	}

	return false, nil
}

func UnexposeAllPorts(container *types.ContainerJSON) error {
	container.HostConfig.PortBindings = natting.PortMap{}
	return nil
}

func GetAllPorts(container types.ContainerJSON) []string {
	ports := []string{}

	// prevent publishing all ports automatically
	container.HostConfig.PublishAllPorts = false

	for port, _ := range container.HostConfig.PortBindings {
		ports = append(ports, port.Port())
	}
	
	return ports
}

func IsConnectedToNetwork(containerConfig types.ContainerJSON, networkName string) (bool, error) {
	if(containerConfig.NetworkSettings == nil) {
		utils.Error("IsConnectedToNetwork: NetworkSettings is nil", nil)
		return false, nil
	}

	for name, _ := range containerConfig.NetworkSettings.Networks {
		if name == networkName {
			return true, nil
		}
	}

	return false, nil
}

func IsConnectedToASecureCosmosNetwork(containerConfig types.ContainerJSON) (bool, error) {
	if(containerConfig.NetworkSettings == nil) {
		utils.Error("IsConnectedToASecureCosmosNetwork: NetworkSettings is nil", nil)
		return false, nil
	}

	if(!HasLabel(containerConfig, "cosmos-network-name")) {
		return false, nil
	}

	for name, _ := range containerConfig.NetworkSettings.Networks {
		if name == GetLabel(containerConfig, "cosmos-network-name") {
			return true, nil
		}
	}

	return false, nil
}
