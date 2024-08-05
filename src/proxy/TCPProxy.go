package proxy

import (
    "io"
    "net"
    "sync"
    "strings"
    "strconv"

    "github.com/azukaar/cosmos-server/src/utils"
)

var (
    activeProxies map[string]chan bool
    proxiesLock   sync.Mutex
)

func GetActiveProxies() map[string]chan bool {
    return activeProxies
}

func handleClient(client net.Conn, server net.Conn, stop chan bool) {
    defer client.Close()
    defer server.Close()

    // Forward data between client and server, and watch for stop signal
    done := make(chan struct{})
    go func() {
        io.Copy(server, client)
        done <- struct{}{}
    }()
    go func() {
        io.Copy(client, server)
        done <- struct{}{}
    }()

    select {
    case <-stop:
        return
    case <-done:
        return
    }
}

func startProxy(listenAddr string, target string, stop chan bool) {
    listener, err := net.Listen("tcp", listenAddr)
    if err != nil {
        utils.Error("Failed to listen on " + listenAddr, err)
        return
    }
    defer listener.Close()

    utils.Log("Proxy listening on "+listenAddr+", forwarding to " + target)

    for {
        select {
        case <-stop:
            return
        default:
            client, err := listener.Accept()
            if err != nil {
                utils.Error("Failed to accept connection: %v", err)
                continue
            }

            server, err := net.Dial("tcp", target)
            if err != nil {
                utils.Error("Failed to connect to server: %v", err)
                client.Close()
                continue
            }

            go handleClient(client, server, stop)
        }
    }
}

type PortsPair struct {
    From string
    To string
}

func initInternalPortProxy(ports []PortsPair, destination string) {
    proxiesLock.Lock()
    defer proxiesLock.Unlock()

    // Initialize activeProxies map if it's nil
    if activeProxies == nil {
        activeProxies = make(map[string]chan bool)
    }

    // Stop any existing proxies that are not in the new list
    for port, stop := range activeProxies {
        found := false
        for _, p := range ports {
            if p.From == port {
                found = true
                break
            }
        }
        if !found {
            close(stop)
            delete(activeProxies, port)
        }
    }

    // Start new proxies for ports in the list that aren't already running
    for _, port := range ports {
        if _, exists := activeProxies[port.From]; !exists {
            utils.Log("Network Starting internal proxy for port " + port.From)
            stop := make(chan bool)
            activeProxies[port.From] = stop
            go startProxy(":"+port.From, destination+":"+port.To, stop)
        }
    }
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
    for _, a := range slice {
        if a == item {
            return true
        }
    }
    return false
}

func InitInternalTCPProxy() {
    utils.Log("Network: Initializing internal TCP proxy")
    
    config := utils.GetMainConfig()
    expectedPorts := []PortsPair{}
	isHTTPS := utils.IsHTTPS
	HTTPPort := config.HTTPConfig.HTTPPort
	HTTPSPort := config.HTTPConfig.HTTPSPort
	routes := config.HTTPConfig.ProxyConfig.Routes
    tunnels := []utils.ProxyRouteConfig{}
    if config.ConstellationConfig.Enabled {
        tunnels = config.ConstellationConfig.Tunnels
    }
	targetPort := HTTPPort

    allowHTTPLocal := config.HTTPConfig.AllowHTTPLocalIPAccess

	if isHTTPS {
		targetPort = HTTPSPort
	}

    routesList := []utils.ProxyRouteConfig{}
    routesList = append(routesList, tunnels...)
    routesList = append(routesList, routes...)

	for _, route := range routesList {
		if route.UseHost && strings.Contains(route.Host, ":") {
			hostname := route.Host
			port := strings.Split(hostname, ":")[1]
            // if port is a number
            if _, err := strconv.Atoi(port); err == nil {
                portpair := PortsPair{port, targetPort}
                if allowHTTPLocal && utils.IsLocalIP(route.Host) {
                    portpair.To = HTTPPort
                }
    			expectedPorts = append(expectedPorts, portpair)
            }
		}
	}

	// append hostname port (if it's not the same as the HTTP/HTTPS port?)
    // Should this be removed?
	hostname := config.HTTPConfig.Hostname
	if strings.Contains(hostname, ":") {
		hostnameport := strings.Split(hostname, ":")[1]
        if hostnameport != targetPort {
            portpair := PortsPair{hostnameport, targetPort}
    		expectedPorts = append(expectedPorts, portpair)
        }
	}

    initInternalPortProxy(expectedPorts, "localhost")
}