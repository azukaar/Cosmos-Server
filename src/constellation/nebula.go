package constellation

import (
	"github.com/azukaar/cosmos-server/src/utils"
	"os/exec"
	"os"
	"fmt"
	"errors"
	"runtime"
	"sync"
	"bufio"
	"gopkg.in/yaml.v2"
	"strings"
	"io/ioutil"
	"io"
	"net"
	"encoding/json"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
	"github.com/slackhq/nebula"
	nebulaConfig "github.com/slackhq/nebula/config"
)

var logBuffer *lumberjack.Logger

var (
	control    *nebula.Control
	nebulaLock sync.Mutex
)

func certBinaryToRun() string {
	if runtime.GOARCH == "arm" || runtime.GOARCH == "arm64" {
		return "./nebula-arm-cert"
	}
	return "./nebula-cert"
}

// cosmosLogHook is a logrus hook that routes nebula log output to the Cosmos logging system
type cosmosLogHook struct {
	logBuffer *lumberjack.Logger
}

func (h *cosmosLogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *cosmosLogHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}
	// Route to Cosmos VPN logger (parses level= from the formatted line)
	utils.VPNWithLevel(line)
	// Write to nebula.log file
	if h.logBuffer != nil {
		h.logBuffer.Write([]byte(line))
	}
	return nil
}

func startNebula() error {
	nebulaLock.Lock()
	defer nebulaLock.Unlock()

	utils.Log("Starting nebula sub-process...")

	if control != nil {
		return errors.New("nebula is already running")
	}

	// copy nebula.yml to nebula-temp.yml
	source, err := os.Open(utils.CONFIGFOLDER + "nebula.yml")
	if err != nil {
		return fmt.Errorf("failed to open nebula.yml: %w", err)
	}
	defer source.Close()

	destination, err := os.Create(utils.CONFIGFOLDER + "nebula-temp.yml")
	if err != nil {
		return fmt.Errorf("failed to create nebula-temp.yml: %w", err)
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		return fmt.Errorf("failed to copy nebula config: %w", err)
	}

	// Initialize log buffer
	logBuffer = &lumberjack.Logger{
		Filename:   utils.CONFIGFOLDER + "nebula.log",
		MaxSize:    1, // megabytes
		MaxBackups: 1,
		MaxAge:     15, //days
		Compress:   false,
	}

	UpdateFirewallBlockedClients()
	AdjustDNS(logBuffer)
	ExportLighthouseFromDB()

	// Read the final config
	configData, err := ioutil.ReadFile(utils.CONFIGFOLDER + "nebula-temp.yml")
	if err != nil {
		return fmt.Errorf("failed to read nebula-temp.yml: %w", err)
	}

	// Create logrus logger with Cosmos hook
	l := logrus.New()
	l.SetOutput(ioutil.Discard) // suppress default output, hook handles it
	l.SetLevel(logrus.DebugLevel)
	l.AddHook(&cosmosLogHook{logBuffer: logBuffer})

	// Load nebula config
	c := nebulaConfig.NewC(l)
	err = c.LoadString(string(configData))
	if err != nil {
		return fmt.Errorf("failed to load nebula config: %w", err)
	}

	// Create nebula instance (retry to handle TUN device not yet released by kernel)
	for i := 0; i < 10; i++ {
		control, err = nebula.Main(c, false, "", l, nil)
		if err == nil {
			break
		} else {
			utils.Warn("Failed to create nebula instance, retrying... Error was: " + err.Error())
		}
		if i < 9 {
			time.Sleep(1 * time.Second)
			// Reload config for each retry since NewC is consumed
			c = nebulaConfig.NewC(l)
			c.LoadString(string(configData))
		}
	}
	if err != nil {
		return fmt.Errorf("failed to create nebula instance: %w", err)
	}

	// Actually start the nebula service (brings up TUN interface)
	control.Start()

	NebulaStarted = true

	utils.Log("Constellation: nebula started successfully")
	return nil
}

func stop() {
	// Grab control reference and clear state under lock, then shut down outside the lock
	// to avoid holding nebulaLock during a potentially slow shutdown.
	nebulaLock.Lock()
	ctrl := control
	control = nil
	NebulaStarted = false
	nebulaLock.Unlock()

	if ctrl != nil {
		ctrl.Stop()

		// After ctrl.Stop() closes the socket FDs, nebula's ListenOut goroutine
		// may still be blocked in a raw recvmsg() syscall. The kernel keeps the
		// socket alive (and port bound) as long as recvmsg holds a reference via
		// fdget(). Send a dummy UDP packet to wake up the blocked goroutine so it
		// can detect the closed FD and exit, releasing the port.
		wakeConn, err := net.Dial("udp4", "127.0.0.1:4242")
		if err == nil {
			wakeConn.Write([]byte{0})
			wakeConn.Close()
		}

		// Wait for the TUN device to be released by the kernel
		for i := 0; i < 50; i++ {
			if _, err := net.InterfaceByName("nebula1"); err != nil {                                                                  
                utils.Log("Nebula TUN device released")                 
				break // interface is gone
			}
			time.Sleep(500 * time.Millisecond)
		}
		utils.Log("Stopped nebula.")
	}

	cachedCurrentDevice = nil
	CachedDevices = map[string]utils.ConstellationDevice{}
	CachedDeviceNames = map[string]string{}
}

var restartMutex sync.Mutex

func RestartNebula() {
	restartMutex.Lock()
	defer restartMutex.Unlock()

	utils.Log("Restarting Constellation...")
	cachedCurrentDevice = nil
	StopNATS()
	CloseNATSClient()
	stop()
	utils.Log("Constellation Init...")
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
	os.RemoveAll(utils.CONFIGFOLDER + "jetstream")

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
	config.ConstellationConfig.DNSDisabled = false
	config.ConstellationConfig.FirewallBlockedClients = []string{}
	config.ConstellationConfig.ThisDeviceName = ""
	config.ConstellationConfig.ConstellationHostname = strings.Join(GetDefaultHostnames(), ",")

	if config.Licence == "" {
		config.ServerToken = ""
	}

	utils.SetBaseMainConfig(config)

	cachedCurrentDevice = nil
	CachedDevices = map[string]utils.ConstellationDevice{}
	CachedDeviceNames = map[string]string{}

	utils.SoftRestartServer()

	time.Sleep(2 * time.Second)

	return nil
}

func GetAllDevicesEvenBlocked() ([]utils.ConstellationDevice, error) {
	c, closeDb, err := utils.GetEmbeddedCollection(utils.GetRootAppId(), "devices")
    defer closeDb()
	if err != nil {
		return []utils.ConstellationDevice{}, err
	}

	var devices []utils.ConstellationDevice

	cursor, err := c.Find(nil, map[string]interface{}{})
	defer cursor.Close(nil)
	cursor.All(nil, &devices)

	if err != nil {
		return []utils.ConstellationDevice{}, err
	}

	return devices, nil
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

func cleanIp(ip string) string {
	return strings.Split(ip, "/")[0]
}

func ExportDefaultConfigToYAML(outputPath string) error {
	// Combine defaultConfig
	finalConfig := NebulaDefaultConfig

	hostnames := []string{}
	hsraw := strings.Split(utils.GetMainConfig().ConstellationConfig.ConstellationHostname, ",")
	for _, hostname := range hsraw {
		// trim
		hostname = strings.TrimSpace(hostname)
		if hostname != "" {
			hostnames = append(hostnames, hostname + ":4242")
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

func getYAMLClientConfig(name, configPath, capki, cert, key, APIKey string, device utils.ConstellationDevice, lite bool, getLicence bool) (string, error) {
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
		hostnames := []string{}
		hs, _ := GetCurrentDeviceHostname()
		hsraw := strings.Split(hs, ",")
		for _, hostname := range hsraw {
			// trim
			hostname = strings.TrimSpace(hostname)
			if hostname != "" {
				hostnames = append(hostnames, hostname + ":4242")
			}
		}

		for _, l := range lh {
			staticHostMap[cleanIp(l.IP)] = []string{}

			for _, hostname := range strings.Split(l.PublicHostname, ",") {
				if device.IP == l.IP {
					continue
				}
				hostname = strings.TrimSpace(hostname)
				staticHostMap[cleanIp(l.IP)] = append(staticHostMap[cleanIp(l.IP)].([]string), hostname + ":" + l.Port)
			}
		}
	} else {
		return "", errors.New("static_host_map not found in nebula.yml")
	}

	// set lightHouse
	if lighthouseMap, ok := configMap["lighthouse"].(map[interface{}]interface{}); ok {
		lighthouseMap["am_lighthouse"] = device.IsLighthouse

		lighthouseMap["hosts"] = []string{}

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
		relayMap["use_relays"] = !(device.IsRelay && device.IsLighthouse)
		relayMap["relays"] = []string{}

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


	configMap["cstln_device_name"] = name
	configMap["cstln_public_hostname"] = device.PublicHostname
	configMap["cstln_api_key"] = APIKey
	configMap["cstln_cosmos_node"] = device.CosmosNode
	configMap["cstln_is_exit_node"] = device.IsExitNode
	configMap["cstln_is_relay"] = device.IsRelay
	configMap["cstln_is_lighthouse"] = device.IsLighthouse
	configMap["cstln_ip"] = device.IP
	configMap["cstln_config_endpoint"] = utils.GetServerURL("")
	configMap["cstln_ip_range"] = utils.GetMainConfig().ConstellationConfig.IPRange

	if getLicence && device.CosmosNode == 0 {
		// get client licence
		utils.Log("Creating client license for " + name)
		lic, err := utils.FBL.CreateClientLicense(name)
		if err != nil {
			return "", err
		}
		configMap["cstln_licence"] = lic
		utils.Log("Client license created for " + name)
	}

	if getLicence && device.CosmosNode > 0 {
		configMap["cstln_server_licence"] = utils.GetMainConfig().Licence
	}

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

func GetCertFingerprint(certPath string) (string, error) {
	// nebula-cert print -json
	var cmd *exec.Cmd

	cmd = exec.Command(certBinaryToRun(),
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

func GetConfigAttribute(configPath string, attr string) (string, error) {
	// Read the YAML file
	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
			utils.Error("Error reading config file: "+configPath, err)
			return "", err
	}

	// Parse YAML into a map
	var config map[string]interface{}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
			utils.Error("Error parsing YAML file: "+configPath, err)
			return "", err
	}

	// Split the attribute path by dots to support nested attributes
	attrs := strings.Split(attr, ".")

	// Navigate through the nested structure
	var value interface{} = config
	for _, key := range attrs {
			// Convert current value to map for nested navigation
			m, ok := value.(map[string]interface{})
			if !ok {
					return "", fmt.Errorf("invalid path: %s is not a map", key)
			}

			// Get next value
			value, ok = m[key]
			if !ok {
					return "", fmt.Errorf("attribute not found: %s", attr)
			}
	}

	// Convert final value to string
	strValue, ok := value.(string)
	if !ok {
			return "", fmt.Errorf("attribute %s is not a string", attr)
	}

	return strValue, nil
}

func generateNebulaCert(name, ip, PK string, saveToFile bool) (string, string, string, error) {
	// Run the nebula-cert command
	var cmd *exec.Cmd

	ip = ip + "/24"

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
		cmd = exec.Command(certBinaryToRun(),
			"sign",
			"-ca-crt", utils.CONFIGFOLDER + "ca.crt",
			"-ca-key", utils.CONFIGFOLDER + "ca.key",
			"-subnets", "0.0.0.0/0",
			"-name", name,
			"-ip", ip,
		)
	} else {
		// write PK to temp.cert
		err := ioutil.WriteFile("./temp.key", []byte(PK), 0644)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to write temp.key: %s", err)
		}
		cmd = exec.Command(certBinaryToRun(),
			"sign",
			"-ca-crt", utils.CONFIGFOLDER + "ca.crt",
			"-ca-key", utils.CONFIGFOLDER + "ca.key",
			"-subnets", "0.0.0.0/0",
			"-name", name,
			"-ip", ip,
			"-in-pub", "./temp.key",
		)
		// delete temp.key
		defer os.Remove("./temp.key")
	}

	// Get pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
			return "", "", "", fmt.Errorf("failed to create stdout pipe: %s", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
			return "", "", "", fmt.Errorf("failed to create stderr pipe: %s", err)
	}

	// Start command
	err = cmd.Start()
	if err != nil {
			return "", "", "", fmt.Errorf("failed to start nebula-cert: %s", err)
	}

	// Create scanner for stdout
	stdoutScanner := bufio.NewScanner(stdout)
	go func() {
			for stdoutScanner.Scan() {
					utils.VPN(stdoutScanner.Text())
			}
	}()

	// Create scanner for stderr
	stderrScanner := bufio.NewScanner(stderr)
	go func() {
			for stderrScanner.Scan() {
					utils.Error("nebula-cert error", errors.New(stderrScanner.Text()))
			}
	}()

	// Wait for command to complete
	err = cmd.Wait()
	if err != nil {
			return "", "", "", fmt.Errorf("nebula-cert failed: %s", err)
	}

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

func generateNebulaCACert(name string) error {
	// Clean up existing files
	for _, file := range []string{"./ca.key", "./ca.crt"} {
			if _, err := os.Stat(file); err == nil {
					if err := os.Remove(file); err != nil {
							return fmt.Errorf("failed to remove existing %s: %s", file, err)
					}
			}
	}

	// Run the nebula-cert command to generate CA certificate and key
	cmd := exec.Command(
		certBinaryToRun(),
		"ca",
		"-name",
		"-duration", "87600h", // 10 years
		"\""+name+"\"",
	)
	utils.Debug(cmd.String())

	// Get pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
			return fmt.Errorf("failed to create stdout pipe: %s", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
			return fmt.Errorf("failed to create stderr pipe: %s", err)
	}

	// Start command
	if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start nebula-cert: %s", err)
	}

	// Handle stdout based on logging level
	go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				utils.VPN(scanner.Text())
			}
	}()

	// Handle stderr
	go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
					utils.Error("nebula-cert error", errors.New(scanner.Text()))
			}
	}()

	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
			return fmt.Errorf("nebula-cert failed: %s", err)
	}

	// Move files to config folder with error handling
	for _, moveCmd := range []struct{src, dst string}{
			{"./ca.crt", utils.CONFIGFOLDER + "ca.crt"},
			{"./ca.key", utils.CONFIGFOLDER + "ca.key"},
	} {
			cmd := exec.Command("mv", moveCmd.src, moveCmd.dst)

			// Get pipes for move command
			stdout, err := cmd.StdoutPipe()
			if err != nil {
					return fmt.Errorf("failed to create stdout pipe for move: %s", err)
			}
			stderr, err := cmd.StderrPipe()
			if err != nil {
					return fmt.Errorf("failed to create stderr pipe for move: %s", err)
			}

			// Start move command
			if err := cmd.Start(); err != nil {
					return fmt.Errorf("failed to start move command: %s", err)
			}

			// Handle stdout and stderr for move command
			go func() {
					scanner := bufio.NewScanner(stdout)
					for scanner.Scan() {
							utils.VPN(scanner.Text())
					}
			}()

			go func() {
					scanner := bufio.NewScanner(stderr)
					for scanner.Scan() {
							utils.Error("move command error", errors.New(scanner.Text()))
					}
			}()

			// Wait for move command to complete
			if err := cmd.Wait(); err != nil {
					return fmt.Errorf("failed to move %s to %s: %s", moveCmd.src, moveCmd.dst, err)
			}
	}

	return nil
}

func GetDeviceIp(device string) string {
	return CachedDeviceNames[device]
}

func populateIPTableMasquerade() {
	config := utils.GetMainConfig()
	isExitNode, err := GetCurrentDeviceIsExitNode()

	if isExitNode && err == nil {
		utils.Log("Constellation: Exit node enabled, configuring iptables masquerade...")

		// Enable IP forwarding
		cmd := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1")
		if err := cmd.Run(); err != nil {
			utils.Error("Constellation: Failed to enable IP forwarding", err)
			// Continue anyway
		} else {
			utils.Log("Constellation: IP forwarding enabled")
		}

		// Detect network interface
		var iface string
		if config.ConstellationConfig.OverrideNebulaExitNodeInterface != "" {
			iface = config.ConstellationConfig.OverrideNebulaExitNodeInterface
			utils.Log("Constellation: Using manual interface override: " + iface)
		} else {
			// Detect default interface using ip route
			cmd := exec.Command("sh", "-c", "ip route get 1.1.1.1 | grep -oP 'dev \\K\\S+'")
			output, err := cmd.Output()
			if err != nil {
				utils.Error("Constellation: Failed to detect network interface", err)
				return
			}
			iface = strings.TrimSpace(string(output))
			utils.Log("Constellation: Detected network interface: " + iface)
		}

		if iface == "" {
			utils.Error("Constellation: No network interface detected", nil)
			return
		}

		// Check if rules already exist
		cmd = exec.Command("iptables-save")
		output, err := cmd.Output()
		if err != nil {
			utils.Error("Constellation: Failed to check existing iptables rules", err)
			// Continue anyway
		} else if strings.Contains(string(output), "COSMOS-CLOUD-EXIT-NODE") {
			utils.Log("Constellation: IPTables rules already exist, skipping")
			return
		}

		// Add iptables rules with comment markers
		// Using a unique comment for easy identification and removal
		rules := []string{
			fmt.Sprintf("iptables -t nat -A POSTROUTING -s "+utils.GetMainConfig().ConstellationConfig.IPRange+" -o %s -m comment --comment 'COSMOS-CLOUD-EXIT-NODE' -j MASQUERADE", iface),
			fmt.Sprintf("iptables -A FORWARD -i nebula1 -o %s -m comment --comment 'COSMOS-CLOUD-EXIT-NODE' -j ACCEPT", iface),
			fmt.Sprintf("iptables -A FORWARD -i %s -o nebula1 -m state --state RELATED,ESTABLISHED -m comment --comment 'COSMOS-CLOUD-EXIT-NODE' -j ACCEPT", iface),
		}

		for _, rule := range rules {
			cmd := exec.Command("sh", "-c", rule)
			if err := cmd.Run(); err != nil {
				utils.Error("Constellation: Failed to add iptables rule: "+rule, err)
				// Continue anyway
			}
		}

		utils.Log("Constellation: IPTables masquerade rules added")
	} else {
		// Remove rules with our comment marker
		utils.Log("Constellation: Exit node disabled, removing iptables masquerade rules...")

		// Get current rules
		cmd := exec.Command("iptables-save")
		output, err := cmd.Output()
		if err != nil {
			utils.Error("Constellation: Failed to get iptables rules", err)
			return
		}

		rules := string(output)

		// Check if our rules exist
		if !strings.Contains(rules, "COSMOS-CLOUD-EXIT-NODE") {
			utils.Log("Constellation: No masquerade rules found, nothing to remove")
			return
		}

		// Remove rules with our comment marker from nat table
		cmd = exec.Command("sh", "-c", "iptables-save -t nat | grep 'COSMOS-CLOUD-EXIT-NODE' | grep '^-A' | sed 's/-A/-D/' | xargs -r -L1 iptables -t nat")
		if err := cmd.Run(); err != nil {
			utils.Error("Constellation: Error removing NAT rules", err)
		} else {
			utils.Log("Constellation: NAT rules removed")
		}

		// Remove rules with our comment marker from filter table
		cmd = exec.Command("sh", "-c", "iptables-save | grep 'COSMOS-CLOUD-EXIT-NODE' | grep '^-A' | sed 's/-A/-D/' | xargs -r -L1 iptables")
		if err := cmd.Run(); err != nil {
			utils.Error("Constellation: Error removing FORWARD rules", err)
		} else {
			utils.Log("Constellation: FORWARD rules removed")
		}
	}
}

var inited = false
func InitPingLighthouses() {
	if inited {
		return
	}
	inited = true
	for {
		PingLighthouses()
		time.Sleep(1 * time.Minute)
	}
}

func PingLighthouses() {
	lighthouses, err := GetAllLightHouses()
	if err != nil {
		utils.Error("Constellation: Failed to get lighthouses for pinging", err)
		return
	}

	for _, lh := range lighthouses {
		go pingLighthouse(lh, 0)
	}
}

func pingLighthouse(lh utils.ConstellationDevice, retries int) {
	cmd := exec.Command("ping", "-c", "1", "-W", "2", cleanIp(lh.IP))
	err := cmd.Run()
	if err != nil {
		if retries < 5 {
			time.Sleep(10 * time.Second)
			pingLighthouse(lh, retries+1)
		} else {
			utils.Warn("Constellation: Lighthouse " + lh.IP + " (" + cleanIp(lh.IP) + ") is unreachable after 10 retries")
		}
	} else {
		utils.Debug("Constellation: Lighthouse " + lh.IP + " (" + cleanIp(lh.IP) + ") is reachable")
	}
}

func GetCurrentDeviceName() (string, error) {
	config := utils.GetMainConfig()
	name := config.ConstellationConfig.ThisDeviceName
	if name == "" {
		nebulaFile, err := ioutil.ReadFile(utils.CONFIGFOLDER + "nebula.yml")
		if err != nil {
			utils.Error("GetCurrentDeviceName: error while reading nebula.yml", err)
			return "", err
		}

		configMap := make(map[string]interface{})
		err = yaml.Unmarshal(nebulaFile, &configMap)
		if err != nil {
			utils.Error("GetCurrentDeviceName: Invalid new config file for resync", err)
			return "", err
		}

		if configMap["cstln_device_name"] == nil {
			return "", errors.New("Invalid new config file for resync")
		}

		deviceName := configMap["cstln_device_name"].(string)

		return deviceName, nil
	}
	return name, nil
}

var cachedCurrentDevice *utils.ConstellationDevice

func GetCurrentDevice() (utils.ConstellationDevice, error) {
	if cachedCurrentDevice != nil {
		return *cachedCurrentDevice, nil
	}

	name, err := GetCurrentDeviceName()
	if err != nil {
		return utils.ConstellationDevice{}, err
	}

	device, exists := CachedDevices[name]
	if !exists {
		nebulaFile, err := ioutil.ReadFile(utils.CONFIGFOLDER + "nebula.yml")
		if err != nil {
			utils.Error("GetCurrentDeviceName: error while reading nebula.yml", err)
			return utils.ConstellationDevice{}, err
		}

		configMap := make(map[string]interface{})
		err = yaml.Unmarshal(nebulaFile, &configMap)
		if err != nil {
			utils.Error("GetCurrentDevice: Invalid new config file for resync", err)
			return utils.ConstellationDevice{}, err
		}

		device = utils.ConstellationDevice{}

		if configMap["cstln_device_name"] != nil  {
			device.DeviceName = configMap["cstln_device_name"].(string)
		} else {
			return utils.ConstellationDevice{}, errors.New("Invalid new config file for resync")
		}

		if configMap["cstln_ip"] != nil  {
			device.IP = configMap["cstln_ip"].(string)
		}

		if configMap["cstln_public_hostname"] != nil  {
			device.PublicHostname = configMap["cstln_public_hostname"].(string)
		}

		if configMap["cstln_cosmos_node"] != nil  {
			device.CosmosNode = configMap["cstln_cosmos_node"].(int)
		}

		if configMap["cstln_is_exit_node"] != nil  {
			device.IsExitNode = configMap["cstln_is_exit_node"].(bool)
		}

		if configMap["cstln_is_relay"] != nil  {
			device.IsRelay = configMap["cstln_is_relay"].(bool)
		}

		if configMap["cstln_is_lighthouse"] != nil  {
			device.IsLighthouse = configMap["cstln_is_lighthouse"].(bool)
		}

		if configMap["cstln_api_key"] != nil  {
			device.APIKey = configMap["cstln_api_key"].(string)
		} else {
			return utils.ConstellationDevice{}, errors.New("Invalid new config file for resync")
		}
	}

	cachedCurrentDevice = &device
	return device, nil
}

func GetCurrentDeviceAPIKey() (string, error) {
	device, err := GetCurrentDevice()
	if err != nil {
		return "", errors.New("current device not found in cache")
	}
	return device.APIKey, nil
}

func GetCurrentDeviceIP() (string, error) {
	device, err := GetCurrentDevice()
	if err != nil {
		return "", errors.New("current device not found in cache")
	}
	return device.IP, nil
}

func GetCurrentDeviceIsLoadbalancer() (bool, error) {
	device, err := GetCurrentDevice()
	if err != nil {
		return false, errors.New("current device not found in cache")
	}
	return device.IsLoadBalancer, nil
}

func GetCurrentDeviceIsRelay() (bool, error) {
	device, err := GetCurrentDevice()
	if err != nil {
		return false, errors.New("current device not found in cache")
	}
	return device.IsRelay, nil
}

func GetCurrentDeviceIsExitNode() (bool, error) {
	device, err := GetCurrentDevice()
	if err != nil {
		return false, errors.New("current device not found in cache")
	}
	return device.IsExitNode, nil
}

func GetCurrentDeviceHostname() (string, error) {
	device, err := GetCurrentDevice()
	if err != nil {
		if utils.GetMainConfig().ConstellationConfig.ConstellationHostname != "" {
			return utils.GetMainConfig().ConstellationConfig.ConstellationHostname, nil
		}
		return "", errors.New("current device not found in cache")
	}
	return device.PublicHostname, nil
}

func GetAllLighthouseIPFromTempConfig() ([]string, error) {
	nebulaFile, err := ioutil.ReadFile(utils.CONFIGFOLDER + "nebula-temp.yml")
	if err != nil {
		utils.Error("GetAllLighthouseIPFromConfig: error while reading nebula.yml", err)
		return []string{}, err
	}

	configMap := make(map[string]interface{})
	err = yaml.Unmarshal(nebulaFile, &configMap)
	if err != nil {
		utils.Error("GetAllLighthouseIPFromConfig: Invalid new config file for resync", err)
		return []string{}, err
	}

	lhIPs := []string{}

	if lighthouseMap, ok := configMap["lighthouse"].(map[interface{}]interface{}); ok {
		if hosts, ok := lighthouseMap["hosts"].([]interface{}); ok {
			for _, host := range hosts {
				lhIPs = append(lhIPs, host.(string))
			}
		}
	} else {
		return []string{}, errors.New("lighthouse not found in nebula.yml")
	}

	return lhIPs, nil
}
