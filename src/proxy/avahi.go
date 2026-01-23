package proxy

import (
	"context"
	"fmt"
	"sync"
	"time"
	"net"
	"strings"
	"strconv"

	"github.com/godbus/dbus/v5"
	"github.com/holoplot/go-avahi"
	"github.com/miekg/dns"

	"github.com/azukaar/cosmos-server/src/utils"
)

const (
	AVAHI_DNS_CLASS_IN   = uint16(0x01)
	AVAHI_DNS_TYPE_CNAME = uint16(0x05)
	MAX_ENTRIES_PER_GROUP = 15
)

type Publisher struct {
	dbusConn        *dbus.Conn
	avahiServer     *avahi.Server
	entryGroups     []*avahi.EntryGroup
	fqdn            string
	rdataField      []byte
	mu              sync.Mutex
	cnames          []string
	published       bool
}

func NewPublisher() (*Publisher, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		utils.Error("[mDNS] failed to connect to system bus", err)
		return nil, err
	}

	server, err := avahi.ServerNew(conn)
	if err != nil {
		utils.Error("[mDNS] failed to create Avahi server", err)
		return nil, err
	}

	avahiFqdn, err := server.GetHostNameFqdn()
	if err != nil {
		utils.Error("[mDNS] failed to get FQDN from Avahi", err)
		return nil, err
	}

	// group, err := server.EntryGroupNew()
	// if err != nil {
	// 	utils.Error("[mDNS] failed to create entry group", err)
	// 	return nil, err
	// }

	fqdn := dns.Fqdn(avahiFqdn)

	rdataField := make([]byte, len(fqdn)+1)
	_, err = dns.PackDomainName(fqdn, rdataField, 0, nil, false)
	if err != nil {
		utils.Error("[mDNS] failed to pack FQDN into RDATA", err)
		return nil, err
	}

	return &Publisher{
		dbusConn:        conn,
		avahiServer:     server,
		entryGroups:     make([]*avahi.EntryGroup, 0),
		fqdn:            fqdn,
		rdataField:      rdataField,
	}, nil
}

func (p *Publisher) Fqdn() string {
	return p.fqdn
}

func (p *Publisher) createNewGroup() (*avahi.EntryGroup, error) {
	group, err := p.avahiServer.EntryGroupNew()
	if err != nil {
		utils.Error("[mDNS] failed to create new entry group", err)
		return nil, err
	}
	p.entryGroups = append(p.entryGroups, group)
	return group, nil
}

func (p *Publisher) UpdateCNAMES(cnames []string, ttl uint32) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// If the CNAMEs haven't changed and we've already published, do nothing
	if p.published && stringSlicesEqual(p.cnames, cnames) {
		return nil
	}

	// Reset all existing groups
	for _, group := range p.entryGroups {
		if err := group.Reset(); err != nil {
			if err2 := group.Free(); err2 != nil {
				utils.Error("[mDNS] failed to reset entry group", err)
				utils.Error("[mDNS] failed to free entry group ", err2)
			}
		}
	}

	p.entryGroups = make([]*avahi.EntryGroup, 0)

	ifaces, err := net.Interfaces()
	if err != nil {
		utils.Error("[mDNS] failed to get network interfaces", err)
		return err
	}

	var currentGroup *avahi.EntryGroup
	entriesInCurrentGroup := 0

	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		if strings.HasPrefix(iface.Name, "docker") || 
		   strings.HasPrefix(iface.Name, "br-") || 
		   strings.HasPrefix(iface.Name, "veth") || 
		   strings.HasPrefix(iface.Name, "virbr") {
			continue
		}

		portStr, _ := strconv.Atoi(utils.GetServerPort())

		for _, cname := range cnames {
			// Create new group if needed
			if currentGroup == nil || entriesInCurrentGroup >= MAX_ENTRIES_PER_GROUP {
				var err error
				currentGroup, err = p.createNewGroup()
				if err != nil {
					return err
				}
				entriesInCurrentGroup = 0
				utils.Log(fmt.Sprintf("[mDNS] Created new entry group, total groups: %d", len(p.entryGroups)))
			}

			utils.Log(fmt.Sprintf("[mDNS] Adding service for: %s on interface %s (group %d, entry %d)", 
				cname, iface.Name, len(p.entryGroups)-1, entriesInCurrentGroup))

			err := currentGroup.AddService(
				int32(iface.Index),
				avahi.ProtoUnspec,
				0,
				cname,
				"_http._tcp",
				"",
				"",
				uint16(portStr),
				nil,
			)
			if err != nil {
				utils.Error(fmt.Sprintf("[mDNS] failed to add service to entry group for interface %s", iface.Name), err)
				continue
			}
			entriesInCurrentGroup++

			err = currentGroup.AddRecord(
				int32(iface.Index),
				avahi.ProtoUnspec,
				0,
				cname,
				AVAHI_DNS_CLASS_IN,
				AVAHI_DNS_TYPE_CNAME,
				ttl,
				p.rdataField,
			)
			if err != nil {
				utils.Error(fmt.Sprintf("[mDNS] failed to add CNAME record to entry group for interface %s", iface.Name), err)
				continue
			}
			entriesInCurrentGroup++

			// Commit group if it's full
			if entriesInCurrentGroup >= MAX_ENTRIES_PER_GROUP {
				if err := currentGroup.Commit(); err != nil {
					utils.Error("[mDNS] failed to commit full entry group", err)
					return err
				}
			}
		}
	}

	// Commit the last group if it has any entries
	if currentGroup != nil && entriesInCurrentGroup > 0 {
		if err := currentGroup.Commit(); err != nil {
			utils.Error("[mDNS] failed to commit final entry group", err)
			return err
		}
	}

	p.cnames = cnames
	p.published = true
	return nil
}


func (p *Publisher) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	for _, group := range p.entryGroups {
		err := group.Reset()
		if err != nil {
			utils.Error("[mDNS] failed to reset entry group during close", err)
		}
	}
	p.entryGroups = nil
	
	if p.avahiServer != nil {
		p.avahiServer.Close()
	}
	if p.dbusConn != nil {
		p.dbusConn.Close()
	}
	p.published = false
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func PublishAllMDNSFromConfig() {
	config := utils.GetMainConfig()
	utils.Log("[mDNS] Publishing routes to mDNS")

	publisherMu.Lock()
	defer publisherMu.Unlock()

	var err error
	if publisher == nil {
		publisher, err = NewPublisher()
		if err != nil {
			utils.Error("[mDNS] failed to start mDNS (*.local domains). Install Avahi to solve this issue.", err)
			return
		}
	}

	routes := utils.GetAllHostnames(false, true)

	localRoutes := []string{}

	if config.NewInstall {
		localRoutes = []string{
			"setup-cosmos.local",
		}
	} else {
		for _, route := range routes {
			if len(route) > 6 && route[len(route)-6:] == ".local" {
				localRoutes = append(localRoutes, route)
			}
		}
	}

	utils.Log("[mDNS] Publishing the following routes to mDNS: " + fmt.Sprint(localRoutes))

	if len(localRoutes) == 0 {
		utils.Log("[mDNS] No .local domains to publish")
		return
	}

	err = publisher.UpdateCNAMES(localRoutes, 60)
	if err != nil {
		utils.Error("[mDNS] failed to update CNAMEs", err)
	}
}

func RestartMDNS() {
	publisherMu.Lock()
	defer publisherMu.Unlock()

	if publisher != nil {
		publisher.Close()
		publisher = nil
	}

	PublishAllMDNSFromConfig()
}

var publisher *Publisher
var publisherMu sync.Mutex

func publishing(ctx context.Context, publisher *Publisher, ttl, interval uint32) error {
	resendDuration := time.Duration(interval) * time.Second
	ticker := time.NewTicker(resendDuration)
	defer ticker.Stop()

	publishFunc := func() error {
		publisherMu.Lock()
		cnamesToPublish := publisher.cnames
		publisherMu.Unlock()

		err := publisher.UpdateCNAMES(cnamesToPublish, ttl)
		if err != nil {
			utils.Error("[mDNS] failed to update CNAMEs", err)
		} else {
			utils.Debug("[mDNS] Successfully published CNAMEs: " + fmt.Sprint(cnamesToPublish))
		}
		return err
	}

	// Publish immediately
	if err := publishFunc(); err != nil {
		return err
	}

	for {
		select {
		case <-ticker.C:
			if err := publishFunc(); err != nil {
				utils.Error("[mDNS] Failed to publish CNAMEs in ticker", err)
				// Continue the loop instead of returning, to keep trying
				continue
			}
		case <-ctx.Done():
			utils.Log("[mDNS] context is done, closing publisher")
			publisher.Close()
			return nil
		}
	}
}
