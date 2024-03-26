package main

import (
	"io/ioutil"
	"net/http"
	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/azukaar/cosmos-server/src/storage"
	"github.com/azukaar/cosmos-server/src/docker"
	"github.com/azukaar/cosmos-server/src/proxy"
	"os"
	"path/filepath"
	"encoding/json"

	"github.com/jasonlvhit/gocron"
)

type Version struct {
	Version string `json:"version"`
}

func GetCosmosVersion() string {
	ex, err := os.Executable()
	if err != nil {
			panic(err)
	}
	exPath := filepath.Dir(ex)

	pjs, errPR := os.Open(exPath + "/meta.json")
	if errPR != nil {
		utils.Error("checkVersion", errPR)
		return ""
	}

	packageJson, _ := ioutil.ReadAll(pjs)

	utils.Debug("checkVersion" + string(packageJson))

	var version Version
	errJ := json.Unmarshal(packageJson, &version)
	if errJ != nil {
		utils.Error("checkVersion", errJ)
		return ""
	}

	return version.Version

}

func checkVersion() {
	utils.NewVersionAvailable = false

	myVersion := GetCosmosVersion()
	if myVersion == "" {
		utils.Error("checkVersion - Could not get version", nil)
		return
	}
	
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

func runSnapRAIDSync(snap utils.SnapRAIDConfig) {
	err := storage.RunSnapRAIDSync(snap)
	if err != nil {
		utils.Error("runSnapRAIDSync", err)
	}
}

func runSnapRAIDScrub(snap utils.SnapRAIDConfig) {
	err := storage.RunSnapRAIDScrub(snap)
	if err != nil {
		utils.Error("runSnapRAIDScrub", err)
	}
}

func CRON() {
	go func() {
		// TODO: change to new CRON executor, wth customizable maintenance schedules
		
		s := gocron.NewScheduler()
		s.Every(2).Hours().Do(func() {
			go RunBackup()
		})
		s.Every(1).Hours().Do(utils.CleanBannedIPs)
		s.Every(1).Hours().Do(proxy.CleanUp)
		s.Every(1).Day().At("2:00").Do(func() {
			checkVersion()
			utils.CleanupByDate("notifications")
			utils.CleanupByDate("events")
			imageCleanUp()
			checkCerts()
			checkUpdatesAvailable()
		})

		s.Start()
	}()
}