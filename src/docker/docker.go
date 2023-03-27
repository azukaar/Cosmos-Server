package docker

import (
	"context"
	"fmt"
	"errors"

	"github.com/azukaar/cosmos-server/src/utils" 

	"github.com/docker/docker/client"
	natting "github.com/docker/go-connections/nat"
	"github.com/docker/docker/api/types/container"
	network "github.com/docker/docker/api/types/network"
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

func connect() error {
	if DockerClient == nil {
		ctx := context.Background()
		client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			return err
		}
		defer client.Close()

		DockerClient = client
		DockerContext = ctx

		ping, err := DockerClient.Ping(DockerContext)
		if ping.APIVersion != "" && err == nil {
			utils.Log("Docker Connected")
		} else {
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

func runContainer(imagename string, containername string, port string, inputEnv []string) error {
	errD := connect()
	if errD != nil {
		utils.Error("Docker Connect", errD)
		return errD
	}

	// Define a PORT opening
	newport, err := natting.NewPort("tcp", port)
	if err != nil {
		fmt.Println("Unable to create docker port")
		return err
	}

	// Configured hostConfig: 
	// https://godoc.org/github.com/docker/docker/api/types/container#HostConfig
	hostConfig := &container.HostConfig{
		PortBindings: natting.PortMap{
			newport: []natting.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: port,
				},
			},
		},
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
		LogConfig: container.LogConfig{
			Type:   "json-file",
			Config: map[string]string{},
		},
	}

	// Define Network config (why isn't PORT in here...?:
	// https://godoc.org/github.com/docker/docker/api/types/network#NetworkingConfig
	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{},
	}
	gatewayConfig := &network.EndpointSettings{
		Gateway: "gatewayname",
	}
	networkConfig.EndpointsConfig["bridge"] = gatewayConfig

	// Define ports to be exposed (has to be same as hostconfig.portbindings.newport)
	exposedPorts := map[natting.Port]struct{}{
		newport: struct{}{},
	}

	// Configuration 
	// https://godoc.org/github.com/docker/docker/api/types/container#Config
	config := &container.Config{
		Image:        imagename,
		Env: 		  inputEnv,
		ExposedPorts: exposedPorts,
		Hostname:     fmt.Sprintf("%s-hostnameexample", imagename),
	}

	//archi := runtime.GOARCH

	// Creating the actual container. This is "nil,nil,nil" in every example.
	cont, err := DockerClient.ContainerCreate(
		context.Background(),
		config,
		hostConfig,
		networkConfig,
		nil,
		containername,
	)

	if err != nil {
		utils.Error("Docker Container Create", err)
		return err
	}

	// Run the actual container 
	DockerClient.ContainerStart(context.Background(), cont.ID, types.ContainerStartOptions{})
	utils.Log("Container created " + cont.ID)


	return nil
}

func EditContainer(containerID string, newConfig types.ContainerJSON) (string, error) {
	errD := connect()
	if errD != nil {
		return "", errD
	}
	utils.Log("Container updating " + containerID)

	// get container informations
	// https://godoc.org/github.com/docker/docker/api/types#ContainerJSON
	oldContainer, err := DockerClient.ContainerInspect(DockerContext, containerID)

	if err != nil {
		return "", err
	}

	// if no name, use the same one, that will force Docker to create a hostname if not set
	newName := oldContainer.Name
	newConfig.Config.Hostname = newName

	// stop and remove container
	stopError := DockerClient.ContainerStop(DockerContext, containerID, container.StopOptions{})
	if stopError != nil {
		return "", stopError
	}

	removeError := DockerClient.ContainerRemove(DockerContext, containerID, types.ContainerRemoveOptions{})
	if removeError != nil {
		return "", removeError
	}

	utils.Log("Container stopped " + containerID)

	// recreate container with new informations
	createResponse, createError := DockerClient.ContainerCreate(
		DockerContext,
		newConfig.Config,
		newConfig.HostConfig,
		nil,
		nil,
		newName,
	)

	runError := DockerClient.ContainerStart(DockerContext, createResponse.ID, types.ContainerStartOptions{})

	if runError != nil {
		return "", runError
	}

	utils.Log("Container recreated " + createResponse.ID)

	if createError != nil {
		// attempt to restore container
		_, restoreError := DockerClient.ContainerCreate(DockerContext, oldContainer.Config, nil, nil, nil, oldContainer.Name)
		if restoreError != nil {
			utils.Error("Failed to restore Docker Container after update failure", restoreError)
		}

		return "", createError
	}
	

	return createResponse.ID, nil
}

func ListContainers() ([]types.Container, error) {
	errD := connect()
	if errD != nil {
		return nil, errD
	}

	containers, err := DockerClient.ContainerList(DockerContext, types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}

	// for _, container := range containers {
	// 	fmt.Println("ID - ", container.ID)
	// 	fmt.Println("ID - ", container.Names)
	// 	fmt.Println("ID - ", container.Image)
	// 	fmt.Println("ID - ", container.Command)
	// 	fmt.Println("ID - ", container.State)
	// 	fmt.Println("Ports - ", container.Ports)
	// 	fmt.Println("HostConfig - ", container.HostConfig)
	// 	fmt.Println("ID - ", container.Labels)
	// 	fmt.Println("NetworkSettings - ", container.NetworkSettings)
	// 	if(container.NetworkSettings.Networks["cosmos-network"] != nil) {
	// 		fmt.Println("IP COSMOS - ", container.NetworkSettings.Networks["cosmos-network"].IPAddress);
	// 	}
	// 	if(container.NetworkSettings.Networks["bridge"] != nil) {
	// 		fmt.Println("IP bridge - ", container.NetworkSettings.Networks["bridge"].IPAddress);
	// 	}
	// }

	return containers, nil
}

func AddLabels(containerConfig types.ContainerJSON, labels map[string]string) error {
	for key, value := range labels {
		containerConfig.Config.Labels[key] = value
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