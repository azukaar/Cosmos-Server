package docker

import (
	"fmt"
	"sync"
	"time"
	"strings"

	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/docker/docker/api/types"
)

type Cache struct {
	data map[string]cacheItem
	mu   sync.Mutex
}

type cacheItem struct {
	ipAddress string
	expires   time.Time
}

func NewCache() *Cache {
	return &Cache{
		data: make(map[string]cacheItem),
	}
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, found := c.data[key]
	if !found || time.Now().After(item.expires) {
		if found {
			delete(c.data, key)
		}
		return "", false
	}
	return item.ipAddress, true
}

func (c *Cache) Set(key string, value string, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = cacheItem{
		ipAddress: value,
		expires:   time.Now().Add(duration),
	}
}

func _getContainerIPByName(containerName string) (string, error) {
	errD := Connect()
	if errD != nil {
		return "", errD
	}

	container, err := DockerClient.ContainerInspect(DockerContext, containerName)
	if err != nil {
		return "", err
	}

	// Prioritize "host"
	// if net, ok := container.NetworkSettings.Networks["host"]; ok && net.IPAddress != "" {
	// 	return "localhost", nil
	// }

	// Next, prioritize "bridge"
	// if net, ok := container.NetworkSettings.Networks["bridge"]; ok && net.IPAddress != "" {
	// 	return net.IPAddress, nil
	// }

	// if container is in network mode host
	if container.HostConfig.NetworkMode == "host" {
		return "localhost", nil
	} else if container.HostConfig.NetworkMode == "bridge" || container.HostConfig.NetworkMode == "default" || container.HostConfig.NetworkMode == "" {
		// if container is in network mode bridge or default or not set
		
		for _, net := range container.NetworkSettings.Networks {
			// if this network is using the bridge driver 
			if net.IPAddress != "" {
				return net.IPAddress, nil
			}
		}
	} else if strings.HasPrefix(string(container.HostConfig.NetworkMode), "container:") {
		// if container is in network mode container:<container_name>

		// get the container name
		otherContainerName := strings.TrimPrefix(string(container.HostConfig.NetworkMode), "container:")
		// get the ip of the other container
		ip, err := _getContainerIPByName(otherContainerName)
		if err != nil {
			return "", err
		}
		return ip, nil
	} else {
		// if container is in network mode <network_name>

		// get the network name
		networkName := string(container.HostConfig.NetworkMode)
		// get the network
		network, err := DockerClient.NetworkInspect(DockerContext, networkName, types.NetworkInspectOptions{})
		if err != nil {
			return "", err
		}
		// get the ip of the container in the network
		for _, containerEP := range network.Containers {
			if containerEP.Name == container.Name {
				return network.IPAM.Config[0].Gateway, nil
			}
		}
	}

	// Finally, if nothing, return the IP of the first network we find
	for _, net := range container.NetworkSettings.Networks {
		if net.IPAddress != "" {
			return net.IPAddress, nil
		}
	}

	return "", fmt.Errorf("no IP address found for container %s", containerName)
}

var cache = NewCache()

func GetContainerIPByName(containerName string) (ip string, err error) {
	// Check cache first
	ip, found := cache.Get(containerName)
	if !found {
		ip, err = _getContainerIPByName(containerName)
		if err != nil {
			utils.Error("Docker - Cannot get container IP", err)
			return "", err
		}

		cache.Set(containerName, ip, 10*time.Second)
	}

	utils.Debug("Docker - Docker IP " + containerName + " : " + ip)
	return ip, nil
}

func DoesContainerExist(containerName string) bool {
	errD := Connect()
	if errD != nil {
		return false
	}

	_, err := DockerClient.ContainerInspect(DockerContext, containerName)
	
	if err != nil {
		return false
	}

	return true
}