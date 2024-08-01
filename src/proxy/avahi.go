package proxy

import (

	"context"
	"fmt"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/holoplot/go-avahi"
	"github.com/miekg/dns"

	"github.com/azukaar/cosmos-server/src/utils"
)

/**
*  https://github.com/grishy/go-avahi-cname/tree/main
*  LICENCE: MIT License
**/

const (
	// https://github.com/lathiat/avahi/blob/v0.8/avahi-common/defs.h#L343
	AVAHI_DNS_CLASS_IN = uint16(0x01)
	// https://github.com/lathiat/avahi/blob/v0.8/avahi-common/defs.h#L331
	AVAHI_DNS_TYPE_CNAME = uint16(0x05)
)

type Publisher struct {
	dbusConn        *dbus.Conn
	avahiServer     *avahi.Server
	avahiEntryGroup *avahi.EntryGroup
	fqdn            string
	rdataField      []byte
}

// NewPublisher creates a new service for Publisher.
func NewPublisher() (*Publisher, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		utils.Error("failed to connect to system bus", err)
		return nil, err
	}

	server, err := avahi.ServerNew(conn)
	if err != nil {
		utils.Error("failed to create Avahi server", err)
		return nil, err
	}

	avahiFqdn, err := server.GetHostNameFqdn()
	if err != nil {
		utils.Error("failed to get FQDN from Avahi", err)
		return nil, err
	}

	group, err := server.EntryGroupNew()
	if err != nil {
		utils.Error("failed to create entry group", err)
		return nil, err
	}

	fqdn := dns.Fqdn(avahiFqdn)

	// RDATA: a variable length string of octets that describes the resource. CNAME in our case
	// Plus 1 because it will add a null byte at the end.
	rdataField := make([]byte, len(fqdn)+1)
	_, err = dns.PackDomainName(fqdn, rdataField, 0, nil, false)
	if err != nil {
		utils.Error("failed to pack FQDN into RDATA", err)
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

// Fqdn returns the fully qualified domain name from Avahi.
func (p *Publisher) Fqdn() string {
	return p.fqdn
}

// PublishCNAMES send via Avahi-daemon CNAME records with the provided TTL.
func (p *Publisher) PublishCNAMES(cnames []string, ttl uint32) error {
	// Reset the entry group to remove all records.
	// Because we can't update records without it after the `Commit`.
	if err := p.avahiEntryGroup.Reset(); err != nil {
		utils.Error("failed to reset entry group", err)
		return err
	}

	for _, cname := range cnames {
		err := p.avahiEntryGroup.AddRecord(
			avahi.InterfaceUnspec,
			avahi.ProtoUnspec,
			uint32(0), // From Avahi Python bindings https://gist.github.com/gdamjan/3168336#file-avahi-alias-py-L42
			cname,
			AVAHI_DNS_CLASS_IN,
			AVAHI_DNS_TYPE_CNAME,
			ttl,
			p.rdataField,
		)
		if err != nil {
			utils.Error("failed to add record to entry group", err)
			return err
		}
	}

	if err := p.avahiEntryGroup.Commit(); err != nil {
		utils.Error("failed to commit entry group", err)
		return err
	}

	return nil
}

// Close associated resources.
func (p *Publisher) Close() {
	p.avahiServer.Close() // It also closes the DBus connection and free the entry group
}


var publishingCnames []string
// publishLoop handles the continuous publishing of CNAME records.
func publishing(ctx context.Context, publisher *Publisher, ttl, interval uint32) error {
	resendDuration := time.Duration(interval) * time.Second
	ticker := time.NewTicker(resendDuration)
	defer ticker.Stop()

	// Publish immediately
	// https://github.com/golang/go/issues/17601
	if err := publisher.PublishCNAMES(publishingCnames, ttl); err != nil {
		utils.Error("failed to publish CNAMEs", err)
		return err
	}

	for {
		select {
		case <-ticker.C:
			if err := publisher.PublishCNAMES(publishingCnames, ttl); err != nil {
				utils.Error("failed to publish CNAMEs", err)
				return err
			}
		case <-ctx.Done():
			utils.Log("context is done, closing publisher")
			publisher.Close()
			return nil
		}
	}
}

var publisher *Publisher
func PublishAllMDNSFromConfig() {
	config := utils.GetMainConfig()
	utils.Log("Publishing routes to mDNS")
	publishingCnames = []string{}
	newPub := false
	var err error

	if publisher == nil {
		publisher, err = NewPublisher()
		newPub = true
	}
	
	if err != nil {
		utils.MajorError("failed to start mDNS (*.local domains). Install Avahi to solve this issue.", err)
	} else {
		routes := utils.GetAllHostnames(false, true)
		
		// only keep .local domains
		localRoutes := []string{}

		if config.NewInstall {
			localRoutes = []string{
				"setup-cosmos.local",
			}
		} else {
			for _, route := range routes {
				if route[len(route)-6:] == ".local" {
					localRoutes = append(localRoutes, route)
				}
			}	
		}
		

		utils.Log("Publishing the following routes to mDNS: " + fmt.Sprint(localRoutes))
		
		publishingCnames = localRoutes

		if newPub {
			go publishing(
				context.Background(),
				publisher,
				60,
				5,
			)
		} else {
			err := publisher.PublishCNAMES(publishingCnames, 60)
			if err != nil {
				utils.MajorError("failed to publish CNAMEs", err)
			}
		}
	}
}
