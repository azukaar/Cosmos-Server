package proxy

import (
    "io"
    "net"
    "sync"
    "strings"
    "strconv"
    "time"


    "github.com/azukaar/cosmos-server/src/utils"
    "github.com/azukaar/cosmos-server/src/docker"
    "github.com/azukaar/cosmos-server/src/constellation"
)

var (
    activeProxies = make(map[string]*ProxyInfo)
    proxiesLock   sync.Mutex
)

var (
    netDial                    = net.Dial
    dockerGetContainerIPByName = docker.GetContainerIPByName
    tcpShieldMiddleware        = TCPSmartShieldMiddleware
)

func stripTargetProtocol(target string) string {
    parts := strings.Split(target, "://")
    if len(parts) > 1 {
        return parts[1]
    }
    return parts[0]
}

type ProxyInfo struct {
    stop    chan bool
    stopped chan bool
}

func GetActiveProxies() []string {
    proxiesLock.Lock()
    defer proxiesLock.Unlock()

    var ports []string
    for port := range activeProxies {
        ports = append(ports, port)
    }
    return ports
}

func handleClient(client net.Conn, server net.Conn, proxyInfo *ProxyInfo) {
    defer client.Close()
    defer server.Close()

    // Forward data between client and server, and watch for stop signal.
    // Buffered so the second copy goroutine never parks on send.
    done := make(chan struct{}, 2)
    go func() {
        io.Copy(server, client)
        done <- struct{}{}
    }()
    go func() {
        io.Copy(client, server)
        done <- struct{}{}
    }()

    select {
    case <-proxyInfo.stop:
        return
    case <-done:
        return
    }
}

var mapToTCPUDP = map[string]string{
    "http":     "tcp",
    "https":    "tcp",
    "ftp":      "tcp",
    "sftp":     "tcp",
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

// targetProtocol splits a target into listener protocol and bare host:port.
func targetProtocol(target string) (protocol string, hostPort string) {
    parts := strings.Split(target, "://")
    if len(parts) == 1 {
        return "tcp", parts[0]
    }
    protocol = parts[0]
    if mapped, ok := mapToTCPUDP[protocol]; ok {
        protocol = mapped
    }
    return protocol, parts[1]
}

func startProxy(pair PortsPair, proxyInfo *ProxyInfo) {
    listenAddr := pair.From
    listenProtocol, target := targetProtocol(pair.To)

    var err error
    var listener net.Listener
    var packetConn net.PacketConn

    // Ensure we always signal completion, even on early errors
    defer func() {
        if listener != nil {
            listener.Close()
        }
        if packetConn != nil {
            packetConn.Close()
        }
        close(proxyInfo.stopped)
    }()

    switch listenProtocol {
    case "tcp":
        listener, err = net.Listen("tcp", listenAddr)
    case "udp":
        packetConn, err = net.ListenPacket("udp", listenAddr)
    default:
        utils.Error("[SocketProxy] Unsupported protocol: "+listenProtocol, nil)
        return
    }

    if err != nil {
        utils.Error("[SocketProxy] Failed to listen on "+listenAddr, err)
        return
    }

    utils.Log("[SocketProxy] Proxy listening on "+listenAddr+", forwarding to "+listenProtocol+"://"+target)

    if listenProtocol == "tcp" {
        utils.Debug("[SocketProxy] Starting TCP proxy on " + listenAddr + " -> " + target)
        handleTCPProxy(listener, target, pair, proxyInfo)
    } else {
        handleUDPProxy(packetConn, target, pair.Targets, proxyInfo, pair.route, listenAddr)
    }
}

func handleTCPProxy(listener net.Listener, target string, pair PortsPair, proxyInfo *ProxyInfo) {
    listenAddr := pair.From
    route := pair.route

    for {
        acceptChan := make(chan net.Conn)
        acceptErrChan := make(chan error)
        go func() {
            conn, err := listener.Accept()
            if err != nil {
                acceptErrChan <- err
                return
            }
            acceptChan <- conn
        }()

        select {
        case <-proxyInfo.stop:
            utils.Log("[SocketProxy] Stopping TCP proxy on " + listenAddr)
            return
        case err := <-acceptErrChan:
            utils.Error("[SocketProxy] Failed to accept TCP connection", err)
        case client := <-acceptChan:
            var shieldedClient net.Conn
            // this listener is the only place the client's real address exists;
            // restricted/whitelisted routes must be enforced here
            if !pair.isHTTPProxy || route.RestrictToConstellation || len(route.WhitelistInboundIPs) > 0 {
                shieldedClient = tcpShieldMiddleware("proxy-"+listenAddr, route)(client)
                if shieldedClient == nil {
                    // the shield skips Close on the abuse-counter path
                    client.Close()
                    continue
                }
            } else {
                shieldedClient = client
            }

            utils.Debug("[SocketProxy] New TCP connection accepted on " + listenAddr)

            stickyKey := ""
            if route.LBStickyMode {
                ip, _ := utils.SplitIP(client.RemoteAddr().String())
                stickyKey = ip
            }
            selected := DefaultTunnelLB.SelectTarget(pair.Targets, route.Name, route.LBMode, route.LBStickyMode, stickyKey)
            dialTarget := target
            release := func() {}

            if selected != nil {
                // a tunnel hop: the peer owns that container, nothing to wake here
                dialTarget = stripTargetProtocol(selected.TargetURL)
            } else if pair.lazyContainer != "" {
                // wake and resolve per connection: the IP changes on restart
                if err := lazyWakeForDial(pair.lazyContainer, pair.lazyPort); err != nil {
                    if isLazyStarting(err) {
                        utils.Debug("[SocketProxy] Container " + pair.lazyContainer + " is still starting, dropping the connection on " + listenAddr)
                    } else {
                        utils.Error("[SocketProxy] Failed to wake container "+pair.lazyContainer, err)
                    }
                    shieldedClient.Close()
                    continue
                }

                targetIP, err := dockerGetContainerIPByName(pair.lazyContainer)
                if err != nil {
                    utils.Error("[SocketProxy] Can't find container "+pair.lazyContainer, err)
                    shieldedClient.Close()
                    continue
                }

                utils.Debug("[SocketProxy] Dockerless socket Target IP: " + targetIP)
                dialTarget = targetIP + ":" + pair.lazyPort
                // held for the connection's lifetime so the idle reaper can't stop the container
                release = lazyTrackConn(pair.lazyContainer)
            }

            server, err := netDial("tcp", dialTarget)
            if err != nil {
                utils.Error("[SocketProxy] Failed to connect to TCP server", err)
                release()
                shieldedClient.Close()
                continue
            }
            go func() {
                defer release()
                handleClient(shieldedClient, server, proxyInfo)
            }()
        }
    }
}

func handleUDPProxy(packetConn net.PacketConn, target string, targets []utils.TunnelTarget, proxyInfo *ProxyInfo, route utils.ProxyRouteConfig, listenAddr string) {
    targetAddr, err := net.ResolveUDPAddr("udp", target)
    if err != nil {
        utils.Error("[SocketProxy] Failed to resolve UDP target address", err)
        return
    }

    buffer := make([]byte, 65507) // Max UDP packet size

    for {
        select {
        case <-proxyInfo.stop:
            utils.Log("[SocketProxy] Stopping UDP proxy on " + listenAddr)
            return
        default:
            packetConn.SetReadDeadline(time.Now().Add(5 * time.Second))
            n, remoteAddr, err := packetConn.ReadFrom(buffer)
            if err != nil {
                if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
                    continue
                }
                utils.Error("[SocketProxy] Failed to read UDP packet", err)
                continue
            }

            // Apply UDP shield middleware if needed
            shieldedBuffer := UDPSmartShieldMiddleware("proxy-"+listenAddr, route)(buffer[:n], remoteAddr)
            if shieldedBuffer == nil {
                continue
            }

            stickyKey := ""
            if route.LBStickyMode {
                ip, _ := utils.SplitIP(remoteAddr.String())
                stickyKey = ip
            }
            selected := DefaultTunnelLB.SelectTarget(targets, route.Name, route.LBMode, route.LBStickyMode, stickyKey)
            dialAddr := targetAddr
            if selected != nil {
                resolved, err := net.ResolveUDPAddr("udp", stripTargetProtocol(selected.TargetURL))
                if err == nil {
                    dialAddr = resolved
                }
            }

            _, err = packetConn.WriteTo(buffer[:n], dialAddr)
            if err != nil {
                utils.Error("[SocketProxy] Failed to forward UDP packet to server", err)
                continue
            }

            // Handle response from target
            go handleUDPResponse(packetConn, remoteAddr, dialAddr, proxyInfo)
        }
    }
}

func handleUDPResponse(packetConn net.PacketConn, clientAddr net.Addr, targetAddr *net.UDPAddr, proxyInfo *ProxyInfo) {
    buffer := make([]byte, 65507)
    for {
        packetConn.SetReadDeadline(time.Now().Add(30 * time.Second))
        n, _, err := packetConn.ReadFrom(buffer)
        if err != nil {
            if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
                return // No more responses from target
            }
            utils.Error("[SocketProxy] Failed to read UDP response from target", err)
            return
        }

        _, err = packetConn.WriteTo(buffer[:n], clientAddr)
        if err != nil {
            utils.Error("[SocketProxy] Failed to forward UDP response to client", err)
            return
        }
    }
}

// socketDestination decides what a non-HTTP route's listener forwards to.
func socketDestination(route utils.ProxyRouteConfig, destination string) (dest string, lazyContainer string, lazyPort string) {
    if route.Mode != "SERVAPP" || (utils.IsInsideContainer && !utils.IsHostNetwork) {
        return destination, "", ""
    }

    containerName, containerPort := lazyTargetContainer(route)
    if containerName == "" {
        utils.Error("[SocketProxy] Create socket Route: no container name in target "+route.Target, nil)
        return destination, "", ""
    }

    if protocol, _ := targetProtocol(destination); protocol == "tcp" {
        return destination, containerName, containerPort
    }

    targetIP, err := dockerGetContainerIPByName(containerName)
    if err != nil {
        utils.Error("[SocketProxy] Can't find container", err)
        return destination, "", ""
    }

    utils.Debug("[SocketProxy] Dockerless socket Target IP: " + targetIP)
    destination = targetIP + ":" + containerPort
    if scheme := strings.Split(route.Target, "://"); len(scheme) > 1 {
        destination = scheme[0] + "://" + destination
    }
    return destination, "", ""
}

type PortsPair struct {
    From string
    To string
    Targets []utils.TunnelTarget
    isHTTPProxy bool
    route utils.ProxyRouteConfig
    // dockerless SERVAPP TCP routes keep the container name to wake it and resolve its IP per connection
    lazyContainer string
    lazyPort string
}

func initInternalPortProxy(ports []PortsPair) {
    proxiesLock.Lock()
    defer proxiesLock.Unlock()
    
    for _, port := range ports {
        utils.Log("[SocketProxy] Network Starting internal proxy for port " + port.From)
        proxyInfo := &ProxyInfo{
            stop:    make(chan bool),
            stopped: make(chan bool),
        }
        activeProxies[port.From] = proxyInfo

        go startProxy(port, proxyInfo)
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

    var wg sync.WaitGroup
    for port, proxyInfo := range activeProxies {
        wg.Add(1)
        go func(port string, proxyInfo *ProxyInfo) {
            defer wg.Done()
            close(proxyInfo.stop)
            
            // Wait for the proxy to stop (with a timeout)
            select {
            case <-time.After(30 * time.Second):
                utils.Error("[SocketProxy] Timeout waiting for proxy on port " + port + " to stop", nil)
            case <-proxyInfo.stopped:
                utils.Log("[SocketProxy] Proxy on port " + port + " stopped successfully")
            }
        }(port, proxyInfo)
    }

    // Wait for all goroutines to finish
    wg.Wait()

    // Clear the map of active proxies
    activeProxies = make(map[string]*ProxyInfo)
}

func InitInternalSocketProxy() {
    utils.Log("[SocketProxy] Initializing internal TCP/UDP proxy")
    
    config := utils.GetMainConfig()
    claims := []PortClaim{}
	isHTTPS := utils.IsHTTPS
	HTTPPort := config.HTTPConfig.HTTPPort
	HTTPSPort := config.HTTPConfig.HTTPSPort
	routes := config.HTTPConfig.ProxyConfig.Routes


	remoteTunnels := constellation.GetLocalTunnelCache()

    remoteListRoutes := []utils.ProxyRouteConfig{}
    for _, shares := range config.RemoteStorage.Shares {
        route := shares.Route
        remoteListRoutes = append(remoteListRoutes, route)
    }

	targetPort := HTTPPort

    allowHTTPLocal := config.HTTPConfig.AllowHTTPLocalIPAccess

	if isHTTPS {
		targetPort = HTTPSPort
	}

    addPortPair := func(route utils.ProxyRouteConfig, targets []utils.TunnelTarget, isLocal bool) {
        if !route.UseHost || !strings.Contains(route.Host, ":") || route.Disabled {
            return
        }
        hostname := route.Host
        port := strings.Split(hostname, ":")[1]
        if _, err := strconv.Atoi(port); err != nil {
            return
        }
        destination := ""
        isHTTPProxy := strings.HasPrefix(route.Target, "http://") || strings.HasPrefix(route.Target, "https://")
        sourceHostname := hostname
        lazyContainer := ""
        lazyPort := ""
        var pairTargets []utils.TunnelTarget
        if isHTTPProxy {
            destination = "localhost:" + targetPort
            sourceHostname = ":" + port
        } else {
            destination = route.Target
            pairTargets = targets

            destination, lazyContainer, lazyPort = socketDestination(route, destination)
        }

        portpair := PortsPair{
            From: sourceHostname,
            To: destination,
            Targets: pairTargets,
            isHTTPProxy: isHTTPProxy,
            route: route,
            lazyContainer: lazyContainer,
            lazyPort: lazyPort,
        }

        if isHTTPProxy && allowHTTPLocal && utils.IsLocalIP(route.Host) {
            portpair.To = "localhost:" + HTTPPort
        }

        claims = append(claims, PortClaim{Pair: portpair, IsLocal: isLocal})
    }

    // Local routes first: they take precedence over tunnel advertisements.
    for _, route := range remoteListRoutes {
        addPortPair(route, nil, true)
    }
	for _, route := range routes {
        addPortPair(route, nil, true)
	}
	for _, tunnel := range remoteTunnels {
		if !tunnel.Route.Disabled {
            addPortPair(tunnel.Route, tunnel.Targets, false)
		}
	}

    expectedPorts, shadowed, conflicts := ResolvePortClaims(claims)

    for _, port := range shadowed {
        utils.Debug("[SocketProxy] Port " + port + " is served locally, ignoring the tunnel advertisement for it")
    }
    for _, port := range conflicts {
        utils.MajorError("[SocketProxy] Duplicate port detected: "+port+". Multiple routes are trying to use the same port.", nil)
    }

    StopAllProxies()

    initInternalPortProxy(expectedPorts)
}

// PortClaim is one route's bid for a listening port.
type PortClaim struct {
	Pair    PortsPair
	IsLocal bool
}

// ResolvePortClaims picks one route per port: a local claim always beats a
// tunnel advertisement for the same port.
func ResolvePortClaims(claims []PortClaim) (accepted []PortsPair, shadowed []string, conflicts []string) {
	claimedLocally := map[string]bool{}
	claimed := map[string]bool{}

	for _, claim := range claims {
		if !claim.IsLocal {
			continue
		}
		port := claim.Pair.From
		if claimed[port] {
			conflicts = append(conflicts, port)
			continue
		}
		claimed[port] = true
		claimedLocally[port] = true
		accepted = append(accepted, claim.Pair)
	}

	for _, claim := range claims {
		if claim.IsLocal {
			continue
		}
		port := claim.Pair.From
		if claimedLocally[port] {
			shadowed = append(shadowed, port)
			continue
		}
		if claimed[port] {
			conflicts = append(conflicts, port)
			continue
		}
		claimed[port] = true
		accepted = append(accepted, claim.Pair)
	}

	return accepted, shadowed, conflicts
}