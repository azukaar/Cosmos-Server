package constellation

import (
	"github.com/azukaar/cosmos-server/src/utils" 
	"os"
	"time"
	"strings"
)

var NebulaStarted = false
var CachedDeviceNames = map[string]string{}

func Init() {
	ConstellationInitLock.Lock()
	defer ConstellationInitLock.Unlock()
	
	NebulaStarted = false

	var err error

	// if date is > 1st of January 2024
	timeNow := time.Now()
	if  timeNow.Year() > 2024 || (timeNow.Year() == 2024 && timeNow.Month() > 9) {
		utils.Error("Constellation: this preview version has expired, please update to use the lastest version of Constellation.", nil)
		// disable constellation
		configFile := utils.ReadConfigFromFile()
		configFile.ConstellationConfig.Enabled = false
		utils.SetBaseMainConfig(configFile)
		return
	}
	
	// if Constellation is enabled
	if utils.GetMainConfig().ConstellationConfig.Enabled {
		if !utils.GetMainConfig().ConstellationConfig.SlaveMode {
			InitConfig()
			
			utils.Log("Initializing Constellation module...")

			// if no hostname yet, set default one
			if utils.GetMainConfig().ConstellationConfig.ConstellationHostname == "" {
				utils.Log("Constellation: no hostname found, setting default one...")
				hostnames, _ := utils.ListIps(true)
				httpHostname := utils.GetMainConfig().HTTPConfig.Hostname
				if(utils.IsDomain(httpHostname)) {
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

			// check if ca.crt exists
			if _, err = os.Stat(utils.CONFIGFOLDER + "ca.crt"); os.IsNotExist(err) {
				utils.Log("Constellation: ca.crt not found, generating...")
				// generate ca.crt
				
				errG := generateNebulaCACert("Cosmos - " + utils.GetMainConfig().ConstellationConfig.ConstellationHostname)
				if errG != nil {
					utils.Error("Constellation: error while generating ca.crt", errG)
				}
			}

			// check if cosmos.crt exists
			if _, err := os.Stat(utils.CONFIGFOLDER + "cosmos.crt"); os.IsNotExist(err) {
				utils.Log("Constellation: cosmos.crt not found, generating...")
				// generate cosmos.crt
				_,_,_,errG := generateNebulaCert("cosmos", "192.168.201.1/24", "", true)
				if errG != nil {
					utils.Error("Constellation: error while generating cosmos.crt", errG)
				}
			}

			// export nebula.yml
			utils.Log("Constellation: exporting nebula.yml...")
			err := ExportConfigToYAML(utils.GetMainConfig().ConstellationConfig, utils.CONFIGFOLDER + "nebula.yml")

			if err != nil {
				utils.Error("Constellation: error while exporting nebula.yml", err)
			}

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
							CachedDeviceNames[device.DeviceName] = device.IP
							utils.Debug("Constellation: device name cached: " + device.DeviceName + " -> " + device.IP)
						}
	
						utils.Log("Constellation: device names cache populated")
					}
				}
			}
		} else {
			SlaveConfigSync()
		}
		
		// start nebula
		utils.Log("Constellation: starting nebula...")
		err = startNebulaInBackground()
		if err != nil {
			utils.Error("Constellation: error while starting nebula", err)
		}

		utils.Log("Constellation module initialized")
	}
}