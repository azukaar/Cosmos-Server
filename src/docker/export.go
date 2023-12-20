package docker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"bytes"
	"gopkg.in/yaml.v2"
	"os"

	"github.com/azukaar/cosmos-server/src/utils"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	volumeTypes "github.com/docker/docker/api/types/volume"
)

var ExportError = "" 

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
	containers, err := DockerClient.ContainerList(DockerContext, types.ContainerListOptions{})
	if err != nil {
		utils.MajorError("ExportDocker - Cannot list containers", err)
		ExportError = "Export Docker - Cannot list containers - " + err.Error()
		return
	}

	
	// Convert the containers into your custom format
	var services = make(map[string]ContainerCreateRequestContainer)
	for _, container := range containers {
		// Fetch detailed info of each container
		detailedInfo, err := DockerClient.ContainerInspect(DockerContext, container.ID)
		if err != nil {
			utils.MajorError("Export Docker - Cannot inspect container" + container.Names[0], err)
			ExportError = "Export Docker - Cannot inspect container" + container.Names[0] + " - " + err.Error()
			return
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
			SysctlsMap:       detailedInfo.HostConfig.Sysctls,
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
		
		services[strings.TrimPrefix(detailedInfo.Name, "/")] = service
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

func exportService(containerName string) (string, error) {
    // Connect to Docker
    err := Connect()
    if err != nil {
        return "", fmt.Errorf("exportService - cannot connect - %s", err.Error())
    }

		utils.Log("exportService - export service from container " + containerName)

    // Inspect the specified container
    container, err := DockerClient.ContainerInspect(DockerContext, containerName)
    if err != nil {
        return "", fmt.Errorf("exportService - cannot inspect container %s - %s", containerName, err.Error())
    }

    // Check for docker-compose project label
    projectName, projectOk := container.Config.Labels["com.docker.compose.project"]
    composeFileLocation, fileLocOk := container.Config.Labels["com.docker.compose.project.config_files"]

    var services []types.ContainerJSON
    var networksMap = make(map[string]types.NetworkResource)
    var volumesMap = make(map[string]volumeTypes.Volume)

    if projectOk {
				utils.Log("exportService - part of stack " + projectName)
        // List all containers
        containers, err := DockerClient.ContainerList(DockerContext, types.ContainerListOptions{})
        if err != nil {
            return "", fmt.Errorf("exportService - cannot list containers - %s", err.Error())
        }

        // Export all containers in the stack
        for _, c := range containers {
            if c.Labels["com.docker.compose.project"] == projectName {
                service, err := DockerClient.ContainerInspect(DockerContext, c.ID)
                if err != nil {
                    return "", fmt.Errorf("exportService - cannot inspect container %s - %s", c.ID, err.Error())
                }
                services = append(services, service)
            }
        }

        // List and export all networks in the stack
        networkFilter := filters.NewArgs()
        networkFilter.Add("label", "com.docker.compose.project="+projectName)
        networks, err := DockerClient.NetworkList(DockerContext, types.NetworkListOptions{Filters: networkFilter})
        if err != nil {
            return "", fmt.Errorf("exportService - cannot list networks - %s", err.Error())
        }
        for _, network := range networks {
            networksMap[network.Name] = network
        }

        // List and export all volumes in the stack
        // volumeFilter.Add("label", "com.docker.compose.project="+projectName)
        volumeFilter := volumeTypes.ListOptions{
					Filters: filters.NewArgs(),
				}

				volumeFilter.Filters.Add("label", "com.docker.compose.project="+projectName)
				
        volumes, err := DockerClient.VolumeList(DockerContext, volumeFilter)
        if err != nil {
            return "", fmt.Errorf("exportService - cannot list volumes - %s", err.Error())
        }
        for _, volume := range volumes.Volumes {
            volumesMap[volume.Name] = *volume
        }
    } else {
				utils.Log("exportService - not part of stack")
        // If not part of a stack, export only the specified container
        services = append(services, container)
    }

    // Convert the services, networks, and volumes to docker-compose format
    dockerCompose := make(map[string]interface{})
    dockerCompose["version"] = "3"
    composeServices := make(map[string]interface{})
    composeNetworks := make(map[string]interface{})
    composeVolumes := make(map[string]interface{})

    for _, service := range services {
        composeService := ContainerCreateRequestContainer{
					Name:         strings.TrimPrefix(service.Name, "/"),
					Image:        service.Config.Image,
					Environment:  service.Config.Env,
					Labels:       service.Config.Labels,
					Command:      strings.Join(service.Config.Cmd, " "),
					Entrypoint:   strings.Join(service.Config.Entrypoint, " "),
					WorkingDir:   service.Config.WorkingDir,
					User:         service.Config.User,
					Tty:          service.Config.Tty,
					StdinOpen:    service.Config.OpenStdin,
					Hostname:     service.Config.Hostname,
					Domainname:   service.Config.Domainname,
					MacAddress:   service.NetworkSettings.MacAddress,
					NetworkMode:  string(service.HostConfig.NetworkMode),
					StopSignal:   service.Config.StopSignal,
					HealthCheck:  ContainerCreateRequestContainerHealthcheck {
					},
					DNS:              service.HostConfig.DNS,
					DNSSearch:        service.HostConfig.DNSSearch,
					ExtraHosts:       service.HostConfig.ExtraHosts,
					SecurityOpt:      service.HostConfig.SecurityOpt,
					StorageOpt:       service.HostConfig.StorageOpt,
					Sysctls:          service.HostConfig.Sysctls,
					Isolation:        string(service.HostConfig.Isolation),
					CapAdd:           service.HostConfig.CapAdd,
					CapDrop:          service.HostConfig.CapDrop,
					SysctlsMap:       service.HostConfig.Sysctls,
					Privileged:       service.HostConfig.Privileged,
					
					// StopGracePeriod:  int(service.HostConfig.StopGracePeriod.Seconds()),
					
					// Ports
					Ports: func() []string {
							ports := []string{}
							for port, binding := range service.NetworkSettings.Ports {
									for _, b := range binding {
											ports = append(ports, fmt.Sprintf("%s:%s:%s/%s", b.HostIP, b.HostPort, port.Port(), port.Proto()))
									}
							}
							return ports
					}(),

					// Volumes
					Volumes: func() []mount.Mount {
							mounts := []mount.Mount{}
							for _, m := range service.Mounts {
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
							for netName, _ := range service.NetworkSettings.Networks {
									networks[netName] = ContainerCreateRequestServiceNetwork{
											// Aliases:     netConfig.Aliases,
											// IPV4Address: netConfig.IPAddress,
											// IPV6Address: netConfig.GlobalIPv6Address,
									}
							}
							return networks
					}(),

					DependsOn:      []string{},  // This is not directly available from inspect. It's part of docker-compose.
					RestartPolicy:  string(service.HostConfig.RestartPolicy.Name),
					Devices:        func() []string {
							var devices []string
							for _, device := range service.HostConfig.Devices {
									devices = append(devices, fmt.Sprintf("%s:%s", device.PathOnHost, device.PathInContainer))
							}
							return devices
					}(),
					Expose:         []string{},  // This information might need to be derived from other properties
				}

				// healthcheck
				if service.Config.Healthcheck != nil {
					composeService.HealthCheck.Test = service.Config.Healthcheck.Test
					composeService.HealthCheck.Interval = int(service.Config.Healthcheck.Interval.Seconds())
					composeService.HealthCheck.Timeout = int(service.Config.Healthcheck.Timeout.Seconds())
					composeService.HealthCheck.Retries = service.Config.Healthcheck.Retries
					composeService.HealthCheck.StartPeriod = int(service.Config.Healthcheck.StartPeriod.Seconds())
				}

				// user UID/GID
				if service.Config.User != "" {
					parts := strings.Split(service.Config.User, ":")
					if len(parts) == 2 {
						uid, err := strconv.Atoi(parts[0])
						if err != nil {
							composeService.UID = uid
						}
						gid, err := strconv.Atoi(parts[1])
						if err != nil {
							composeService.GID = gid
						}
					}
				}


        composeServices[strings.TrimPrefix(service.Name, "/")] = composeService
    }

    for _, network := range networksMap {
        composeNetwork := map[string]interface{}{
            "driver": network.Driver,
						"labels": network.Labels,
						"name": network.Name,
						// "internal": network.Internal,
						// "attachable": network.Attachable,
						// "enable_ipv6": network.EnableIPv6,
						// "ipam": network.IPAM,
        }
				if network.Labels != nil && network.Labels["com.docker.compose.project"] == projectName {
					composeNetwork["internal"] = network.Internal
					composeNetwork["attachable"] = network.Attachable
					composeNetwork["enable_ipv6"] = network.EnableIPv6
					composeNetwork["labels"] = network.Labels
					composeNetwork["ipam"] = network.IPAM
				} else {
					composeNetwork["external"] = true
				}
	
				composeNetworks[network.Name] = composeNetwork
    }

    for _, volume := range volumesMap {
        composeVolume := map[string]interface{}{
						"name": volume.Name,
            "driver": volume.Driver,
						"driver_opts": volume.Options,
						"labels": volume.Labels,
        }
				if volume.Labels != nil && volume.Labels["com.docker.compose.project"] == projectName {
				} else {
					composeVolume["external"] = true
				}
        composeVolumes[volume.Name] = composeVolume
    }

    dockerCompose["services"] = composeServices
    dockerCompose["networks"] = composeNetworks
    dockerCompose["volumes"] = composeVolumes

    // Marshal to YAML
    data, err := yaml.Marshal(dockerCompose)
    if err != nil {
        return "", fmt.Errorf("exportService - cannot marshal to YAML - %s", err.Error())
    }

    // Write to the file location specified in the label, if available
    if fileLocOk {
				utils.Log("exportService - file location label " + composeFileLocation)
        err = ioutil.WriteFile(composeFileLocation, data, 0644)
        if err != nil {
            return "", fmt.Errorf("exportService - cannot write to file %s - %s", composeFileLocation, err.Error())
        }
    } else {
				utils.Log("exportService - no file location label")
        // If no file location label is present, return the data as a string
        return string(data), nil
    }

    return composeFileLocation, nil
}
