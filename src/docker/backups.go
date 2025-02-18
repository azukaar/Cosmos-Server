package docker

import (
	"strings"
	"path/filepath"
	"fmt"
	"time"
	
	contstuff "github.com/docker/docker/api/types/container"

	"github.com/azukaar/cosmos-server/src/utils"
)

// arePathsRelated checks if either path is a parent of the other or if they are the same
func arePathsRelated(path1, path2 string) bool {
	// Clean and normalize both paths
	path1 = filepath.Clean(path1)
	path2 = filepath.Clean(path2)

	// If paths are equal, return true
	if path1 == path2 {
		return true
	}

	// Check if path1 is a parent of path2
	rel1, err1 := filepath.Rel(path1, path2)
	if err1 == nil && !strings.HasPrefix(rel1, "..") {
		return true
	}

	// Check if path2 is a parent of path1
	rel2, err2 := filepath.Rel(path2, path1)
	if err2 == nil && !strings.HasPrefix(rel2, "..") {
		return true
	}

	return false
}

// getStackContainers returns all containers that are part of the same stack as the given container
func getStackContainers(containerID string) ([]string, error) {
	var stackContainers []string
	
	// Get the container details to find stack labels
	container, err := DockerClient.ContainerInspect(DockerContext, containerID)
	if err != nil {
		return nil, err
	}

	// Check for compose project name (stack name)
	stackName := ""
	for label, value := range container.Config.Labels {
		if label == "com.docker.compose.project" || label == "cosmos-stack" {
			stackName = value
			break
		}
	}

	// If no stack found, return just this container
	if stackName == "" {
		return []string{containerID}, nil
	}

	// List all containers
	containers, err := ListContainers()
	if err != nil {
		return nil, err
	}

	// Find all containers with the same stack name
	for _, cont := range containers {
		fullContainer, err := DockerClient.ContainerInspect(DockerContext, cont.ID)
		if err != nil {
			continue
		}

		for label, value := range fullContainer.Config.Labels {
			if (label == "com.docker.compose.project" || label == "cosmos-stack") && value == stackName {
				stackContainers = append(stackContainers, cont.ID)
				break
			}
		}
	}

	return stackContainers, nil
}

// GetContainersUsingPath returns a list of container IDs that have volumes or binds
// that are either equal to or contained within the specified path
func GetContainersUsingPath(path string) ([]string, error) {
	var affectedContainers []string
	seenContainers := make(map[string]bool)

	// Connect to Docker if not already connected
	errD := Connect()
	if errD != nil {
		return nil, errD
	}

	// List all containers
	containers, err := ListContainers()
	if err != nil {
		return nil, err
	}

	// Check each container
	for _, container := range containers {
		// Skip containers that are stopped
		if container.State == "exited" || container.State == "dead" || container.State == "created" || container.State == "removing" || container.State == "paused" {
			continue
		}

		// Get detailed container info
		fullContainer, err := DockerClient.ContainerInspect(DockerContext, container.ID)
		if err != nil {
			continue
		}

		// Check mounts (both volumes and binds)
		isAffected := false
		for _, mount := range fullContainer.Mounts {
			var sourcePath string
			
			switch mount.Type {
			case "bind":
				sourcePath = mount.Source
			case "volume":
				// For volumes, we need to check in /var/lib/docker/volumes
				sourcePath = filepath.Join("/var/lib/docker/volumes", mount.Name, "_data")
			default:
				continue
			}

			utils.Debug("Checking container " + container.ID + " mount " + sourcePath + " against " + path)

			if arePathsRelated(path, sourcePath) {
				isAffected = true
				break
			}
		}

		if isAffected {
			// Get all containers in the same stack
			stackContainers, err := getStackContainers(container.ID)
			if err != nil {
				continue
			}

			// Add all stack containers to the result list (avoiding duplicates)
			for _, stackContainer := range stackContainers {
				if !seenContainers[stackContainer] {
					affectedContainers = append(affectedContainers, stackContainer)
					seenContainers[stackContainer] = true
				}
			}
		}
	}

	return affectedContainers, nil
}


// StopContainers stops the given containers and waits for them to stop
func StopContainers(containerIDs []string) error {
	errD := Connect()
	if errD != nil {
		return errD
	}

	for _, containerID := range containerIDs {
		// Get container state
		container, err := DockerClient.ContainerInspect(DockerContext, containerID)
		if err != nil {
			return fmt.Errorf("failed to inspect container %s: %v", containerID, err)
		}

		// Skip if already stopped
		if !container.State.Running {
			continue
		}

		// Stop the container with a timeout
		err = DockerClient.ContainerStop(DockerContext, containerID, contstuff.StopOptions{})
		if err != nil {
			return fmt.Errorf("failed to stop container %s: %v", containerID, err)
		}

		// Wait until container is actually stopped
		for {
			container, err := DockerClient.ContainerInspect(DockerContext, containerID)
			if err != nil {
				return fmt.Errorf("failed to inspect container %s: %v", containerID, err)
			}

			if !container.State.Running {
				break
			}

			time.Sleep(100 * time.Millisecond)
		}
	}

	return nil
}

// StartContainers starts the given containers and waits for them to be running
func StartContainers(containerIDs []string) error {
	errD := Connect()
	if errD != nil {
		return errD
	}

	for _, containerID := range containerIDs {
		// Get container state
		container, err := DockerClient.ContainerInspect(DockerContext, containerID)
		if err != nil {
			return fmt.Errorf("failed to inspect container %s: %v", containerID, err)
		}

		// Skip if already running
		if container.State.Running {
			continue
		}

		// Start the container
		err = DockerClient.ContainerStart(DockerContext, containerID, contstuff.StartOptions{})
		if err != nil {
			return fmt.Errorf("failed to start container %s: %v", containerID, err)
		}
	}

	return nil
}