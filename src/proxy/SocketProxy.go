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

var mapToTCPUDP = map[string]string{
    "http":     "tcp",
    "https":    "tcp",
    "ftp":      "tcp",
    "ssh":      "tcp",
    "telnet":   "tcp",
    "smtp":     "tcp",
    "pop3":     "tcp",
    "imap":     "tcp",
    "dns":      "udp", // Note: DNS uses both UDP and TCP, but primarily UDP
    "ntp":      "udp",
    "snmp":     "udp",
    "tftp":     "udp",
    "dhcp":     "udp",
    "rdp":      "tcp",
    "sip":      "udp", // Note: SIP can use both UDP and TCP
    "mysql":    "tcp",
    "postgresql": "tcp",
    "mongodb":  "tcp",
    "ldap":     "tcp",
    "vnc":      "tcp",
    "rtsp":     "tcp", // Note: RTSP can use both TCP and UDP
    "xmpp":     "tcp",
    "irc":      "tcp",
    "mqtt":     "tcp",
    "quic":     "udp", // QUIC is a transport protocol that runs over UDP
}


func startProxy(listenAddr string, target string, stop chan bool, isHTTPProxy bool, route utils.ProxyRouteConfig) {
    // remove any protocol
    destinationArr := strings.Split(listenAddr, "://")
    listenProtocol := "tcp"

    if len(destinationArr) > 1 {
        listenAddr = destinationArr[1]
        listenProtocol = destinationArr[0]
        if protocol, ok := mapToTCPUDP[listenProtocol]; ok {
            listenProtocol = protocol
        }
    } else {
        listenAddr = destinationArr[0]
    }

    listener, err := net.Listen(listenProtocol, listenAddr)

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
                utils.Error("Failed to accept connection", err)
                continue
            }

            var shieldedClient net.Conn

            if !isHTTPProxy {
                // Apply TCP Smart Shield
                shieldedClient = TCPSmartShieldMiddleware("proxy-"+listenAddr, route)(client)
                if shieldedClient == nil {
                    // Connection was blocked by the shield
                    continue
                }
            } else {
                shieldedClient = client
            }

            server, err := net.Dial(listenProtocol, target)
            if err != nil {
                utils.Error("Failed to connect to server", err)
                shieldedClient.Close()
                continue
            }

            go handleClient(shieldedClient, server, stop)
        }
    }
}

type PortsPair struct {
    From string
    To string
    isHTTPProxy bool
    route utils.ProxyRouteConfig
}

func initInternalPortProxy(ports []PortsPair) {
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
            go startProxy(":"+port.From, port.To, stop, port.isHTTPProxy, port.route)
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

func StopAllProxies() {
    proxiesLock.Lock()
    defer proxiesLock.Unlock()

    for _, stop := range activeProxies {
        close(stop)
    }
    activeProxies = make(map[string]chan bool)
}

func InitInternalSocketProxy() {
    utils.Log("Network: Initializing internal TCP/UDP proxy")
    
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
                destination := ""
                isHTTPProxy := strings.HasPrefix(route.Target, "http://") || strings.HasPrefix(route.Target, "https://")

                if isHTTPProxy {
                    destination = "localhost:" + targetPort
                } else {
                    destination = route.Target
                }
                
                portpair := PortsPair{port, destination, isHTTPProxy, route}

                if isHTTPProxy && allowHTTPLocal && utils.IsLocalIP(route.Host) {
                    portpair.To = HTTPPort
                }

    			expectedPorts = append(expectedPorts, portpair)
            }
		}
	}

    StopAllProxies()
    
    initInternalPortProxy(expectedPorts)
}