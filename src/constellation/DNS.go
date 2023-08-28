package constellation

import (
	"time"
	"strconv"
	"strings"
	"io/ioutil"
	"fmt"

	"github.com/miekg/dns"
	"github.com/azukaar/cosmos-server/src/utils" 
)

var DNSBlacklist = []string{}

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
	originalHostname := hostnames[0]

	specialQuery := false

	// if lighthouse-cosmos.constellation is the query, return originalHostname's external lookup
	for i, q := range r.Question {
		if strings.HasSuffix(q.Name, "lighthouse-cosmos.constellation.") {
			utils.Debug("DNS Overwrite lighthouse-cosmos.constellation with " + originalHostname)
			
			// Create a deep copy of the original request.
			modifiedRequest := r.Copy()
			
			client := new(dns.Client)
			
			// Modify only the copied request.
			modifiedRequest.Question[i].Name = originalHostname + "."
			
			externalResponse, time, err := externalLookup(client, modifiedRequest, DNSFallback)
			if err != nil {
				utils.Error("Failed to forward query:", err)
				return
			}
			utils.Debug("DNS Forwarded DNS query to "+DNSFallback+" in " + time.String())
			
			for _, rr := range externalResponse.Answer {
				if aRecord, ok := rr.(*dns.A); ok {
						// 2. Replace the hostname with "lighthouse-cosmos.constellation".
						modifiedString := fmt.Sprintf("lighthouse-cosmos.constellation. A %s", aRecord.A.String())
		
						// 3. Convert the string back into a dns.RR.
						newRR, err := dns.NewRR(modifiedString)
						if err != nil {
								utils.Error("Failed to convert string into dns.RR:", err)
								return
						}
		
						// Replace the response RR with the new RR.
						r.Answer = append(r.Answer, newRR)
				}
			}

			m = r
			
			specialQuery = true
		}
	} 
	
	if !specialQuery {
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
		
		if !customHandled {
			// map[string]string customEntries
			customDNSEntries := config.ConstellationConfig.CustomDNSEntries

			// Overwrite local hostnames with custom entries
			for _, q := range r.Question {
				for hostname, ip := range customDNSEntries {
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
			// Block blacklisted domains
			for _, q := range r.Question {
				for _, hostname := range DNSBlacklist {
					if strings.HasSuffix(q.Name, hostname + ".") {
						if q.Qtype == dns.TypeA {
							utils.Debug("DNS Block " + hostname)
							rr, _ := dns.NewRR(q.Name + " A 0.0.0.0")
							m.Answer = append(m.Answer, rr)
						}
						
						customHandled = true
					}
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
	}

	w.WriteMsg(m)
}

func InitDNS() {
	config := utils.GetMainConfig()
	DNSPort := config.ConstellationConfig.DNSPort
	DNSBlockBlacklist := config.ConstellationConfig.DNSBlockBlacklist

	if DNSPort == "" {
		DNSPort = "53"
	}

	if DNSBlockBlacklist {
		DNSBlacklist = []string{}
		blacklistPath := utils.CONFIGFOLDER + "dns-blacklist.txt"

		utils.Log("Loading DNS blacklist from " + blacklistPath)

		fileExist := utils.FileExists(blacklistPath)
		if fileExist {
			DNSBlacklistRaw, err := ioutil.ReadFile(blacklistPath)
			if err != nil {
				utils.Error("Failed to load DNS blacklist", err)
			} else {
				DNSBlacklist = strings.Split(string(DNSBlacklistRaw), "\n")
				utils.Log("Loaded " + strconv.Itoa(len(DNSBlacklist)) + " domains to block")
			}
		} else {
			utils.Log("No DNS blacklist found")
		}
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
