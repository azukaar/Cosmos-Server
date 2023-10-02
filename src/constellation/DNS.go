package constellation

import (
	"time"
	"strconv"
	"strings"
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

func handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
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
	
	if !customHandled {
		customDNSEntries := config.ConstellationConfig.CustomDNSEntries

		// Overwrite local hostnames with custom entries
		for _, q := range r.Question {
			for _, entry := range customDNSEntries {
				hostname := entry.Key
				ip := entry.Value

				if strings.HasSuffix(q.Name, hostname + ".") && q.Qtype == dns.TypeA {
					utils.Debug("DNS Overwrite " + hostname + " with " + ip)
					rr, _ := dns.NewRR(q.Name + " A " + ip)
					m.Answer = append(m.Answer, rr)
					customHandled = true
				}
			}
		}
	}

	if !customHandled {
		// Overwrite local hostnames with Constellation IP
		for _, q := range r.Question {
			utils.Debug("DNS Question " + q.Name)
			for _, hostname := range hostnames {
				if strings.HasSuffix(q.Name, hostname + ".") && q.Qtype == dns.TypeA {
					utils.Debug("DNS Overwrite " + hostname + " with 192.168.201.1")
					rr, _ := dns.NewRR(q.Name + " A 192.168.201.1")
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

func loadRawBlockList(DNSBlacklistRaw string) {
	DNSBlacklistArray := strings.Split(string(DNSBlacklistRaw), "\n")
	for _, domain := range DNSBlacklistArray {
		if domain != "" && !strings.HasPrefix(domain, "#") {
			splitDomain := strings.Split(domain, " ")
			if len(splitDomain) == 1 && isDomain(splitDomain[0]) {
				DNSBlacklist[splitDomain[0]] = true
			} else if len(splitDomain) == 2 {
				if isDomain(splitDomain[0]) {
					DNSBlacklist[splitDomain[0]] = true
				} else if isDomain(splitDomain[1]) {
					DNSBlacklist[splitDomain[1]] = true
				}
			}
		}
	}
}

func InitDNS() {
	config := utils.GetMainConfig()
	DNSPort := config.ConstellationConfig.DNSPort
	DNSBlockBlacklist := config.ConstellationConfig.DNSBlockBlacklist

	if DNSPort == "" {
		DNSPort = "53"
	}

	if DNSBlockBlacklist {
		DNSBlacklist = map[string]bool{}
		blacklistPath := utils.CONFIGFOLDER + "dns-blacklist.txt"

		utils.Log("Loading DNS blacklist from " + blacklistPath)

		fileExist := utils.FileExists(blacklistPath)
		if fileExist {
			DNSBlacklistRaw, err := ioutil.ReadFile(blacklistPath)
			if err != nil {
				utils.Error("Failed to load DNS blacklist", err)
			} else {
				loadRawBlockList(string(DNSBlacklistRaw))
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
				loadRawBlockList(DNSBlacklistRaw)
			}
		}
		
		utils.Log("Loaded " + strconv.Itoa(len(DNSBlacklist)) + " domains")
	}

	if(config.ConstellationConfig.DNS) {
		go (func() {
			dns.HandleFunc(".", handleDNSRequest)
			server := &dns.Server{Addr: ":" + DNSPort, Net: "udp"}

			utils.Log("Starting DNS server on :" + DNSPort)
			if err := server.ListenAndServe(); err != nil {
				utils.Fatal("Failed to start server: %s\n", err)
			}
		})()
	}
}
