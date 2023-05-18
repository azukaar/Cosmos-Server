package docker

import (
	"github.com/azukaar/cosmos-server/src/utils" 
	"io"
	"os"
	// "github.com/docker/docker/client"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/mount"
	
	"runtime"
	"golang.org/x/sys/cpu"
)

type VolumeMount struct {
	Destination string
	Volume   *types.Volume
}

func NewDB() (string, error) {
	id := utils.GenerateRandomString(3)
	mongoUser := "cosmos-" + utils.GenerateRandomString(5) 
	mongoPass := utils.GenerateRandomString(24)
	monHost := "cosmos-mongo-" + id
	
	imageName := "mongo:latest"

	// if CPU is missing AVX, use 4.4
	if runtime.GOARCH == "amd64" && !cpu.X86.HasAVX {
		utils.Warn("CPU does not support AVX. Using mongo 4.4")
		imageName = "mongo:4.4"
	}

	err := RunContainer(
		imageName,
		monHost,
		[]string{
			"MONGO_INITDB_ROOT_USERNAME=" + mongoUser,
			"MONGO_INITDB_ROOT_PASSWORD=" + mongoPass,
		},
		[]VolumeMount{
			{
				Destination: "/data/db",
				Volume: &types.Volume{
					Name: "cosmos-mongo-data-" + id,
				},
			},
			{
				Destination: "/data/configdb",
				Volume: &types.Volume{
					Name: "cosmos-mongo-config-" + id,
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	return "mongodb://"+mongoUser+":"+mongoPass+"@"+monHost+":27017", nil
}

func RunContainer(imagename string, containername string, inputEnv []string, volumes []VolumeMount) error {
	errD := Connect()
	if errD != nil {
		utils.Error("Docker Connect", errD)
		return errD
	}

	pull, errPull := DockerClient.ImagePull(DockerContext, imagename, types.ImagePullOptions{})
	if errPull != nil {
		utils.Error("Docker Pull", errPull)
		return errPull
	}
	io.Copy(os.Stdout, pull)

	var mounts []mount.Mount

	for _, volume := range volumes {
		mount := mount.Mount{
			Type:   mount.TypeVolume,
			Source: volume.Volume.Name,
			Target: volume.Destination,
		}
		mounts = append(mounts, mount)
	}

	// Define a PORT opening
	// newport, err := natting.NewPort("tcp", port)
	// if err != nil {
	// 	fmt.Println("Unable to create docker port")
	// 	return err
	// }

	// Configured hostConfig: 
	// https://godoc.org/github.com/docker/docker/api/types/container#HostConfig
	hostConfig := &container.HostConfig{
		// PortBindings: natting.PortMap{
		// 	newport: []natting.PortBinding{
		// 		{
		// 			HostIP:   "0.0.0.0",
		// 			HostPort: port,
		// 		},
		// 	},
		// },
		Mounts : mounts,
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
		// LogConfig: container.LogConfig{
		// 	Type:   "json-file",
		// 	Config: map[string]string{},
		// },
	}

	// Define Network config
	// https://godoc.org/github.com/docker/docker/api/types/network#NetworkingConfig
	

	
	// networkConfig := &network.NetworkingConfig{
	// 	EndpointsConfig: map[string]*network.EndpointSettings{},
	// }
	// gatewayConfig := &network.EndpointSettings{
	// 	Gateway: "gatewayname",
	// }
	// networkConfig.EndpointsConfig["bridge"] = gatewayConfig

	// Define ports to be exposed (has to be same as hostconfig.portbindings.newport)
	// exposedPorts := map[natting.Port]struct{}{
	// 	newport: struct{}{},
	// }

	// Configuration 
	// https://godoc.org/github.com/docker/docker/api/types/container#Config
	config := &container.Config{
		Image:    imagename,
		Env: 		  inputEnv,
		Hostname: containername,
		Labels: map[string]string{
			"cosmos-force-network-secured": "true",
		},
		// ExposedPorts: exposedPorts,
	}

	//archi := runtime.GOARCH

	// Creating the actual container. This is "nil,nil,nil" in every example.
	cont, err := DockerClient.ContainerCreate(
		DockerContext,
		config,
		hostConfig,
		nil,
		nil,
		containername,
	)

	if err != nil {
		utils.Error("Docker Container Create", err)
		return err
	}

	// Run the actual container 
	DockerClient.ContainerStart(DockerContext, cont.ID, types.ContainerStartOptions{})
	utils.Log("Container created " + cont.ID)

	return nil
}