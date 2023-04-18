package docker

import (
	"os"
	"time"
	"errors"

	"github.com/azukaar/cosmos-server/src/utils" 
	
	"github.com/docker/docker/api/types"
	network "github.com/docker/docker/api/types/network"
	natting "github.com/docker/go-connections/nat"
)

// use a semaphore lock
var DockerNetworkLock = make(chan bool, 1)

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
	newNetHan, err := DockerClient.NetworkCreate(DockerContext, newNeworkName, types.NetworkCreate{
		Attachable: true,
	})

	utils.Debug("New network created: " + newNeworkName + " with id " + newNetHan.ID)

	if err != nil {
		utils.Error("Docker Network Create", err)
		return "", err
	}

	//if running in Docker, connect to main network
	// utils.Debug("HOSTNAME: " + os.Getenv("HOSTNAME"))
	// if os.Getenv("HOSTNAME") != "" {
	// 	err := DockerClient.NetworkConnect(DockerContext, newNeworkName, os.Getenv("HOSTNAME"), &network.EndpointSettings{})
	
	// 	if err != nil {
	// 		utils.Error("Docker Network Connect", err)
	// 		return "", err
	// 	}
	// }

	return newNeworkName, nil
}

func ConnectToSecureNetwork(containerConfig types.ContainerJSON) (bool, error) {
	errD := Connect()
	if errD != nil {
		utils.Error("Docker Connect", errD)
		return false, errD
	}

	needsRestart := false
	netName := "" 

	if(!HasLabel(containerConfig, "cosmos-network-name")) {
		newNetwork, errNC := CreateCosmosNetwork()
		if errNC != nil {
			utils.Error("DockerSecureNetworkCreate", errNC)
			return false, errNC
		}

		utils.Log("Added label to new secure network for container: " + newNetwork)
		
		AddLabels(containerConfig, map[string]string{
			"cosmos-network-name": newNetwork,
		})
		
		needsRestart = true
		netName = newNetwork
		} else {
			netName = GetLabel(containerConfig, "cosmos-network-name")
			
			//if network doesn't exists
			_, err := DockerClient.NetworkInspect(DockerContext, netName, types.NetworkInspectOptions{})
			if err != nil {
				utils.Error("Container tries to connect to a non existing Cosmos network, resetting", err)
				RemoveLabels(containerConfig, []string{"cosmos-network-name"})
				return ConnectToSecureNetwork(containerConfig)
			}
	
		// else check if already connected
		if(IsConnectedToNetwork(containerConfig, netName)) {
			return false, nil
		}
	}
	
	// at this point network is created and container is not connected to it
	
	errCo := ConnectToNetworkSync(netName, containerConfig.ID)
	
	utils.Log("Created and connected to secure network: " + netName)

	if errCo != nil {
		utils.Error("ConnectToSecureNetworkConnect", errCo)
		return false, errCo
	}

	return needsRestart, nil
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

func IsConnectedToNetwork(containerConfig types.ContainerJSON, networkName string) bool {
	if(containerConfig.NetworkSettings == nil) {
		return false
	}

	for name, _ := range containerConfig.NetworkSettings.Networks {
		if name == networkName {
			return true
		}
	}

	return false
}

func IsConnectedToASecureCosmosNetwork(self types.ContainerJSON, containerConfig types.ContainerJSON) (bool, error) {
	if(containerConfig.NetworkSettings == nil) {
		utils.Error("IsConnectedToASecureCosmosNetwork: NetworkSettings is nil", nil)
		return false, nil
	}

	if(!HasLabel(containerConfig, "cosmos-network-name")) {
		return false, nil
	}

	for name, _ := range containerConfig.NetworkSettings.Networks {
		if name == GetLabel(containerConfig, "cosmos-network-name") {
			if os.Getenv("HOSTNAME") != "" {
				if(!IsConnectedToNetwork(self, name)) {
					err := DockerClient.NetworkConnect(DockerContext, name, os.Getenv("HOSTNAME"), &network.EndpointSettings{})
					if err != nil {
						utils.Error("Docker Network Connect EXISTING ", err)
						return false, err
					}
				}
			}
			return true, nil
		}
	}

	return false, nil
}

func ConnectToNetworkIfNotConnected(containerConfig types.ContainerJSON, networkName string) error {
	if(IsConnectedToNetwork(containerConfig, networkName)) {
		return nil
	}
	return ConnectToNetworkSync(networkName, containerConfig.ID)
}

func ConnectToNetworkSync(networkName string, containerID string) error {
	err := DockerClient.NetworkConnect(DockerContext, networkName, containerID, &network.EndpointSettings{})
	if err != nil {
		utils.Error("ConnectToNetworkSync", err)
		return err
	}

	// wait for connection to be established
	retries := 10
	for {
		newContainer, err := DockerClient.ContainerInspect(DockerContext, containerID)
		if err != nil {
			utils.Error("ConnectToNetworkSync", err)
			return err
		}
		
		if(IsConnectedToNetwork(newContainer, networkName)) {
			break
		}

		retries--
		if retries == 0 {
			return errors.New("ConnectToNetworkSync: Timeout")
		}
		time.Sleep(500 * time.Millisecond)
	}
	

	utils.Log("Connected "+containerID+" to network: " + networkName)

	return nil
}

func NetworkCleanUp() {
	return
	DockerNetworkLock <- true
	defer func() { <-DockerNetworkLock }()

	config := utils.GetMainConfig()
	
	utils.Log("Cleaning up orphan networks...")

	// list every network
	networks, err := DockerClient.NetworkList(DockerContext, types.NetworkListOptions{})
	if err != nil {
		utils.Error("NetworkCleanUpList", err)
		return
	}


	// check if network is empty or has only self as container
	for _, networkHollow := range networks {
		utils.Debug("Checking network: " + networkHollow.Name)

		if(networkHollow.Name == "bridge" || networkHollow.Name == "host" || networkHollow.Name == "none") {
			continue
		}
		
		// inspect network because the Docker API is a complete mess :)
		network, err := DockerClient.NetworkInspect(DockerContext, networkHollow.ID, types.NetworkInspectOptions{})
		if err != nil {
			utils.Error("NetworkCleanUpInspect", err)
			continue
		}

		if(len(network.Containers) > 1) {
			continue
		}

		utils.Debug("Ready to Check network: " + network.Name)

		if(config.DockerConfig.SkipPruneNetwork){
			utils.Debug("Skipping network prune")
		}
		
		if(!config.DockerConfig.SkipPruneNetwork && len(network.Containers) == 0) {
			utils.Log("Removing orphan network: " + network.Name)
			err := DockerClient.NetworkRemove(DockerContext, network.ID)
			if err != nil {
				utils.Error("DockerNetworkCleanupRemove", err)
			}
			continue
		}

		self := os.Getenv("HOSTNAME")
		if self == "" {
			utils.Warn("Skipping zombie network cleanup because not a docker cosmos container")
			continue
		}

		utils.Debug("Checking self name: " + self)
		utils.Debug("Checking non-empty network: " + network.Name)
		
		containsCosmos := false
		for _, container := range network.Containers {
			utils.Debug("Checking name: " + container.Name)
			if(container.Name == self) {
				containsCosmos = true
			}
		}
		
		if(containsCosmos) {
			utils.Log("Disconnecting and removing zombie network: " + network.Name)
			err := DockerClient.NetworkDisconnect(DockerContext, network.ID, self, true)
			if err != nil {
				utils.Error("DockerNetworkCleanupDisconnect", err)
			}
			err = DockerClient.NetworkRemove(DockerContext, network.ID)
			if err != nil {
				utils.Error("DockerNetworkCleanupRemove", err)
			}
		}
	}
}
