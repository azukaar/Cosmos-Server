package docker

import (
	"time"

	"github.com/docker/docker/client"
)

// ContainerLiveness is a point-in-time snapshot of whether a container is up and how long the current run has lasted.
type ContainerLiveness struct {
	// Exists is false when docker has no such container (a definite answer, unlike an error).
	Exists bool
	// Running is true only for state "running".
	Running bool
	// Uptime is how long the CURRENT run has lasted. Zero unless Running.
	Uptime time.Duration
}

// GetContainerLiveness reports a container's live state, uncached; a missing container is (Exists: false, nil), not an error.
func GetContainerLiveness(containerName string) (ContainerLiveness, error) {
	if err := Connect(); err != nil {
		return ContainerLiveness{}, err
	}

	container, err := DockerClient.ContainerInspect(DockerContext, containerName)
	if err != nil {
		if client.IsErrNotFound(err) {
			return ContainerLiveness{Exists: false}, nil
		}
		return ContainerLiveness{}, err
	}

	if container.State == nil {
		return ContainerLiveness{Exists: true}, nil
	}
	if container.State.Status != "running" {
		return ContainerLiveness{Exists: true}, nil
	}

	startedAt, perr := time.Parse(time.RFC3339Nano, container.State.StartedAt)
	if perr != nil {
		// running but uptime unknown: return an error rather than guess
		return ContainerLiveness{}, perr
	}

	return ContainerLiveness{
		Exists:  true,
		Running: true,
		Uptime:  time.Since(startedAt),
	}, nil
}
