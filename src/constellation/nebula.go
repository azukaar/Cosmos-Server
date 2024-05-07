package constellation

import (
	"github.com/azukaar/cosmos-server/src/utils" 
	"os/exec"
	"os"
	"fmt"
	"errors"
	"runtime"
	"sync"
	"gopkg.in/yaml.v2"
	"strings"
	"io/ioutil"
	"strconv"
	"encoding/json"
	"io"

	"github.com/natefinch/lumberjack"
)

var logBuffer *lumberjack.Logger

var (
	process    *exec.Cmd
	ProcessMux sync.Mutex
	ConstellationInitLock sync.Mutex
)

func binaryToRun() string {
	if runtime.GOARCH == "arm" || runtime.GOARCH == "arm64" {
		return "./nebula-arm"
	}
	return "./nebula"
}

func startNebulaInBackground() error {
	ProcessMux.Lock()
	defer ProcessMux.Unlock()

	if process != nil {
		return errors.New("nebula is already running")
	}

	logBuffer = &lumberjack.Logger{
		Filename:   utils.CONFIGFOLDER+"nebula.log",
		MaxSize:    1, // megabytes
		MaxBackups: 1,
		MaxAge:     15, //days
		Compress:   false,
	}

	process = exec.Command(binaryToRun(), "-config", utils.CONFIGFOLDER+"nebula.yml")

	// Set up multi-writer for stderr
	process.Stderr = io.MultiWriter(logBuffer, os.Stderr)

	if utils.LoggingLevelLabels[utils.GetMainConfig().LoggingLevel] == utils.DEBUG {
		// Set up multi-writer for stdout if in debug mode
		process.Stdout = io.MultiWriter(logBuffer, os.Stdout)
	} else {
		process.Stdout = io.MultiWriter(logBuffer)
	}

	// Start the process in the background
	if err := process.Start(); err != nil {
		return err
	}

	NebulaStarted = true

	utils.Log(fmt.Sprintf("%s started with PID %d\n", binaryToRun(), process.Process.Pid))
	return nil
}

func stop() error {
	ProcessMux.Lock()
	defer ProcessMux.Unlock()

	if process == nil {
		return nil
	}

	if err := process.Process.Kill(); err != nil {
		return err
	}
	process = nil
	utils.Log("Stopped nebula.")
	return nil
}

func RestartNebula() {
	stop()
	Init()
}

func ResetNebula() error {
	stop()
	utils.Log("Resetting nebula...")
	os.RemoveAll(utils.CONFIGFOLDER + "nebula.yml")
	os.RemoveAll(utils.CONFIGFOLDER + "ca.crt")
	os.RemoveAll(utils.CONFIGFOLDER + "ca.key")
	os.RemoveAll(utils.CONFIGFOLDER + "cosmos.crt")
	os.RemoveAll(utils.CONFIGFOLDER + "cosmos.key")
	// remove everything in db

	c, closeDb, err := utils.GetEmbeddedCollection(utils.GetRootAppId(), "devices")
  defer closeDb()
	if err != nil {
			return err
	}

	_, err = c.DeleteMany(nil, map[string]interface{}{})
	if err != nil {
		return err
	}
	
	config := utils.ReadConfigFromFile()
	config.ConstellationConfig.Enabled = false
	config.ConstellationConfig.SlaveMode = false
	config.ConstellationConfig.DNSDisabled = false

	utils.SetBaseMainConfig(config)

	Init()

	return nil
}

func GetAllLightHouses() ([]utils.ConstellationDevice, error) {
	c, closeDb, err := utils.GetEmbeddedCollection(utils.GetRootAppId(), "devices")
  defer closeDb()
	if err != nil {
		return []utils.ConstellationDevice{}, err
	}

	var devices []utils.ConstellationDevice

	cursor, err := c.Find(nil, map[string]interface{}{
		"IsLighthouse": true,
		"Blocked": false,
	})
	defer cursor.Close(nil)
	cursor.All(nil, &devices)

	if err != nil {
		return []utils.ConstellationDevice{}, err
	}

	return devices, nil
}

func GetBlockedDevices() ([]utils.ConstellationDevice, error) {
	c, closeDb, err := utils.GetEmbeddedCollection(utils.GetRootAppId(), "devices")
  defer closeDb()
	if err != nil {
		return []utils.ConstellationDevice{}, err
	}

	var devices []utils.ConstellationDevice

	cursor, err := c.Find(nil, map[string]interface{}{
		"Blocked": true,
	})
	defer cursor.Close(nil)
	cursor.All(nil, &devices)

	if err != nil {
		return []utils.ConstellationDevice{}, err
	}

	return devices, nil
}

func cleanIp(ip string) string {
	return strings.Split(ip, "/")[0]
}

func ExportConfigToYAML(overwriteConfig utils.ConstellationConfig, outputPath string) error {
	// Combine defaultConfig and overwriteConfig
	finalConfig := NebulaDefaultConfig

	if !overwriteConfig.PrivateNode {
		hostnames := []string{}
		hsraw := strings.Split(utils.GetMainConfig().ConstellationConfig.ConstellationHostname, ",")
		for _, hostname := range hsraw {
			// trim
			hostname = strings.TrimSpace(hostname)
			hostnames = append(hostnames, hostname + ":4242")
		}

		finalConfig.StaticHostMap = map[string][]string{
			"192.168.201.1": hostnames,
		}
	} else {
		finalConfig.StaticHostMap = map[string][]string{}
	}

	// for each lighthouse
	lh, err := GetAllLightHouses()
	if err != nil {
		return err
	}

	for _, l := range lh {
		finalConfig.StaticHostMap[cleanIp(l.IP)] = []string{
			l.PublicHostname + ":" + l.Port,
		}
	}

	// add blocked devices
	blockedDevices, err := GetBlockedDevices()
	if err != nil {
		return err
	}

	for _, d := range blockedDevices {
		finalConfig.PKI.Blocklist = append(finalConfig.PKI.Blocklist, d.Fingerprint)
	}
	
	finalConfig.Lighthouse.AMLighthouse = !overwriteConfig.PrivateNode

	// add other lighthouses 
	finalConfig.Lighthouse.Hosts = []string{}
	for _, l := range lh {
		finalConfig.Lighthouse.Hosts = append(finalConfig.Lighthouse.Hosts, cleanIp(l.IP))
	}

	finalConfig.Relay.AMRelay = overwriteConfig.NebulaConfig.Relay.AMRelay

	finalConfig.Relay.Relays = []string{}
	for _, l := range lh {
		if l.IsRelay {
			finalConfig.Relay.Relays = append(finalConfig.Relay.Relays, cleanIp(l.IP))
		}
	}

	// Marshal the combined config to YAML
	yamlData, err := yaml.Marshal(finalConfig)
	if err != nil {
		return err
	}

	// delete nebula.yml if exists
	if _, err := os.Stat(outputPath); err == nil {
		os.Remove(outputPath)
	}

	// Write YAML data to the specified file
	yamlFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer yamlFile.Close()

	_, err = yamlFile.Write(yamlData)
	if err != nil {
		return err
	}

	return nil
}

func getYAMLClientConfig(name, configPath, capki, cert, key, APIKey string, device utils.ConstellationDevice, lite bool) (string, error) {
	utils.Log("Exporting YAML config for " + name + " with file " + configPath)

	// Read the YAML config file
	yamlData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return "", err
	}
	
	// Unmarshal the YAML data into a map interface
	var configMap map[string]interface{}
	err = yaml.Unmarshal(yamlData, &configMap)
	if err != nil {
		return "", err
	}

	lh, err := GetAllLightHouses()
	if err != nil {
		return "", err
	}

	if staticHostMap, ok := configMap["static_host_map"].(map[interface{}]interface{}); ok {
		if !utils.GetMainConfig().ConstellationConfig.PrivateNode {
			// staticHostMap["192.168.201.1"] = []string{
			// 	utils.GetMainConfig().ConstellationConfig.ConstellationHostname + ":4242",
			// }

			hostnames := []string{}
			hsraw := strings.Split(utils.GetMainConfig().ConstellationConfig.ConstellationHostname, ",")
			for _, hostname := range hsraw {
				// trim
				hostname = strings.TrimSpace(hostname)
				hostnames = append(hostnames, hostname + ":4242")
			}

			staticHostMap["192.168.201.1"] = hostnames
		}
		
		for _, l := range lh {
			staticHostMap[cleanIp(l.IP)] = []string{
				l.PublicHostname + ":" + l.Port,
			}
		}
	} else {
		return "", errors.New("static_host_map not found in nebula.yml")
	}

	// set lightHouse
	if lighthouseMap, ok := configMap["lighthouse"].(map[interface{}]interface{}); ok {
		lighthouseMap["am_lighthouse"] = device.IsLighthouse
		
		lighthouseMap["hosts"] = []string{}
		if !utils.GetMainConfig().ConstellationConfig.PrivateNode {
			lighthouseMap["hosts"] = append(lighthouseMap["hosts"].([]string), "192.168.201.1")
		}

		for _, l := range lh {
			if cleanIp(l.IP) != cleanIp(device.IP) {
				lighthouseMap["hosts"] = append(lighthouseMap["hosts"].([]string), cleanIp(l.IP))
			}
		}
	} else {
		return "", errors.New("lighthouse not found in nebula.yml")
	}

	if pkiMap, ok := configMap["pki"].(map[interface{}]interface{}); ok {
		pkiMap["ca"] = capki
		pkiMap["cert"] = cert
		pkiMap["key"] = key
	} else {
		return "", errors.New("pki not found in nebula.yml")
	}

	if relayMap, ok := configMap["relay"].(map[interface{}]interface{}); ok {
		relayMap["am_relay"] = device.IsRelay && device.IsLighthouse
		relayMap["relays"] = []string{}
		if utils.GetMainConfig().ConstellationConfig.NebulaConfig.Relay.AMRelay {
			relayMap["relays"] = append(relayMap["relays"].([]string), "192.168.201.1")
		}

		for _, l := range lh {
			if l.IsRelay && l.IsLighthouse && cleanIp(l.IP) != cleanIp(device.IP) {
				relayMap["relays"] = append(relayMap["relays"].([]string), cleanIp(l.IP))
			}
		}
	} else {
		return "", errors.New("relay not found in nebula.yml")
	}
	
	if listen, ok := configMap["listen"].(map[interface{}]interface{}); ok {
		if device.Port != "" {
			listen["port"] = device.Port
		} else {
			listen["port"] = "4242"
		}
	} else {
		return "", errors.New("listen not found in nebula.yml")
	}

	configEndpoint := utils.GetServerURL("") + "/api/constellation/config-sync"

	configMap["cstln_device_name"] = name
	configMap["cstln_local_dns_address"] = "192.168.201.1"
	configMap["cstln_public_hostname"] = device.PublicHostname
	configMap["cstln_api_key"] = APIKey
	configMap["cstln_config_endpoint"] = configEndpoint

	// list routes with a tunnel property matching the device name
	routesList := utils.GetMainConfig().HTTPConfig.ProxyConfig.Routes
	tunnels := []utils.ProxyRouteConfig{}

	for _, route := range routesList {
		if route.TunnelVia == name {
			if route.UseHost {
				route.OverwriteHostHeader = route.Host
			}
			port := ""
			protocol := ""

			if strings.Contains(route.Target, "://") {
				protocol = strings.Split(route.Target, "://")[0] + "://"
			}

			if protocol == "http://" || protocol == "https://" {
				if utils.IsHTTPS {
					protocol = "https://"
					port = ":" + utils.GetMainConfig().HTTPConfig.HTTPSPort
				} else {
					protocol = "http://"
					port = ":" + utils.GetMainConfig().HTTPConfig.HTTPPort
				}
			} else if route.UseHost {
				// extract port from target
				if strings.Contains(route.Host, ":") {
					_port := strings.Split(route.Host, ":")[1]
						// if port is a number
						if _, err := strconv.Atoi(_port); err == nil {
							port = ":" + _port
					}
				}
			}
			
			route.AcceptInsecureHTTPSTarget = true

			route.Target = protocol + "192.168.201.1" + port

			tunnels = append(tunnels, route)
		}

	}

	configMap["cstln_tunnels"] = tunnels

	// lighten the config for QR Codes
	// remove tun, firewall, punchy and logging
	if(lite) {
		delete(configMap, "tun")
		delete(configMap, "firewall")
		delete(configMap, "punchy")
		delete(configMap, "logging")
		delete(configMap, "listen")
		delete(configMap, "cstln_tunnels")
	}

	// delete blocked pki
	delete(configMap["pki"].(map[interface{}]interface{}), "blocklist")

	// export configMap as YML
	yamlData, err = yaml.Marshal(configMap)
	if err != nil {
		return "", err
	}

	return string(yamlData), nil
}

func getCApki() (string, error) {
	// read config/ca.crt
	caCrt, err := ioutil.ReadFile(utils.CONFIGFOLDER + "ca.crt")
	if err != nil {
		return "", err
	}

	return string(caCrt), nil
}

func killAllNebulaInstances() error {
	ProcessMux.Lock()
	defer ProcessMux.Unlock()

	cmd := exec.Command("ps", "-e", "-o", "pid,command")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, binaryToRun()) {
			fields := strings.Fields(line)
			if len(fields) > 1 {
				pid := fields[0]
				pidInt, _ := strconv.Atoi(pid)
				process, err := os.FindProcess(pidInt)
				if err != nil {
					return err
				}
				err = process.Kill()
				if err != nil {
					return err
				}
				utils.Log(fmt.Sprintf("Killed Nebula instance with PID %s\n", pid))
			}
		}
	}

	return nil
}

func GetCertFingerprint(certPath string) (string, error) {
	// nebula-cert print -json 
	var cmd *exec.Cmd
	
	cmd = exec.Command(binaryToRun() + "-cert",
		"print",
		"-json",
		"-path", certPath,
	)

	// capture and parse output
	output, err := cmd.CombinedOutput()
	if err != nil {
		utils.Error("Error while printing cert", err)
	}

	var certInfo map[string]interface{}
	err = json.Unmarshal(output, &certInfo)
	if err != nil {
		utils.Error("Error while unmarshalling cert information", err)
		return "", err
	}

	// Extract fingerprint, replace "fingerprint" with the actual key where the fingerprint is stored in the JSON output
	fingerprint, ok := certInfo["fingerprint"].(string)
	if !ok {
		utils.Error("Fingerprint not found or not a string", nil)
		return "", errors.New("fingerprint not found or not a string")
	}

	return fingerprint, nil
}

func generateNebulaCert(name, ip, PK string, saveToFile bool) (string, string, string, error) {
	// Run the nebula-cert command
	var cmd *exec.Cmd
	
	// Read the generated certificate and key files
	certPath := fmt.Sprintf("./%s.crt", name)
	keyPath := fmt.Sprintf("./%s.key", name)

	
	// if the temp exists, delete it
	if _, err := os.Stat(certPath); err == nil {
		os.Remove(certPath)
	}
	if _, err := os.Stat(keyPath); err == nil {
		os.Remove(keyPath)
	}

	if(PK == "") {
		cmd = exec.Command(binaryToRun() + "-cert",
			"sign",
			"-ca-crt", utils.CONFIGFOLDER + "ca.crt",
			"-ca-key", utils.CONFIGFOLDER + "ca.key",
			"-name", name,
			"-ip", ip,
		)
	} else {
		// write PK to temp.cert
		err := ioutil.WriteFile("./temp.key", []byte(PK), 0644)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to write temp.key: %s", err)
		}
		cmd = exec.Command(binaryToRun() + "-cert",
			"sign",
			"-ca-crt", utils.CONFIGFOLDER + "ca.crt",
			"-ca-key", utils.CONFIGFOLDER + "ca.key",
			"-name", name,
			"-ip", ip,
			"-in-pub", "./temp.key",
		)
		// delete temp.key
		defer os.Remove("./temp.key")
	}

	utils.Debug(cmd.String())

	cmd.Stderr = os.Stderr
	
	if utils.LoggingLevelLabels[utils.GetMainConfig().LoggingLevel] == utils.DEBUG {
		cmd.Stdout = os.Stdout
	} else {
		cmd.Stdout = nil
	}
	
	cmd.Run()

	if cmd.ProcessState.ExitCode() != 0 {
		return "", "", "", fmt.Errorf("nebula-cert exited with an error, check the Cosmos logs")
	}


	utils.Debug("Reading certificate from " + certPath)
	utils.Debug("Reading key from " + keyPath)

	fingerprint, err := GetCertFingerprint(certPath)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get certificate fingerprint: %s", err)
	}

	certContent, errCert := ioutil.ReadFile(certPath)
	if errCert != nil {
		return "", "", "", fmt.Errorf("failed to read certificate file: %s", errCert)
	}

	keyContent, errKey := ioutil.ReadFile(keyPath)
	if errKey != nil {
		return "", "", "", fmt.Errorf("failed to read key file: %s", errKey)
	}

	if saveToFile {
		cmd = exec.Command("mv", certPath, utils.CONFIGFOLDER + name + ".crt")
		utils.Debug(cmd.String())
		cmd.Run()
		cmd = exec.Command("mv", keyPath, utils.CONFIGFOLDER + name + ".key")
		utils.Debug(cmd.String())
		cmd.Run()
	} else {
		// Delete the generated certificate and key files
		if err := os.Remove(certPath); err != nil {
			return "", "", "", fmt.Errorf("failed to delete certificate file: %s", err)
		}

		if err := os.Remove(keyPath); err != nil {
			return "", "", "", fmt.Errorf("failed to delete key file: %s", err)
		}
	}

	return string(certContent), string(keyContent), fingerprint, nil
}

func generateNebulaCACert(name string) (error) {
	// if ca.key exists, delete it
	if _, err := os.Stat("./ca.key"); err == nil {
		os.Remove("./ca.key")
	}
	if _, err := os.Stat("./ca.crt"); err == nil {
		os.Remove("./ca.crt")
	}
	
	// Run the nebula-cert command to generate CA certificate and key
	cmd := exec.Command(binaryToRun() + "-cert", "ca", "-name", "\""+name+"\"")

	utils.Debug(cmd.String())

	cmd.Stderr = os.Stderr
	
	if utils.LoggingLevelLabels[utils.GetMainConfig().LoggingLevel] == utils.DEBUG {
		cmd.Stdout = os.Stdout
	} else {
		cmd.Stdout = nil
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("nebula-cert error: %s", err)
	}
	
	// copy to /config/ca.*
	cmd = exec.Command("mv", "./ca.crt", utils.CONFIGFOLDER + "ca.crt")
	cmd.Run()
	cmd = exec.Command("mv", "./ca.key", utils.CONFIGFOLDER + "ca.key")
	cmd.Run()

	return nil
}