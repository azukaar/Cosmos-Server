package main

import (
	"math/rand"
	"time"
	"context"
	"os"
	"strings"
	"fmt"
	"log"
	"path/filepath"

	"github.com/azukaar/cosmos-server/src/docker"
	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/azukaar/cosmos-server/src/authorizationserver"
	"github.com/azukaar/cosmos-server/src/market"
	"github.com/azukaar/cosmos-server/src/constellation"
	"github.com/azukaar/cosmos-server/src/metrics"
	"github.com/azukaar/cosmos-server/src/storage"
	"github.com/azukaar/cosmos-server/src/cron"
	"github.com/azukaar/cosmos-server/src/proxy"
	"github.com/azukaar/cosmos-server/src/backups"
	
	"github.com/kardianos/service"
)


// Program structure for the service
type Program struct{}

// Start implements service.Interface
func (p *Program) Start(s service.Service) error {
	fmt.Println("Starting Cosmos Server service")
	go p.run()
	return nil
}

// Stop implements service.Interface
func (p *Program) Stop(s service.Service) error {
	fmt.Println("Stopping Cosmos Server service")
	return nil
}

func (p *Program) run() {
	cosmos()
}

func getCosmosEnvVars() map[string]string {
	envVars := make(map[string]string)
	for _, env := range os.Environ() {
		if pair := strings.SplitN(env, "=", 2); len(pair) == 2 {
			if strings.HasPrefix(pair[0], "COSMOS_") || pair[0] == "TZ" || pair[0] == "CONFIG_FILE" || pair[0] == "ACME_STAGING" {
				envVars[pair[0]] = pair[1]
			}
		}
	}
	return envVars
}

func HandleCLIArgs() bool {
	args := os.Args[1:]

	if len(args) > 0 && args[0] == "rclone" {
			args = args[1:]
			
			storage.RunRClone(args)

			return true
	} else if len(args) > 0 && args[0] == "service" {
		// Get the executable's directory
		execPath, err := os.Executable()
		if err != nil {
			log.Fatal("Failed to get executable path:", err)
			return true
		}
		
		workingDir := filepath.Dir(execPath)

		const systemdScript = `[Unit]
Description={{.Description}}
ConditionFileIsExecutable={{.Path|cmdEscape}}
{{range $i, $dep := .Dependencies}} 
{{$dep}} {{end}}

[Service]
StartLimitInterval=10
StartLimitBurst=5
ExecStart={{.Path|cmdEscape}}{{range .Arguments}} {{.|cmd}}{{end}}
{{if .ChRoot}}RootDirectory={{.ChRoot|cmd}}{{end}}
{{if .WorkingDirectory}}WorkingDirectory={{.WorkingDirectory|cmdEscape}}{{end}}
{{if .UserName}}User={{.UserName}}{{end}}
{{if .ReloadSignal}}ExecReload=/bin/kill -{{.ReloadSignal}} "$MAINPID"{{end}}
{{if .PIDFile}}PIDFile={{.PIDFile|cmd}}{{end}}
{{if and .LogOutput .HasOutputFileSupport -}}
StandardOutput=file:{{.LogDirectory}}/{{.Name}}.out
StandardError=file:{{.LogDirectory}}/{{.Name}}.err
{{- end}}
{{if gt .LimitNOFILE -1 }}LimitNOFILE={{.LimitNOFILE}}{{end}}
{{if .Restart}}Restart={{.Restart}}{{end}}
{{if .SuccessExitStatus}}SuccessExitStatus={{.SuccessExitStatus}}{{end}}
RestartSec=2
EnvironmentFile=-/etc/sysconfig/{{.Name}}

{{range $k, $v := .EnvVars -}}
Environment={{$k}}={{$v}}
{{end -}}

[Install]
WantedBy=multi-user.target`

		svcConfig := &service.Config{
			Name:        "CosmosCloud",
			DisplayName: "Cosmos Cloud",
			Description: "Cosmos Cloud service",
			WorkingDirectory: workingDir,
			Executable:  filepath.Join(workingDir, "start.sh"), 
			EnvVars: getCosmosEnvVars(),
			Option: service.KeyValue{
				"SystemdScript": systemdScript,
			},
		}

		// fmt.Println("Service mode: ", svcConfig)

		prg := &Program{}
		s, err := service.New(prg, svcConfig)
		if err != nil {
			log.Fatal(err)
		}
		
		// Handle service management commands
		switch args[1] {
		case "install":
			err = s.Install()
			if err != nil {
				log.Fatal(err)
				return true
			}
			log.Fatal("Service installed successfully")
			
			// Print captured environment variables for verification
			envVars := getCosmosEnvVars()
			if len(envVars) > 0 {
				fmt.Println("\nCaptured environment variables:")
				for k, v := range envVars {
					fmt.Printf("%s=%s\n", k, v)
				}
			}
		case "uninstall":
			err = s.Uninstall()
			if err != nil {
				log.Fatal(err)
				return true
			}
			fmt.Println("Service uninstalled successfully")
		case "start":
			err = s.Start()
			if err != nil {
				log.Fatal(err)
				return true
			}
			fmt.Println("Service started successfully")
		case "stop":
			err = s.Stop()
			if err != nil {
				log.Fatal(err)
				return true
			}
			fmt.Println("Service stopped successfully")
		case "status":
			status, err := s.Status()
			if err != nil {
				log.Fatal(err)
				return true
			}
			fmt.Printf("Service status: %v\n", status)
		default:
			fmt.Println("Unknown command")
		}
		
		return true
	}

	return false
}

func main() {
	docker.IsInsideContainer()

	if os.Getenv("COSMOS_CONFIG_FOLDER") != "" {
		utils.CONFIGFOLDER = os.Getenv("COSMOS_CONFIG_FOLDER")
	} else if utils.IsInsideContainer {
		utils.CONFIGFOLDER = "/config/"
	}

	if HandleCLIArgs() {
		return
	}
	cosmos()
}

func cosmos() {
	utils.InitLogs()

	if _, err := os.Stat(utils.CONFIGFOLDER); os.IsNotExist(err) {
		err := os.MkdirAll(utils.CONFIGFOLDER, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	if utils.IsInsideContainer {
		utils.Log("Running inside Docker container")
	} else {
		utils.Log("Running outside Docker container")
	}

	i, _ := utils.ListInterfaces(false)
	utils.Log("Interfaces are: " + strings.Join(i, ", "))
	

	utils.Log("------------------------------------------")
	utils.Log("Starting Cosmos-Server version " + GetCosmosVersion())
	utils.Log("------------------------------------------")
	
	// utils.ReBootstrapContainer = docker.BootstrapContainerFromTags
	utils.PushShieldMetrics = metrics.PushShieldMetrics
	utils.GetContainerIPByName = docker.GetContainerIPByName
	utils.DoesContainerExist = docker.DoesContainerExist
	utils.CheckDockerNetworkMode = docker.CheckDockerNetworkMode

	rand.Seed(time.Now().UnixNano())
	
	LoadConfig()

	utils.RemovePIDFile()

	utils.CheckHostNetwork()
	
	go CRON()

	cleanUpUpdateFiles()

	docker.ExportDocker()

	docker.DockerListenEvents()

	docker.BootstrapAllContainersFromTags()

	docker.RemoveSelfUpdater()

	go func() {
		time.Sleep(180 * time.Second)
		checkUpdatesAvailable()
	}()

	version, err := docker.DockerClient.ServerVersion(context.Background())
	if err == nil {
		utils.Log("Docker API version: " + version.APIVersion)
	}

	config := utils.GetMainConfig()

	proxy.InitSocketShield()
	proxy.InitUDPShield()

	if !config.NewInstall {
		MigratePre013()
		MigratePre014()

		utils.CheckInternet()

		docker.CheckPuppetDB()

		utils.InitDBBuffers()

		utils.Log("Starting monitoring services...")

		metrics.Init()

		utils.Log("Starting market services...")

		market.Init()
		
		utils.Log("Starting OpenID services...")

		authorizationserver.Init()

		utils.Log("Starting constellation services...")

		utils.InitFBL()

		constellation.Init()

		utils.InitRemoteStorage = storage.InitRemoteStorage
		
		if utils.FBL.LValid && !utils.FBL.IsCosmosNode {
			utils.ProxyRClone = storage.InitRemoteStorage()
		}

		storage.InitSnapRAIDConfig()
		
		// Has to be done last, so scheduler does not re-init
		cron.Init()

		
		utils.InitBackups = backups.InitBackups
		if utils.FBL.LValid && !utils.FBL.IsCosmosNode {
			go backups.InitBackups()
		}

		utils.Log("Starting server...")
	}

	StartServer()
}
