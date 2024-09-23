package proxy

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/holoplot/go-avahi"
	"github.com/miekg/dns"

	"github.com/azukaar/cosmos-server/src/utils"
)

const (
	AVAHI_DNS_CLASS_IN   = uint16(0x01)
	AVAHI_DNS_TYPE_CNAME = uint16(0x05)
)

type Publisher struct {
	dbusConn        *dbus.Conn
	avahiServer     *avahi.Server
	avahiEntryGroup *avahi.EntryGroup
	fqdn            string
	rdataField      []byte
	mu              sync.Mutex
	cnames          []string
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

	group, err := server.EntryGroupNew()
	if err != nil {
		utils.Error("[mDNS] failed to create entry group", err)
		return nil, err
	}

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
		avahiEntryGroup: group,
		fqdn:            fqdn,
		rdataField:      rdataField,
	}, nil
}

func (p *Publisher) Fqdn() string {
	return p.fqdn
}

func (p *Publisher) UpdateCNAMES(cnames []string, ttl uint32) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.avahiEntryGroup.Reset(); err != nil {
		utils.Error("[mDNS] failed to reset entry group", err)
		return err
	}

	for _, cname := range cnames {
		err := p.avahiEntryGroup.AddRecord(
			avahi.InterfaceUnspec,
			avahi.ProtoUnspec,
			uint32(0),
			cname,
			AVAHI_DNS_CLASS_IN,
			AVAHI_DNS_TYPE_CNAME,
			ttl,
			p.rdataField,
		)
		if err != nil {
			utils.Error("[mDNS] failed to add record to entry group", err)
			return err
		}
	}

	if err := p.avahiEntryGroup.Commit(); err != nil {
		utils.Error("[mDNS] failed to commit entry group", err)
		return err
	}

	p.cnames = cnames
	return nil
}

func (p *Publisher) Close() {
	p.avahiServer.Close()
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

		go func() {
			err := publishing(context.Background(), publisher, 60, 5)
			if err != nil {
				utils.Error("[mDNS] mDNS publishing loop failed", err)
			}
		}()
	}

	routes := utils.GetAllHostnames(false, true)

	// only keep .local domains
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

	// if empty
	if len(localRoutes) == 0 {
		utils.Log("[mDNS] No .local domains to publish")
		return
	}

	err = publisher.UpdateCNAMES(localRoutes, 60)
	if err != nil {
		utils.Error("[mDNS] failed to update CNAMEs", err)
	}
}