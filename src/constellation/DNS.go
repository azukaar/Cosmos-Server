package constellation

import (
	"context"
	"time"
	"sort"
	"strconv"
	"strings"
	"sync"
	"io/ioutil"

	"github.com/miekg/dns"
	"github.com/azukaar/cosmos-server/src/utils"
)

var DNSBlacklist = map[string]bool{}

func externalLookup(client *dns.Client, r *dns.Msg, serverAddr string) (*dns.Msg, time.Duration, error) {
	rCopy := r.Copy() // Create a copy of the request to forward
	rCopy.Id = dns.Id() // Assign a new ID for the forwarded request
	
	// Enable DNSSEC
	rCopy.SetEdns0(4096, true)
	rCopy.CheckingDisabled = false
	rCopy.MsgHdr.AuthenticatedData = true

	return client.Exchange(rCopy, serverAddr)
}

// isAddrQuery reports whether q asks for an address record. AAAA is claimed and
// answered empty rather than forwarded, so a client can't bypass us over IPv6.
func isAddrQuery(q dns.Question) bool {
	return q.Qtype == dns.TypeA || q.Qtype == dns.TypeAAAA
}

// dnsAnswerTTL is how long a Cosmos answer may be cached: the addresses follow
// the nodes, and the first answers after a start come before the cluster view
// is complete. The resolver of a server caches them like any client.
const dnsAnswerTTL = 60

// answerA appends an A record, skipping malformed IPs instead of packing a nil
// RR, and skipping duplicates so overlapping hostnames only answer once
func answerA(m *dns.Msg, name string, ip string) {
	ip = cleanIp(ip)
	if ip == "" {
		utils.Error("DNS: no IP to answer " + name + " with", nil)
		return
	}

	for _, existing := range m.Answer {
		if a, ok := existing.(*dns.A); ok && a.Hdr.Name == name && a.A.String() == ip {
			return
		}
	}

	rr, err := dns.NewRR(name + " A " + ip)
	if err != nil {
		utils.Error("DNS: bad A record for "+name+" -> "+ip, err)
		return
	}

	rr.Header().Ttl = dnsAnswerTTL
	m.Answer = append(m.Answer, rr)
}

func hostOnly(host string) string {
	return strings.Split(host, ":")[0]
}

// tunneledHostnames lists hostnames this node only serves as a tunnel backend
// (nil on a load balancer, which answers for itself).
func tunneledHostnames() map[string]bool {
	if isLB, err := GetCurrentDeviceIsLoadbalancer(); err != nil || isLB {
		return nil
	}

	config := utils.GetMainConfig()
	tunneled := map[string]bool{}
	local := map[string]bool{
		hostOnly(config.HTTPConfig.Hostname): true,
	}

	for _, route := range config.HTTPConfig.ProxyConfig.Routes {
		if !route.UseHost || !usableHost(route.Host) {
			continue
		}
		if route.Tunnel == "" {
			local[hostOnly(route.Host)] = true
			continue
		}
		// overridden exit serves only TunneledHost; the origin still serves Host
		if tunnelHostOverridden(route) {
			local[hostOnly(route.Host)] = true
			tunneled[hostOnly(route.TunneledHost)] = true
			continue
		}
		tunneled[hostOnly(route.Host)] = true
	}

	for host := range local {
		delete(tunneled, host)
	}

	return tunneled
}

// loadBalancerIPs returns the constellation IPs of the known load balancers,
// deduplicated and sorted so the answer is stable across queries.
func loadBalancerIPs() []string {
	devices, _ := deviceCacheSnapshot()

	seen := map[string]bool{}
	ips := []string{}
	for _, device := range devices {
		ip := cleanIp(device.IP)
		if !device.IsLoadBalancer || ip == "" || seen[ip] {
			continue
		}
		seen[ip] = true
		ips = append(ips, ip)
	}

	sort.Strings(ips)
	return ips
}

// dnsAnswerIPs picks what to answer hostname with. A tunneled hostname on a
// non-load-balancer node must point at the load balancers, which hold the
// tunnel cache and provide rotation/failover. Falls back to this node.
func dnsAnswerIPs(hostname string, thisIp string, tunneled map[string]bool) []string {
	if !tunneled[hostname] {
		return []string{thisIp}
	}
	if ips := loadBalancerIPs(); len(ips) > 0 {
		return ips
	}
	return []string{thisIp}
}

func handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	if len(r.Question) == 0 {
		m := new(dns.Msg)
		m.SetRcode(r, dns.RcodeFormatError)
		w.WriteMsg(m)
		return
	}

	config := utils.GetMainConfig()
	DNSFallback := config.ConstellationConfig.DNSFallback

	if DNSFallback == "" {
		DNSFallback = "8.8.8.8:53"
	}

	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true

	customHandled := false

	// []string hostnames
	hostnames := utils.GetAllHostnames(true)

	utils.Debug("DNS Request from " + w.RemoteAddr().String() + " for " + r.Question[0].Name)
	
	if !customHandled {
		customDNSEntries := config.ConstellationConfig.CustomDNSEntries

		// Overwrite local hostnames with custom entries; exact entries win
		// over wildcard ones so "*.example.com" can't shadow "app.example.com"
		for _, q := range r.Question {
			if !isAddrQuery(q) {
				continue
			}
			matched := []utils.ConstellationDNSEntry{}
			for _, entry := range customDNSEntries {
				if matchesCustomEntry(q.Name, entry.Key) {
					matched = append(matched, entry)
				}
			}
			exact := []utils.ConstellationDNSEntry{}
			for _, entry := range matched {
				if !strings.ContainsAny(entry.Key, "*?") {
					exact = append(exact, entry)
				}
			}
			if len(exact) > 0 {
				matched = exact
			}
			for _, entry := range matched {
				if q.Qtype == dns.TypeA {
					utils.Debug("DNS Overwrite " + entry.Key + " with " + entry.Value)
					answerA(m, q.Name, entry.Value)
				}
				customHandled = true
			}
		}
	}

	// Overwrite local hostnames with their Constellation IP. hostnames already
	// includes the tunnel routes this node load-balances, so a server that was
	// selected by DNS always answers for itself and keeps the request local.
	if !customHandled {
		thisIp, err := GetCurrentDeviceIP()
		if err != nil {
			utils.Error("[constellation] Failed to get current device IP for DNS handling", err)
		} else {
			tunneled := tunneledHostnames()
			// an overridden TunneledHost is absent from GetAllHostnames (the exit
			// serves it, not this node) — add it so the origin answers it with the
			// load balancer IPs instead of forwarding the query upstream
			for host := range tunneled {
				hostnames = append(hostnames, host)
			}
			for _, q := range r.Question {
				utils.Debug("DNS Question " + q.Name)
				if !isAddrQuery(q) {
					continue
				}

				// most specific match wins; on a tie local beats cluster
				best := ""
				for _, hostname := range hostnames {
					if matchesDomain(q.Name, hostname) && len(hostname) > len(best) {
						best = hostname
					}
				}
				clusterBest, clusterEntry := clusterDNSLookup(q.Name)

				ips := []string{}
				if best != "" && len(best) >= len(clusterBest) {
					ips = dnsAnswerIPs(best, thisIp, tunneled)
				} else if clusterBest != "" {
					best = clusterBest
					ips = clusterEntry.IPs
				} else {
					continue
				}

				if q.Qtype == dns.TypeA {
					for _, ip := range ips {
						utils.Debug("DNS Overwrite " + best + " with " + ip)
						answerA(m, q.Name, ip)
					}
				}
				customHandled = true
			}
		}
	}
	
	if !customHandled {
		// Overwrite Constellation devices with Constellation IP
		_, cachedNames := deviceCacheSnapshot()
		for _, q := range r.Question {
			utils.Debug("DNS Question " + q.Name)
			qName := deviceQueryName(q.Name)
			for deviceName, ip := range cachedNames {
				procDeviceName := strings.ReplaceAll(deviceName, " ", "-")
				
				if matchesDomain(qName, procDeviceName) && isAddrQuery(q) {
					if q.Qtype == dns.TypeA {
						utils.Debug("DNS Overwrite " + procDeviceName + " with its IP")
						answerA(m, q.Name, ip)
					}
					customHandled = true
				}
			}
		}
	}

	if !customHandled {
		// Block blacklisted domains
		for _, q := range r.Question {
			noDot := strings.ToLower(strings.TrimSuffix(q.Name, "."))
			if isBlacklisted(noDot) {
				if q.Qtype == dns.TypeA {
					utils.Debug("DNS Block " + noDot)
					answerA(m, q.Name, "0.0.0.0")
				}
				
				customHandled = true
			}
		}
	}

	// If not custom handled, use external DNS
	if !customHandled {
		client := new(dns.Client)
		externalResponse, time, err := externalLookup(client, r, DNSFallback)
		if err != nil {
			utils.Error("Failed to forward query:", err)
			m.SetRcode(r, dns.RcodeServerFailure)
			w.WriteMsg(m)
			return
		}
		utils.Debug("DNS Forwarded DNS query to "+DNSFallback+" in " + time.String())
		
		externalResponse.Id = r.Id

		m = externalResponse
	}

	w.WriteMsg(m)
}

func isDomain(domain string) bool {
	// contains . and at least a letter and no special characters invalid in a domain
	if strings.Contains(domain, ".") && strings.ContainsAny(domain, "abcdefghijklmnopqrstuvwxyz") && !strings.ContainsAny(domain, " !@#$%^&*()+=[]{}\\|;:'\",/<>?") {
		return true
	}
	return false
}

func loadRawBlockList(blacklist map[string]bool, DNSBlacklistRaw string) {
	DNSBlacklistArray := strings.Split(string(DNSBlacklistRaw), "\n")
	for _, domain := range DNSBlacklistArray {
		if domain != "" && !strings.HasPrefix(domain, "#") {
			splitDomain := strings.Split(domain, " ")
			if len(splitDomain) == 1 && isDomain(splitDomain[0]) {
				blacklist[splitDomain[0]] = true
			} else if len(splitDomain) == 2 {
				if isDomain(splitDomain[0]) {
					blacklist[splitDomain[0]] = true
				} else if isDomain(splitDomain[1]) {
					blacklist[splitDomain[1]] = true
				}
			}
		}
	}
}

var DNSStarted = false
var dnsServer *dns.Server
var dnsStarting = false
var dnsMux sync.Mutex

// isCurrentDNSServer reports whether s is still the active server (false after StopDNS)
func isCurrentDNSServer(s *dns.Server) bool {
	dnsMux.Lock()
	defer dnsMux.Unlock()
	return dnsServer == s
}

// constellationDNSAddress is the address this node serves the Constellation DNS
// on, empty when it serves none. It follows the config rather than DNSStarted,
// so that a reload of the DNS does not read as a node without DNS.
func constellationDNSAddress() (string, error) {
	config := utils.GetMainConfig().ConstellationConfig
	if !config.Enabled || config.DNSDisabled {
		return "", nil
	}
	device, err := GetCurrentDevice()
	if err != nil {
		return "", err
	}
	// a client node runs no DNS
	if device.CosmosNode == 0 {
		return "", nil
	}
	return device.IP, nil
}

func StopDNS() {
	dnsMux.Lock()
	server := dnsServer
	dnsServer = nil
	DNSStarted = false
	dnsMux.Unlock()

	if server != nil {
		utils.Log("Stopping Constellation DNS")
		revertHostDNS()
		// bounded shutdown so a hung handler can never block stop()/RestartNebula
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.ShutdownContext(ctx); err != nil {
			// includes the benign "server not started" case when stopping mid-bind
			utils.Warn("Failed to stop DNS server: " + err.Error())
		}
	}
}

// ReloadDNS rebinds the DNS server against the current config. Reaction of the
// dns op-log domain, replacing the old restart-everything sync.
func ReloadDNS() {
	StopDNS()
	InitDNS()
}

func InitDNS() {
	dnsMux.Lock()
	if dnsStarting || DNSStarted || dnsServer != nil {
		dnsMux.Unlock()
		return
	}
	// claim before the slow blacklist work so a concurrent InitDNS cannot double-start
	dnsStarting = true
	dnsMux.Unlock()

	utils.Log("Initializing Constellation DNS setup")
	
	
	config := utils.GetMainConfig()
	DNSPort := config.ConstellationConfig.DNSPort
	DNSBlockBlacklist := config.ConstellationConfig.DNSBlockBlacklist

	if DNSPort == "" {
		DNSPort = "53"
	}

	if DNSBlockBlacklist {
		// build into a local map and swap once, so live handlers never see a half-built list
		newBlacklist := map[string]bool{}
		blacklistPath := utils.CONFIGFOLDER + "dns-blacklist.txt"

		utils.Log("Loading DNS blacklist from " + blacklistPath)

		fileExist := utils.FileExists(blacklistPath)
		if fileExist {
			DNSBlacklistRaw, err := ioutil.ReadFile(blacklistPath)
			if err != nil {
				utils.Error("Failed to load DNS blacklist", err)
			} else {
				loadRawBlockList(newBlacklist, string(DNSBlacklistRaw))
			}
		} else {
			utils.Log("No DNS blacklist found")
		}

		// download additional blocklists from config.DNSAdditionalBlocklists []string
		for _, url := range config.ConstellationConfig.DNSAdditionalBlocklists {
			utils.Log("Downloading DNS blacklist from " + url)
			DNSBlacklistRaw, err := utils.DownloadFile(url)
			if err != nil {
				utils.Error("Failed to download DNS blacklist", err)
			} else {
				loadRawBlockList(newBlacklist, DNSBlacklistRaw)
			}
		}

		DNSBlacklist = newBlacklist

		utils.Log("Loaded " + strconv.Itoa(len(DNSBlacklist)) + " domains")
	}

	if config.ConstellationConfig.DNSDisabled {
		dnsMux.Lock()
		dnsStarting = false
		dnsMux.Unlock()
		return
	}

	utils.Log("Initializing Constellation DNS")

	go (func() {
		currIp, err := GetCurrentDeviceIP()
		if err != nil {
			utils.Error("Constellation DNS: Failed to get current device IP", err)
			dnsMux.Lock()
			dnsStarting = false
			dnsMux.Unlock()
			return
		}

		dns.HandleFunc(".", handleDNSRequest)
		server := &dns.Server{Addr: currIp + ":" + DNSPort, Net: "udp"}

		// only report started once the socket is actually bound
		server.NotifyStartedFunc = func() {
			dnsMux.Lock()
			// a StopDNS raced the bind: shut this instance down instead of running as a zombie
			if dnsServer != server {
				dnsMux.Unlock()
				go server.Shutdown()
				return
			}
			DNSStarted = true
			dnsMux.Unlock()
			utils.Log("Constellation DNS started!")
			go applyHostDNS(server, currIp)
		}

		dnsMux.Lock()
		dnsServer = server
		dnsStarting = false
		dnsMux.Unlock()

		utils.Log("Starting DNS server on :" + DNSPort)

		err = server.ListenAndServe();
		retries := 0

		for err != nil && retries < 4 && isCurrentDNSServer(server) {
			time.Sleep(time.Duration(2 * (retries + 1)) * time.Second)
			err = server.ListenAndServe();
			retries++
			utils.Debug("Retrying to start DNS server")
		}

		if err != nil && isCurrentDNSServer(server) {
			utils.MajorError("Failed to start DNS server", err)
		}

		dnsMux.Lock()
		died := dnsServer == server
		if died {
			dnsServer = nil
			DNSStarted = false
		}
		dnsMux.Unlock()
		if died {
			// the server went away on its own: do not leave this server resolving through it
			revertHostDNS()
		}
	})()
}
