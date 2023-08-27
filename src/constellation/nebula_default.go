package constellation

import (
	"github.com/azukaar/cosmos-server/src/utils" 
)

var NebulaDefaultConfig utils.NebulaConfig

func InitConfig() {
	NebulaDefaultConfig = utils.NebulaConfig {
		PKI: struct {
			CA   string `yaml:"ca"`
			Cert string `yaml:"cert"`
			Key  string `yaml:"key"`
		}{
			CA:   utils.CONFIGFOLDER + "ca.crt",
			Cert: utils.CONFIGFOLDER + "cosmos.crt",
			Key:  utils.CONFIGFOLDER + "cosmos.key",
		},
		StaticHostMap: map[string][]string{
			
		},
		Lighthouse: struct {
			AMLighthouse bool     `yaml:"am_lighthouse"`
			Interval     int      `yaml:"interval"`
			Hosts        []string `yaml:"hosts"`
		}{
			AMLighthouse: true,
			Interval:     60,
			Hosts:        []string{},
		},
		Listen: struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
		}{
			Host: "0.0.0.0",
			Port: 4242,
		},
		Punchy: struct {
			Punch bool `yaml:"punch"`
			Respond bool `yaml:"respond"`
		}{
			Punch: true,
			Respond: true,
		},
		Relay: struct {
			AMRelay   bool `yaml:"am_relay"`
			UseRelays bool `yaml:"use_relays"`
		}{
			AMRelay:   true,
			UseRelays: true,
		},
		TUN: struct {
			Disabled            bool     `yaml:"disabled"`
			Dev                 string   `yaml:"dev"`
			DropLocalBroadcast bool     `yaml:"drop_local_broadcast"`
			DropMulticast       bool     `yaml:"drop_multicast"`
			TxQueue             int      `yaml:"tx_queue"`
			MTU                 int      `yaml:"mtu"`
			Routes              []string `yaml:"routes"`
			UnsafeRoutes        []string `yaml:"unsafe_routes"`
		}{
			Disabled:            false,
			Dev:                 "nebula1",
			DropLocalBroadcast: false,
			DropMulticast:       false,
			TxQueue:             500,
			MTU:                 1300,
			Routes:              nil,
			UnsafeRoutes:        nil,
		},
		Logging: struct {
			Level  string `yaml:"level"`
			Format string `yaml:"format"`
		}{
			Level:  "info",
			Format: "text",
		},
		Firewall: struct {
			OutboundAction string                    `yaml:"outbound_action"`
			InboundAction  string                    `yaml:"inbound_action"`
			Conntrack      utils.NebulaConntrackConfig `yaml:"conntrack"`
			Outbound       []utils.NebulaFirewallRule  `yaml:"outbound"`
			Inbound        []utils.NebulaFirewallRule  `yaml:"inbound"`
		}{
			OutboundAction: "drop",
			InboundAction:  "drop",
			Conntrack: utils.NebulaConntrackConfig{
				TCPTimeout:     "12m",
				UDPTimeout:     "3m",
				DefaultTimeout: "10m",
			},
			Outbound: []utils.NebulaFirewallRule {
				utils.NebulaFirewallRule {
					Host: "any",
					Port: "any",
					Proto: "any",
				},
			},
			Inbound: []utils.NebulaFirewallRule {
				utils.NebulaFirewallRule {
					Host: "any",
					Port: "any",
					Proto: "any",
				},
			},
		},
	}
}