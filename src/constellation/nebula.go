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
)

var (
	process    *exec.Cmd
	processMux sync.Mutex
)

func binaryToRun() string {
	if runtime.GOARCH == "arm" || runtime.GOARCH == "arm64" {
		return "./nebula-arm"
	}
	return "./nebula"
}

func startNebulaInBackground() error {
	processMux.Lock()
	defer processMux.Unlock()

	if process != nil {
		return errors.New("nebula is already running")
	}

	process = exec.Command(binaryToRun(), "-config", utils.CONFIGFOLDER + "nebula.yml")

	process.Stderr = os.Stderr
	
	if utils.LoggingLevelLabels[utils.GetMainConfig().LoggingLevel] == utils.DEBUG {
		process.Stdout = os.Stdout
	} else {
		process.Stdout = nil
	}

	// Start the process in the background
	if err := process.Start(); err != nil {
		return err
	}

	utils.Log(fmt.Sprintf("%s started with PID %d\n", binaryToRun(), process.Process.Pid))
	return nil
}

func stop() error {
	processMux.Lock()
	defer processMux.Unlock()

	if process == nil {
		return errors.New("nebula is not running")
	}

	if err := process.Process.Kill(); err != nil {
		return err
	}
	process = nil
	utils.Log("Stopped nebula.")
	return nil
}

func restart() error {
	if err := stop(); err != nil {
		return err
	}
	return startNebulaInBackground()
}

func ExportConfigToYAML(overwriteConfig utils.ConstellationConfig, outputPath string) error {
	// Combine defaultConfig and overwriteConfig
	finalConfig := NebulaDefaultConfig

	finalConfig.StaticHostMap = map[string][]string{
		"192.168.201.0": []string{utils.GetMainConfig().HTTPConfig.Hostname + ":4242"},
	}

	// Marshal the combined config to YAML
	yamlData, err := yaml.Marshal(finalConfig)
	if err != nil {
		return err
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

func getYAMLClientConfig(name, configPath string) (string, error) {
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

	// set lightHouse to false
	if lighthouseMap, ok := configMap["lighthouse"].(map[interface{}]interface{}); ok {
		lighthouseMap["am_lighthouse"] = false

		lighthouseMap["hosts"] = []string{
			"192.168.201.0",
		}
	} else {
		return "", errors.New("lighthouse not found in nebula.yml")
	}

	if pkiMap, ok := configMap["pki"].(map[interface{}]interface{}); ok {
		pkiMap["ca"] = "ca.crt"
		pkiMap["cert"] = name + ".crt"
		pkiMap["key"] = name + ".key"
	} else {
		return "", errors.New("pki not found in nebula.yml")
	}

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
	processMux.Lock()
	defer processMux.Unlock()

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

func generateNebulaCert(name, ip string, saveToFile bool) (string, string, error) {
	// Run the nebula-cert command
	cmd := exec.Command(binaryToRun() + "-cert",
		 "sign",
		 "-ca-crt", utils.CONFIGFOLDER + "ca.crt",
		 "-ca-key", utils.CONFIGFOLDER + "ca.key",
		 "-name", name,
		 "-ip", ip,
	)
	utils.Debug(cmd.String())

	cmd.Stderr = os.Stderr
	
	if utils.LoggingLevelLabels[utils.GetMainConfig().LoggingLevel] == utils.DEBUG {
		cmd.Stdout = os.Stdout
	} else {
		cmd.Stdout = nil
	}
	
	cmd.Run()

	if cmd.ProcessState.ExitCode() != 0 {
		return "", "", fmt.Errorf("nebula-cert exited with an error, check the Cosmos logs")
	}

	// Read the generated certificate and key files
	certPath := fmt.Sprintf("./%s.crt", name)
	keyPath := fmt.Sprintf("./%s.key", name)

	utils.Debug("Reading certificate from " + certPath)
	utils.Debug("Reading key from " + keyPath)

	certContent, errCert := ioutil.ReadFile(certPath)
	if errCert != nil {
		return "", "", fmt.Errorf("failed to read certificate file: %s", errCert)
	}

	keyContent, errKey := ioutil.ReadFile(keyPath)
	if errKey != nil {
		return "", "", fmt.Errorf("failed to read key file: %s", errKey)
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
			return "", "", fmt.Errorf("failed to delete certificate file: %s", err)
		}

		if err := os.Remove(keyPath); err != nil {
			return "", "", fmt.Errorf("failed to delete key file: %s", err)
		}
	}

	return string(certContent), string(keyContent), nil
}

func generateNebulaCACert(name string) (error) {
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