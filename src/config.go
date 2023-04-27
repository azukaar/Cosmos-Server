package main

import (
	"encoding/json"
	"github.com/azukaar/cosmos-server/src/utils"
)

func LoadConfig() utils.Config {
	config := utils.ReadConfigFromFile()

	// check if config is valid
	utils.Log("Validating config file...")
	err := utils.Validate.Struct(config)
	if err != nil {
		utils.Fatal("Reading Config File: " + err.Error(), err)
	}

	utils.LoadBaseMainConfig(config)
	
	configJson, _ := json.MarshalIndent(config, "", "  ")
	utils.Debug("Loaded Configuration " + (string)(configJson))

	return config
}