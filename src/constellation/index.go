package constellation

import (
	"github.com/azukaar/cosmos-server/src/utils" 
	"os"
)

func Init() {
	// if Constellation is enabled
	if utils.GetMainConfig().ConstellationConfig.Enabled {
		InitConfig()
		
		utils.Log("Initializing Constellation module...")

		// check if ca.crt exists
		if _, err := os.Stat(utils.CONFIGFOLDER + "ca.crt"); os.IsNotExist(err) {
			utils.Log("Constellation: ca.crt not found, generating...")
			// generate ca.crt
			generateNebulaCACert("Cosmos - " + utils.GetMainConfig().HTTPConfig.Hostname)
		}

		// check if cosmos.crt exists
		if _, err := os.Stat(utils.CONFIGFOLDER + "cosmos.crt"); os.IsNotExist(err) {
			utils.Log("Constellation: cosmos.crt not found, generating...")
			// generate cosmos.crt
			generateNebulaCert("cosmos", "192.168.201.1/24", "", true)
		}

		// export nebula.yml
		utils.Log("Constellation: exporting nebula.yml...")
		err := ExportConfigToYAML(utils.GetMainConfig().ConstellationConfig, utils.CONFIGFOLDER + "nebula.yml")

		if err != nil {
			utils.Error("Constellation: error while exporting nebula.yml", err)
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