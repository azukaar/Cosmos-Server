package constellation

import (
	"context"
	"time"
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

// matchesDomain reports whether qName matches hostname exactly or on a label boundary
func matchesDomain(qName string, hostname string) bool {
	return qName == hostname + "." || strings.HasSuffix(qName, "." + hostname + ".")
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
	hostnames := utils.GetAllHostnames(false, true)

	utils.Debug("DNS Request from " + w.RemoteAddr().String() + " for " + r.Question[0].Name)
	
	if !customHandled {
		customDNSEntries := config.ConstellationConfig.CustomDNSEntries

		// Overwrite local hostnames with custom entries
		for _, q := range r.Question {
			for _, entry := range customDNSEntries {
				hostname := entry.Key
				ip := entry.Value

				if matchesDomain(q.Name, hostname) && q.Qtype == dns.TypeA {
					utils.Debug("DNS Overwrite " + hostname + " with " + ip)
					rr, _ := dns.NewRR(q.Name + " A " + ip)
					m.Answer = append(m.Answer, rr)
					customHandled = true
				}
			}
		}
	}

	// Overwrite local hostnames with their Constellation IP. Prioritize local URLs over tunnels
	if !customHandled {
		thisIp, err := GetCurrentDeviceIP()
		if err != nil {
			utils.Error("[constellation] Failed to get current device IP for DNS handling", err)
		} else {
			for _, q := range r.Question {
				utils.Debug("DNS Question " + q.Name)
				for _, hostname := range hostnames {
					if matchesDomain(q.Name, hostname) && q.Qtype == dns.TypeA {
						utils.Debug("DNS Overwrite " + hostname + " with " + thisIp)
						rr, _ := dns.NewRR(q.Name + " A " + thisIp)
						m.Answer = append(m.Answer, rr)
						customHandled = true
					}
				}
			}
		}

		remoteTunnels := GetLocalTunnelCache()
		currentName, err := GetCurrentDeviceName()
		if err != nil {
			utils.Error("[constellation] Failed to get current device name for DNS handling", err)
		} else {
			for _, q := range r.Question {
				for _, tunnel := range remoteTunnels {
					for _, target := range tunnel.Targets {
						if target.DeviceName == currentName {
							continue
						}
						destination := CachedDeviceNames[target.DeviceName]
						if destination != "" {
							if matchesDomain(q.Name, tunnel.Route.Host) && q.Qtype == dns.TypeA {
								utils.Debug("DNS Overwrite " + tunnel.Route.Host + " with " + destination)
								rr, _ := dns.NewRR(q.Name + " A " + destination)
								m.Answer = append(m.Answer, rr)
								customHandled = true
							}
						}
					}
				}
			}
		}
	}
	
	if !customHandled {
		// Overwrite Constellation devices with Constellation IP
		for _, q := range r.Question {
			utils.Debug("DNS Question " + q.Name)
			for deviceName, ip := range CachedDeviceNames {
				procDeviceName := strings.ReplaceAll(deviceName, " ", "-")
				
				if matchesDomain(q.Name, procDeviceName) && q.Qtype == dns.TypeA {
					utils.Debug("DNS Overwrite " + procDeviceName + " with its IP")
					rr, _ := dns.NewRR(q.Name + " A " + ip)
					m.Answer = append(m.Answer, rr)
					customHandled = true
				}
			}
		}
	}

	if !customHandled {
		// Block blacklisted domains
		for _, q := range r.Question {
			noDot := strings.TrimSuffix(q.Name, ".")
			if DNSBlacklist[noDot] {
				if q.Qtype == dns.TypeA {
					utils.Debug("DNS Block " + noDot)
					rr, _ := dns.NewRR(q.Name + " A 0.0.0.0")
					m.Answer = append(m.Answer, rr)
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

func StopDNS() {
	dnsMux.Lock()
	server := dnsServer
	dnsServer = nil
	DNSStarted = false
	dnsMux.Unlock()

	if server != nil {
		utils.Log("Stopping Constellation DNS")
		// bounded shutdown so a hung handler can never block stop()/RestartNebula
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.ShutdownContext(ctx); err != nil {
			// includes the benign "server not started" case when stopping mid-bind
			utils.Warn("Failed to stop DNS server: " + err.Error())
		}
	}
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
		if dnsServer == server {
			dnsServer = nil
			DNSStarted = false
		}
		dnsMux.Unlock()
	})()
}
