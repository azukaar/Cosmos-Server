package main

import (
	"os"
	"regexp"
	"encoding/json"
	"./utils"
)

func LoadConfig() utils.Config {
	configFile := utils.GetConfigFileName()
	utils.Log("Using config file: " + configFile)
	if utils.CreateDefaultConfigFileIfNecessary() {
		utils.LoadBaseMainConfig(utils.DefaultConfig)
		return utils.DefaultConfig
	}

	file, err := os.Open(configFile)
	if err != nil {
		utils.Fatal("Opening Config File: ", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := utils.Config{}
	err = decoder.Decode(&config)
	// check file is not empty
	if err != nil {
		// check error is not empty 
		if err.Error() == "EOF" {
			utils.Fatal("Reading Config File: File is empty.", err)
		}

		// get error string 
		errString := err.Error()

		// replace string in error
		m1 := regexp.MustCompile(`json: cannot unmarshal ([A-Za-z\.]+) into Go struct field ([A-Za-z\.]+) of type ([A-Za-z\.]+)`)
		errString = m1.ReplaceAllString(errString, "Invalid JSON in config file.\n > Field $2 is wrong.\n > Type is $1 Should be $3")
		utils.Fatal("Reading Config File: " + errString, err)
	}

	utils.LoadBaseMainConfig(config)

	return config
}