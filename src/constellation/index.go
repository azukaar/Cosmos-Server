package constellation

import (
	"github.com/azukaar/cosmos-server/src/utils" 
	"strings"
)

var NebulaStarted = false
var CachedDeviceNames = map[string]string{}
var CachedDevices = map[string]utils.ConstellationDevice{}

func resyncConstellationNodes() {
	go SendNewDBSyncMessage()
}

func InitHostname() {
	// if no hostname yet, set default one
	if utils.GetMainConfig().ConstellationConfig.ConstellationHostname == "" {
		utils.Log("Constellation: no hostname found, setting default one...")
		hostnames, _ := utils.ListIps(true)
		httpHostname := utils.GetMainConfig().HTTPConfig.Hostname
		if(utils.IsDomain(httpHostname) && !utils.IsLocalDomain(httpHostname)) {
			hostnames = append(hostnames, "vpn." + httpHostname)
		} else if httpHostname != "127.0.0.1" && httpHostname != "localhost" {
			hostnames = append(hostnames, httpHostname)
		}
		configFile := utils.ReadConfigFromFile()
		configFile.ConstellationConfig.ConstellationHostname = strings.Join(hostnames, ", ")
		utils.SetBaseMainConfig(configFile)
	} else {
		utils.Log("Constellation: hostname found: " + utils.GetMainConfig().ConstellationConfig.ConstellationHostname)
	}
}

func Init() {
	InitConfig()

	utils.ResyncConstellationNodes = resyncConstellationNodes
	utils.ConstellationSlaveIPWarning = ""

	ConstellationInitLock.Lock()
	defer ConstellationInitLock.Unlock()
	
	NebulaStarted = false

	var err error

	// Debug step
	utils.GetAllTunnelHostnames()
	
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
								utils.Debug("Constellation: device name cached: " + publicHostname + " -> " + device.IP)
							}
						}
					}

					utils.Log("Constellation: device names cache populated")
				}
			}
		}
		
		if !utils.GetMainConfig().ConstellationConfig.SlaveMode {
			if !utils.FBL.LValid {
				utils.MajorError("Constellation: No valid licence found to use Constellation. Disabling.", nil)
				// disable constellation
				configFile := utils.ReadConfigFromFile()
				configFile.ConstellationConfig.Enabled = false
				configFile.AdminConstellationOnly = false
				utils.SetBaseMainConfig(configFile)
				return
			}

			utils.Log("Initializing Constellation module...")

			// export nebula.yml
			utils.Log("Constellation: exporting nebula.yml...")
			err := ExportConfigToYAML(utils.GetMainConfig().ConstellationConfig, utils.CONFIGFOLDER + "nebula.yml")

			if err != nil {
				utils.Error("Constellation: error while exporting nebula.yml", err)
			}

		}

		// Does not work because of Digital Ocean's floating IP's gateway system
		// if utils.GetMainConfig().ConstellationConfig.SlaveMode {
		// 	// check if the IP
		// 	absoluteConfigPath, _ := filepath.Abs(utils.CONFIGFOLDER + "nebula.yml")
		// 	cip, _ := GetConfigAttribute(absoluteConfigPath, "cstln_public_hostname")
		// 	myIps, _ := utils.ListIps(true)
		// 	contained := utils.StringArrayContains(myIps, cip)

		// 	if cip != "" && !contained {
		// 		utils.ConstellationSlaveIPWarning = "This device's public IP is NOT the IP the server is currently using! Hostname: " + cip + " - But only those IPs are associated with this server: " + strings.Join(myIps, ", ") + ". If you are using a public VPS with a floating IP, that IP has not been setup in your netplan config!"
		// 		utils.MajorError("Constellation: " + utils.ConstellationSlaveIPWarning, nil)
		// 	}
		// }

		// read isExitNode from config, if true, add masquerade to iptable
		populateIPTableMasquerade()

		// start nebula
		utils.Log("Constellation: starting nebula...")
		err = startNebulaInBackground()
		if err != nil {
			utils.Error("Constellation: error while starting nebula", err)
		}
		
		/*if utils.GetMainConfig().ConstellationConfig.SlaveMode {
			go (func() {
				InitNATSClient()

				var err error
				retries := 0
				needRestart := false
				needRestart, err = SlaveConfigSync("")
				for err != nil && retries < 4 {
					time.Sleep(time.Duration(2 * (retries + 1)) * time.Second)
					needRestart, err = SlaveConfigSync("")
					retries++
					utils.Debug("Retrying to sync slave config")
				}
				if err != nil {
					utils.MajorError("Failed to sync slave config", err)
				} else {
					utils.Log("Slave config synced")
					if needRestart {
						utils.Warn("Slave config has changed, restarting Nebula...")
						go (func() {		
							time.Sleep(1 * time.Second)		
							RestartNebula()
							utils.RestartHTTPServer()
						})()
					}
				}
			})()
		} else {*/
			go InitDNS()
			go StartNATS()
		//}
		
		go InitPingLighthouses()

		utils.Log("Constellation module initialized")
	}
}