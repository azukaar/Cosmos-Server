package storage 

import (
	"os"
	"errors"
	"strings"
	"strconv"
	
	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/azukaar/cosmos-server/src/cron"
)

func findParent(disks []BlockDevice, path string) (BlockDevice, error) {
	for _, d := range disks {
		if d.MountPoint == path {
			return d, nil
		}
		if len(d.Children) > 0 {
			if p, err := findParent(d.Children, path); err == nil {
				return p, nil
			}
		}
	}
	return BlockDevice{}, errors.New("Not found")
}

// Create a SnapRAID configuration
func CreateSnapRAID(raidOptions utils.SnapRAIDConfig, editRaid string) error {
	data := raidOptions.Data
	parity := raidOptions.Parity
	
	// utils.Log("[STORAGE] Create SnapRAID " + strings.Join(data, ":") + " with " + strings.Join(data, ":"))
	
	// make sure name is at least 3 characters and contains only letters and numbers
	if len(raidOptions.Name) < 3 || strings.ContainsAny(raidOptions.Name, "./!?@#$%^&*()_+=-{}[]|\\:;\"'<>,") {
		return errors.New("Name must be at least 3 characters and contain only letters and numbers")
	}

	// Check config
	count := 0
	for _, d := range data {
		if d != "" {
			count++
		}
	}
	if count <= 1 {
		return errors.New("At least 2 data disks are required")
	}

	if len(parity) == 0 {
		return errors.New("At least 1 parity disk is required")
	}

	// check if two data are actually on the same disk
	mountPoints, err := ListMounts()
	if err != nil {
		return err
	}

	devices := map[string]bool{}

	for _, d1 := range data {
		for _, m := range mountPoints {
			if m.Path == d1 {
				if !devices[m.Device] {
					devices[m.Device] = true
				} else {
					return errors.New("Do not add different mountpoints that are on the same disk. It will cause data loss")
				}
			}
		}
	}

	// check if 2 partitions are on the same disk
	disks, err := ListDisks()
	if err != nil {
		return err
	}

	disksUsed := map[string]bool{}
	for _, d1 := range data {
		parent, err := findParent(disks, d1)
		if err != nil {
			return errors.New("Make sure you are using mountpoints that are disks or full disks partitions")
		}
		if !disksUsed[parent.Name] {
			disksUsed[parent.Name] = true
		} else {
			return errors.New("Do not add different partitions that are on the same disk. It will cause data loss")
		}
	}

	// Check if already exist
	if editRaid == "" {
		for _, r := range utils.GetMainConfig().Storage.SnapRAIDs {
			if r.Name == raidOptions.Name {
				return errors.New("Name already exist")
			}
		}
	}

	// Check if partiy is the biggest disk
	// Check if parity is not in data
	for _, p := range parity {
		for _, d := range data {
			if p == d {
				return errors.New("Parity disk cannot be a data disk")
			}
			pd, err := findParent(disks, p)
			if err != nil {
				return errors.New("Parity disk must be a disk")
			}
			dd, err := findParent(disks, d)
			if err != nil {
				return errors.New("Data disk must be a disk")
			}
			if dd.Size.Int64 > pd.Size.Int64 {
				return errors.New("Parity disk must be the biggest disk")
			}
		}
	}

	// save to config 
	config := utils.ReadConfigFromFile()
	// remove the old one
	if editRaid != "" {
		for i, r := range config.Storage.SnapRAIDs {
			if r.Name == editRaid {
				config.Storage.SnapRAIDs = append(config.Storage.SnapRAIDs[:i], config.Storage.SnapRAIDs[i+1:]...)
				break
			}
		}
	}

	// add the new one
	config.Storage.SnapRAIDs = append(config.Storage.SnapRAIDs, raidOptions)
	utils.SetBaseMainConfig(config)
	
	utils.TriggerEvent(
		"cosmos.settings",
		"Settings updated",
		"success",
		"",
		map[string]interface{}{
			"from": "SnapRAID configuration created",
	})

	InitSnapRAIDConfig()

	return nil
}

func DeleteSnapRAID(name string) error {
	utils.ConfigLock.Lock()
	defer utils.ConfigLock.Unlock()

	os.Remove(utils.CONFIGFOLDER + "snapraid/" + name + ".conf")
	
	
	config := utils.ReadConfigFromFile()
	for i, r := range config.Storage.SnapRAIDs {
		if r.Name == name {
			// remove the snapraid.content and snapraid.parity files, and lock fiels
			for _, d := range r.Data {
				os.Remove(d + "/snapraid.content")
				os.Remove(d + "/snapraid.parity")
				os.Remove(d + "/snapraid.content.lock")
				os.Remove(d + "/snapraid.parity.lock")
			}

			config.Storage.SnapRAIDs = append(config.Storage.SnapRAIDs[:i], config.Storage.SnapRAIDs[i+1:]...)
			utils.SetBaseMainConfig(config)
			utils.Log("Deleted SnapRAID " + name)
			return nil
		}
	}
	return errors.New("SnapRAID not found")
}

func InitSnapRAIDConfig() {
	config := utils.GetMainConfig()
	snaps := config.Storage.SnapRAIDs

	// reset the scheduler
	cron.ResetScheduler("SnapRAID")
	
	// remove the folder if it exists
	if _, err := os.Stat(utils.CONFIGFOLDER + "snapraid"); err == nil {
		os.RemoveAll(utils.CONFIGFOLDER + "snapraid")
	}

	os.MkdirAll(utils.CONFIGFOLDER + "snapraid", 0755)

	for _, raidOptions := range snaps {
		// create or overwrite the file
		file, err := os.OpenFile(utils.CONFIGFOLDER + "snapraid/" + raidOptions.Name + ".conf", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			utils.MajorError("Failed to create SnapRAID config file", err)
			return
		}
		defer file.Close()

		// write the configuration
		for _di, d := range raidOptions.Parity {
			di := strconv.Itoa(_di)
			if _di == 0 {
				file.WriteString("parity " + d + "/snapraid.parity\n")
			} else {
				file.WriteString(di + "-parity " + d + "/snapraid."+di+"-parity\n")
			}
		}

		// file.WriteString("content " + utils.CONFIGFOLDER + "snapraid/" + raidOptions.Name + ".conf\n")
		
		for name, d := range raidOptions.Data {
			file.WriteString("content " + d + "/snapraid.content\n")
			file.WriteString("data " + name + " " + d + "\n")
		}
		
		// Init scheduler
		
		cron.RegisterJob(cron.ConfigJob{
			Scheduler: "SnapRAID",
			Name: "SnapRAID sync " + raidOptions.Name,
			Crontab: raidOptions.SyncCrontab,
			Cancellable: true,
			Job: cron.JobFromCommand("snapraid", "sync", "-c", utils.CONFIGFOLDER + "snapraid/" + raidOptions.Name + ".conf"),
		})
		
		cron.RegisterJob(cron.ConfigJob{
			Scheduler: "SnapRAID",
			Name: "SnapRAID scrub " + raidOptions.Name,
			Crontab: raidOptions.ScrubCrontab,
			Cancellable: true,
			Job: cron.JobFromCommand("snapraid", "scrub", "-c", utils.CONFIGFOLDER + "snapraid/" + raidOptions.Name + ".conf"),
		})
	}	
}

func ToggleSnapRAID(name string, enable bool) error {
	utils.ConfigLock.Lock()
	defer utils.ConfigLock.Unlock()

	config := utils.ReadConfigFromFile()
	for i, r := range config.Storage.SnapRAIDs {
		if r.Name == name {
			config.Storage.SnapRAIDs[i].Enabled = enable
			utils.SetBaseMainConfig(config)
			utils.Log("Toggled SnapRAID " + name)
			return nil
		}
	}
	return errors.New("SnapRAID not found")
}

func RunSnapRAIDSync(raid utils.SnapRAIDConfig) error {
	utils.Log("[STORAGE] Running SnapRAID sync for " + raid.Name)
	cron.ManualRunJob("SnapRAID", "SnapRAID sync " + raid.Name)

	return nil
}

func RunSnapRAIDFix(raid utils.SnapRAIDConfig) error {
	utils.Log("[STORAGE] Running SnapRAID fix for " + raid.Name)

	cron.RunOneTimeJob(cron.ConfigJob{
		Scheduler: "SnapRAID",
		Name: "SnapRAID fix " + raid.Name,
		Cancellable: true,
		Job: cron.JobFromCommand("snapraid", "fix", "-c", utils.CONFIGFOLDER + "snapraid/" + raid.Name + ".conf"),
	})

	return nil
}

func RunSnapRAIDScrub(raid utils.SnapRAIDConfig) error  {
	utils.Log("[STORAGE] Running SnapRAID scrub for " + raid.Name)
	cron.ManualRunJob("SnapRAID", "SnapRAID scrub " + raid.Name)

	return nil
}

func RunSnapRAIDStatus(raid utils.SnapRAIDConfig) (string, error)  {
	out, err := utils.Exec("snapraid", "status", "-c", utils.CONFIGFOLDER + "snapraid/" + raid.Name + ".conf")

	if err != nil {
		utils.MajorError("Failed to status " + raid.Name, err)
		return "", err
	}

	return out, nil
}