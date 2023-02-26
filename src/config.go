package main

import (
	"log"
	"os"
	"regexp"
	"./proxy"
	"encoding/json"
)

type Config struct {
	HTTPConfig HTTPConfig
}

var defaultConfig = Config{
	HTTPConfig: HTTPConfig{
		TLSCert: "localcert.crt",
		TLSKey: "localcert.key",
		GenerateMissingTLSCert: true,
		HTTPPort: "80",
		HTTPSPort: "443",
		ProxyConfig: proxy.Config{
			Routes: []proxy.RouteConfig{},
		},
	},
}

func GetConfig() Config {
	configFile := os.Getenv("CONFIG_FILE")
	
	if configFile == "" {
		configFile = "/cosmos.config.json"
	}

	log.Println("Using config file: " + configFile)

	// if file does not exist, create it
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.Println("Config file does not exist. Creating default config file.")
		file, err := os.Create(configFile)
		if err != nil {
			log.Fatal("[ERROR] Creating Default Config File: " + err.Error())
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		err = encoder.Encode(defaultConfig)
		if err != nil {
			log.Fatal("[ERROR] Writing Default Config File: " + err.Error())
		}

		return defaultConfig
	}

	file, err := os.Open(configFile)
	if err != nil {
		log.Fatal("[ERROR] Opening Config File: " + err.Error())
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := Config{}
	err = decoder.Decode(&config)
	// check file is not empty
	if err != nil {
		// check error is not empty 
		if err.Error() == "EOF" {
			log.Fatal("[ERROR] Reading Config File: File is empty.")
		}

		// get error string 
		errString := err.Error()

		// replace string in error
		m1 := regexp.MustCompile(`json: cannot unmarshal ([A-Za-z\.]+) into Go struct field ([A-Za-z\.]+) of type ([A-Za-z\.]+)`)
		errString = m1.ReplaceAllString(errString, "Invalid JSON in config file.\n > Field $2 is wrong.\n > Type is $1 Should be $3")
		log.Fatal("[ERROR] Reading Config File: " + errString)
	}

	return config
}