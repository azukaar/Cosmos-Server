package constellation

import (
	"github.com/azukaar/cosmos-server/src/utils" 
	"os"
	"time"
	"strings"
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

var NebulaStarted = false
var CachedDeviceNames = map[string]string{}
var CachedDevices = map[string]utils.ConstellationDevice{}
var DeviceName = ""
var APIKey = ""

func Init() {
	ConstellationInitLock.Lock()
	defer ConstellationInitLock.Unlock()
	
	NebulaStarted = false

	var err error

	// Debug step
	utils.GetAllTunnelHostnames()
	
	// if Constellation is enabled
	if utils.GetMainConfig().ConstellationConfig.Enabled {
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
		}

		// cache device name and api key
		nebulaFile, err := ioutil.ReadFile(utils.CONFIGFOLDER + "nebula.yml")
		if err == nil {
			configMap := make(map[string]interface{})
			err = yaml.Unmarshal(nebulaFile, &configMap)
			if err != nil {
				utils.Error("Constellation: error while unmarshalling nebula.yml", err)
			} else {
				if configMap["cstln_device_name"] == nil || configMap["cstln_api_key"] == nil {
					utils.Warn("Constellation: device name or api key not found in nebula.yml")
					DeviceName = ""
					APIKey = ""
				} else {
					DeviceName = configMap["cstln_device_name"].(string)
					APIKey = configMap["cstln_api_key"].(string)
				}
			}
		} else {
			utils.Error("Constellation: error while reading nebula.yml", err)
			DeviceName = ""
			APIKey = ""
		}

		// start nebula
		utils.Log("Constellation: starting nebula...")
		err = startNebulaInBackground()
		if err != nil {
			utils.Error("Constellation: error while starting nebula", err)
		}
		
		if utils.GetMainConfig().ConstellationConfig.SlaveMode {
			go (func() {
				ConstellationInitLock.Lock()
				defer ConstellationInitLock.Unlock()
				
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
						ConstellationInitLock.Unlock()
						RestartNebula()
						ConstellationInitLock.Lock()
						utils.RestartHTTPServer()
					}
				}
			})()
		} else {
			go InitDNS()
			go StartNATS()
		}

		utils.Log("Constellation module initialized")
	}
}