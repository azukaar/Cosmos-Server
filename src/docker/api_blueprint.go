package docker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"bufio"
	"strconv"
	"os"
	"io/ioutil"
	"os/user"
	"errors"
	"reflect"
	"github.com/docker/go-connections/nat"
	"github.com/docker/go-units"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	conttype "github.com/docker/docker/api/types/container"
	doctype "github.com/docker/docker/api/types"
	strslice "github.com/docker/docker/api/types/strslice"
	volumetype "github.com/docker/docker/api/types/volume"

	"github.com/azukaar/cosmos-server/src/utils"
)

type ContainerCreateRequestServiceNetwork struct {
	Aliases []string `json:"aliases,omitempty"`
	IPV4Address string `json:"ipv4_address,omitempty"`
	IPV6Address string `json:"ipv6_address,omitempty"`
}

type ContainerCreateRequestContainerHealthcheck struct {
	Test        []string `json:"test"`
	Interval int `json:"interval"`
	Timeout int `json:"timeout"`
	Retries int `json:"retries"`
	StartPeriod int `json:"start_period"`
}

type ContainerCreateRequestContainerDependsOnCont struct {
	Condition string `json:"condition"`
	Restart string `json:"restart"`
}

type ContainerCreateRequestContainer struct {
	Name 			string            `json:"container_name"`
	Image       string            `json:"image" validate:"required"`
	Environment []string `json:"environment"`
	Labels      map[string]string `json:"labels"`
	Ports       []string          `json:"ports"`
	Volumes     []mount.Mount          `json:"volumes"`
	Networks    map[string]ContainerCreateRequestServiceNetwork `json:"networks"`
	Routes 			[]utils.ProxyRouteConfig          `json:"routes"`
	Links       []string  `json:"links,omitempty"`

	RestartPolicy  string            `json:"restart,omitempty"`
	Devices        []string          `json:"devices"`
	Expose 		     []string          `json:"expose"`
	DependsOn      map[string]ContainerCreateRequestContainerDependsOnCont `json:"depends_on,omitempty"`
	Tty            bool              `json:"tty,omitempty"`
	StdinOpen      bool              `json:"stdin_open,omitempty"`

	Command string `json:"command,omitempty"`
	Entrypoint string `json:"entrypoint,omitempty"`
	Runtime string `json:"runtime,omitempty"`
	WorkingDir string `json:"working_dir,omitempty"`
	User string `json:"user,omitempty"`
	UID int `json:"uid,omitempty"`
	GID int `json:"gid,omitempty"`
	Hostname string `json:"hostname,omitempty"`
	Domainname string `json:"domainname,omitempty"`
	MacAddress string `json:"mac_address,omitempty"`
	Privileged bool `json:"privileged,omitempty"`
	NetworkMode string `json:"network_mode,omitempty"`
	StopSignal string `json:"stop_signal,omitempty"`
	StopGracePeriod int `json:"stop_grace_period,omitempty"`
	HealthCheck ContainerCreateRequestContainerHealthcheck `json:"healthcheck,omitempty"`
	DNS []string `json:"dns,omitempty"`
	DNSSearch []string `json:"dns_search,omitempty"`
	ExtraHosts []string `json:"extra_hosts,omitempty"`
	SecurityOpt []string `json:"security_opt,omitempty"`
	StorageOpt map[string]string `json:"storage_opt,omitempty"`
	Sysctls map[string]string `json:"sysctls,omitempty"`
	Isolation string `json:"isolation,omitempty"`

	CapAdd []string `json:"cap_add,omitempty"`
	CapDrop []string `json:"cap_drop,omitempty"`

	// Resource constraints
	MemLimit string `json:"mem_limit,omitempty"`
	MemReservation string `json:"mem_reservation,omitempty"`
	CPUs float64 `json:"cpus,omitempty"`
	CPUShares int64 `json:"cpu_shares,omitempty"`
	CpusetCpus string `json:"cpuset_cpus,omitempty"`

	PostInstall []string `json:"post_install,omitempty"`
}

type ContainerCreateRequestVolume struct {
	// name must be unique
	Name string `json:"name"`
	Driver string `json:"driver"`
	Source string `json:"source"`
	Target string `json:"target"`
	Labels map[string]string `json:"labels,omitempty"`
}

type ContainerCreateRequestNetworkIPAMConfig struct {
	Subnet string `json:"subnet"`
	Gateway string `json:"gateway"`
}

type ContainerCreateRequestNetwork struct {
	// name must be unique
	Name string `json:"name"`
	Driver string `json:"driver"`
	Attachable bool `json:"attachable"`
	Internal bool `json:"internal"`
	EnableIPv6 bool `json:"enable_ipv6"`
	Labels map[string]string `json:"labels"`
	IPAM struct {
		Driver string `json:"driver"`
		Config []ContainerCreateRequestNetworkIPAMConfig `json:"config"`
	} `json:"ipam"`
}

type DockerServiceCreateRequest struct {
	Services map[string]ContainerCreateRequestContainer `json:"services"`
	Volumes map[string]ContainerCreateRequestVolume `json:"volumes"`
	Networks map[string]ContainerCreateRequestNetwork `json:"networks"`
}

type DockerServiceCreateRollback struct {
	// action: disconnect, remove, etc...
	Action string `json:"action"`
	// type: container, volume, network
	Type string `json:"type"`
	// name: container name, volume name, network name
	Name string `json:"name"`
	// was: container old settings
	Was doctype.ContainerJSON `json:"was"`
}

func Rollback(actions []DockerServiceCreateRollback , OnLog func(string)) {
	for i := len(actions) - 1; i >= 0; i-- {
		action := actions[i]
		switch action.Type {
		case "container":
			if action.Action == "remove" {
				utils.Log(fmt.Sprintf("Removing container %s...", action.Name))

				DockerClient.ContainerKill(DockerContext, action.Name, "SIGKILL")
				err := DockerClient.ContainerRemove(DockerContext, action.Name, conttype.RemoveOptions{})
		
				if err != nil {
					utils.Error("Rollback: Container", err)
					OnLog(utils.DoErr("Rollback: Container %s", err))
				} else {
					utils.Log(fmt.Sprintf("Rolled back container %s", action.Name))
					OnLog(fmt.Sprintf("Rolled back container %s\n", action.Name))
				}	
			} else if action.Action == "revert" {
				utils.Log(fmt.Sprintf("Reverting container %s...", action.Name))

				// Edit Container
				_, err := EditContainer(action.Name, action.Was, false)
	
				if err != nil {
					utils.Error("Rollback: Container", err)
					OnLog(utils.DoErr("Rollback: Container %s", err))
				} else {
					utils.Log(fmt.Sprintf("Rolled back container %s", action.Name))
					OnLog(fmt.Sprintf("Rolled back container %s\n", action.Name))
				}	
			} else if action.Action == "restore" {
				utils.Log(fmt.Sprintf("Restoring container %s...", action.Name))

				// Edit Container
				_, err := EditContainer("", action.Was, true)
	
				if err != nil {
					utils.Error("Rollback: Container", err)
					OnLog(utils.DoErr("Rollback: Container %s", err))
				} else {
					utils.Log(fmt.Sprintf("Rolled back container %s", action.Name))
					OnLog(fmt.Sprintf("Rolled back container %s\n", action.Name))
				}	
			}
		case "volume":
			utils.Log(fmt.Sprintf("Removing volume %s...", action.Name))

			err := DockerClient.VolumeRemove(DockerContext, action.Name, true)
			if err != nil {
				utils.Error("Rollback: Volume", err)
				OnLog(utils.DoErr("Rollback: Volume %s", err))
			} else {
				utils.Log(fmt.Sprintf("Rolled back volume %s", action.Name))
				OnLog(fmt.Sprintf("Rolled back volume %s\n", action.Name))
			}
		case "network":
			utils.Log(fmt.Sprintf("Removing network %s...", action.Name))

			if utils.IsInsideContainer {
				DockerClient.NetworkDisconnect(DockerContext, action.Name, os.Getenv("HOSTNAME"), true)
			}
			err := DockerClient.NetworkRemove(DockerContext, action.Name)
			if err != nil {
				utils.Error("Rollback: Network", err)
				OnLog(utils.DoErr("Rollback: Network %s", err))
			} else {
				utils.Log(fmt.Sprintf("Rolled back network %s", action.Name))
				OnLog(fmt.Sprintf("Rolled back network %s\n", action.Name))
			}
		}
	}
	
	// After all operations
	utils.Error("CreateService", fmt.Errorf("Operation failed. Changes have been rolled back."))
	OnLog("[OPERATION FAILED]. CHANGES HAVE BEEN ROLLEDBACK.\n")
}

// CreateServiceRoute godoc
// @Summary Create a Docker service (compose-like) with networks, volumes, and containers
// @Tags docker
// @Accept json
// @Produce plain
// @Param body body DockerServiceCreateRequest true "Service creation payload"
// @Security BearerAuth
// @Success 200 {string} string "Streamed creation output"
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/docker-service [post]
func CreateServiceRoute(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_RESOURCES) != nil {
		return
	}

	errD := Connect()
	if errD != nil {
		utils.Error("CreateService - connect - ", errD)
		utils.HTTPError(w, "Internal server error: " + errD.Error(), http.StatusInternalServerError, "DS002")
		return
	}

	if req.Method == "POST" {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Transfer-Encoding", "chunked")
		
		flusher, ok := w.(http.Flusher)
		if !ok {
				http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
				return 
		}

		decoder := json.NewDecoder(req.Body)
		var serviceRequest DockerServiceCreateRequest
		err := decoder.Decode(&serviceRequest)
		if err != nil {
			utils.Error("CreateService - decode - ", err)
			fmt.Fprintf(w, "[OPERATION FAILED] Bad request: "+err.Error(), http.StatusBadRequest, "DS003")
			flusher.Flush()
			utils.HTTPError(w, "Bad request: " + err.Error(), http.StatusBadRequest, "DS003")
			return
		}

		CreateService(serviceRequest, 
			func (msg string) {
				fmt.Fprintf(w, "%s", msg)
				flusher.Flush()
			},
		)
	} else {
		utils.Error("CreateService: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

// generatePorts is a helper function to generate a slice of ports from a string range.
func generatePorts(portRangeStr string) []string {
	portsStr := strings.Split(portRangeStr, "-")
	if len(portsStr) != 2 {
		return []string{
			portsStr[0],
		}
	}
	start, _ := strconv.Atoi(portsStr[0])
	end, _ := strconv.Atoi(portsStr[1])

	ports := make([]string, end-start+1)
	for i := range ports {
		ports[i] = strconv.Itoa(start + i)
	}

	return ports
}

func CreateService(serviceRequest DockerServiceCreateRequest, OnLog func(string)) error {
	utils.ConfigLock.Lock()
	defer utils.ConfigLock.Unlock()
	
	utils.Log("Starting creation of new service...")
	OnLog("Starting creation of new service...\n")
	
	needsHTTPRestart := false
	config := utils.ReadConfigFromFile()
	configRoutes := config.HTTPConfig.ProxyConfig.Routes

	var rollbackActions []DockerServiceCreateRollback
	var err error

	// Create networks
	for networkToCreateName, networkToCreate := range serviceRequest.Networks {
		utils.Log(fmt.Sprintf("Creating network %s...", networkToCreateName))
		OnLog(fmt.Sprintf("Creating network %s...\n", networkToCreateName))

		// check if network already exists
		exNetworkDef, err := DockerClient.NetworkInspect(DockerContext, networkToCreateName, doctype.NetworkInspectOptions{})

		if err == nil {
			if networkToCreate.Driver == "" {
				networkToCreate.Driver = "bridge"
			}

			if (exNetworkDef.Driver != networkToCreate.Driver) {
				utils.Error("CreateService: Network", err)
				OnLog(utils.DoErr("Network %s already exists with incompatible settings, cannot merge new network into it.\n", networkToCreateName))
				Rollback(rollbackActions, OnLog)
				return err
			} else {
				utils.Warn(fmt.Sprintf("Network %s already exists, skipping creation", networkToCreateName))
				OnLog(utils.DoWarn("Network %s already exists, skipping creation\n", networkToCreateName))
				continue
			}
		}

		ipamConfig := make([]network.IPAMConfig, len(networkToCreate.IPAM.Config))
		if networkToCreate.IPAM.Config != nil {
			for i, config := range networkToCreate.IPAM.Config {
					ipamConfig[i] = network.IPAMConfig{
							Subnet: config.Subnet,
					}
			}
		}
		
		networkPayload := doctype.NetworkCreate{
			Driver:     networkToCreate.Driver,
			Attachable: networkToCreate.Attachable,
			Internal:   networkToCreate.Internal,
			EnableIPv6: networkToCreate.EnableIPv6,
			Labels:     networkToCreate.Labels,
			IPAM: &network.IPAM{
					Driver: networkToCreate.IPAM.Driver,
					Config: ipamConfig,
			},
		}

		_, err = CreateReasonableNetwork(networkToCreateName, networkPayload)

		if err != nil {
			utils.Error("CreateService: Rolling back changes because of -- Network", err)
			OnLog(utils.DoErr("Rolling back changes because of -- Network creation error: %s\n", err.Error()))
			Rollback(rollbackActions, OnLog)
			return err
		}

		rollbackActions = append(rollbackActions, DockerServiceCreateRollback{
			Action: "remove",
			Type:   "network",
			Name:   networkToCreateName,
		})
	
		// Write a response to the client
		utils.Log(fmt.Sprintf("Network %s created", networkToCreateName))
		OnLog(fmt.Sprintf("Network %s created\n", networkToCreateName))
	}

	// Create volumes
	for volumeName, volume := range serviceRequest.Volumes {
		if volume.Name == "" {
			volume.Name = volumeName
		}

		// check if volume already exists
		_, err := DockerClient.VolumeInspect(DockerContext, volume.Name)
		if err == nil {
			utils.Warn(fmt.Sprintf("Volume %s already exists, skipping creation", volume.Name))
			OnLog(utils.DoWarn("Volume %s already exists, skipping creation\n", volume.Name))
			continue
		}

		utils.Log(fmt.Sprintf("Creating volume %s...", volume.Name))
		OnLog(fmt.Sprintf("Creating volume %s...\n", volume.Name))
		
		_, err = DockerClient.VolumeCreate(DockerContext, volumetype.CreateOptions{
			Driver:     volume.Driver,
			Name:       volume.Name,
			Labels:     volume.Labels,
		})

		if err != nil {
			utils.Error("CreateService: Rolling back changes because of -- Volume", err)
			OnLog(utils.DoErr("Rolling back changes because of -- Volume creation error: %s\n", err.Error()))
			Rollback(rollbackActions, OnLog)
			return err
		}

		rollbackActions = append(rollbackActions, DockerServiceCreateRollback{
			Action: "remove",
			Type:   "volume",
			Name:   volume.Name,
		})

		// Write a response to the client
		utils.Log(fmt.Sprintf("Volume %s created", volume.Name))
		OnLog(fmt.Sprintf("Volume %s created\n", volume.Name))
	}

	// pull images
	for _, container := range serviceRequest.Services {
		// Write a response to the client
		utils.Log(fmt.Sprintf("Pulling image %s", container.Image))
		OnLog(fmt.Sprintf("Pulling image %s\n", container.Image))

		out, err := DockerPullImage(container.Image)
		if err != nil {
			utils.Error("CreateService: Rolling back changes because of -- Image pull", err)
			OnLog(utils.DoErr("Rolling back changes because of -- Image pull error: %s\n", err.Error()))
			Rollback(rollbackActions, OnLog)
			return err
		}
		defer out.Close()

		// wait for image pull to finish
		scanner := bufio.NewScanner(out)
		for scanner.Scan() {
			OnLog(fmt.Sprintf("%s\n", scanner.Text()))
		}
		
		// Write a response to the client
		utils.Log(fmt.Sprintf("Image %s pulled", container.Image))
		OnLog(fmt.Sprintf("Image %s pulled\n", container.Image))
	}


	// Create containers
	tempServiceList := make(map[string]ContainerCreateRequestContainer)
	for serviceName, container := range serviceRequest.Services {
		utils.Log(fmt.Sprintf("Checking service %s...", serviceName))
		OnLog(fmt.Sprintf("Checking service %s...\n", serviceName))

		// Default the container name to the compose service key when no explicit
		// container_name is given (standard docker-compose behavior). Cluster
		// deployments routinely omit container_name. Without this, ContainerCreate
		// auto-generates a random name but the rest of the flow (start, rollback,
		// network connect) keeps using the empty container.Name, so ContainerStart
		// hits /containers//start and the daemon returns "page not found".
		if container.Name == "" {
			container.Name = serviceName
		}

		// If container request a Cosmos network, create and attach it
		if strings.ToLower(container.Labels["cosmos-network-name"]) == "auto" {
			utils.Log(fmt.Sprintf("Forcing secure %s...", serviceName))
			OnLog(fmt.Sprintf("Forcing secure %s...\n", serviceName))
	
			newNetwork, errNC := CreateCosmosNetwork(serviceName)
			if errNC != nil {
				utils.Error("CreateService: Network", err)
				OnLog(utils.DoErr("Network %s cant be created\n", newNetwork))
				Rollback(rollbackActions, OnLog)
				return err
			}

			container.Labels["cosmos-network-name"] = newNetwork

			AttachNetworkToCosmos(newNetwork)

			if container.Networks == nil {
				container.Networks = make(map[string]ContainerCreateRequestServiceNetwork)
			}

			container.Networks[newNetwork] = ContainerCreateRequestServiceNetwork{}

			rollbackActions = append(rollbackActions, DockerServiceCreateRollback{
				Action: "remove",
				Type:   "network",
				Name:   newNetwork,
			})
			
			utils.Log(fmt.Sprintf("Created secure network %s", newNetwork))
			OnLog(fmt.Sprintf("Created secure network %s\n", newNetwork))
		} else if container.Labels["cosmos-network-name"] != "" {
			// Container has a declared a Cosmos network, check if it exists and connect to it
			utils.Log(fmt.Sprintf("Checking declared network %s...", container.Labels["cosmos-network-name"]))
			OnLog(fmt.Sprintf("Checking declared network %s...\n", container.Labels["cosmos-network-name"]))

			_, err := DockerClient.NetworkInspect(DockerContext, container.Labels["cosmos-network-name"], doctype.NetworkInspectOptions{})
			if err == nil {
				utils.Log(fmt.Sprintf("Connecting to declared network %s...", container.Labels["cosmos-network-name"]))
				OnLog(fmt.Sprintf("Connecting to declared network %s...\n", container.Labels["cosmos-network-name"]))
	
				AttachNetworkToCosmos(container.Labels["cosmos-network-name"])
			}
		}

		utils.Log(fmt.Sprintf("Creating container %s...", container.Name))
		OnLog(fmt.Sprintf("Creating container %s...\n", container.Name))

		containerConfig := &conttype.Config{
			Image:        container.Image,
			Env:          container.Environment,
			Labels:       container.Labels,
			ExposedPorts: nat.PortSet{},
			WorkingDir:   container.WorkingDir,
			User:         container.User,
			Hostname:     container.Hostname,
			Domainname:   container.Domainname,
			MacAddress:   container.MacAddress,
			StopSignal:   container.StopSignal,
			StopTimeout:  &container.StopGracePeriod,
			Tty:          container.Tty,
			OpenStdin:    container.StdinOpen,
		}

		// Tag multi-service compose stacks with cosmos.stack (+ .main) so the
		// UI groups them and the dependents cascade can scope by stack. Single
		// services stay untagged.
		if len(serviceRequest.Services) > 1 {
			if containerConfig.Labels == nil {
				containerConfig.Labels = make(map[string]string)
			}
			if containerConfig.Labels["cosmos.stack"] == "" && containerConfig.Labels["com.docker.compose.project"] == "" {
				containerConfig.Labels["cosmos.stack"] = serviceName
			}
			// mark the first service as the stack main
			first := ""
			for k := range serviceRequest.Services {
				first = k
				break
			}
			if containerConfig.Labels["cosmos.stack.main"] == "" && serviceName == first {
				containerConfig.Labels["cosmos.stack.main"] = "true"
			}
		}

		// Persist depends_on in compose's com.docker.compose.depends_on label
		// so runtime restart/recreate paths and compose-imported stacks work.
		if len(container.DependsOn) > 0 {
			deps := make(map[string]dependsOnEntry, len(container.DependsOn))
			for depName, depCfg := range container.DependsOn {
				cond := depCfg.Condition
				if cond == "" {
					cond = DepConditionStarted
				}
				deps[depName] = dependsOnEntry{
					Condition: cond,
					Restart:   depCfg.Restart == "true" || depCfg.Restart == "always",
				}
			}
			SetDependsOnLabels(containerConfig, deps)
		}

		// check if there's an empty TZ env, if so, replace it with the host's TZ
		if containerConfig.Env != nil {
			for i, env := range containerConfig.Env {
				if strings.HasPrefix(env, "TZ=") {
					if strings.TrimPrefix(env, "TZ=") == "auto" {
						if os.Getenv("TZ") != "" {
							containerConfig.Env[i] = "TZ=" + os.Getenv("TZ")
						} else {
							containerConfig.Env = append(containerConfig.Env[:i], containerConfig.Env[i+1:]...)
						}
					}
				}
			}
		}
		
		if container.Command != "" {
			containerConfig.Cmd = strings.Fields(container.Command)
		}

		if container.Entrypoint != "" {
			containerConfig.Entrypoint = strslice.StrSlice(strings.Fields(container.Entrypoint))
		}

		// For Expose / Ports
		
		for _, expose := range container.Expose {
			exposePort := nat.Port(expose)
			containerConfig.ExposedPorts[exposePort] = struct{}{}
		}

		PortBindings := nat.PortMap{}
		finalPorts := []string{}

		for _, portRaw := range container.Ports {
			portStuff := strings.Split(portRaw, "/")

			if len(portStuff) == 1 {
				portStuff = append(portStuff, "tcp")
			}

			port, protocol := portStuff[0], portStuff[1]
			
			hostPorts := []string{}
			containerPorts := []string{}

			ports := strings.Split(port, ":")

			hostPorts = generatePorts(ports[len(ports)-2])
			containerPorts = generatePorts(ports[len(ports)-1])

			// compose "ip:hostPort:containerPort": everything before the last two segments is the bind address
			ipExposed := ""
			if len(ports) > 2 {
				ipExposed = strings.Join(ports[0:len(ports)-2], ":")
			}

			for i := 0; i < utils.Max(len(hostPorts), len(containerPorts)); i++ {
				hostPort := hostPorts[i%len(hostPorts)]
				containerPort := containerPorts[i%len(containerPorts)]
				
				finalPorts = append(finalPorts, fmt.Sprintf("%s@%s:%s/%s", ipExposed, hostPort, containerPort, protocol))
			}
		}

		utils.Debug(fmt.Sprintf("Final ports: %s", finalPorts))
		
		hostPortsBound := make(map[string]bool)

		for _, portRaw := range finalPorts {
			portStuff := strings.Split(portRaw, "/")
			ipport := strings.Split(portStuff[0], "@")

			ipdest := ""
			portdef := ipport[0]
			if len(ipport) > 1 {
				portdef = ipport[1]
				ipdest = ipport[0]
			}

			port, protocol := portdef, portStuff[1]

			if protocol == "" {
				protocol = "tcp"
			}

			nextPort := strings.Split(port, ":")
			hostPort, contPort := nextPort[0], nextPort[1]

			contPort = contPort + "/" + protocol
			
			if hostPortsBound[hostPort + "/" + protocol] {
				utils.Warn("Port " + hostPort + " already bound, skipping")
				continue
			}

			// Get the existing bindings for this container port, if any
			bindings := PortBindings[nat.Port(contPort)]

			// Append a new PortBinding to the slice of bindings
			pb := nat.PortBinding {
				HostPort: hostPort,
			}

			if ipdest != "" {
				pb.HostIP = ipdest
			}

			bindings = append(bindings, pb)

			// Update the port bindings for this container port
			PortBindings[nat.Port(contPort)] = bindings

			// Mark the container port as exposed
			containerConfig.ExposedPorts[nat.Port(contPort)] = struct{}{}

			hostPortsBound[hostPort + "/" + protocol] = true
		}

		// Create missing folders for bind mounts
		for _, newmount := range container.Volumes {
			if newmount.Type == mount.TypeBind {
				newSource := newmount.Source

				if utils.IsInsideContainer {
					if _, err := os.Stat("/mnt/host"); os.IsNotExist(err) {
						utils.Error("CreateService: Unable to create directory for bind mount in the host directory. Please mount the host / in Cosmos with  -v /:/mnt/host to enable folder creations, or create the bind folder yourself", err)
						OnLog(utils.DoErr("Unable to create directory for bind mount in the host directory. Please mount the host / in Cosmos with  -v /:/mnt/host to enable folder creations, or create the bind folder yourself: %s\n", err.Error()))
					
						continue
					} else {
						newSource = "/mnt/host" + newSource
					}
				}
						
				utils.Log(fmt.Sprintf("Checking directory %s for bind mount", newSource))
				OnLog(fmt.Sprintf("Checking directory %s for bind mount\n", newSource))

				if _, err := os.Stat(newSource); os.IsNotExist(err) {
					utils.Log(fmt.Sprintf("Not found. Creating directory %s for bind mount", newSource))
					OnLog(fmt.Sprintf("Not found. Creating directory %s for bind mount\n", newSource))
	
					err := os.MkdirAll(newSource, 0750)

					if err != nil {
						utils.Error("CreateService: Unable to create directory for bind mount. Make sure parent directories exist, and that Cosmos has permissions to create directories in the host directory", err)
						OnLog(utils.DoErr("Unable to create directory for bind mount. Make sure parent directories exist, and that Cosmos has permissions to create directories in the host directory: %s\n", err.Error()))
						Rollback(rollbackActions, OnLog)
						return err
					}

					if container.UID != 0 {
						// Change the ownership of the directory to the container.UID
						err = os.Chown(newSource, container.UID, container.GID)
						if err != nil {
							utils.Error("CreateService: Unable to change ownership of directory", err)
							OnLog(utils.DoErr("%s", "Unable to change ownership of directory: " + err.Error()))
						}
					} else if container.User != "" && strings.Contains(container.User, ":") { 
						uidgid := strings.Split(container.User, ":")
						uid, _ := strconv.Atoi(uidgid[0])
						gid, _ := strconv.Atoi(uidgid[1])
						err = os.Chown(newSource, uid, gid)
						if err != nil {
							utils.Error("CreateService: Unable to change ownership of directory", err)
							OnLog(utils.DoErr("%s", "Unable to change ownership of directory: " + err.Error()))
						}
					} else if container.User != "" {
						// Change the ownership of the directory to the container.User
						userInfo, err := user.Lookup(container.User)
						if err != nil {
							utils.Error("CreateService: Unable to lookup user", err)
							OnLog(utils.DoErr("%s", "Unable to lookup user " + container.User + ". " +err.Error()))
						} else {
							uid, _ := strconv.Atoi(userInfo.Uid)
							gid, _ := strconv.Atoi(userInfo.Gid)
							err = os.Chown(newSource, uid, gid)
							if err != nil {
								utils.Error("CreateService: Unable to change ownership of directory", err)
								OnLog(utils.DoErr("%s", "Unable to change ownership of directory: " + err.Error()))
							}
						}	
					}
				}
			}
		}

		// Parse resource constraints
		var memLimit, memReservation int64
		if container.MemLimit != "" {
			memLimit, err = units.RAMInBytes(container.MemLimit)
			if err != nil {
				utils.Error("CreateService: Invalid mem_limit", err)
				OnLog(utils.DoErr("Invalid mem_limit value: %s\n", err.Error()))
				Rollback(rollbackActions, OnLog)
				return err
			}
		}
		if container.MemReservation != "" {
			memReservation, err = units.RAMInBytes(container.MemReservation)
			if err != nil {
				utils.Error("CreateService: Invalid mem_reservation", err)
				OnLog(utils.DoErr("Invalid mem_reservation value: %s\n", err.Error()))
				Rollback(rollbackActions, OnLog)
				return err
			}
		}

		hostConfig := &conttype.HostConfig{
			PortBindings: PortBindings,
			Mounts:       container.Volumes,
			RestartPolicy: conttype.RestartPolicy{
				Name: conttype.RestartPolicyMode(container.RestartPolicy),
			},
			Privileged:   container.Privileged,
			NetworkMode:  conttype.NetworkMode(container.NetworkMode),
			DNS:         container.DNS,
			DNSSearch:   container.DNSSearch,
			ExtraHosts:  container.ExtraHosts,
			SecurityOpt: container.SecurityOpt,
			StorageOpt:  container.StorageOpt,
			Sysctls:     container.Sysctls,
			Isolation:   conttype.Isolation(container.Isolation),
			CapAdd:      container.CapAdd,
			CapDrop:     container.CapDrop,
			Resources: conttype.Resources{
				Memory:            memLimit,
				MemoryReservation: memReservation,
				NanoCPUs:          int64(container.CPUs * 1e9),
				CPUShares:         container.CPUShares,
				CpusetCpus:        container.CpusetCpus,
			},
		}

		// normalize container/service refs to stable container:<name>
		if containerConfig.Labels["cosmos-force-network-mode"] == "" {
			if NetworkModeContainerRef(string(hostConfig.NetworkMode)) || NetworkModeServiceRef(string(hostConfig.NetworkMode)) {
				normalized := ContainerRefToName(string(hostConfig.NetworkMode))
				containerConfig.Labels["cosmos-force-network-mode"] = normalized
				if normalized != string(hostConfig.NetworkMode) {
					hostConfig.NetworkMode = conttype.NetworkMode(normalized)
				}
			}
		} else {
			normalized := ContainerRefToName(containerConfig.Labels["cosmos-force-network-mode"])
			hostConfig.NetworkMode = conttype.NetworkMode(normalized)
			containerConfig.Labels["cosmos-force-network-mode"] = normalized
			utils.Debug("Forcing network mode to " + string(hostConfig.NetworkMode))
		}


		// docker must not auto-restart a container Cosmos put to sleep
		if IsLazyLabels(containerConfig.Labels) {
			hostConfig.RestartPolicy = conttype.RestartPolicy{Name: conttype.RestartPolicyMode("no")}
		}

		if container.Runtime != "" {
			hostConfig.Runtime = strings.Join(strings.Fields(container.Runtime), " ")
		}		

		// For Healthcheck
		if len(container.HealthCheck.Test) > 0 {
			containerConfig.Healthcheck = &conttype.HealthConfig{
				Test: container.HealthCheck.Test,
				Interval: time.Duration(container.HealthCheck.Interval) * time.Second,
				Timeout: time.Duration(container.HealthCheck.Timeout) * time.Second,
				StartPeriod: time.Duration(container.HealthCheck.StartPeriod) * time.Second,
				Retries: container.HealthCheck.Retries,
			}
		}

		// For Devices
		devices := []conttype.DeviceMapping{}

		for _, device := range container.Devices {
			deviceSplit := strings.Split(device, ":")
			devices = append(devices, conttype.DeviceMapping{
				PathOnHost:        deviceSplit[0],
				PathInContainer:   deviceSplit[1],
				CgroupPermissions: "rwm", // This can be "r", "w", "m", or any combination
			})
		}

		hostConfig.Devices = devices

		networkingConfig := &network.NetworkingConfig{
			EndpointsConfig: make(map[string]*network.EndpointSettings),
		}

		// check if container exist
		existingContainer, err := DockerClient.ContainerInspect(DockerContext, container.Name)
		if err == nil {		
			
			// Edit Container
			oldConfig := doctype.ContainerJSON{}
			oldConfig.ContainerJSONBase = new(doctype.ContainerJSONBase)
			oldConfig.Config = existingContainer.Config
			oldConfig.HostConfig = existingContainer.HostConfig
			oldConfig.Name = existingContainer.Name
			oldConfig.NetworkSettings = existingContainer.NetworkSettings

			utils.Warn("CreateService: Container " + container.Name + " already exist, overwriting.")
			OnLog(utils.DoWarn("%s", "Container " + container.Name + " already exist, overwriting.\n"))
	
			// stop the container 
			utils.Log("CreateService: Stopping container: " + container.Name)
			OnLog("Stopping container: " + container.Name + "\n")
			err = DockerClient.ContainerStop(DockerContext, container.Name, conttype.StopOptions{})
			if err != nil {
				utils.Error("CreateService: Rolling back changes because of -- Container", err)
				OnLog(utils.DoErr("%s", "Rolling back changes because of -- Container creation error: "+err.Error()))
				Rollback(rollbackActions, OnLog)
				return err
			}

			// remove the container
			utils.Log("CreateService: Removing container: " + container.Name)
			OnLog("Removing container: " + container.Name + "\n")
			err = DockerClient.ContainerRemove(DockerContext, container.Name, conttype.RemoveOptions{})
			if err != nil {
				utils.Error("CreateService: Rolling back changes because of -- Container", err)
				OnLog(utils.DoErr("%s", "Rolling back changes because of -- Container creation error: "+err.Error()))
				Rollback(rollbackActions, OnLog)
				return err
			}

			// check if there are persistent env var
			if containerConfig.Labels["cosmos-persistent-env"] != "" {
				// split env vars
				envVars := strings.Split(containerConfig.Labels["cosmos-persistent-env"], ",")
				// get existing env vars
				existingEnvVars := existingContainer.Config.Env
				// loop through env vars
				for _, envVar := range envVars {
					envVar = strings.TrimSpace(envVar)
					
					// check if env var already exist
					exists := false
					existingEnvVarValue := ""
					for _, existingEnvVar := range existingEnvVars {
						if strings.HasPrefix(existingEnvVar, envVar + "=") {
							exists = true
							existingEnvVarValue = existingEnvVar
							break
						}
					}
					// if it exist, copy value to new container
					if exists {
						wasReplace := false
						for i, newEnvVar := range containerConfig.Env {
							if strings.HasPrefix(newEnvVar, envVar + "=") {
								containerConfig.Env[i] = envVar + "=" + strings.TrimPrefix(existingEnvVarValue, envVar + "=")
								wasReplace = true
								break
							}
						}
						if !wasReplace {
							containerConfig.Env = append(containerConfig.Env, envVar + "=" + strings.TrimPrefix(existingEnvVarValue, envVar + "="))
						}
					}
				}
			}
			
			rollbackActions = append(rollbackActions, DockerServiceCreateRollback{
				Action: "restore",
				Type:   "container",
				Was: oldConfig,
			})
		}
		_, err = DockerClient.ContainerCreate(DockerContext, containerConfig, hostConfig, networkingConfig, nil, container.Name)

		if err != nil {
			utils.Error("CreateService: Rolling back changes because of -- Container", err)
			OnLog(utils.DoErr("%s", "Rolling back changes because of -- Container creation error: "+err.Error()))
			Rollback(rollbackActions, OnLog)
			return err
		}
	
		rollbackActions = append(rollbackActions, DockerServiceCreateRollback{
			Action: "remove",
			Type:   "container",
			Name:   container.Name,
		})
	

		// connect to networks
		for netName, netConfig := range container.Networks {
			utils.Log("CreateService: Connecting to network: " + netName)
			err = DockerClient.NetworkConnect(DockerContext, netName, container.Name, &network.EndpointSettings{
				Aliases:     netConfig.Aliases,
				IPAddress:   netConfig.IPV4Address,
				GlobalIPv6Address: netConfig.IPV6Address,
			})
			if err != nil && !strings.Contains(err.Error(), "already exists in network") {
				utils.Error("CreateService: Rolling back changes because of -- Network Connection -- ", err)
				OnLog(utils.DoErr("%s", "Rolling back changes because of -- Network connection error: "+err.Error()))
				Rollback(rollbackActions, OnLog)
				return err
			} else if err != nil && strings.Contains(err.Error(), "already exists in network") {
				utils.Warn("CreateService: Container " + container.Name + " already connected to network " + netName + ", skipping.")
				OnLog(utils.DoWarn("Container %s already connected to network %s, skipping.", container.Name, netName))			
			}
		}

		// add routes 
		for _, route := range container.Routes {
			// check if route already exists
			exists := false
			existsAt := 0
			for destRouteIndex, configRoute := range configRoutes {
				if configRoute.Name == route.Name {
					exists = true
					existsAt = destRouteIndex
					break
				}
			}

			if !exists {
				needsHTTPRestart = true
				configRoutes = append([]utils.ProxyRouteConfig{(utils.ProxyRouteConfig)(route)}, configRoutes...)
			} else {
				// utils.Error("CreateService: Rolling back changes because of -- Route already exist", nil)
				// OnLog(utils.DoErr("Rolling back changes because of -- Route already exist"))
				// Rollback(rollbackActions, OnLog)
				// return errors.New("Route already exist")

				//overwrite route
				if !reflect.DeepEqual(configRoutes[existsAt], (utils.ProxyRouteConfig)(route)) {
					needsHTTPRestart = true
				}
				configRoutes[existsAt] = (utils.ProxyRouteConfig)(route)
				utils.Warn("CreateService: Route " + route.Name + " already exist, overwriting.")
				OnLog(utils.DoWarn("%s", "Route " + route.Name + " already exist, overwriting.\n"))
			}
		}
		

		// Create the networks for links
		for _, targetContainer := range container.Links {
			if targetContainer == "" {
				continue
			}
			
			if strings.Contains(targetContainer, ":") {
				err = errors.New("Link network cannot contain ':' please use container name only")
				utils.Error("CreateService: Rolling back changes because of -- Link network", err)
				OnLog(utils.DoErr("%s", "Rolling back changes because of -- Link network creation error: "+err.Error()))
				Rollback(rollbackActions, OnLog)
				return err
			}

			err = CreateLinkNetwork(container.Name, targetContainer)
			if err != nil {
				utils.Error("CreateService: Rolling back changes because of -- Link network", err)
				OnLog(utils.DoErr("%s", "Rolling back changes because of -- Link network creation error: "+err.Error()))
				Rollback(rollbackActions, OnLog)
				return err
			}
		}
		
		// Write a response to the client
		utils.Log(fmt.Sprintf("Container %s created", container.Name))
		OnLog(fmt.Sprintf("Container %s created", container.Name))

		tempServiceList[serviceName] = ContainerCreateRequestContainer{
			Name:        container.Name,
			DependsOn:   container.DependsOn,
			NetworkMode: string(hostConfig.NetworkMode),
			PostInstall: container.PostInstall,
		}
	}

	// re-order containers dpeneding on depends_on
	startOrder, mustStart, err := ReOrderServices(tempServiceList)
	if err != nil {
		utils.Error("CreateService: Rolling back changes because of -- Container", err)
		OnLog(utils.DoErr("%s", "Rolling back changes because of -- Container creation error: "+err.Error()))
		Rollback(rollbackActions, OnLog)
		return err
	}

	// Start containers in dependency order, waiting for non-started
	// conditions (service_healthy / service_completed_successfully).
	for _, container := range startOrder {
		if len(container.DependsOn) > 0 {
			for depName, depCfg := range container.DependsOn {
				cond := depCfg.Condition
				if cond == "" {
					cond = DepConditionStarted
				}
				utils.Log(fmt.Sprintf("Waiting for dependency %s (%s) before starting %s", depName, cond, container.Name))
				OnLog(fmt.Sprintf("Waiting for dependency %s (%s) before starting %s\n", depName, cond, container.Name))
				if err := WaitForDepCondition(DockerContext, depName, cond); err != nil {
					utils.Error("CreateService: Start Container", err)
					OnLog(utils.DoErr("Rolling back changes because of -- dependency wait error: "+err.Error()))
					Rollback(rollbackActions, OnLog)
					return err
				}
			}
		}

		err = DockerClient.ContainerStart(DockerContext, container.Name, conttype.StartOptions{})
		if err != nil {
			utils.Error("CreateService: Start Container", err)
			OnLog(utils.DoErr("%s", "Rolling back changes because of -- Container start error" + container.Name + " : "+err.Error()))
			Rollback(rollbackActions, OnLog)
			return err
		}

		// Write a response to the client
		utils.Log(fmt.Sprintf("Container %s initiated", container.Name))
		OnLog(fmt.Sprintf("Container %s initiated", container.Name))

		utils.Log(fmt.Sprintf("Waiting for container %s to start", container.Name))
		OnLog(fmt.Sprintf("Waiting for container %s to start", container.Name))

		if len(container.PostInstall) > 0 || mustStart {
			// wait for container to start
			retries := 0
			for {
				time.Sleep(1 * time.Second)
				inspect, _ := DockerClient.ContainerInspect(DockerContext, container.Name)
				if inspect.State.Running {
					break
				}

				retries++

				if retries > 30 {
					utils.Error("CreateService: Start Container", fmt.Errorf("Container %s did not start", container.Name))
					OnLog(utils.DoErr("%s", "Rolling back changes because of -- Container start error" + container.Name + " : Container did not start"))
					Rollback(rollbackActions, OnLog)
					return fmt.Errorf("Container %s did not start", container.Name)
				}
			}
		}

		time.Sleep(1 * time.Second)

		// if post install
		if len(container.PostInstall) > 0 {
			utils.Log(fmt.Sprintf("Container %s started. Running %d post install commands...", container.Name, len(container.PostInstall)))
			OnLog(fmt.Sprintf("Container %s started. Running %d post install commands...\n", container.Name, len(container.PostInstall)))
			
			// run post install commands
			for _, cmd := range container.PostInstall {
				utils.Log(fmt.Sprintf("Running post install command: %s", cmd))
				OnLog(fmt.Sprintf("Running post install command: %s", cmd))
			
				// setup the execution of command
				execResponse, err := DockerClient.ContainerExecCreate(DockerContext, container.Name, doctype.ExecConfig{
					Cmd:          []string{"/bin/sh", "-c", cmd},
					AttachStdout: true,
					AttachStderr: true,
				})
			
				if err != nil {
					utils.Error("CreateService: Post Install", err)
					OnLog(utils.DoErr("%s", "Rolling back changes because of -- Post install error: "+err.Error()))
					Rollback(rollbackActions, OnLog)
					return err
				}
			
				// attach to the exec instance
				response, err := DockerClient.ContainerExecAttach(DockerContext, execResponse.ID, doctype.ExecStartCheck{})
				if err != nil {
					utils.Error("CreateService: Post Install", err)
					OnLog(utils.DoErr("%s", "Rolling back changes because of -- Post install error: "+err.Error()))
					Rollback(rollbackActions, OnLog)
					return err
				}
				// read the output
				out, _ := ioutil.ReadAll(response.Reader)
				response.Close()
				OnLog(fmt.Sprintf("----> %s", out))
			}

			// restart container
			DockerClient.ContainerRestart(DockerContext, container.Name, conttype.StopOptions{})
		}
		
	}
	
	// Save the route configs 
	config.HTTPConfig.ProxyConfig.Routes = configRoutes
	utils.SaveConfigTofile(config)
	
	if needsHTTPRestart {
		utils.RestartHTTPServer()
	}

	// After all operations
	utils.Log("CreateService: Operation succeeded. SERVICE STARTED")
	OnLog("\n")
	OnLog(utils.DoSuccess("[OPERATION SUCCEEDED]. SERVICE STARTED\n"))

	servicesNames := []string{}
	for _, service := range serviceRequest.Services {
		servicesNames = append(servicesNames, service.Name)
	}

	utils.TriggerEvent(
		"cosmos.docker.compose.create",
		"Service created",
		"success",
		"",
		map[string]interface{}{
			"services": servicesNames,
	})

	return nil
}

func ReOrderServices(serviceMap map[string]ContainerCreateRequestContainer) ([]ContainerCreateRequestContainer, bool, error) {
	startOrder := []ContainerCreateRequestContainer{}
	mustStart := false

	// depends_on keys are compose service names; startOrder is keyed by
	// container name. network_mode is NOT a depends_on edge in compose: it is
	// resolved at create and Docker enforces the target exists, so we only add
	// a soft in-batch ordering constraint (external targets never hard-error).
	nameByService := map[string]string{}
	for key, svc := range serviceMap {
		// container.Name defaults to the service key when container_name is
		// unset (see CreateService), so this mapping is usually identity.
		nameByService[key] = svc.Name
	}

	// also allow matching by container name
	serviceByName := map[string]string{}
	for key, svc := range serviceMap {
		serviceByName[svc.Name] = key
	}

	// Resolve a dependency reference to a container name. Dependencies are
	// declared by service key (compose semantics); fall back to matching the
	// container name directly for robustness.
	resolveDep := func(dep string) string {
		if name, ok := nameByService[dep]; ok && name != "" {
			return name
		}
		return dep
	}

	// inBatch reports whether a service/container name is part of this compose
	// creation batch (by service key OR by container name).
	inBatch := func(name string) bool {
		if _, ok := serviceMap[name]; ok {
			return true
		}
		if _, ok := serviceByName[name]; ok {
			return true
		}
		return false
	}

	for len(serviceMap) > 0 {
		// Keep track of whether we've added any services in this iteration
		changed := false

		for name, service := range serviceMap {
			dependencies := service.DependsOn
			if dependencies == nil {
				dependencies = make(map[string]ContainerCreateRequestContainerDependsOnCont)
			}

			// Soft ordering constraint: if network_mode references another
			// service that is part of this batch, that service must start
			// first. Unlike depends_on this never hard-fails for an external
			// target: if the referenced container is not in the batch, Docker
			// enforces it exists at create time (compose parity).
			if nm := string(service.NetworkMode); strings.HasPrefix(nm, "container:") || strings.HasPrefix(nm, "service:") {
				target := strings.TrimPrefix(strings.TrimPrefix(nm, "container:"), "service:")
				if target != "" && inBatch(target) {
					if _, ok := dependencies[target]; !ok {
						dependencies[target] = ContainerCreateRequestContainerDependsOnCont{
							Condition: "service_started",
						}
					}
				}
			}

			// If there are no dependencies, we can add this service to startOrder
			// Check if all dependencies are already in startOrder
			allDependenciesStarted := true
			for dependency, dependencyDetails := range dependencies {
				depName := resolveDep(dependency)
				dependencyStarted := false
				for _, startedService := range startOrder {
					if startedService.Name == depName {
						dependencyStarted = true

						if dependencyDetails.Condition == "service_healthy" || dependencyDetails.Condition == "service_started" {
							mustStart = true
						}

						break
					}
				}
				// A dependency on a service not in this compose batch (e.g. an
				// external/named container): the target must already exist; we
				// treat it as satisfied (Docker will fail create if it doesn't).
				if !dependencyStarted {
					if !inBatch(dependency) {
						dependencyStarted = true
					}
				}
				if !dependencyStarted {
					allDependenciesStarted = false
					break
				}
			}

			// If all dependencies are started, we can add this service to startOrder
			if allDependenciesStarted {
				utils.Debug(fmt.Sprintf("ReOrderServices:  adding: %s", name))
				startOrder = append(startOrder, service)
				delete(serviceMap, name)
				changed = true
			}
		}

		// If we haven't added any services in this iteration, then there must be a circular dependency
		if !changed {
			break
		}
	}

	// If there are any services left in serviceMap, they couldn't be started due to unsatisfied dependencies or circular dependencies
	if len(serviceMap) > 0 {
		errorMessage := "Could not start all services due to unsatisfied dependencies or circular dependencies:\n"
		for name, _ := range serviceMap {
			errorMessage += "Could not start service: " + name + "\n"
			errorMessage += "Unsatisfied dependencies:\n"

			for dependency, _ := range serviceMap[name].DependsOn {
				if inBatch(dependency) {
					errorMessage += dependency + "\n"
				}
			}
		}
		return nil, false, errors.New(errorMessage)
	}

	return startOrder, mustStart, nil
}