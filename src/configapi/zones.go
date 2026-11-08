package configapi

import (
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/azukaar/cosmos-server/src/constellation"
	"github.com/azukaar/cosmos-server/src/dnsrecords"
	"github.com/azukaar/cosmos-server/src/utils"
)

// ZoneCertStatus is the certificate of a zone without its key
type ZoneCertStatus struct {
	Hosts      []string  `json:"hosts"`
	ValidUntil time.Time `json:"validUntil"`
	IssuedBy   string    `json:"issuedBy,omitempty"`
}

// ZoneStatus is a zone as served: explicit or derived from a hostname, with
// the hostnames filed under it and the certificate each one is served with.
type ZoneStatus struct {
	utils.DNSZoneConfig
	Derived bool `json:"derived"`
	// IssuesCertificate is true when the zone gets a certificate of its own
	// through its DNS provider; its hostnames use this node's one otherwise
	IssuesCertificate bool                             `json:"issuesCertificate"`
	Hosts             []string                         `json:"hosts"`
	Certificate       *ZoneCertStatus                  `json:"certificate,omitempty"`
	HostCertificates  map[string]utils.HostCertificate `json:"hostCertificates"`
	// RecordsSupported is true when the DNS provider of the zone can manage records
	RecordsSupported bool `json:"recordsSupported"`
	// Records is the DynDNS state of the zone, only known by the node writing
	// the records (this node when standalone, the cluster leader otherwise)
	Records *dnsrecords.ZoneStatus `json:"records,omitempty"`
}

// AdvertisedAddressOption is one choice for HTTPConfig.AdvertisedAddress
type AdvertisedAddressOption struct {
	Value   string `json:"value"`
	Address string `json:"address"`
}

type AdvertisedAddressesResponse struct {
	Status string                    `json:"status"`
	Data   []AdvertisedAddressOption `json:"data"`
}

type ZonesResponse struct {
	Status string       `json:"status"`
	Data   []ZoneStatus `json:"data"`
}

func ZonesRoute(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		ZonesApiList(w, req)
	} else {
		utils.Error("Zones: Method not allowed "+req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "ZN001")
	}
}

func ZonesIdRoute(w http.ResponseWriter, req *http.Request) {
	if req.Method == "PUT" {
		ZonesApiSet(w, req)
	} else if req.Method == "DELETE" {
		ZonesApiDelete(w, req)
	} else {
		utils.Error("Zones: Method not allowed "+req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "ZN001")
	}
}

// ZonesApiList godoc
// @Summary List the DNS zones
// @Description Lists every zone: the explicit ones, and one derived zone per hostname no explicit zone covers. Each zone comes with its hostnames (cluster-wide) and the certificate each hostname is served with. DNS provider credentials are only included with the credentials permission.
// @Tags zones
// @Produce json
// @Security BearerAuth
// @Success 200 {object} ZonesResponse
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Router /api/zones [get]
func ZonesApiList(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_CONFIGURATION_READ) != nil {
		return
	}

	canReadCredentials := utils.HasPermission(req, utils.PERM_CREDENTIALS_READ)
	config := utils.GetMainConfig()
	hosts := append(utils.GetAllHostnames(true), utils.ClusterHostnames()...)

	recordsStatus := dnsrecords.GetStatus()

	zones := []ZoneStatus{}
	for _, resolved := range utils.ResolveZones(config, hosts) {
		zone := ZoneStatus{
			DNSZoneConfig:     resolved.DNSZoneConfig,
			Derived:           resolved.Derived,
			IssuesCertificate: !resolved.Derived && utils.IssuesZoneCert(resolved.DNSZoneConfig),
			Hosts:             resolved.Hosts,
			HostCertificates:  map[string]utils.HostCertificate{},
		}
		if zone.Hosts == nil {
			zone.Hosts = []string{}
		}
		sort.Strings(zone.Hosts)

		if !canReadCredentials {
			zone.DNSChallengeConfig = map[string]string{}
		}
		// the private key of a provided certificate never leaves the server
		zone.TLSKey = ""

		if cert, ok := config.HTTPConfig.ZoneCerts[resolved.Zone]; ok && !resolved.Derived {
			zone.Certificate = &ZoneCertStatus{
				Hosts:      cert.Hosts,
				ValidUntil: cert.ValidUntil,
				IssuedBy:   cert.IssuedBy,
			}
		}

		for _, host := range zone.Hosts {
			zone.HostCertificates[host] = utils.GetHostCertificate(config, host)
		}

		zone.RecordsSupported = dnsrecords.Supported(resolved.DNSChallengeProvider)
		if status, ok := recordsStatus[resolved.Zone]; ok && !resolved.Derived {
			zone.Records = &status
		}

		zones = append(zones, zone)
	}

	json.NewEncoder(w).Encode(ZonesResponse{
		Status: "OK",
		Data:   zones,
	})
}

// ZonesApiSet godoc
// @Summary Create or replace a DNS zone
// @Description Creates the explicit zone, or replaces it. A zone is a domain with its HTTPS setup (certificate mode, DNS challenge provider, wildcard) and its DynDNS setup. Every hostname under the domain uses it, unless a longer zone matches. The certificate mode is LETSENCRYPT, SELFSIGNED, PROVIDED (TLSCert and TLSKey in PEM format) or DISABLED (plain HTTP), and no server-wide mode overrides it. Replicated to the whole Constellation. Credentials sent empty by a client without the credentials permission keep their current value, and so does the private key of a provided certificate, which is never returned.
// @Tags zones
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param zone path string true "Zone name, e.g. example.com"
// @Param request body utils.DNSZoneConfig true "Zone configuration"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 409 {object} utils.HTTPErrorResult
// @Router /api/zones/{zone} [put]
func ZonesApiSet(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_CONFIGURATION) != nil {
		return
	}

	name := strings.ToLower(strings.TrimSpace(mux.Vars(req)["zone"]))

	var zone utils.DNSZoneConfig
	if err := json.NewDecoder(req.Body).Decode(&zone); err != nil {
		utils.Error("ZonesSet: Invalid request", err)
		utils.HTTPError(w, "Invalid request", http.StatusBadRequest, "ZN002")
		return
	}
	zone.Zone = name

	if err := utils.Validate.Struct(zone); err != nil || !strings.Contains(name, ".") || strings.HasPrefix(name, "*") {
		utils.Error("ZonesSet: Invalid zone "+name, err)
		utils.HTTPError(w, "Invalid zone: expected a domain name such as example.com", http.StatusBadRequest, "ZN003")
		return
	}
	if zone.HTTPSCertificateMode == "" {
		zone.HTTPSCertificateMode = utils.HTTPSCertModeList["LETSENCRYPT"]
	}
	if _, known := utils.HTTPSCertModeList[zone.HTTPSCertificateMode]; !known {
		utils.HTTPError(w, "Invalid certificate mode for a zone: LETSENCRYPT, SELFSIGNED, PROVIDED or DISABLED", http.StatusBadRequest, "ZN004")
		return
	}
	// only Let's Encrypt gets anything out of the DNS challenge; DynDNS keeps its provider whatever the mode
	if zone.HTTPSCertificateMode != utils.HTTPSCertModeList["LETSENCRYPT"] {
		zone.UseWildcardCertificate = false
	}
	if zone.HTTPSCertificateMode != utils.HTTPSCertModeList["PROVIDED"] {
		zone.TLSCert = ""
		zone.TLSKey = ""
	}
	if zone.DNSChallengeProvider == "" {
		if zone.UseWildcardCertificate {
			utils.HTTPError(w, "A wildcard certificate needs a DNS provider", http.StatusBadRequest, "ZN005")
			return
		}
		zone.ManageRecords = false
	}
	if zone.ManageRecords && !dnsrecords.Supported(zone.DNSChallengeProvider) {
		utils.HTTPError(w, "Records cannot be managed through this DNS provider. Supported: "+strings.Join(dnsrecords.SupportedProviders(), ", "), http.StatusBadRequest, "ZN008")
		return
	}
	if !zone.ManageRecords {
		zone.WildcardRecord = false
	}

	utils.ConfigLock.Lock()
	zones := append([]utils.DNSZoneConfig{}, utils.ReadConfigFromFile().HTTPConfig.DNSZones...)
	utils.ConfigLock.Unlock()

	replaced := false
	for i, existing := range zones {
		if existing.Zone == name {
			if !utils.HasPermission(req, utils.PERM_CREDENTIALS_READ) && existing.DNSChallengeProvider == zone.DNSChallengeProvider {
				zone.DNSChallengeConfig = existing.DNSChallengeConfig
			}
			// the key is never sent back to a client: an empty one keeps the current certificate
			if zone.HTTPSCertificateMode == utils.HTTPSCertModeList["PROVIDED"] && zone.TLSKey == "" && existing.TLSKey != "" {
				zone.TLSKey = existing.TLSKey
				if zone.TLSCert == "" {
					zone.TLSCert = existing.TLSCert
				}
			}
			zones[i] = zone
			replaced = true
		}
	}
	if !replaced {
		zones = append(zones, zone)
	}
	if zone.HTTPSCertificateMode == utils.HTTPSCertModeList["PROVIDED"] {
		if _, err := tls.X509KeyPair([]byte(zone.TLSCert), []byte(zone.TLSKey)); err != nil {
			utils.Error("ZonesSet: Invalid certificate for "+name, err)
			utils.HTTPError(w, "Invalid certificate: the certificate and its private key are expected in PEM format, and must belong together ("+err.Error()+")", http.StatusBadRequest, "ZN009")
			return
		}
	}
	sort.SliceStable(zones, func(i, j int) bool { return zones[i].Zone < zones[j].Zone })

	if err := constellation.PublishDomainOp(constellation.DomainDNSZones, zones); err != nil {
		utils.HTTPStoreError(w, err, "ZN006")
		return
	}

	utils.TriggerEvent(
		"cosmos.settings",
		"DNS zone saved",
		"success",
		"zone@"+name,
		map[string]interface{}{
			"zone": name,
		},
	)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
	})
}

// ZonesApiDelete godoc
// @Summary Delete a DNS zone
// @Description Deletes the explicit zone. Its hostnames fall back to the next matching zone, or to a derived zone each. Replicated to the whole Constellation.
// @Tags zones
// @Produce json
// @Security BearerAuth
// @Param zone path string true "Zone name"
// @Success 200 {object} utils.APIResponse
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 404 {object} utils.HTTPErrorResult
// @Failure 409 {object} utils.HTTPErrorResult
// @Router /api/zones/{zone} [delete]
func ZonesApiDelete(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_CONFIGURATION) != nil {
		return
	}

	name := strings.ToLower(strings.TrimSpace(mux.Vars(req)["zone"]))

	utils.ConfigLock.Lock()
	current := utils.ReadConfigFromFile().HTTPConfig.DNSZones
	utils.ConfigLock.Unlock()

	zones := []utils.DNSZoneConfig{}
	for _, existing := range current {
		if existing.Zone != name {
			zones = append(zones, existing)
		}
	}
	if len(zones) == len(current) {
		utils.HTTPError(w, "Zone not found", http.StatusNotFound, "ZN007")
		return
	}

	if err := constellation.PublishDomainOp(constellation.DomainDNSZones, zones); err != nil {
		utils.HTTPStoreError(w, err, "ZN006")
		return
	}

	utils.TriggerEvent(
		"cosmos.settings",
		"DNS zone deleted",
		"success",
		"zone@"+name,
		map[string]interface{}{
			"zone": name,
		},
	)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
	})
}

// ZonesApiAddresses godoc
// @Summary List the addresses this server can advertise
// @Description Lists the choices for HTTPConfig.AdvertisedAddress, the address the managed DNS records of this server point at: the detected public IP (empty value, the default), each network interface (iface:<name>, for a server published on its LAN), with the address each one currently resolves to.
// @Tags zones
// @Produce json
// @Security BearerAuth
// @Success 200 {object} AdvertisedAddressesResponse
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Router /api/zones-addresses [get]
func ZonesApiAddresses(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_CONFIGURATION_READ) != nil {
		return
	}

	if req.Method != "GET" {
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "ZN001")
		return
	}

	publicIP, _ := dnsrecords.DetectPublicIP()
	options := []AdvertisedAddressOption{{Value: "", Address: publicIP}}

	interfaces, _ := net.Interfaces()
	for _, iface := range interfaces {
		name := iface.Name
		if strings.HasPrefix(name, "nebula") || strings.HasPrefix(name, "docker") || strings.HasPrefix(name, "br-") || strings.HasPrefix(name, "veth") || strings.HasPrefix(name, "virbr") || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if address, err := dnsrecords.InterfaceAddress(name); err == nil {
			options = append(options, AdvertisedAddressOption{Value: "iface:" + name, Address: address})
		}
	}

	json.NewEncoder(w).Encode(AdvertisedAddressesResponse{
		Status: "OK",
		Data:   options,
	})
}
