package storage 

import (
	"os"
	"errors"
	"strings"

	"github.com/azukaar/cosmos-server/src/utils"
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
func CreateSnapRAID(raidOptions utils.SnapRAIDConfig) error {
	data := raidOptions.Data
	parity := raidOptions.Parity
	
	utils.Log("[STORAGE] Create SnapRAID " + strings.Join(data, ":") + " with " + strings.Join(data, ":"))
	
	// make sure name is at least 3 characters and contains only letters and numbers
	if len(raidOptions.Name) < 3 || strings.ContainsAny(raidOptions.Name, "./!?@#$%^&*()_+=-{}[]|\\:;\"'<>,") {
		return errors.New("Name must be at least 3 characters and contain only letters and numbers")
	}

	// check config
	if len(data) <= 1 {
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

	// TODO: Check if partiy is the biggest disk

	// TODO: Check if file already exist / config already exist

	// TODO: Check if parity is not in data

	// create the snapraid folder if it doesn't exist
	if _, err := os.Stat(utils.CONFIGFOLDER + "snapraid"); os.IsNotExist(err) {
		os.MkdirAll(utils.CONFIGFOLDER + "snapraid", 0755)
	}

	// export name.conf
	file, err := os.Create(utils.CONFIGFOLDER + "snapraid/" + raidOptions.Name + ".conf")
	if err != nil {
		return err
	}

	// write the configuration
	for _, d := range parity {
		file.WriteString("parity " + d + "/snapraid.parity\n")
	}

	// file.WriteString("content " + utils.CONFIGFOLDER + "snapraid/" + raidOptions.Name + ".conf\n")
	
	for _, d := range data {
		file.WriteString("content " + d + "/snapraid.content\n")
		file.WriteString("data " + d + "\n")
	}

	// save to config 
	config := utils.ReadConfigFromFile()
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

	// TODO: plan a sync

	return nil
}
