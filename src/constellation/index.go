package constellation

import (
	"github.com/azukaar/cosmos-server/src/utils" 
	"os"
	"time"
)

func Init() {
	var err error

	// if date is > 1st of January 2024
	timeNow := time.Now()
	if  timeNow.Year() > 2024 || (timeNow.Year() == 2024 && timeNow.Month() > 1) {
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