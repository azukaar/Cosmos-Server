package main

import (
	"io/ioutil"
	"net/http"
	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/azukaar/cosmos-server/src/docker"
	"os"
	"path/filepath"
	"encoding/json"

	"github.com/jasonlvhit/gocron"
)

type Version struct {
	Version string `json:"version"`
}

func checkVersion() {
	utils.NewVersionAvailable = false

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
	
	response, err := http.Get("https://cosmos-cloud.io/versions/" + myVersion)
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

	cp, errc := utils.CompareSemver(myVersion, string(body))

	if errc != nil {
		utils.Error("checkVersion", errc)
		return
	}

	if cp == -1 {
		utils.Log("New version available: " + string(body))
		utils.NewVersionAvailable = true
	} else {
		utils.Log("No new version available")
	}
}

func checkUpdatesAvailable() {
	utils.UpdateAvailable = docker.CheckUpdatesAvailable()
}

func checkCerts() {
	config := utils.GetMainConfig()
	HTTPConfig := config.HTTPConfig

	if (
		HTTPConfig.HTTPSCertificateMode == utils.HTTPSCertModeList["SELFSIGNED"] ||
		HTTPConfig.HTTPSCertificateMode == utils.HTTPSCertModeList["LETSENCRYPT"]) {
		utils.Log("Checking certificates for renewal")
		if !CertificateIsExpiredSoon(HTTPConfig.TLSValidUntil) {
			utils.Log("Certificates are not valid anymore, renewing")
			RestartServer()
		}
	}
}

func CRON() {
	go func() {
		s := gocron.NewScheduler()
		s.Every(2).Hours().Do(func() {
			go RunBackup()
		})
		s.Every(1).Day().At("00:00").Do(checkVersion)
		s.Every(1).Day().At("01:00").Do(checkCerts)
		s.Every(1).Day().At("02:00").Do(checkUpdatesAvailable)
		s.Every(1).Hours().Do(utils.CleanBannedIPs)
		s.Every(1).Day().At("00:00").Do(func() {
			utils.CleanupByDate("notifications")
			utils.CleanupByDate("events")
			imageCleanUp()
		})
		s.Start()
	}()
}