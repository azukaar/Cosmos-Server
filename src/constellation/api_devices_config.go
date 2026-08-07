package constellation

import (
	"fmt"
	"github.com/azukaar/cosmos-server/src/utils"
)

func compareConfigs(configMap, configMapNew map[string]interface{}) bool {
	configMapStr := fmt.Sprintf("%+v", configMap)
	configMapNewStr := fmt.Sprintf("%+v", configMapNew)

	if string(configMapStr) == string(configMapNewStr) {
		return true
	} else {
		return false
	}
}


func setDefaultConstConfig(configMap map[string]interface{}) map[string]interface{} {
	utils.Debug("setDefaultConstConfig: Setting default values for client nebula.yml")

	if _, exists := configMap["logging"]; !exists {
			configMap["logging"] = map[string]interface{}{
					"format": "text",
					"level":  "info",
			}
	}

	if _, exists := configMap["listen"]; !exists {
			configMap["listen"] = map[string]interface{}{
					"host": "0.0.0.0",
					"port": 4242,
			}
	}

	if _, exists := configMap["punchy"]; !exists {
			configMap["punchy"] = map[string]interface{}{
					"punch":    true,
					"respond":  true,
			}
	}

	if _, exists := configMap["tun"]; !exists {
			configMap["tun"] = map[string]interface{}{
					"dev":                  "nebula1",
					"disabled":             false,
					"drop_local_broadcast": false,
					"drop_multicast":       false,
					"mtu":                  1300,
					"routes":               []interface{}{},
					"tx_queue":             500,
					"unsafe_routes":        []interface{}{},
			}
	}

	if _, exists := configMap["firewall"]; !exists {
			configMap["firewall"] = map[string]interface{}{
					"conntrack": map[string]interface{}{
							"default_timeout": "10m",
							"tcp_timeout":     "12m",
							"udp_timeout":     "3m",
					},
					"inbound": []interface{}{
							map[string]interface{}{
									"host":  "any",
									"port":  "any",
									"proto": "any",
							},
					},
					"inbound_action": "drop",
					"outbound": []interface{}{
							map[string]interface{}{
									"host":  "any",
									"port":  "any",
									"proto": "any",
							},
					},
					"outbound_action": "drop",
			}
	}

	return configMap
}
