package docker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"bytes"
	"errors"
	"gopkg.in/yaml.v2"
	"os"

	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/docker/docker/api/types"

	conttype "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
)

var ExportError = "" 

func ExportContainer(containerID string) (ContainerCreateRequestContainer, error)  {
		// Fetch detailed info of each container
		detailedInfo, err := DockerClient.ContainerInspect(DockerContext, containerID)
		if err != nil {
			ExportError = "Export Docker - Cannot inspect container" + containerID + " - " + err.Error()
			return ContainerCreateRequestContainer{}, errors.New(ExportError)
		}

		// Map the detailedInfo to your ContainerCreateRequestContainer struct
		// Here's a simplified example, you'd need to handle all the fields
		service := ContainerCreateRequestContainer{
			Name:         strings.TrimPrefix(detailedInfo.Name, "/"),
			Image:        detailedInfo.Config.Image,
			Environment:  detailedInfo.Config.Env,
			Labels:       detailedInfo.Config.Labels,
			Command:      strings.Join(detailedInfo.Config.Cmd, " "),
			Entrypoint:   strings.Join(detailedInfo.Config.Entrypoint, " "),
			WorkingDir:   detailedInfo.Config.WorkingDir,
			User:         detailedInfo.Config.User,
			Tty:          detailedInfo.Config.Tty,
			StdinOpen:    detailedInfo.Config.OpenStdin,
			Hostname:     detailedInfo.Config.Hostname,
			Domainname:   detailedInfo.Config.Domainname,
			MacAddress:   detailedInfo.NetworkSettings.MacAddress,
			NetworkMode:  string(detailedInfo.HostConfig.NetworkMode),
			StopSignal:   detailedInfo.Config.StopSignal,
			HealthCheck:  ContainerCreateRequestContainerHealthcheck {
			},
			DNS:              detailedInfo.HostConfig.DNS,
			DNSSearch:        detailedInfo.HostConfig.DNSSearch,
			ExtraHosts:       detailedInfo.HostConfig.ExtraHosts,
			SecurityOpt:      detailedInfo.HostConfig.SecurityOpt,
			StorageOpt:       detailedInfo.HostConfig.StorageOpt,
			Sysctls:          detailedInfo.HostConfig.Sysctls,
			Isolation:        string(detailedInfo.HostConfig.Isolation),
			CapAdd:           detailedInfo.HostConfig.CapAdd,
			CapDrop:          detailedInfo.HostConfig.CapDrop,
			Privileged:       detailedInfo.HostConfig.Privileged,
			
			// StopGracePeriod:  int(detailedInfo.HostConfig.StopGracePeriod.Seconds()),
			
			// Ports
			Ports: func() []string {
					ports := []string{}
					for port, binding := range detailedInfo.NetworkSettings.Ports {
							for _, b := range binding {
									ports = append(ports, fmt.Sprintf("%s:%s:%s/%s", b.HostIP, b.HostPort, port.Port(), port.Proto()))
							}
					}
					return ports
			}(),

			// Volumes
			Volumes: func() []mount.Mount {
					mounts := []mount.Mount{}
					for _, m := range detailedInfo.Mounts {
						mount := mount.Mount{
							Type:        m.Type,
							Source:      m.Source,
							Target:      m.Destination,
							ReadOnly:    !m.RW,
							// Consistency: mount.Consistency(m.Consistency),
						}

						if m.Type == "volume" {
							nodata := strings.Split(strings.TrimSuffix(m.Source, "/_data"), "/")
							mount.Source = nodata[len(nodata)-1]
						}

						mounts = append(mounts, mount)
					}
					return mounts
			}(),
			// Networks
			Networks: func() map[string]ContainerCreateRequestServiceNetwork {
					networks := make(map[string]ContainerCreateRequestServiceNetwork)
					for netName, _ := range detailedInfo.NetworkSettings.Networks {
							networks[netName] = ContainerCreateRequestServiceNetwork{
									// Aliases:     netConfig.Aliases,
									// IPV4Address: netConfig.IPAddress,
									// IPV6Address: netConfig.GlobalIPv6Address,
							}
					}
					return networks
			}(),

			DependsOn:      []string{},  // This is not directly available from inspect. It's part of docker-compose.
			RestartPolicy:  string(detailedInfo.HostConfig.RestartPolicy.Name),
			Devices:        func() []string {
					var devices []string
					for _, device := range detailedInfo.HostConfig.Devices {
							devices = append(devices, fmt.Sprintf("%s:%s", device.PathOnHost, device.PathInContainer))
					}
					return devices
			}(),
			Expose:         []string{},  // This information might need to be derived from other properties
		}

		// healthcheck
		if detailedInfo.Config.Healthcheck != nil {
			service.HealthCheck.Test = detailedInfo.Config.Healthcheck.Test
			service.HealthCheck.Interval = int(detailedInfo.Config.Healthcheck.Interval.Seconds())
			service.HealthCheck.Timeout = int(detailedInfo.Config.Healthcheck.Timeout.Seconds())
			service.HealthCheck.Retries = detailedInfo.Config.Healthcheck.Retries
			service.HealthCheck.StartPeriod = int(detailedInfo.Config.Healthcheck.StartPeriod.Seconds())
		}

		// user UID/GID
		if detailedInfo.Config.User != "" {
			parts := strings.Split(detailedInfo.Config.User, ":")
			if len(parts) == 2 {
				uid, err := strconv.Atoi(parts[0])
				if err != nil {
					service.UID = uid
				}
				gid, err := strconv.Atoi(parts[1])
				if err != nil {
					service.GID = gid
				}
			}
		}

		//expose 
		// for _, port := range detailedInfo.Config.ExposedPorts {
			
		// }

		return service, nil
}

func ExportDocker() {
	config := utils.GetMainConfig()
	if config.NewInstall {
		return
	}

	ExportError = "" 
	
	errD := Connect()
	if errD != nil {
		ExportError = "Export Docker - cannot connect - " + errD.Error()
		utils.MajorError("ExportDocker - connect - ", errD)
		return
	}

	finalBackup := DockerServiceCreateRequest{}
	
	// List containers
	containers, err := DockerClient.ContainerList(DockerContext, conttype.ListOptions{})
	if err != nil {
		utils.MajorError("ExportDocker - Cannot list containers", err)
		ExportError = "Export Docker - Cannot list containers - " + err.Error()
		return
	}

	
	// Convert the containers into your custom format
	var services = make(map[string]ContainerCreateRequestContainer)

	for _, container := range containers {	
		service, err := ExportContainer(container.ID)
		if err != nil {
			utils.MajorError("ExportDocker - Cannot export container", err)
			return
		}

		services[strings.TrimPrefix(service.Name, "/")] = service
	}

	// List networks
	networks, err := DockerClient.NetworkList(DockerContext, types.NetworkListOptions{})
	if err != nil {
		utils.MajorError("Export Docker - Cannot list networks", err)
		ExportError = "Export Docker - Cannot list networks - " + err.Error()
		return
	}

	finalBackup.Networks = make(map[string]ContainerCreateRequestNetwork)

	// Convert the networks into custom format
	for _, network := range networks {
		if network.Name == "bridge" || network.Name == "host" || network.Name == "none" {
			continue
		}

		// Fetch detailed info of each network
		detailedInfo, err := DockerClient.NetworkInspect(DockerContext, network.ID, types.NetworkInspectOptions{})
		if err != nil {
			utils.MajorError("Export Docker - Cannot inspect network", err)
			ExportError = "Export Docker - Cannot inspect network - " + err.Error()
			return
		}

		// Map the detailedInfo to ContainerCreateRequestContainer struct
		network := ContainerCreateRequestNetwork{
			Name:         detailedInfo.Name,
			Driver:       detailedInfo.Driver,
			Internal:     detailedInfo.Internal,
			Attachable:   detailedInfo.Attachable,
			EnableIPv6:   detailedInfo.EnableIPv6,
			Labels:       detailedInfo.Labels,
		}

		network.IPAM.Driver = detailedInfo.IPAM.Driver
		for _, config := range detailedInfo.IPAM.Config {
			network.IPAM.Config = append(network.IPAM.Config, ContainerCreateRequestNetworkIPAMConfig{
				Subnet:  config.Subnet,
				Gateway: config.Gateway,
			})
		}

		finalBackup.Networks[detailedInfo.Name] = network
	}

	// remove cosmos from services
	if os.Getenv("HOSTNAME") != "" {
		cosmos := services[os.Getenv("HOSTNAME")]
		delete(services, os.Getenv("HOSTNAME"))

		// export separately cosmos
		// Create a buffer to hold the JSON output
		var buf bytes.Buffer

		// Create a new yaml encoder that writes to the buffer
		encoder := yaml.NewEncoder(&buf)

		// Set escape HTML to false to avoid escaping special characters
		// encoder.SetEscapeHTML(false)
		//format
		// encoder.SetIndent("", "  ")

		// Use the encoder to write the structured data to the buffer
		toExport := map[string]map[string]ContainerCreateRequestContainer {
			"services": map[string]ContainerCreateRequestContainer {
				os.Getenv("HOSTNAME"): cosmos,
			},
		}

		err = encoder.Encode(toExport)
		if err != nil {
				utils.MajorError("Export Docker - Cannot marshal docker backup", err)
				ExportError = "Export Docker - Cannot marshal docker backup - " + err.Error()
		}

		// The JSON data is now in buf.Bytes()
		yamlData := buf.Bytes()

		// Write the JSON data to a file
		err = ioutil.WriteFile(utils.CONFIGFOLDER + "cosmos.docker-compose.yaml", yamlData, 0644)
		if err != nil {
				utils.MajorError("Export Docker - Cannot save docker backup", err)
				ExportError = "Export Docker - Cannot save docker backup - " + err.Error()
		}
	}

	// Convert the services map to your finalBackup struct
	finalBackup.Services = services

	// Create a buffer to hold the JSON output
	var buf bytes.Buffer

	// Create a new JSON encoder that writes to the buffer
	encoder := json.NewEncoder(&buf)

	// Set escape HTML to false to avoid escaping special characters
	encoder.SetEscapeHTML(false)
	//format
	encoder.SetIndent("", "  ")

	// Use the encoder to write the structured data to the buffer
	err = encoder.Encode(finalBackup)
	if err != nil {
			utils.MajorError("Export Docker - Cannot marshal docker backup", err)
			ExportError = "Export Docker - Cannot marshal docker backup - " + err.Error()
	}

	// The JSON data is now in buf.Bytes()
	jsonData := buf.Bytes()

	// Write the JSON data to a file
	err = ioutil.WriteFile(utils.CONFIGFOLDER + "backup.cosmos-compose.json", jsonData, 0644)
	if err != nil {
		utils.MajorError("Export Docker - Cannot save docker backup", err)
		ExportError = "Export Docker - Cannot save docker backup - " + err.Error()
	}
}