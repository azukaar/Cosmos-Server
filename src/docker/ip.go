package docker

import (
	"fmt"
	"sync"
	"time"

	"github.com/azukaar/cosmos-server/src/utils"
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
	container, err := DockerClient.ContainerInspect(DockerContext, containerName)
	if err != nil {
		return "", err
	}

	// Prioritize "host"
	if net, ok := container.NetworkSettings.Networks["host"]; ok && net.IPAddress != "" {
		return net.IPAddress, nil
	}

	// Next, prioritize "bridge"
	if net, ok := container.NetworkSettings.Networks["bridge"]; ok && net.IPAddress != "" {
		return net.IPAddress, nil
	}

	// Finally, return the IP of the first network we find
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