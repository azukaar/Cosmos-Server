package docker

import (
	"github.com/azukaar/cosmos-server/src/utils" 
	"io"
	"os"
	"net/http"


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

func NewDB(w http.ResponseWriter, req *http.Request) (string, error) {
	id := utils.GenerateRandomString(3)
	mongoUser := "cosmos-" + utils.GenerateRandomString(5) 
	mongoPass := utils.GenerateRandomString(24)
	monHost := "cosmos-mongo-" + id
	
	imageName := "mongo:latest"
	
	//if ARM use amd64/mongo
	if runtime.GOARCH == "arm64" {
		utils.Warn("ARM64 detected. Using ARM mongo 4.4")
		imageName = "amd64/mongo:4.4"

	// if CPU is missing AVX, use 4.4
	} else if runtime.GOARCH == "amd64" && !cpu.X86.HasAVX {
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

	hostConfig := &container.HostConfig{
		Mounts : mounts,
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
	}
	
	config := &container.Config{
		Image:    imagename,
		Env: 		  inputEnv,
		Hostname: containername,
		Labels: map[string]string{
			"cosmos-force-network-secured": "true",
		},
		// ExposedPorts: exposedPorts,
	}

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

	DockerClient.ContainerStart(DockerContext, cont.ID, types.ContainerStartOptions{})
	utils.Log("Container created " + cont.ID)

	return nil
}