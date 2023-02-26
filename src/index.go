package main

import (
	"log"
)

func main() {
	  log.Println("Starting...")

		config := GetConfig()

		defer StopServer()
		StartServer(config.HTTPConfig)
}