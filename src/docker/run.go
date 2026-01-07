package docker

import (
	"github.com/azukaar/cosmos-server/src/utils" 
	"net/http"
	"fmt"
	"errors"

	// "github.com/docker/docker/client"
	// "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	
	"runtime"
	"golang.org/x/sys/cpu"
)

func CheckPuppetDB() {
	config := utils.GetMainConfig()
	if config.Database.PuppetMode {
		utils.Log("Puppet mode enabled. Checking for DB...")
		err := utils.DB()
		if err != nil {
			utils.Error("Puppet mode enabled. DB not found. Recreating DB...", err)
			service, err := RunDB(config.Database)
			if err != nil {
				utils.Error("Puppet mode enabled. DB not found. Error while recreating DB...", err)
				return
			}

			err = CreateService(service,
				func (msg string) {
					utils.Log(msg)
				},
			)

			if err != nil {
				utils.Error("Puppet mode enabled. DB not found. Error while recreating DB...", err)
				return
			}
		}
	}
}

func RunDB(db utils.DatabaseConfig) (DockerServiceCreateRequest, error) {
	imageName := "mongo:" + db.Version
	
	//if ARM use arm64v8/mongo
	if runtime.GOARCH == "arm64" {
		utils.Warn("ARM64 detected. Using ARM mongo 4.4.18")
		imageName = "arm64v8/mongo:4.4.18"
	// if CPU is missing AVX, use 4.4
	} else if runtime.GOARCH == "amd64" && !cpu.X86.HasAVX {
		utils.Warn("CPU does not support AVX. Using mongo 4.4")
		imageName = "mongo:4.4"
	}

	service := DockerServiceCreateRequest{
		Services: map[string]ContainerCreateRequestContainer {},
	}
	
	service.Services[db.Hostname] = ContainerCreateRequestContainer{
		Name: db.Hostname,
		Image: imageName,
		RestartPolicy: "always",
		Command: []string{
			"mongod",
			"--wiredTigerCacheSizeGB=0.25",
		},
		Environment: []string{
			"MONGO_INITDB_ROOT_USERNAME=" + db.Username,
			"MONGO_INITDB_ROOT_PASSWORD=" + db.Password,
		},
		Labels: map[string]string{
			"cosmos-auto-update": "true",
		},
		Networks: map[string]ContainerCreateRequestServiceNetwork{},
		Volumes: []mount.Mount{
			{
				Type:   mount.TypeVolume,
				Source: db.DbVolume,
				Target: "/data/db",
			},
			{
				Type:   mount.TypeVolume,
				Source: db.ConfigVolume,
				Target: "/data/configdb",
			},
		},
	};

	if utils.IsInsideContainer && !utils.IsHostNetwork {
		newNetwork, errNC := CreateCosmosNetwork(db.Hostname)
		if errNC != nil {
			return DockerServiceCreateRequest{}, errNC
		}
	
		service.Services[db.Hostname].Labels["cosmos-network-name"] = newNetwork
		service.Services[db.Hostname].Networks[newNetwork] = ContainerCreateRequestServiceNetwork{}
 		
		AttachNetworkToCosmos(newNetwork)
	}
	
	return service, nil
}

func NewDB(w http.ResponseWriter, req *http.Request) (utils.DatabaseConfig, error) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Transfer-Encoding", "chunked")
	
	flusher, ok := w.(http.Flusher)
	if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return utils.DatabaseConfig{}, errors.New("Streaming unsupported!")
	}

	fmt.Fprintf(w, "NewInstall: Create DB\n")
	flusher.Flush()

	id := utils.GenerateRandomString(3)
	mongoUser := "cosmos-" + utils.GenerateRandomString(5) 
	mongoPass := utils.GenerateRandomString(24)
	monHost := "cosmos-mongo-" + id
	
	imageVersion := "6"

	dbConf := utils.DatabaseConfig {
		PuppetMode: true,
		Hostname: monHost,
		DbVolume: "cosmos-mongo-data-" + id,
		ConfigVolume: "cosmos-mongo-config-" + id,
		Version: imageVersion,
		Username: mongoUser,
		Password: mongoPass,
	}

	service, err := RunDB(dbConf)

	if err != nil {
		utils.Error("NewDB: Error while creating new DB", err)
		fmt.Fprintf(w, "NewDB: Error while creating new DB: %s\n", err)
		flusher.Flush()
		return dbConf, err
	}
	
	err = CreateService(service, 
		func (msg string) {
			utils.Log(msg)
			fmt.Fprintf(w, msg + "\n")
			flusher.Flush()
		},
	)

	if err != nil {
		utils.Error("NewDB: Error while creating new DB", err)
		fmt.Fprintf(w, "NewDB: Error while creating new DB: %s\n", err)
		flusher.Flush()
		return dbConf, err
	}

	// return "mongodb://"+mongoUser+":"+mongoPass+"@"+monHost+":27017", nil
	return dbConf, nil
}
