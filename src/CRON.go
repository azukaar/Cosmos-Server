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
	"github.com/Masterminds/semver"
)

type Version struct {
	Version string `json:"version"`
}

// compareSemver compares two semantic version strings.
// Returns:
//   0 if v1 == v2
//   1 if v1 > v2
//  -1 if v1 < v2
//   error if there's a problem parsing either version string
func compareSemver(v1, v2 string) (int, error) {
	ver1, err := semver.NewVersion(v1)
	if err != nil {
		utils.Error("compareSemver 1 " + v1, err)
		return 0, err
	}

	ver2, err := semver.NewVersion(v2)
	if err != nil {
		utils.Error("compareSemver 2 " + v2, err)
		return 0, err
	}

	return ver1.Compare(ver2), nil
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

	cp, errc := compareSemver(myVersion, string(body))

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
		s.Every(1).Day().At("00:00").Do(checkVersion)
		s.Every(1).Day().At("01:00").Do(checkCerts)
		s.Every(6).Hours().Do(checkUpdatesAvailable)
		s.Every(1).Hours().Do(utils.CleanBannedIPs)
		s.Start()
	}()
}