package constellation

import (
	"github.com/azukaar/cosmos-server/src/utils" 
	"strings"
)

var NebulaStarted = false
var CachedDeviceNames = map[string]string{}
var CachedDevices = map[string]utils.ConstellationDevice{}
var needToSyncCA = false

func resyncConstellationNodes() {
	SendNewDBSyncMessage()
}

func GetDefaultHostnames() []string {
	hostnames, _ := utils.ListIps(true)
	httpHostname := utils.GetMainConfig().HTTPConfig.Hostname
	// Strip port if present
	if colonIndex := strings.LastIndex(httpHostname, ":"); colonIndex != -1 {
		httpHostname = httpHostname[:colonIndex]
	}
	if(utils.IsDomain(httpHostname) && !utils.IsLocalDomain(httpHostname)) {
		hostnames = append(hostnames, "vpn." + httpHostname)
	} else if httpHostname != "127.0.0.1" && httpHostname != "localhost" {
		hostnames = append(hostnames, httpHostname)
	}
	return hostnames
}

func InitHostname() {
	// if no hostname yet, set default one
	if utils.GetMainConfig().ConstellationConfig.ConstellationHostname == "" {
		utils.Log("Constellation: no hostname found, setting default one...")
		hostnames := GetDefaultHostnames()
		configFile := utils.ReadConfigFromFile()
		configFile.ConstellationConfig.ConstellationHostname = strings.Join(hostnames, ", ")
		utils.SetBaseMainConfig(configFile)
	} else {
		utils.Log("Constellation: hostname found: " + utils.GetMainConfig().ConstellationConfig.ConstellationHostname)
	}
}

func ConstellationConnected() bool {
	return utils.GetMainConfig().ConstellationConfig.Enabled && NebulaStarted
}

func IsConstellationIP(ip string) bool {
	if !ConstellationConnected() {
		return false
	}

	// Check if the IP exists in the cached device IPs
	for _, deviceIP := range CachedDeviceNames {
		if deviceIP == ip {
			return true
		}
	}

	return false
}

func Init() {
	utils.Log("Initializing Constellation module...")

	InitConfig()
	InitHostname()

	utils.IsConstellationIP = IsConstellationIP

	utils.ResyncConstellationNodes = resyncConstellationNodes

	NebulaStarted = false

	var err error
	
	// if Constellation is enabled
	if utils.GetMainConfig().ConstellationConfig.Enabled {
		// populate CachedDeviceNames
		utils.Log("Constellation: populating device names cache...")
		c, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "devices")
		defer closeDb()

		if errCo != nil {
			utils.Error("Database Connect", errCo)
		} else {
			cursor, err := c.Find(nil, map[string]interface{}{})
			defer cursor.Close(nil)

			if err != nil {
				utils.Error("DeviceList: Error fetching devices", err)
			} else {
				var devices []utils.ConstellationDevice

				if err = cursor.All(nil, &devices); err != nil {
					utils.Error("DeviceList: Error decoding devices", err)
				} else {
					for _, device := range devices {
						if device.Blocked {
							continue
						}
						CachedDeviceNames[device.DeviceName] = device.IP
						CachedDevices[device.DeviceName] = device
						utils.Debug("Constellation: device name cached: " + device.DeviceName + " -> " + device.IP)

						if device.PublicHostname != "" {
							publicHostnames := strings.Split(device.PublicHostname, ",")
							for _, publicHostname := range publicHostnames {
								CachedDeviceNames[strings.TrimSpace(publicHostname)] = device.IP
								CachedDevices[strings.TrimSpace(publicHostname)] = device
								utils.Debug("Constellation: public hostname cached: " + publicHostname + " -> " + device.IP)
							}
						}
					}

					// If current device is not in cache, populate from nebula.yml
					currentDeviceName, errName := GetCurrentDeviceName()
					if errName == nil && currentDeviceName != "" {
						if _, exists := CachedDevices[currentDeviceName]; !exists {
							utils.Log("Constellation: current device not in cache, populating from config...")
							currentDevice, errDevice := GetCurrentDevice()
							if errDevice == nil {
								CachedDeviceNames[currentDeviceName] = currentDevice.IP
								CachedDevices[currentDeviceName] = currentDevice
								utils.Debug("Constellation: current device cached: " + currentDeviceName + " -> " + currentDevice.IP)
							}
						}
					}

					utils.Log("Constellation: device names cache populated")
				}
			}
		}

		// if !utils.GetMainConfig().ConstellationConfig.SlaveMode {
		// 	if !utils.FBL.LValid {
		// 		utils.MajorError("Constellation: No valid licence found to use Constellation. Disabling.", nil)
		// 		// disable constellation
		// 		configFile := utils.ReadConfigFromFile()
		// 		configFile.ConstellationConfig.Enabled = false
		// 		configFile.AdminConstellationOnly = false
		// 		utils.SetBaseMainConfig(configFile)
		// 		return
		// 	}

		// 	utils.Log("Initializing Constellation module...")
		// }

		// read isExitNode from config, if true, add masquerade to iptable
		populateIPTableMasquerade()

		// start nebula
		utils.Log("Constellation: starting nebula...")
		err = startNebula()
		if err != nil {
			utils.Error("Constellation: error while starting nebula", err)
			return
		}
	
		go InitDNS()
		go StartNATS()
		
		go InitPingLighthouses()

		utils.Log("Constellation module initialized")
	}
}