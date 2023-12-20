package docker

import (
	"github.com/azukaar/cosmos-server/src/utils" 
	"io"
	"os"
	"net/http"
	"fmt"
	"errors"

	// "github.com/docker/docker/client"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/mount"
	volumeType "github.com/docker/docker/api/types/volume"
	
	"runtime"
	"golang.org/x/sys/cpu"
)

type VolumeMount struct {
	Destination string
	Volume   *volumeType.Volume
}

func NewDB(w http.ResponseWriter, req *http.Request) (string, error) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Transfer-Encoding", "chunked")
	
	flusher, ok := w.(http.Flusher)
	if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return "", errors.New("Streaming unsupported!")
	}

	fmt.Fprintf(w, "NewInstall: Create DB\n")
	flusher.Flush()

	id := utils.GenerateRandomString(3)
	mongoUser := "cosmos-" + utils.GenerateRandomString(5) 
	mongoPass := utils.GenerateRandomString(24)
	monHost := "cosmos-mongo-" + id
	
	imageName := "mongo:latest"
	
	//if ARM use arm64v8/mongo
	if runtime.GOARCH == "arm64" {
		utils.Warn("ARM64 detected. Using ARM mongo 4.4.18")
		imageName = "arm64v8/mongo:4.4.18"

	// if CPU is missing AVX, use 4.4
	} else if runtime.GOARCH == "amd64" && !cpu.X86.HasAVX {
		utils.Warn("CPU does not support AVX. Using mongo 4.4")
		imageName = "mongo:4.4"
	}

	service := DockerServiceCreateRequest{
		Services: map[string]ContainerCreateRequestContainer {},
	}

	service.Services[monHost] = ContainerCreateRequestContainer{
		Name: monHost,
		Image: imageName,
		RestartPolicy: "always",
		Environment: []string{
			"MONGO_INITDB_ROOT_USERNAME=" + mongoUser,
			"MONGO_INITDB_ROOT_PASSWORD=" + mongoPass,
		},
		Labels: map[string]string{
			"cosmos-force-network-secured": "true",
		},
		Volumes: []mount.Mount{
			{
				Type:   mount.TypeVolume,
				Source: "cosmos-mongo-data-" + id,
				Target: "/data/db",
			},
			{
				Type:   mount.TypeVolume,
				Source: "cosmos-mongo-config-" + id,
				Target: "/data/configdb",
			},
		},
	};

	err := CreateService(service, 
		func (msg string) {
			utils.Log(msg)
			fmt.Fprintf(w, msg + "\n")
			flusher.Flush()
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

	pull, errPull := DockerPullImage(imagename)
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
