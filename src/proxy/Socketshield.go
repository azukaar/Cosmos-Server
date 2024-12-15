package proxy

import (
	"fmt"
	"net"
	"sync"
	"time"
	"strings"

	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/azukaar/cosmos-server/src/metrics"
)

type TCPSmartShieldState struct {
	sync.Mutex
	Connections []*TCPConnectionWrapper
	Bans        []*UserBan
}

type TCPUserUsedBudget struct {
	ClientID      string
	Time          float64
	Packets       int64
	Bytes         int64
	Simultaneous  int
}

type TCPConnectionWrapper struct {
	Conn          net.Conn
	TimeStarted   time.Time
	TimeEnded     time.Time
	ClientID      string
	Packets       int64
	Bytes         int64
	IsOver        bool
	ShieldID      string
	Policy        utils.SmartShieldPolicy
	IsPrivileged  bool
	ThrottleUntil time.Duration
	mutex         sync.Mutex
	Route 			 utils.ProxyRouteConfig
}

// Implement the missing methods to satisfy the net.Conn interface
func (w *TCPConnectionWrapper) LocalAddr() net.Addr {
	return w.Conn.LocalAddr()
}

func (w *TCPConnectionWrapper) RemoteAddr() net.Addr {
	return w.Conn.RemoteAddr()
}

func (w *TCPConnectionWrapper) SetDeadline(t time.Time) error {
	return w.Conn.SetDeadline(t)
}

func (w *TCPConnectionWrapper) SetReadDeadline(t time.Time) error {
	return w.Conn.SetReadDeadline(t)
}

func (w *TCPConnectionWrapper) SetWriteDeadline(t time.Time) error {
	return w.Conn.SetWriteDeadline(t)
}

var socketShield TCPSmartShieldState

func (shield *TCPSmartShieldState) GetServerConnections(shieldID string) int {
	shield.Lock()
	defer shield.Unlock()
	connections := 0

	for _, conn := range shield.Connections {
		if !conn.IsOver && conn.ShieldID == shieldID {
			connections++
		}
	}

	return connections
}

func CleanUpSocket() {
	socketShield.Lock()
	defer socketShield.Unlock()

	shieldSize := len(socketShield.Connections)

	for i := len(socketShield.Connections) - 1; i >= 0; i-- {
		request := socketShield.Connections[i]
		oneHourAgo := time.Now().Add(-time.Hour)
		if(request.IsOver && request.TimeEnded.Before(oneHourAgo)) {
			socketShield.Connections = append(socketShield.Connections[:i], socketShield.Connections[i+1:]...)
		}
	}

	utils.Log("SmartShield: Cleaned up " + fmt.Sprintf("%d", shieldSize-len(socketShield.Connections)) + " old requests")
}

func (shield *TCPSmartShieldState) GetUserUsedBudgets(shieldID string, clientID string) TCPUserUsedBudget {
	shield.Lock()
	defer shield.Unlock()

	userConsumed := TCPUserUsedBudget{
		ClientID:     clientID,
		Time:         0,
		Packets:      0,
		Bytes:        0,
		Simultaneous: 0,
	}

	for _, conn := range shield.Connections {
		if conn.ShieldID != shieldID || conn.ClientID != clientID {
			continue
		}

		if conn.IsOver {
			userConsumed.Time += conn.TimeEnded.Sub(conn.TimeStarted).Seconds()
		} else {
			userConsumed.Time += time.Now().Sub(conn.TimeStarted).Seconds()
			userConsumed.Simultaneous++
		}
		userConsumed.Packets += conn.Packets
		userConsumed.Bytes += conn.Bytes
	}

	return userConsumed
}

func (shield *TCPSmartShieldState) IsAllowedToConnect(shieldID string, policy utils.SmartShieldPolicy, userConsumed TCPUserUsedBudget) bool {
	if(!policy.Enabled) {
		return true
	}
	
	shield.Lock()
	defer shield.Unlock()

	globalShieldState.Lock()
	defer globalShieldState.Unlock()

	clientID := userConsumed.ClientID

	if clientID == "192.168.1.1" || clientID == "192.168.0.1" || clientID == "192.168.0.254" || clientID == "172.17.0.1" {
		return true
	}

	nbTempBans := 0
	nbStrikes := 0

	for _, ban := range globalShieldState.bans {
		if ban.ClientID != clientID {
			continue
		}

		switch ban.BanType {
		case PERM:
			return false
		case TEMP:
			if ban.time.Add(4 * time.Hour).After(time.Now()) {
				return false
			} else if ban.time.Add(72 * time.Hour).After(time.Now()) {
				nbTempBans++
			}
		case STRIKE:
			if ban.time.Add(time.Hour).After(time.Now()) {
				return false
			} else if ban.time.Add(24 * time.Hour).After(time.Now()) {
				nbStrikes++
			}
		}
	}

	if nbTempBans >= 3 {
		globalShieldState.bans = append(globalShieldState.bans, &UserBan{
			ClientID: clientID,
			BanType:  PERM,
			time:     time.Now(),
			shieldID: shieldID,
		})
		utils.Warn(fmt.Sprintf("TCP User %s has been banned permanently: %+v", clientID, userConsumed))
		return false
	} else if nbStrikes >= 3 {
		globalShieldState.bans = append(globalShieldState.bans, &UserBan{
			ClientID: clientID,
			BanType:  TEMP,
			time:     time.Now(),
			shieldID: shieldID,
		})
		utils.Warn(fmt.Sprintf("TCP User %s has been banned temporarily: %+v", clientID, userConsumed))
		return false
	}

	if (userConsumed.Time > policy.PerUserTimeBudget*float64(policy.PolicyStrictness)) ||
		(userConsumed.Packets > int64(policy.PerUserRequestLimit*1000*policy.PolicyStrictness)) ||
		(userConsumed.Bytes > policy.PerUserByteLimit*int64(policy.PolicyStrictness)) ||
		(userConsumed.Simultaneous > policy.PerUserSimultaneous*policy.PolicyStrictness) {
		globalShieldState.bans = append(globalShieldState.bans, &UserBan{
			ClientID: clientID,
			BanType:  STRIKE,
			time:     time.Now(),
			reason:   fmt.Sprintf("%+v out of %+v", userConsumed, policy),
			shieldID: shieldID,
		})
		utils.Warn(fmt.Sprintf("TCP User %s has received a strike: %+v", clientID, userConsumed))
		return false
	}

	utils.Debug(fmt.Sprintf("TCPSmartShield: User %s is allowed to connect", clientID))

	return true
}

func TCPSmartShieldWrapper(conn net.Conn, shieldID string, route utils.ProxyRouteConfig, policy utils.SmartShieldPolicy) *TCPConnectionWrapper {
	clientID, _, _ := net.SplitHostPort(conn.RemoteAddr().String())

	wrapper := &TCPConnectionWrapper{
		Conn:         conn,
		TimeStarted:  time.Now(),
		ClientID:     clientID,
		ShieldID:     shieldID,
		Policy:       policy,
		Route:        route,
		IsPrivileged: false, // You may want to implement a way to determine privileged connections
	}

	socketShield.Lock()
	socketShield.Connections = append(socketShield.Connections, wrapper)
	socketShield.Unlock()
	
	utils.TriggerEvent(
		"cosmos.socket-proxy.opened." + wrapper.Route.Name,
		"Socket Proxy " + wrapper.Route.Name + " Opened for " + wrapper.ClientID,
		"success",
		"route@" + wrapper.Route.Name,
		map[string]interface{}{
		"clientID": wrapper.ClientID,
	})

	return wrapper
}


func (w *TCPConnectionWrapper) Write(b []byte) (int, error) {
	if w.ThrottleUntil > 0 {
		time.Sleep(w.ThrottleUntil)
	}

	n, err := w.Conn.Write(b)

	w.mutex.Lock()
	w.Bytes += int64(n)
	w.Packets++
	w.mutex.Unlock()

	// utils.Debug(fmt.Sprintf("TCPSmartShield: Wrote %d bytes and %d packets. ThrottleUntil: %v", w.Bytes, w.Packets, w.ThrottleUntil))
	return n, err
}

func (w *TCPConnectionWrapper) Read(b []byte) (int, error) {
	if w.ThrottleUntil > 0 {
		time.Sleep(w.ThrottleUntil)
	}

	n, err := w.Conn.Read(b)

	w.mutex.Lock()
	w.Bytes += int64(n)
	w.Packets++
	w.mutex.Unlock()

	return n, err
}

func (shield *TCPSmartShieldState) EnforceBudget() {
	shield.Lock()
	defer shield.Unlock()

	globalShieldState.Lock()
	defer globalShieldState.Unlock()

	for _, conn := range shield.Connections {
		if conn.IsOver || !conn.Policy.Enabled {
			continue
		}

		policy := conn.Policy

		shield.Unlock()
		userConsumed := shield.GetUserUsedBudgets(conn.ShieldID, conn.ClientID)
		shield.Lock()

		throttle := 0

		overReq := int64(policy.PerUserRequestLimit * 1000) - userConsumed.Packets
		overReqRatio := float64(overReq) / float64((policy.PerUserRequestLimit * 1000))
		if overReq < 0 {
			newThrottle := int(float64(300) * -overReqRatio)
			if newThrottle > throttle {
				throttle = newThrottle
			}
		}
		
		overByte := policy.PerUserByteLimit - userConsumed.Bytes
		overByteRatio := float64(overByte) / float64(policy.PerUserByteLimit)
		if overByte < 0 {
			newThrottle := int(float64(40) * -overByteRatio)
			if newThrottle > throttle {
				throttle = newThrottle
			}
		}
	
		overSim := policy.PerUserSimultaneous - userConsumed.Simultaneous
		overSimRatio := float64(overSim) / float64(policy.PerUserSimultaneous)
		if overSim < 0 {
			newThrottle := int(float64(50) * -overSimRatio)
			if newThrottle > throttle {
				throttle = newThrottle
			}
		}

		utils.Debug(fmt.Sprintf("TCPSmartShield: Ratio: %f, %f, %f", overReqRatio, overByteRatio, overSimRatio))

		conn.ThrottleUntil = time.Duration(throttle * int(time.Millisecond))

		if !conn.Policy.Enabled {
			continue
		}

		if (userConsumed.Time > policy.PerUserTimeBudget*float64(policy.PolicyStrictness)) ||
			(userConsumed.Packets > int64(policy.PerUserRequestLimit*1000*policy.PolicyStrictness)) ||
			(userConsumed.Bytes > policy.PerUserByteLimit*int64(policy.PolicyStrictness)) {

			utils.Warn(fmt.Sprintf("TCPSmartShield: Kicking out user %s due to exceeded budget", conn.ClientID))
			
			conn.Close()

			utils.TriggerEvent(
				"cosmos.proxy.shield.abuse." + conn.Route.Name,
				"Socket Shield " + conn.Route.Name + " Abuse by " + conn.ClientID,
				"warning",
				"route@" + conn.Route.Name,
				map[string]interface{}{
				"route": conn.Route.Name,
				"consumed": userConsumed,
				"lastBan": GetLastBan(conn.ClientID, false),
				"clientID": conn.ClientID,
			})

			globalShieldState.bans = append(globalShieldState.bans, &UserBan{
				ClientID: conn.ClientID,
				BanType:  STRIKE,
				time:     time.Now(),
				reason:   fmt.Sprintf("%+v out of %+v", userConsumed, policy),
				shieldID: conn.ShieldID,
			})
		}
	}
}

func StartBudgetEnforcer() {
	go func() {
		for {
			time.Sleep(10 * time.Second)
			socketShield.EnforceBudget()
		}
	}()
}

func IPInRange(ip, cidr string) (bool, error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false, err
	}
	parsedIP := net.ParseIP(ip)
	return ipnet.Contains(parsedIP), nil
}

func (w *TCPConnectionWrapper) Close() error {
	if !w.IsOver {
		w.TimeEnded = time.Now()
		w.IsOver = true

		utils.TriggerEvent(
			"cosmos.socket-proxy.closed." + w.Route.Name,
			"Socket Proxy " + w.Route.Name + " Closed for " + w.ClientID,
			"success",
			"route@" + w.Route.Name,
			map[string]interface{}{
			"route": w.Route.Name,
			"consumed": w.Bytes,
			"packets": w.Packets,
			"time": w.TimeEnded.Sub(w.TimeStarted).Seconds(),
			"clientID": w.ClientID,
		})

		go metrics.PushRequestMetrics(w.Route, 0, w.TimeStarted, w.Bytes)
	}
	
	return w.Conn.Close()
}

func InitSocketShield() {
	StartBudgetEnforcer()
}

func TCPSmartShieldMiddleware(shieldID string, route utils.ProxyRouteConfig) func(net.Conn) net.Conn {
	policy := route.SmartShield

	if policy.Enabled {
		if(policy.PerUserTimeBudget == 0) {
			policy.PerUserTimeBudget = 2 * 60 * 60 * 1000 // 2 hours
		}
		if(policy.PerUserRequestLimit == 0) {
			policy.PerUserRequestLimit = 12000 // 150 requests per minute
		}
		if(policy.PerUserByteLimit == 0) {
			policy.PerUserByteLimit = 150 * 1024 * 1024 * 1024 // 150GB
		}
		if(policy.PolicyStrictness == 0) {
			policy.PolicyStrictness = 2 // NORMAL
		}
		if(policy.PerUserSimultaneous == 0) {
			policy.PerUserSimultaneous = 24
		}
		if(policy.MaxGlobalSimultaneous == 0) {
			policy.MaxGlobalSimultaneous = 250
		}
		if(policy.PrivilegedGroups == 0) {
			policy.PrivilegedGroups = utils.ADMIN
		}
	}
	
	return func(conn net.Conn) net.Conn {
		clientID, _, _ := net.SplitHostPort(conn.RemoteAddr().String())

		if(utils.GetIPAbuseCounter(clientID) > 275) {
			return nil
		}

		whitelistInboundIPs := route.WhitelistInboundIPs
		restrictToConstellation := route.RestrictToConstellation

		// Whitelist / Constellation check

		isUsingWhitelist := len(whitelistInboundIPs) > 0
		isInWhitelist := false
		isInConstellation := strings.HasPrefix(clientID, "192.168.201.") || strings.HasPrefix(clientID, "192.168.202.")

		for _, ipRange := range whitelistInboundIPs {
			utils.Debug(fmt.Sprintf("Checking if %s is in %s", clientID, ipRange))
			if strings.Contains(ipRange, "/") {
				if ok, _ := IPInRange(clientID, ipRange); ok {
					isInWhitelist = true
					break
				}
			} else if clientID == ipRange {
				isInWhitelist = true
				break
			}
		}

		if restrictToConstellation && !isInConstellation {
			if !isUsingWhitelist || (isUsingWhitelist && !isInWhitelist) {
				utils.PushShieldMetrics("ip-whitelists")
				utils.TriggerEvent(
					"cosmos.proxy.shield.whitelist",
					"Socket Shield IP blocked by whitelist",
					"warning",
					"",
					map[string]interface{}{
						"clientID": clientID,
					},
				)
				utils.IncrementIPAbuseCounter(clientID)
				utils.Error(fmt.Sprintf("Connection from %s is blocked because of restrictions", clientID), nil)
				utils.Debug("Blocked by RestrictToConstellation isInConstellation isUsingWhitelist isInWhitelist")
				conn.Close()
				return nil
			}
		} else if isUsingWhitelist && !isInWhitelist {
			utils.PushShieldMetrics("ip-whitelists")
			utils.TriggerEvent(
				"cosmos.proxy.shield.whitelist",
				"Socket Shield IP blocked by whitelist",
				"warning",
				"",
				map[string]interface{}{
					"clientID": clientID,
				},
			)
			utils.IncrementIPAbuseCounter(clientID)
			utils.Error(fmt.Sprintf("Connection from %s is blocked because of restrictions", clientID), nil)
			utils.Debug("Blocked by isUsingWhitelist isInWhitelist")
			conn.Close()
			return nil
		}

		// Geo check
		
		countryCode, err := utils.GetIPLocation(clientID)
		if err == nil {
			config := utils.GetMainConfig()
			countryBlacklistIsWhitelist := config.CountryBlacklistIsWhitelist
			blockedCountries := config.BlockedCountries

			if countryBlacklistIsWhitelist {
				if countryCode != "" {
					utils.Debug("Country code: " + countryCode)
					blocked := true
					for _, blockedCountry := range blockedCountries {
						if config.ServerCountry != countryCode && countryCode == blockedCountry {
							blocked = false
						}
					}

					if blocked {
						utils.PushShieldMetrics("geo")
						utils.IncrementIPAbuseCounter(clientID)

						utils.TriggerEvent(
							"cosmos.proxy.shield.geo",
							"Proxy Shield Geo blocked",
							"warning",
							"",
							map[string]interface{}{
							"clientID": clientID,
							"country": countryCode,
							"route": route.Name,
						})

						utils.Warn(fmt.Sprintf("Connection from %s is blocked because of geo restrictions", clientID))
						conn.Close()
						return nil
					}
				} else {
					utils.Debug("Missing geolocation information to block IPs")
				}
			} else {
				for _, blockedCountry := range blockedCountries {
					if config.ServerCountry != countryCode && countryCode == blockedCountry {

						utils.PushShieldMetrics("geo")
						utils.IncrementIPAbuseCounter(clientID)

						utils.TriggerEvent(
							"cosmos.proxy.shield.geo",
							"Proxy Shield Geo blocked",
							"warning",
							"",
							map[string]interface{}{
							"clientID": clientID,
							"country": countryCode,
							"route": route.Name,
						})

						utils.Warn(fmt.Sprintf("Connection from %s is blocked because of geo restrictions", clientID))

						conn.Close()
						return nil
					}
				}
			}
		} else {
			utils.Debug("Missing geolocation information to block IPs")
		}
		
		userConsumed := socketShield.GetUserUsedBudgets(shieldID, clientID)

		if !socketShield.IsAllowedToConnect(shieldID, policy, userConsumed) {
			utils.TriggerEvent(
				"cosmos.proxy.shield.abuse." + route.Name,
				"Socket Shield " + route.Name + " Abuse by " + clientID,
				"warning",
				"route@" + route.Name,
				map[string]interface{}{
				"route": route.Name,
				"consumed": userConsumed,
				"lastBan": GetLastBan(clientID, false),
				"clientID": clientID,
			})

			utils.Warn(fmt.Sprintf("TCPSmartShield: Connection blocked for %s due to policy violation", clientID))
			conn.Close()
			return nil
		}

		wrapper := TCPSmartShieldWrapper(conn, shieldID, route, policy)

		if policy.Enabled {
			currentConnections := socketShield.GetServerConnections(shieldID) + 1
			utils.Debug(fmt.Sprintf("TCPSmartShield: Current global connections: %d", currentConnections))
	
			if currentConnections > policy.MaxGlobalSimultaneous {
				utils.TriggerEvent(
					"cosmos.proxy.shield.abuse." + route.Name,
					"Socket Shield " + route.Name + " Abuse by " + clientID,
					"warning",
					"route@" + route.Name,
					map[string]interface{}{
					"route": route.Name,
					"consumed": userConsumed,
					"clientID": clientID,
				})

				utils.Warn(fmt.Sprintf("TCPSmartShield: Too many connections on the server for %s (%d of %d)", shieldID, currentConnections, policy.MaxGlobalSimultaneous))
				wrapper.Close()
				return nil
			}
		}

		return wrapper
	}
}