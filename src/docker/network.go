package docker

import (
	"os"
	"time"
	"errors"
	"sync"
	"net"
	"fmt"

	"github.com/azukaar/cosmos-server/src/utils" 
	
	"github.com/docker/docker/api/types"
	network "github.com/docker/docker/api/types/network"
	natting "github.com/docker/go-connections/nat"
	
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
		
		newNetwork, errNC := CreateCosmosNetwork()
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
		err := ConnectToNetworkSync(GetLabel(containerConfig, "cosmos-network-name"), containerConfig.ID)
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
	err = ConnectToNetworkSync(networkName, containerName)
	if err != nil {
		return err
	}

	err = ConnectToNetworkSync(networkName, container2Name)
	if err != nil {
		// disconnect first container
		DockerClient.NetworkDisconnect(DockerContext, networkName, containerName, true)
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
