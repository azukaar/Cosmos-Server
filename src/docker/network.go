package docker

import (
	"os"
	"time"
	"errors"
	"sync"
	"net"
	"fmt"

	"github.com/azukaar/cosmos-server/src/utils" 
	
	yaml "gopkg.in/yaml.v2"
	"github.com/docker/docker/api/types"
	network "github.com/docker/docker/api/types/network"
	natting "github.com/docker/go-connections/nat"
	composeTypes "github.com/compose-spec/compose-go/v2/types"
	
	"encoding/binary"
)

// use a semaphore lock
var DockerNetworkLock = make(chan bool, 1)

func intToIP(nn uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, nn)
	return ip
}

func ipToInt(ip net.IP) uint32 {
	return binary.BigEndian.Uint32(ip.To4())
}

func getNextSubnet(subnet string) string {
	ip, ipnet, _ := net.ParseCIDR(subnet)
	ipInt := ipToInt(ip.To4())
	maskSize, _ := ipnet.Mask.Size()

	nextIPInt := ipInt + (1 << uint(32-maskSize))
	nextIP := intToIP(nextIPInt)

	return fmt.Sprintf("%s/%d", nextIP, maskSize)
}

func doesSubnetOverlap(networks []types.NetworkResource, subnet string) bool {
	_, ipnet, _ := net.ParseCIDR(subnet)
	for _, network := range networks {
		for _, config := range network.IPAM.Config {
			_, existingNet, _ := net.ParseCIDR(config.Subnet)
			maskSize, _ := ipnet.Mask.Size()
			if existingNet.Contains(ipnet.IP) || existingNet.Contains(intToIP(ipToInt(ipnet.IP) + uint32(1<<(32-maskSize)) - 1)) {
				return true
			}
		}
	}
	return false
}

func findAvailableSubnet() string {
	baseSubnet := "100.0.0.0/29"

	networks, err := DockerClient.NetworkList(DockerContext, types.NetworkListOptions{})
	if err != nil {
		panic(err)
	}

	for doesSubnetOverlap(networks, baseSubnet) {
		baseSubnet = getNextSubnet(baseSubnet)
	}

	return baseSubnet
}

func CreateCosmosNetwork(name string) (string, error) {
	networks, err := DockerClient.NetworkList(DockerContext, types.NetworkListOptions{})
	if err != nil {
		utils.Error("Docker Network List", err)
		return "", err
	}

	newNeworkName := ""
	for {
		newNeworkName = "cosmos-" + name + "-" + utils.GenerateRandomString(3)
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
	
	subnet := findAvailableSubnet()

	// create network
	newNetHan, err := DockerClient.NetworkCreate(DockerContext, newNeworkName, types.NetworkCreate{
		Attachable: true,
		IPAM: &network.IPAM{
			Driver: "default",
			Config: []network.IPAMConfig{
				network.IPAMConfig{
					Subnet: subnet,
				},
			},
		},
	})

	utils.Debug("New network created: " + newNeworkName + " with id " + newNetHan.ID)

	if err != nil {
		utils.Error("Docker Network Create", err)
		return "", err
	}

	return newNeworkName, nil
}

func AttachNetworkToCosmos(newNeworkName string ) error {
	utils.Log("Connecting Cosmos to network " + newNeworkName)
	utils.Debug("HOSTNAME: " + os.Getenv("HOSTNAME"))
	if os.Getenv("HOSTNAME") != "" {
		err := DockerClient.NetworkConnect(DockerContext, newNeworkName, os.Getenv("HOSTNAME"), &network.EndpointSettings{})
	
		if err != nil {
			utils.Error("Docker Network Connect", err)
			return err
		}
		return nil
	}
	return nil
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
		newNetwork, errNC := CreateCosmosNetwork(containerConfig.Name[1:])
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
	
	errCo := ConnectToNetworkLoose(netName, containerConfig.ID, true)
	
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
		utils.Debug("IsConnectedToNetwork - " + containerConfig.Name + ": has no settings")
	}

	for name, _ := range containerConfig.NetworkSettings.Networks {
		if name == networkName {
			return true
			utils.Debug("IsConnectedToNetwork - " + containerConfig.Name + ": is connected to " + networkName)
		}
	}

	utils.Debug("IsConnectedToNetwork - " + containerConfig.Name + ": is NOT connected to " + networkName)
	return false
}

func IsConnectedToASecureCosmosNetwork(self types.ContainerJSON, containerConfig types.ContainerJSON) (bool, error, bool) {
	if(containerConfig.NetworkSettings == nil) {
		utils.Error("IsConnectedToASecureCosmosNetwork: NetworkSettings is nil", nil)
		return false, nil, false
	}

	if(!HasLabel(containerConfig, "cosmos-network-name")) {
		return false, nil, false
	}

	// if network doesn't exists
	_, err := DockerClient.NetworkInspect(DockerContext, GetLabel(containerConfig, "cosmos-network-name"), types.NetworkInspectOptions{})
	if err != nil {
		utils.Error("Container tries to connect to a non existing Cosmos network, replacing it.", err)
		
		newNetwork, errNC := CreateCosmosNetwork(containerConfig.Name[1:])
		if errNC != nil {
			utils.Error("DockerSecureNetworkCreate", errNC)
			return false, errNC, false
		}

		AddLabels(containerConfig, map[string]string{
			"cosmos-network-name": newNetwork,
		})

		utils.Log("Added label to replace faulty secure network for container: " + newNetwork)

		return true, nil, true
	}

	// Check if container is already connected to the network it claimed
	connected := false
	for name, _ := range containerConfig.NetworkSettings.Networks {
		if name == GetLabel(containerConfig, "cosmos-network-name") {
			connected = true
		}
	}
	if(!connected) {
		utils.Warn("Container wasn't connected to the network it claimed, connecting it.")
		err := ConnectToNetworkLoose(GetLabel(containerConfig, "cosmos-network-name"), containerConfig.ID, true)
		if err != nil {
			utils.Error("ReConnectToSecureNetworkConnectFromLabel", err)
			return false, err, false
		}
	}

	// else check if Cosmos is already connected
	if os.Getenv("HOSTNAME") != "" {
		utils.Debug("IsSelfConnectedToASecureCosmosNetwork - " + self.Name + ": is connected to " + GetLabel(containerConfig, "cosmos-network-name"))
		if(!IsConnectedToNetwork(self, GetLabel(containerConfig, "cosmos-network-name"))) {
			err := DockerClient.NetworkConnect(DockerContext, GetLabel(containerConfig, "cosmos-network-name"), os.Getenv("HOSTNAME"), &network.EndpointSettings{})
			if err != nil {
				utils.Error("Docker Network Connect EXISTING ", err)
				return false, err, false
			}
		}
	}
	return true, nil, false
}

func ConnectToNetworkLoose(networkName string, containerID string, connect bool) error {
	utils.Log("ConnectToNetworkLoose: " + networkName + " - " + containerID)

	var err error
	
	if connect {
		err = DockerClient.NetworkConnect(DockerContext, networkName, containerID, &network.EndpointSettings{})
	} else {
		err = DockerClient.NetworkDisconnect(DockerContext, networkName, containerID, true)
	}

	if err != nil {
		utils.Error("ConnectToNetworkLoose", err)
		return err
	}

	// wait for connection to be established
	retries := 10
	for {
		newContainer, err := DockerClient.ContainerInspect(DockerContext, containerID)
		if err != nil {
			utils.Error("ConnectToNetworkLoose", err)
			return err
		}
		
		if(IsConnectedToNetwork(newContainer, networkName)) {
			if connect {
				break
			}
		} else {
			if !connect {
				break
			}
		}

		retries--
		if retries == 0 {
			return errors.New("ConnectToNetworkLoose: Timeout")
		}
		time.Sleep(500 * time.Millisecond)
	}
	

	utils.Log("Connected "+containerID+" to network: " + networkName)

	return nil
}

func ConnectToNetworkCompose(networkName string, containerName string, connect bool) error {
	utils.Log("ConnectToNetworkCompose: " + networkName + " - " + containerName)

	container, err := DockerClient.ContainerInspect(DockerContext, containerName)
	if err != nil {
		return err
	}

	dcWorkingDir := container.Config.Labels["com.docker.compose.project.working_dir"]
	dcServiceName := container.Config.Labels["com.docker.compose.service"]

	project, err := LoadComposeFromName(containerName)
	if err != nil {
		return err
	}

	if project.Services == nil {
		return errors.New("ConnectToNetworkCompose: project.Services is nil")
	}
	
	service := project.Services[dcServiceName]

	composeNetworkName := networkName
	// check if network has a different name in the compose
	for key, network := range project.Networks {
		if network.Name == networkName {
			composeNetworkName = key
		}
	}

	// check if already connected 
	for key, _ := range service.Networks {
		if key == composeNetworkName {
			if connect {
				utils.Log("Already connected "+containerName+" to network: " + composeNetworkName + ". Skipping.")
				break
			} else {
				utils.Log("Disconnecting "+containerName+" from network: " + composeNetworkName)
				delete(service.Networks, key)
				break
			}
		}
	}

	if connect {
		value := composeTypes.ServiceNetworkConfig{}
		service.Networks[composeNetworkName] = &value
		
		// if network does not exist in compose
		if _, ok := project.Networks[composeNetworkName]; !ok {
			// add network as external 
			project.Networks[networkName] = composeTypes.NetworkConfig{
				External: true,
			}
		}
	} 
	
	project.Services[dcServiceName] = service

	if !connect {
		// check if no services is connected to this network
		connected := false
		for _, service := range project.Services {
			for key, _ := range service.Networks {
				if key == composeNetworkName {
					utils.Log("Network still connected to service: " + service.Name)
					connected = true
				}
			}
		}

		if !connected {
			// remove network as external 
			delete(project.Networks, composeNetworkName)
		}
	}

	// Marshal the struct to JSON or yml format.
	data, err := yaml.Marshal(project)
	if err != nil {
		return err
	}

	// Write the data to the file.
	err = SaveComposeFromName(containerName, data)
	if err != nil {
		return err
	}

	ComposeUp(dcWorkingDir, func(message string, outputType int) {
		utils.Debug(message)
	})

	return nil
}

func ConnectToNetwork(networkName string, containerName string) error {
	utils.Log("ConnectToNetwork: " + networkName + " - " + containerName)

	container, err := DockerClient.ContainerInspect(DockerContext, containerName)

	if err != nil {
		return err
	}

	// if label container docker.compose.project 
	if container.Config.Labels["com.docker.compose.project"] != "" {
		utils.Log("UpdateContainer: Updating docker compose container")
		return ConnectToNetworkCompose(networkName, containerName, true)
	} else {
		utils.Log("UpdateContainer: Updating lose container")
		return ConnectToNetworkLoose(networkName, containerName, true)
	}

	return nil
}

func DisconnectToNetwork(networkName string, containerName string) error {
	utils.Log("DisconnectToNetwork: " + networkName + " - " + containerName)

	container, err := DockerClient.ContainerInspect(DockerContext, containerName)

	if err != nil {
		return err
	}

	// if label container docker.compose.project 
	if container.Config.Labels["com.docker.compose.project"] != "" {
		utils.Log("UpdateContainer: Updating docker compose container")
		return ConnectToNetworkCompose(networkName, containerName, false)
	} else {
		utils.Log("UpdateContainer: Updating lose container")
		return ConnectToNetworkLoose(networkName, containerName, false)
	}

	return nil
}


func _debounceNetworkCleanUp() func(string) {
	var mu sync.Mutex
	var timer *time.Timer

	return func(networkId string) {
		if(networkId == "bridge" || networkId == "host" || networkId == "none") {
			return
		}

		mu.Lock()
		defer mu.Unlock()

		if timer != nil {
			timer.Stop()
		}

		timer = time.AfterFunc(30*time.Minute, NetworkCleanUp)
	}
}

func CreateLinkNetwork(containerName string, container2Name string) error {
	subnet := findAvailableSubnet()

	// create network
	networkName := "cosmos-link-" + containerName + "-" + container2Name + "-" + utils.GenerateRandomString(2)
	_, err := DockerClient.NetworkCreate(DockerContext, networkName, types.NetworkCreate{
		CheckDuplicate: true,
		Labels: map[string]string{
			"cosmos-link": "true",
			"cosmos-link-name": networkName,
			"cosmos-link-container1": containerName,
			"cosmos-link-container2": container2Name,
		},
		IPAM: &network.IPAM{
			Config: []network.IPAMConfig{
				network.IPAMConfig{
					Subnet: subnet,
				},
			},
		},
	})

	if err != nil {
		return err
	}

	// connect containers to network
	err = ConnectToNetwork(networkName, containerName)
	if err != nil {
		return err
	}

	err = ConnectToNetwork(networkName, container2Name)
	if err != nil {
		// disconnect first container
		DisconnectToNetwork(networkName, containerName)
		// destroy network
		DockerClient.NetworkRemove(DockerContext, networkName)
		return err
	}

	return nil
}

var DebouncedNetworkCleanUp = _debounceNetworkCleanUp()

func NetworkCleanUp() {
	config := utils.GetMainConfig()
	
	if(config.DockerConfig.SkipPruneNetwork) {
		return
	}

	DockerNetworkLock <- true
	defer func() { <-DockerNetworkLock }()

	
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
		
		if(len(network.Containers) == 0) {
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
			// docker inspect self
			selfContainer, err := DockerClient.ContainerInspect(DockerContext, self)
			if err != nil {
				utils.Error("NetworkCleanUpInspectSelf", err)
				continue
			}

			// check if self is network_mode to this network
			if(string(selfContainer.HostConfig.NetworkMode) == network.Name) {
				utils.Warn("Skipping network cleanup because self is network_mode to this network")
				continue
			}

			utils.Log("Disconnecting and removing zombie network: " + network.Name)
			err = DockerClient.NetworkDisconnect(DockerContext, network.ID, self, true)
			if err != nil {
				utils.Error("DockerNetworkCleanupDisconnect", err)
				continue
			}
			err = DockerClient.NetworkRemove(DockerContext, network.ID)
			if err != nil {
				utils.Error("DockerNetworkCleanupRemove", err)
			}
		}
	}
}

func DeleteNetworkLoose(networkName string) error {
	utils.Log("Deleting network: " + networkName)
	err := DockerClient.NetworkRemove(DockerContext, networkName)
	if err != nil {
		return err
	}
	return nil
}

func DeleteNetworkCompose(networkName, composeProjectDir string) error {
	utils.Log("DeleteNetworkCompose: " + networkName + " - " + composeProjectDir)

	project, err := LoadComposeFromName(composeProjectDir)
	if err != nil {
		return err
	}

	deleted := false
	for key, network := range project.Networks {
		if network.Name == networkName {
			delete(project.Networks, key)
			deleted = true
		}
	}
	if !deleted {
		delete(project.Networks, networkName)
	}

	// Marshal the struct to JSON or yml format.
	data, err := yaml.Marshal(project)
	if err != nil {
		return err
	}

	// Write the data to the file.
	err = SaveComposeFromName(composeProjectDir, data)
	if err != nil {
		return err
	}

	ComposeUp(composeProjectDir, func(message string, outputType int) {
		utils.Debug(message)
	})

	return nil
}

func DeleteNetwork(networkName string) error {
	utils.Log("DeleteNetwork: " + networkName)
	network, err := DockerClient.NetworkInspect(DockerContext, networkName, types.NetworkInspectOptions{})
	if err != nil {
		return err
	}

	composeProject := network.Labels["com.docker.compose.project"]

	if composeProject != "" {
		containers := network.Containers
		// inspect containers until we find the compose file location
		composeProjectDir := ""

		for _, container := range containers {
			utils.Debug("Checking container: " + container.Name)
			containerConfig, err := DockerClient.ContainerInspect(DockerContext, container.Name)
			if err != nil {
				return err
			}

			utils.Debug(containerConfig.Config.Labels["com.docker.compose.project"])
			
			if containerConfig.Config.Labels["com.docker.compose.project"] == composeProject {
				composeProjectDir = containerConfig.Config.Labels["com.docker.compose.project.working_dir"]
				break
			}
		}

		if composeProjectDir == "" {
			// return DeleteNetworkLoose(networkName)
			return errors.New("DeleteNetwork: Couldn't find docker-compose project directory for network " + networkName + ". Aborting to prevent desync.")
		}

		return DeleteNetworkCompose(networkName, composeProjectDir)
	} else {
		return DeleteNetworkLoose(networkName)
	}

	return nil
}