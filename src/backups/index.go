package backups

import (
	"os"
	"path/filepath"
	"github.com/azukaar/cosmos-server/src/cron"
	"github.com/azukaar/cosmos-server/src/utils"
	"time"
)

func InitBackups() {
	execPath, _ := os.Executable()
	resticPath := filepath.Dir(execPath) + "/restic"
	utils.Exec("chmod", "+x", resticPath)

	config := utils.GetMainConfig()
	
	if !utils.FBL.LValid {
		utils.Warn("InitBackups: No valid licence found, not starting module.")
		return
	}
	
	if utils.IsInsideContainer {
		utils.Warn("InitBackups: Currently running in a docker container, disabling backups.")
		return
	}

	Repositories := config.Backup.Backups

	cron.ResetScheduler("Restic")

	// check if "Cosmos Internal Backup" exists and is missing password
	// if so, create it
	internalBackup := ""
	intBack := utils.SingleBackupConfig{}
	password := ""

	for _, repo := range Repositories {
		if repo.Name == "Cosmos Internal Backup" && repo.Password == "" {
			utils.Warn("[Backup] Cosmos Internal Backup is missing password, getting one.")
			internalBackup = repo.Repository
			intBack = repo
		}
	}

	if internalBackup != "" {
		created := false
		for _, repo := range Repositories {
			if repo.Name != "Cosmos Internal Backup" && repo.Repository == internalBackup {
				utils.Log("[Backup] Found backup repository password.")
				password = repo.Password
				created = true
			}
		}

		if !created {
			utils.Log("[Backup] Password not found. Creating one")
			password = utils.GenerateRandomString(16)
			CreateRepository(internalBackup, password)
		}

		
		config.Backup.Backups["Cosmos Internal Backup"] = utils.SingleBackupConfig{
			Repository: internalBackup,
			Password:   password,
			Source:     intBack.Source,
			Crontab:    intBack.Crontab,
			CrontabForget: intBack.CrontabForget,
			RetentionPolicy: intBack.RetentionPolicy,
			Name:       "Cosmos Internal Backup",
		}

		utils.SetBaseMainConfig(config)
		config = utils.GetMainConfig()
	}

	go func() {
		time.Sleep(2 * time.Second) 
		for _, repo := range Repositories {
			err := CheckRepository(repo.Repository, repo.Password)

			if err != nil {
				utils.MajorError("Backups destination unavailable", err)
			} else {
				// create backup job
				CreateBackupJob(BackupConfig{
					Repository: repo.Repository,
					Password:   repo.Password,
					Source:     repo.Source,
					Name:       repo.Name,
					AutoStopContainers: repo.AutoStopContainers,
					Tags:       []string{repo.Name},
					// Exclude:    repo.Exclude,
				}, repo.Crontab)
				
				// create backup job
				CreateForgetJob(BackupConfig{
					Repository: repo.Repository,
					Password:   repo.Password,
					Source:     repo.Source,
					Name:       repo.Name,
					Tags:       []string{repo.Name},
					Retention:  repo.RetentionPolicy,
				}, repo.CrontabForget)
			}
		}
	}()
}