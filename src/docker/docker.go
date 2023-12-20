package docker

import (
	"context"
	"errors"
	"time"
	"bufio"
	"os"
	"os/user"
	"io"
	"fmt"
	"strings"
	"encoding/base64"
	"encoding/json"
	"sync"
	"strconv"
	"runtime"
	"github.com/azukaar/cosmos-server/src/utils" 
	"github.com/docker/cli/cli/config"

	"github.com/docker/docker/client"
	// natting "github.com/docker/go-connections/nat"
	"github.com/docker/docker/api/types/container"
	composeTypes "github.com/compose-spec/compose-go/v2/types"
	mountType "github.com/docker/docker/api/types/mount"
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

func RecreateContainer(containerID string, containerConfig types.ContainerJSON) (string, error) {
	if os.Getenv("HOSTNAME") != ""  && os.Getenv("HOSTNAME") == containerID[1:] {
		err := SelfRecreate()
		if err != nil {
			return "", err
		}
	} else {
		return EditContainer(containerID, containerConfig, false)
	}
	
	utils.TriggerEvent(
		"cosmos.docker.recreate",
		"Cosmos Container Recreate",
		"success",
		"container@" + containerID,
		map[string]interface{}{
			"container": containerID,
	})

	return "", nil
}

func EditContainer(oldContainerID string, newConfig types.ContainerJSON, noLock bool) (string, error) {
	if(oldContainerID != "" && !noLock) {
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

	if(newConfig.HostConfig.NetworkMode != "bridge" &&
		 newConfig.HostConfig.NetworkMode != "default" &&
		 newConfig.HostConfig.NetworkMode != "host" &&
		 newConfig.HostConfig.NetworkMode != "none") {
			if(!HasLabel(newConfig, "cosmos-force-network-mode")) {
				AddLabels(newConfig, map[string]string{"cosmos-force-network-mode": string(newConfig.HostConfig.NetworkMode)})
			} else {
				newConfig.HostConfig.NetworkMode = container.NetworkMode(GetLabel(newConfig, "cosmos-force-network-mode"))
			}
	}
	
	newName := newConfig.Name
	oldContainer := newConfig

	if(oldContainerID != "") {
		utils.Log("EditContainer - inspecting previous container " + oldContainerID)

		// create missing folders
		
		for _, newmount := range newConfig.HostConfig.Mounts {
			if newmount.Type == mountType.TypeBind {
				newSource := newmount.Source

				if os.Getenv("HOSTNAME") != "" {
					if _, err := os.Stat("/mnt/host"); os.IsNotExist(err) {
						utils.Error("EditContainer: Unable to create directory for bind mount in the host directory. Please mount the host / in Cosmos with  -v /:/mnt/host to enable folder creations, or create the bind folder yourself", err)
					} else {
						newSource = "/mnt/host" + newSource
					}
				}
						
				utils.Log(fmt.Sprintf("Checking directory %s for bind mount", newSource))

				if _, err := os.Stat(newSource); os.IsNotExist(err) {
					utils.Log(fmt.Sprintf("Not found. Creating directory %s for bind mount", newSource))
	
					err := os.MkdirAll(newSource, 0755)
					if err != nil {
						utils.Error("EditContainer: Unable to create directory for bind mount", err)
						return "", errors.New("Unable to create directory for bind mount. Make sure parent directories exist, and that Cosmos has permissions to create directories in the host directory")
					}
		
					if newConfig.Config.User != "" {
						// Change the ownership of the directory to the container.User
						userInfo, err := user.Lookup(newConfig.Config.User)
						if err != nil {
							utils.Error("EditContainer: Unable to lookup user", err)
						} else {
							uid, _ := strconv.Atoi(userInfo.Uid)
							gid, _ := strconv.Atoi(userInfo.Gid)
							err = os.Chown(newSource, uid, gid)
							if err != nil {
								utils.Error("EditContainer: Unable to change ownership of directory", err)
							}
						}	
					}
				}
			}
		}

		utils.Log("EditContainer - Container updating. Retriveing currently running " + oldContainerID)

		var err error

		// get container informations
		// https://godoc.org/github.com/docker/docker/api/types#ContainerJSON
		oldContainer, err = DockerClient.ContainerInspect(DockerContext, oldContainerID)

		if err != nil {
			return "", err
		}

		// check if new image exists, if not, pull it
		_, _, errImage := DockerClient.ImageInspectWithRaw(DockerContext, newConfig.Config.Image)
		if errImage != nil {
			utils.Log("EditContainer - Image not found, pulling " + newConfig.Config.Image)
			out, errPull := DockerPullImage(newConfig.Config.Image)
			if errPull != nil {
				utils.Error("EditContainer - Image not found.", errPull)
				return "", errors.New("Image not found. " + errPull.Error())
			}
			defer out.Close()

			// wait for image pull to finish
			scanner := bufio.NewScanner(out)
			for scanner.Scan() {
				utils.Log(scanner.Text())
			}
		}

		// if no name, use the same one, that will force Docker to create a hostname if not set
		newName = oldContainer.Name

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
	
	// only force hostname if network is bridge or default, otherwise it will fail
	if newConfig.HostConfig.NetworkMode == "bridge" || newConfig.HostConfig.NetworkMode == "default" {
		newConfig.Config.Hostname = newName[1:]
	} else {
		// if not, remove hostname because otherwise it will try to keep the old one
		newConfig.Config.Hostname = ""
		// IDK Docker is weird, if you don't erase this it will break
		newConfig.Config.ExposedPorts = nil
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
	if createError != nil {
		utils.Error("EditContainer - Failed to create container", createError)
	}
	
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
		errNet := ConnectToNetworkLoose(networkName, createResponse.ID, true)
		if errNet != nil {
			utils.Error("EditContainer - Failed to connect to network " + networkName, errNet)
		} else {
			utils.Debug("EditContainer - New Container connected to network " + networkName)
		}
	}
	
	utils.Log("EditContainer - Networks Connected. Starting new container " + createResponse.ID)

	runError := DockerClient.ContainerStart(DockerContext, createResponse.ID, types.ContainerStartOptions{})

	if runError != nil {
		utils.Error("EditContainer - Failed to run container", runError)
	}

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
			// attempt kill
			DockerClient.ContainerKill(DockerContext, oldContainerID, "")
			DockerClient.ContainerKill(DockerContext, createResponse.ID, "")
			// attempt remove in case created state
			DockerClient.ContainerRemove(DockerContext, oldContainerID, types.ContainerRemoveOptions{})
			DockerClient.ContainerRemove(DockerContext, createResponse.ID, types.ContainerRemoveOptions{})
		}

		utils.Log("EditContainer - Reverting...")
		// attempt to restore container
		restored, restoreError := EditContainer("", oldContainer, false)

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
	
	// Recreating dependant containers
	utils.Debug("Unlocking EDIT Container")

	if oldContainerID != "" {
		RecreateDepedencies(oldContainerID)
	}

	utils.Log("EditContainer - Container started. All done! " + createResponse.ID)

	return createResponse.ID, nil
}

func RecreateDepedencies(containerID string) {
	containers, err := ListContainers()
	if err != nil {
		utils.Error("RecreateDepedencies", err)
		return
	}

	for _, container := range containers {
		if container.ID == containerID {
			continue
		}

		fullContainer, err := DockerClient.ContainerInspect(DockerContext, container.ID)
		if err != nil {
			utils.Error("RecreateDepedencies", err)
			continue
		}

		// check if network mode contains containerID
		if strings.Contains(string(fullContainer.HostConfig.NetworkMode), containerID) {
			utils.Log("RecreateDepedencies - Recreating " + container.Names[0])
			_, err := EditContainer(container.ID, fullContainer, true)
			if err != nil {
				utils.Error("RecreateDepedencies - Failed to update - ", err)
			}
		}
	}
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

func HasAutoUpdateOn(containerConfig types.ContainerJSON) bool {
	if containerConfig.Config.Labels["cosmos-auto-update"] == "true" {
		return true
	}

	config := utils.ReadConfigFromFile()

	if os.Getenv("HOSTNAME") == containerConfig.Name[1:] && config.AutoUpdate {
		return true
	}

	return false
}

func CheckUpdatesAvailable() map[string]bool {
	result := make(map[string]bool)

	// for each containers
	containers, err := ListContainers()
	if err != nil {
		utils.Error("CheckUpdatesAvailable", err)
		return result
	}

	for _, container := range containers {
		utils.Log("Checking for updates for " + container.Image)
		
		fullContainer, err := DockerClient.ContainerInspect(DockerContext, container.ID)
		if err != nil {
			utils.Error("CheckUpdatesAvailable", err)
			continue
		}

		// check container is running 
		if container.State != "running" {
			utils.Log("Container " + container.Names[0] + " is not running, skipping")
			continue
		}

		rc, err := DockerPullImage(container.Image)
		if err != nil {
			utils.Error("CheckUpdatesAvailable", err)
			continue
		}

		scanner := bufio.NewScanner(rc)
		defer  rc.Close()

		needsUpdate := false

		for scanner.Scan() {
			newStr := scanner.Text()
			// Check if a download has started
			if strings.Contains(newStr, "\"status\":\"Pulling fs layer\"") {
				utils.Log("Updates available for " + container.Image)

				result[container.Names[0]] = true
				if !HasAutoUpdateOn(fullContainer) {
					rc.Close()
					break
				} else {
					needsUpdate = true
				}
			} else if strings.Contains(newStr, "\"status\":\"Status: Image is up to date") {
				utils.Log("No updates available for " + container.Image)
				
				if !HasAutoUpdateOn(fullContainer) {
					rc.Close()
					break
				}
			} else {
				utils.Log(newStr)
			}
		}

		// no new image to pull, see if local image is matching
		if !result[container.Names[0]] && !needsUpdate {
			// check sum of local vs container image
			utils.Log("CheckUpdatesAvailable - Checking local image for change for " + container.Image)
			localImage, _, err := DockerClient.ImageInspectWithRaw(DockerContext, container.Image)
			if err != nil {
				utils.Error("CheckUpdatesAvailable - local image - ", err)
				continue
			}

			if localImage.ID != container.ImageID {
				result[container.Names[0]] = true
				needsUpdate = true
				utils.Log("CheckUpdatesAvailable - Local updates available for " + container.Image)
			} else {
				utils.Log("CheckUpdatesAvailable - No local updates available for " + container.Image)
			}
		}

		if needsUpdate && HasAutoUpdateOn(fullContainer) {
			utils.TriggerEvent(
				"cosmos.docker.container.update",
				"Cosmos Container Update",
				"success",
				"",
				map[string]interface{}{
					"container": container.Names[0][1:],
			})

			utils.WriteNotification(utils.Notification{
				Recipient: "admin",
				Title: "Container Update",
				Message: "Container " + container.Names[0][1:] + " updated to the latest version!",
				Level: "info",
				Link: "/cosmos-ui/servapps/containers/" + container.Names[0][1:],
			})

			utils.Log("Downloaded new update for " + container.Image + " ready to install")
			_, err := RecreateContainer(container.Names[0], fullContainer)
			if err != nil {
				utils.MajorError("Container failed to update", err)
			} else {
				result[container.Names[0]] = false
			}
		}
	}

	return result
}

func RemoveSelfUpdater() error {
	utils.Log("Checking for self updater agent")

	// look for a container with the name cosmos-self-updater-agent
	containers, err := ListContainers()
	if err != nil {
		utils.Error("RemoveSelfUpdater", err)
		return err
	}

	for _, container := range containers {
		if container.Names[0] == "/cosmos-self-updater-agent" {
			utils.Log("Found. Removing self updater agent")
			err := DockerClient.ContainerKill(DockerContext, container.ID, "SIGKILL")
			if err != nil {
				utils.Error("RemoveSelfUpdater", err)
			}
			err = DockerClient.ContainerRemove(DockerContext, container.ID, types.ContainerRemoveOptions{
				Force: true,
			})
			if err != nil {
				utils.Error("RemoveSelfUpdater", err)
				return err
			}
		}
	}

	return nil
}

func SelfRecreate() error {
	utils.Log("SelfRecreate - Starting...")

	if os.Getenv("HOSTNAME") == "" {
		utils.Error("SelfRecreate - not using Docker", nil)
		return errors.New("SelfRecreate - not using Docker")
	}

	// make sure to remove resiude of old self updater
	RemoveSelfUpdater()

	containerName := os.Getenv("HOSTNAME")

	version := "latest"

	// if arm
	if runtime.GOARCH == "arm64" {
		version = "latest-arm64"
	}
	
	service := DockerServiceCreateRequest{
		Services: map[string]ContainerCreateRequestContainer {},
	}

	utils.TriggerEvent(
		"cosmos.internal.self-updater",
		"Cosmos Self Updater",
		"important",
		"",
		map[string]interface{}{
			"action": "recreate",
			"container": containerName,
	})

	service.Services["cosmos-self-updater-agent"] = ContainerCreateRequestContainer{
		Name: "cosmos-self-updater-agent",
		Image: "azukaar/docker-self-updater:" + version,
		RestartPolicy: "no",
		SecurityOpt: []string{
			"label:disable",
		},
		Environment: []string{
			"CONTAINER_NAME=" + containerName,
			"ACTION=recreate",
			"DOCKER_HOST=" + os.Getenv("DOCKER_HOST"),
		},
		Volumes: []mountType.Mount{
			{
				Type: mountType.TypeBind,
				Source: "/var/run/docker.sock",
				Target: "/var/run/docker.sock",
			},
		},
	};

	err := CreateService(service, func (msg string) {})

	if err != nil {
		return err
	}

	return nil
}

func DockerPullImage(image string) (io.ReadCloser, error) {
	utils.Debug("DockerPull - Preparing Pulling image " + image)

	options := types.ImagePullOptions{}

	configfile, err := config.Load(config.Dir())
	if err != nil {
			utils.Error("DockerPull - Read config file error -", err)
	} else {
		slashIndex := strings.Index(image, "/")
		
		if slashIndex >= 1 && strings.ContainsAny(image[:slashIndex], ".:") {
			repoURL := strings.Split(image, "/")[0]
			creds, err := configfile.GetCredentialsStore(repoURL).Get(repoURL)
			
			if err != nil {
				utils.Error("DockerPull - Read config file error -", err)
				} else {	
				encodedJSON, _ := json.Marshal(creds)
				options.RegistryAuth = base64.URLEncoding.EncodeToString(encodedJSON)
			}
		} 
	}
	
	utils.Debug("DockerPull - Starting Pulling image " + image)

	out, errPull := DockerClient.ImagePull(DockerContext, image, options)

	return out, errPull
}

type ContainerStats struct {
	Name      string
	CPUUsage  float64
	MemUsage  uint64
	MemLimit	uint64
	NetworkRx float64
	NetworkTx float64
}

func Stats(container types.Container) (ContainerStats, error) {
		// utils.Debug("StatsAll - Getting stats for " + container.Names[0])
		// utils.Debug("Time: " + time.Now().String())
		
		statsBody, err := DockerClient.ContainerStats(DockerContext, container.ID, false)
		if err != nil {
			return ContainerStats{}, fmt.Errorf("error fetching stats for container %s: %s", container.ID, err)
		}

		defer statsBody.Body.Close()

		stats := types.StatsJSON{}
		if err := json.NewDecoder(statsBody.Body).Decode(&stats); err != nil {
			return ContainerStats{}, fmt.Errorf("error decoding stats for container %s: %s", container.ID, err)
		}

		previousCPU := stats.PreCPUStats.CPUUsage.TotalUsage
		previousSystem := stats.PreCPUStats.SystemUsage

		cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage) - float64(previousCPU)
		systemDelta := float64(stats.CPUStats.SystemUsage) - float64(previousSystem)

		perCore := len(stats.CPUStats.CPUUsage.PercpuUsage)
		if perCore == 0 {
			utils.Debug("StatsAll - Docker CPU PercpuUsage is 0")
			perCore = 1
		}

		// utils.Debug("StatsAll - CPU CPUUsage TotalUsage " + strconv.FormatUint(stats.CPUStats.CPUUsage.TotalUsage, 10))
		// utils.Debug("StatsAll - CPU PreCPUStats TotalUsage " + strconv.FormatUint(stats.PreCPUStats.CPUUsage.TotalUsage, 10))
		// utils.Debug("StatsAll - CPU CPUUsage PercpuUsage " + strconv.Itoa(perCore))
		// utils.Debug("StatsAll - CPU CPUUsage SystemUsage " + strconv.FormatUint(stats.CPUStats.SystemUsage, 10))
		
		// utils.Debug("StatsAll - CPU CPUUsage CPU Delta " + strconv.FormatFloat(cpuDelta, 'f', 6, 64))
		// utils.Debug("StatsAll - CPU CPUUsage System Delta " + strconv.FormatFloat(systemDelta, 'f', 6, 64))

		cpuUsage := 0.0

		if systemDelta > 0 && cpuDelta > 0 {
			cpuUsage = (cpuDelta / systemDelta) * float64(perCore) * 100
			
			// utils.Debug("StatsAll - CPU CPUUsage " + strconv.FormatFloat(cpuUsage, 'f', 6, 64))
		} else {
			utils.Debug("StatsAll - Error calculating CPU usage for " + container.Names[0])
		}

		// memUsage := float64(stats.MemoryStats.Usage) / float64(stats.MemoryStats.Limit) * 100
		
		netRx := 0.0
		netTx := 0.0
		
		for _, net := range stats.Networks {
			netRx += float64(net.RxBytes)
			netTx += float64(net.TxBytes)
		}

		containerStats := ContainerStats{
			Name:      strings.TrimPrefix(container.Names[0], "/"),
			CPUUsage:  cpuUsage * 100,
			MemUsage:  stats.MemoryStats.Usage,
			MemLimit:  stats.MemoryStats.Limit,
			NetworkRx: netRx,
			NetworkTx: netTx,
		}

		return containerStats, nil
	}

	func StatsAll() ([]ContainerStats, error) {
		containers, err := ListContainers()
		if err != nil {
			utils.Error("StatsAll", err)
			return nil, err
		}

		var containerStatsList []ContainerStats
		var wg sync.WaitGroup
		semaphore := make(chan struct{}, 5) // A channel with a buffer size of 5 for controlling parallelism.
	
		for _, container := range containers {
			// If not running, skip this container
			if container.State != "running" {
				continue
			}
	
			wg.Add(1)
			semaphore <- struct{}{} // Acquire a semaphore slot, limiting parallelism.
	
			go func(container types.Container) {
				defer func() {
					<-semaphore // Release the semaphore slot when done.
					wg.Done()
				}()
	
				stat, err := Stats(container)
				if err != nil {
					utils.Error("StatsAll", err)
					return
				}
				containerStatsList = append(containerStatsList, stat)
			}(container)
		}
	
		wg.Wait() // Wait for all goroutines to finish.
	
		return containerStatsList, nil
	}

	func StopContainer(containerName string) {
		err := DockerClient.ContainerStop(DockerContext, containerName, container.StopOptions{})
		if err != nil {
			utils.Error("StopContainer", err)
			return
		}
	}

	func MapEqualToArr(in map[string]string) []string {
		var out []string
		for key, value := range in {
			out = append(out, key + "=" + value)
		}
		return out
	}

	func ArrToMap(in []string) map[string]string {
		out := make(map[string]string)
		for _, value := range in {
			split := strings.Split(value, "=")
			out[split[0]] = split[1]
		}
		return out
	}

	func MappingToLabel(in composeTypes.MappingWithEquals) composeTypes.Labels {
		result := composeTypes.Labels{}
		for key, value := range in {
			result[key] = *value
		}
		return result
	}

	func LabelsToMapping(in composeTypes.Labels) composeTypes.MappingWithEquals {
		result := composeTypes.MappingWithEquals{}
		for key, val := range in {
			value := val
			result[key] = &value
		}
		return result
	}