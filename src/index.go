package main

import (
	"./utils"
	"time"
	"math/rand"
)

func main() {
	  utils.Log("Starting...")

		rand.Seed(time.Now().UnixNano())

		LoadConfig()

		StartServer()
}