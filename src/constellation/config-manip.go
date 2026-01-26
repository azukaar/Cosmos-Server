package constellation

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"
    "encoding/binary"
    "time"

	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/natefinch/lumberjack"
	"gopkg.in/yaml.v2"
)

func pingNebula(target string) error {
    // Resolve UDP address
    udpAddr, err := net.ResolveUDPAddr("udp", target)
    if err != nil {
		return err
    }
    
    // Create UDP connection
    conn, err := net.DialUDP("udp", nil, udpAddr)
    if err != nil {
		return err
    }
    defer conn.Close()
    
    // Create Nebula packet
    // Nebula header structure:
    // Version(1) + Type(1) + Subtype(1) + Reserved(2) + RemoteIndex(4) + MessageCounter(8)
    message := make([]byte, 17)
    
    // Version 1
    message[0] = 1
    
    // Type 1 = Handshake
    message[1] = 1
    
    // Subtype 0 = HandshakeIXPSK0 (initial handshake)
    message[2] = 0
    
    // Reserved (2 bytes) - zeros
    
    // RemoteIndex (4 bytes) - can be 0 for initial
    binary.BigEndian.PutUint32(message[5:9], 0)
    
    // MessageCounter (8 bytes) - can be 1
    binary.BigEndian.PutUint64(message[9:17], 1)
    
    // Send packet
    _, err = conn.Write(message)
    if err != nil {
		return err
    }
    
    // Set read timeout
    conn.SetReadDeadline(time.Now().Add(5 * time.Second))
    
    // Try to read response
    buffer := make([]byte, 1024)
    _, _, err = conn.ReadFromUDP(buffer)
    
    if err != nil {
		return err
    } else {
		return nil
    }
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

func UpdateFirewallBlockedClients() error {
	nebulaYmlPath := utils.CONFIGFOLDER + "nebula-temp.yml"

	// Read the existing nebula-temp.yml file
	yamlData, err := ioutil.ReadFile(nebulaYmlPath)
	if err != nil {
		return fmt.Errorf("failed to read nebula-temp.yml: %w", err)
	}

	// Unmarshal the YAML data into a map
	var configMap map[string]interface{}
	err = yaml.Unmarshal(yamlData, &configMap)
	if err != nil {
		return fmt.Errorf("failed to unmarshal nebula-temp.yml: %w", err)
	}

	// Get the firewall configuration
	firewallMap, ok := configMap["firewall"].(map[interface{}]interface{})
	if !ok {
		return errors.New("firewall not found in nebula-temp.yml")
	}

	// Get all devices from the database to map names to IPs
	c, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "devices")
	if errCo != nil {
		return errCo
	}
	defer closeDb()

	var devices []utils.ConstellationDevice
	cursor, err := c.Find(nil, map[string]interface{}{})
	if err != nil {
		return err
	}
	defer cursor.Close(nil)
	cursor.All(nil, &devices)

	// Create a map of device names to IPs
	deviceIPs := make(map[string]string)
	for _, device := range devices {
		deviceIPs[device.DeviceName] = cleanIp(device.IP)
	}

	// Get the blocked clients list from config
	blockedClients := utils.GetMainConfig().ConstellationConfig.FirewallBlockedClients
	blockedIPs := make(map[string]bool)
	for _, clientName := range blockedClients {
		if ip, exists := deviceIPs[clientName]; exists && ip != "" {
			blockedIPs[ip] = true
			utils.Log("Constellation: Blocking device " + clientName + " (" + ip + ") in firewall")
		}
	}

	// Build new firewall rules
	newInboundRules := []interface{}{}
	newOutboundRules := []interface{}{}

	// Always allow ICMP (ping) from any
	newInboundRules = append(newInboundRules, map[interface{}]interface{}{
		"port":  "any",
		"proto": "icmp",
		"host":  "any",
	})

	// Always allow port 4222 (NATS) from any
	newInboundRules = append(newInboundRules, map[interface{}]interface{}{
		"port":  4222,
		"proto": "any",
		"host":  "any",
	})
	
	// Always allow port 6222 (NATS Cluster) from any
	newInboundRules = append(newInboundRules, map[interface{}]interface{}{
		"port":  6222,
		"proto": "any",
		"host":  "any",
	})
	
	// Always allow port 7422 (NATS Leaf) from any
	newInboundRules = append(newInboundRules, map[interface{}]interface{}{
		"port":  7422,
		"proto": "any",
		"host":  "any",
	})

	// Always allow port 53 (DNS) from any
	newInboundRules = append(newInboundRules, map[interface{}]interface{}{
		"port":  53,
		"proto": "any",
		"host":  "any",
	})

	// Always allow lighthouse (cosmos)
	newInboundRules = append(newInboundRules, map[interface{}]interface{}{
		"port":  "any",
		"proto": "any",
		"host":  "cosmos",
	})

	// Allow outbound to any
	newOutboundRules = append(newOutboundRules, map[interface{}]interface{}{
		"port":  "any",
		"proto": "any",
		"host":  "any",
	})

	for _, device := range devices {
		if device.DeviceName != "" && !blockedIPs[cleanIp(device.IP)] && device.DeviceName != "cosmos" {
			// Allow inbound from this device using its hostname
			newInboundRules = append(newInboundRules, map[interface{}]interface{}{
				"port":  "any",
				"proto": "any",
				"host":  device.DeviceName,
			})
		}
	}

	firewallMap["inbound"] = newInboundRules
	firewallMap["outbound"] = newOutboundRules

	// Marshal back to YAML
	updatedYaml, err := yaml.Marshal(configMap)
	if err != nil {
		return fmt.Errorf("failed to marshal updated config: %w", err)
	}

	// Write back to nebula-temp.yml
	err = ioutil.WriteFile(nebulaYmlPath, updatedYaml, 0644)
	if err != nil {
		return fmt.Errorf("failed to write nebula-temp.yml: %w", err)
	}

	utils.Log("Updated firewall rules in nebula-temp.yml")
	return nil
}

func isIP(str string) bool {
	return net.ParseIP(str) != nil
}

func getDNSRecord(domain string, logger *log.Logger) (string, error) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return "", err
	}

	// debug log all found IPs
	logger.Printf("DEBUG: DNS lookup for %s returned IPs: %v", domain, ips)

	// Prioritize IPv4 over IPv6
	for _, ip := range ips {
		if ip.To4() != nil {
			return ip.String(), nil
		}
	}

	// Fall back to IPv6 if no IPv4 found
	// if len(ips) > 0 {
	// 	return ips[0].String(), nil
	// }

	return "", fmt.Errorf("DNS did not find IP addresses for domain: %s", domain)
}

func AdjustDNS(logBuffer *lumberjack.Logger) error {
	logger := log.New(logBuffer, "", log.LstdFlags)
	nebulaYmlPath := utils.CONFIGFOLDER + "nebula-temp.yml"

	// Read the existing nebula-temp.yml file
	yamlData, err := ioutil.ReadFile(nebulaYmlPath)
	if err != nil {
		logger.Printf("level=error msg=failed to read nebula-temp.yml")
		return fmt.Errorf("failed to read nebula-temp.yml: %w", err)
	}

	// Unmarshal the YAML data into a map
	var configMap map[string]interface{}
	err = yaml.Unmarshal(yamlData, &configMap)
	if err != nil {
		logger.Printf("level=error msg=failed to read nebula-temp.yml")
		return fmt.Errorf("failed to unmarshal nebula-temp.yml: %w", err)
	}

	// Process static_host_map
	if staticHostMap, ok := configMap["static_host_map"].(map[interface{}]interface{}); ok {
		for nebulaIp, destinations := range staticHostMap {
			nebulaIpStr := fmt.Sprintf("%v", nebulaIp)
			newIP := []interface{}{}
			
			// Handle the destinations which could be in different formats
			var destList []interface{}
			switch v := destinations.(type) {
			case []interface{}:
				destList = v
			case []string:
				for _, s := range v {
					destList = append(destList, s)
				}
			default:
				logger.Printf("level=warning msg=Unexpected type for destinations: %T", destinations)
				continue
			}

			for _, destination := range destList {
				destStr := fmt.Sprintf("%v", destination)
				parts := strings.Split(destStr, ":")
				if len(parts) != 2 {
					logger.Printf("level=warning msg=Invalid destination format: %s", destStr)
					continue
				}
				
				originalDestination := parts[0]
				originalPort := parts[1]
				
				// Check if the destination is an IP address
				if !isIP(originalDestination) {
					// get DNS record
					dnsRecord, err := getDNSRecord(originalDestination, logger)
					if err != nil {
						logger.Printf("level=warning msg=Failed to get DNS record for %s: %v. Make sure you create a DNS record for it", originalDestination, err)
					} else {
						newIP = append(newIP, dnsRecord+":"+originalPort)
						logger.Printf("level=info msg=DNS Resolved %s to %s", originalDestination, dnsRecord)
					}
				} else {
					newIP = append(newIP, destStr)
				}
			}
			staticHostMap[nebulaIpStr] = newIP
		}
	}

	// Marshal back to YAML
	updatedYaml, err := yaml.Marshal(configMap)
	if err != nil {
		return fmt.Errorf("failed to marshal updated config: %w", err)
	}

	// Write back to nebula-temp.yml
	err = ioutil.WriteFile(nebulaYmlPath, updatedYaml, 0644)
	if err != nil {
		logger.Printf("level=error msg=failed to write nebula-temp.yml")
		return fmt.Errorf("failed to write nebula-temp.yml: %w", err)
	}

	logger.Printf("level=info msg=Updated DNS entries in nebula-temp.yml")
	return nil
}

func ValidateStaticHosts(logBuffer *lumberjack.Logger) error {
	logger := log.New(logBuffer, "", log.LstdFlags)
	nebulaYmlPath := utils.CONFIGFOLDER + "nebula-temp.yml"

	// Read the existing nebula-temp.yml file
	yamlData, err := ioutil.ReadFile(nebulaYmlPath)
	if err != nil {
		logger.Printf("level=error msg=failed to read nebula-temp.yml")
		return fmt.Errorf("failed to read nebula-temp.yml: %w", err)
	}

	// Unmarshal the YAML data into a map
	var configMap map[string]interface{}
	err = yaml.Unmarshal(yamlData, &configMap)
	if err != nil {
		logger.Printf("level=error msg=failed to read nebula-temp.yml")
		return fmt.Errorf("failed to unmarshal nebula-temp.yml: %w", err)
	}

	// Get static_host_map
	staticHostMap, ok := configMap["static_host_map"].(map[interface{}]interface{})
	if !ok || len(staticHostMap) == 0 {
		logger.Printf("No static hosts to validate")
		return nil
	}

	logger.Printf("Validating %d static host entries...", len(staticHostMap))

	// Track Nebula IPs with no reachable hosts
	completelyUnreachableNebulaIPs := make(map[string]bool)
	updatedHosts := make(map[interface{}][]interface{})
	keptHosts := make(map[string][]string)
	removedHosts := make(map[string][]string)
	totalHostsRemoved := 0

	// Check each static host entry
	for nebulaIP, destinations := range staticHostMap {
		nebulaIPStr := fmt.Sprintf("%v", nebulaIP)

		// Handle the destinations which could be in different formats
		var destList []interface{}
		switch v := destinations.(type) {
		case []interface{}:
			destList = v
		case []string:
			for _, s := range v {
				destList = append(destList, s)
			}
		default:
			logger.Printf("level=warning msg=Unexpected type for destinations: %T", destinations)
			continue
		}

		reachableHosts := []interface{}{}
		unreachableHosts := []string{}

		for _, dest := range destList {
			ipPort := fmt.Sprintf("%v", dest)
			parts := strings.Split(ipPort, ":")
			if len(parts) == 0 {
				continue
			}

			// Try to resolve the hostname/IP
			err := pingNebula(ipPort)
			if err == nil {
				logger.Printf("✓ %s -> %s is resolvable", nebulaIPStr, ipPort)
				reachableHosts = append(reachableHosts, ipPort)
			} else {
				logger.Printf("✗ %s -> %s cannot be resolved", nebulaIPStr, ipPort)
				unreachableHosts = append(unreachableHosts, ipPort)
				totalHostsRemoved++
			}
		}

		// Track changes but don't modify during iteration
		if len(reachableHosts) == 0 {
			// No reachable hosts for this Nebula IP - mark for removal
			logger.Printf("Marking %s as completely unreachable (no valid hosts)", nebulaIPStr)
			completelyUnreachableNebulaIPs[nebulaIPStr] = true
			removedHosts[nebulaIPStr] = unreachableHosts
		} else {
			// Convert reachableHosts to string slice for logging
			reachableStrs := make([]string, len(reachableHosts))
			for i, host := range reachableHosts {
				reachableStrs[i] = fmt.Sprintf("%v", host)
			}
			keptHosts[nebulaIPStr] = reachableStrs

			if len(reachableHosts) < len(destList) {
				// Some hosts unreachable - track update
				updatedHosts[nebulaIP] = reachableHosts
				removedHosts[nebulaIPStr] = unreachableHosts
				logger.Printf("Updated %s: %d/%d hosts reachable", nebulaIPStr, len(reachableHosts), len(destList))
			}
		}
	}

	// Now apply all changes after iteration is complete
	for nebulaIPStr := range completelyUnreachableNebulaIPs {
		// Find the original key in the map (might be string or interface{})
		for key := range staticHostMap {
			if fmt.Sprintf("%v", key) == nebulaIPStr {
				delete(staticHostMap, key)
				break
			}
		}
	}

	for nebulaIP, hosts := range updatedHosts {
		staticHostMap[nebulaIP] = hosts
	}

	if totalHostsRemoved > 0 {
		logger.Printf("Removed %d unreachable host(s) from static_host_map", totalHostsRemoved)
	}

	// Remove completely unreachable Nebula IPs from relay.relays list
	if relayMap, ok := configMap["relay"].(map[interface{}]interface{}); ok {
		if relays, ok := relayMap["relays"].([]interface{}); ok {
			originalCount := len(relays)
			newRelays := []interface{}{}
			for _, ip := range relays {
				ipStr := fmt.Sprintf("%v", ip)
				if !completelyUnreachableNebulaIPs[ipStr] {
					newRelays = append(newRelays, ip)
				}
			}
			relayMap["relays"] = newRelays
			removedCount := originalCount - len(newRelays)
			if removedCount > 0 {
				logger.Printf("level=debug msg=Removed %d completely unreachable IP(s) from relay list", removedCount)
			}
		}
	}

	// Remove completely unreachable Nebula IPs from lighthouse.hosts list
	if lighthouseMap, ok := configMap["lighthouse"].(map[interface{}]interface{}); ok {
		if hosts, ok := lighthouseMap["hosts"].([]interface{}); ok {
			originalCount := len(hosts)
			newHosts := []interface{}{}
			for _, ip := range hosts {
				ipStr := fmt.Sprintf("%v", ip)
				if !completelyUnreachableNebulaIPs[ipStr] {
					newHosts = append(newHosts, ip)
				}
			}
			lighthouseMap["hosts"] = newHosts
			removedCount := originalCount - len(newHosts)
			if removedCount > 0 {
				logger.Printf("level=debug msg=Removed %d completely unreachable IP(s) from relay list", removedCount)
				logger.Printf("Removed %d completely unreachable IP(s) from lighthouse hosts", removedCount)
			}
		}
	}

	remainingNebulaIPs := len(staticHostMap)
	logger.Printf("Static host validation completed: %d Nebula IP(s) with reachable hosts, %d completely unreachable",
		remainingNebulaIPs, len(completelyUnreachableNebulaIPs))

	// Log detailed summary
	if len(keptHosts) > 0 {
		logger.Printf("--- Kept Hosts ---")
		for nebulaIP, hosts := range keptHosts {
			logger.Printf("  %s: %s", nebulaIP, strings.Join(hosts, ", "))
		}
	}

	if len(removedHosts) > 0 {
		logger.Printf("--- Removed Hosts ---")
		for nebulaIP, hosts := range removedHosts {
			logger.Printf("  %s: %s", nebulaIP, strings.Join(hosts, ", "))
		}
	}

	// Marshal back to YAML
	updatedYaml, err := yaml.Marshal(configMap)
	if err != nil {
		logger.Printf("level=error msg=failed to write nebula-temp.yml")
		return fmt.Errorf("failed to marshal updated config: %w", err)
	}

	// Write back to nebula-temp.yml
	err = ioutil.WriteFile(nebulaYmlPath, updatedYaml, 0644)
	if err != nil {
		logger.Printf("level=error msg=failed to write nebula-temp.yml")
		return fmt.Errorf("failed to write nebula-temp.yml: %w", err)
	}

	logger.Printf("Updated nebula-temp.yml after validation")
	return nil
}

func ExportLighthouseFromDB() error {
	currentDevice,err  := GetCurrentDevice()
	if err != nil {
		return err
	}

	currentIP := currentDevice.IP
	
	// Get all lighthouse IPs from the database
	ll, err := GetAllDevicesEvenBlocked()
	if err != nil {
		return err
	}

	if len(ll) == 0 {
		return nil
	}

	// Prepare list of IPs
	var lighthouses []interface{}
	var publicHost map[string][]string = make(map[string][]string)
	var relays []string
	var Blocklist []string
	for _, lh := range ll {
		utils.Debug("[CACA PROUT] Lighthouse from DB: " + lh.IP + " - "  + lh.PublicHostname)
		if lh.Blocked {
			Blocklist = append(Blocklist, lh.Fingerprint)
		} else {
			if lh.IP != ""  && lh.IP != currentIP {
				if(lh.IsLighthouse) {
					lighthouses = append(lighthouses, lh.IP)
				}
				if lh.IsRelay {
					relays = append(relays, lh.IP)
				}
				if lh.PublicHostname != "" {
					for _, hostname := range strings.Split(lh.PublicHostname, ",") {
						hostname = strings.TrimSpace(hostname)
						if hostname != "" {
							if publicHost[lh.IP] == nil {
								publicHost[lh.IP] = make([]string, 0)
							}
							publicHost[lh.IP] = append(publicHost[lh.IP], hostname + ":" + lh.Port)
						}
					}
				}
			}
		}
	}

	nebulaYmlPath := utils.CONFIGFOLDER + "nebula-temp.yml"

	// Read the existing nebula-temp.yml file
	yamlData, err := ioutil.ReadFile(nebulaYmlPath)
	if err != nil {
		return fmt.Errorf("failed to read nebula-temp.yml: %w", err)
	}

	// Unmarshal the YAML data into a map
	var configMap map[string]interface{}
	err = yaml.Unmarshal(yamlData, &configMap)
	if err != nil {
		return fmt.Errorf("failed to unmarshal nebula-temp.yml: %w", err)
	}

	// Update lighthouse.hosts
	lighthouseMap, ok := configMap["lighthouse"].(map[interface{}]interface{})
	if !ok {
		lighthouseMap = make(map[interface{}]interface{})
		configMap["lighthouse"] = lighthouseMap
	}
	lighthouseMap["hosts"] = lighthouses

	// Update relay.relays
	relayMap, ok := configMap["relay"].(map[interface{}]interface{})
	if !ok {
		relayMap = make(map[interface{}]interface{})
		configMap["relay"] = relayMap
	}
	var relayInterfaces []interface{}
	for _, r := range relays {
		relayInterfaces = append(relayInterfaces, r)
	}
	relayMap["relays"] = relayInterfaces

	// set relay am_relay
	relayMap["am_relay"] = currentDevice.IsRelay

	// Update static_host_map with public hostnames
	staticHostMap, ok := configMap["static_host_map"].(map[interface{}]interface{})
	if !ok {
		staticHostMap = make(map[interface{}]interface{})
		configMap["static_host_map"] = staticHostMap
	}
	for ip, hostnames := range publicHost {
		var hostInterfaces []interface{}
		for _, h := range hostnames {
			hostInterfaces = append(hostInterfaces, h)
		}
		staticHostMap[ip] = hostInterfaces
	}

	// Update PKI.Blocklist
	pkiMap, ok := configMap["pki"].(map[interface{}]interface{})
	if !ok {
		pkiMap = make(map[interface{}]interface{})
		configMap["pki"] = pkiMap
	}
	var blocklistInterfaces []interface{}
	for _, b := range Blocklist {
		blocklistInterfaces = append(blocklistInterfaces, b)
	}
	pkiMap["blocklist"] = blocklistInterfaces

	// Marshal back to YAML
	updatedYaml, err := yaml.Marshal(configMap)
	if err != nil {
		return fmt.Errorf("failed to marshal updated config: %w", err)
	}

	// Write back to nebula-temp.yml
	err = ioutil.WriteFile(nebulaYmlPath, updatedYaml, 0644)
	if err != nil {
		return fmt.Errorf("failed to write nebula-temp.yml: %w", err)
	}

	utils.Log("Exported lighthouse IPs to nebula-temp.yml")
	return nil
}