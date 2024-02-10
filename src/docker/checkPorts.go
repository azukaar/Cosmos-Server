package docker

import (
	"github.com/azukaar/cosmos-server/src/utils"

	"os"
	"net"
	"strings"
	"errors"
	"runtime"
	mountType "github.com/docker/docker/api/types/mount"
)

func isPortAvailable(port string) bool {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return false
	}

	ln.Close()
	return true
}

func CheckPorts() error {
	utils.Log("Setup: Checking Docker port mapping ")
	expectedPorts := []string{}
	isHTTPS := utils.IsHTTPS
	config := utils.GetMainConfig()
	HTTPPort := config.HTTPConfig.HTTPPort
	HTTPSPort := config.HTTPConfig.HTTPSPort
	routes := config.HTTPConfig.ProxyConfig.Routes
	expectedPort := HTTPPort
	if isHTTPS {
		expectedPort = HTTPSPort
	}

	for _, route := range routes {
		if route.UseHost && strings.Contains(route.Host, ":") {
			hostname := route.Host
			port := strings.Split(hostname, ":")[1]
			expectedPorts = append(expectedPorts, port)
		}
	}

	// append hostname port 
	hostname := config.HTTPConfig.Hostname
	if strings.Contains(hostname, ":") {
		hostnameport := strings.Split(hostname, ":")[1]
		expectedPorts = append(expectedPorts, hostnameport)
	}

	errD := Connect()
	if errD != nil {
		return errD
	}

	// Get container ID from HOSTNAME environment variable
	containerID := os.Getenv("HOSTNAME")
	if containerID == "" {
		utils.Warn("Not in docker. Skipping port auto-mapping")
		return nil
	}

	// Inspect the container
	inspect, err := DockerClient.ContainerInspect(DockerContext, containerID)
	if err != nil {
		return err
	}

	// Get the ports
	ports := map[string]struct{}{}
	finalPorts := []string{}

	for containerPort, hostConfig := range inspect.NetworkSettings.Ports {
		utils.Debug("Container port: " + containerPort.Port() + "/" + containerPort.Proto())
		
		for _, hostPort := range hostConfig {
			utils.Debug("Host port: " + hostPort.HostPort)
			ports[hostPort.HostPort] = struct{}{}
			finalPorts = append(finalPorts, hostPort.HostPort + ":" + containerPort.Port() + "/" + containerPort.Proto())
		}
	}

	hasChanged := false

	utils.Debug("Expected ports: " + strings.Join(expectedPorts, ", "))

	for _, port := range expectedPorts {
		utils.Debug("Checking port : " + string(port))
		if _, ok := ports[port]; !ok {
			if !isPortAvailable(port) {
				utils.Error("Port "+port+" is added to a URL but it is not available. Skipping for now", nil)
			} else {
				utils.Debug("Port "+port+" is not mapped. Adding it.")
				finalPorts = append(finalPorts, port + ":" + expectedPort)
				hasChanged = true
			}
		}
	}

	if hasChanged {
		utils.Log("Port mapping changed. Needs update.")
		utils.Log("New ports: " + strings.Join(finalPorts, ", "))

		UpdatePorts(finalPorts)
		return nil
	}

	utils.Log("Port mapping not changed.")
	return nil
}


func UpdatePorts(finalPorts []string) error {
	utils.Log("SelUpdatePorts - Starting...")

	if os.Getenv("HOSTNAME") == "" {
		utils.Error("SelUpdatePorts - not using Docker", nil)
		return errors.New("SelUpdatePorts - not using Docker")
	}

	containerName := os.Getenv("HOSTNAME")
	utils.Log("SelUpdatePorts - Container name: " + containerName)

	// make sure to remove resiude of old self updater
	RemoveSelfUpdater()

	version := "latest"

	// if arm
	if runtime.GOARCH == "arm64" {
		version = "latest-arm64"
	}
	
	service := DockerServiceCreateRequest{
		Services: map[string]ContainerCreateRequestContainer {},
	}

	service.Services["cosmos-self-updater-agent"] = ContainerCreateRequestContainer{
		Name: "cosmos-self-updater-agent",
		Image: "azukaar/docker-self-updater:" + version,
		RestartPolicy: "no",
		SecurityOpt: []string{
			"label:disable",
		},
		Environment: []string{
			"CONTAINER_NAME=" + containerName,
			"ACTION=ports",
			"DOCKER_HOST=" + os.Getenv("DOCKER_HOST"),
			"PORTS=" + strings.Join(finalPorts, ","),
		},
		Volumes: []mountType.Mount{
			{
				Type: mountType.TypeBind,
				Source: "/var/run/docker.sock",
				Target: "/var/run/docker.sock",
			},
		},
	};
	
	
	utils.Log("SelUpdatePorts - Creating updater service")
	utils.Log("Creating self-updater service: docker run -d --name cosmos-self-updater-agent -e CONTAINER_NAME=" + containerName + " -e ACTION=ports -e DOCKER_HOST=" + os.Getenv("DOCKER_HOST") + " -e PORTS=" + strings.Join(finalPorts, ",") + " -v /var/run/docker.sock:/var/run/docker.sock azukaar/docker-self-updater:" + version)

	err := CreateService(service, func (msg string) {})

	if err != nil {
		return err
	}

	return nil
}