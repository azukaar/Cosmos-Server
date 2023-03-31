package main

import (
	"github.com/jasonlvhit/gocron"
	"io/ioutil"
	"net/http"
	"github.com/azukaar/cosmos-server/src/utils"
	"os"
	"path/filepath"
	"encoding/json"
)

type Version struct {
	Version string `json:"version"`
}

func checkVersion() {

	ex, err := os.Executable()
	if err != nil {
			panic(err)
	}
	exPath := filepath.Dir(ex)

	pjs, errPR := os.Open(exPath + "/meta.json")
	if errPR != nil {
		utils.Error("checkVersion", errPR)
		return
	}

	packageJson, _ := ioutil.ReadAll(pjs)

	utils.Debug("checkVersion" + string(packageJson))

	var version Version
	errJ := json.Unmarshal(packageJson, &version)
	if errJ != nil {
		utils.Error("checkVersion", errJ)
		return
	}

	myVersion := version.Version
	
	response, err := http.Get("https://comos-technologies.com/versions/" + myVersion)
	if err != nil {
		utils.Error("checkVersion", err)
		return
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		utils.Error("checkVersion", err)
		return
	}

	if string(body) != myVersion {
		utils.Log("New version available: " + string(body))
		// update
	} else {
		utils.Log("No new version available")
	}
}

func CRON() {
	gocron.Every(1).Day().At("00:00").Do(checkVersion)
	<-gocron.Start()
}