package storage

import (
	"fmt"
	"io/ioutil"
	"strings"
	"strconv"
	"os/exec"
	"errors"

	"github.com/azukaar/cosmos-server/src/utils"
)

func ListDisks() ([]DiskInfo, error) {
	var disks []DiskInfo
	listPath := "/sys/block"

	// Read the contents of /sys/block to get disk names
	files, err := ioutil.ReadDir(listPath)
	if err != nil {
			return nil, err
	}
	
	utils.Debug("[STORAGE] Listing " + strconv.Itoa(len(files)) + " disks")

	for _, f := range files {
		name := f.Name()
		path := "/dev/" + name

		// Get the size of the disk
		size, err := GetDiskSize(name)
		if err != nil {
				return nil, err
		}

		use, err := GetDiskUsage(path)
		if err != nil {
			return nil, err
		}

		disks = append(disks, DiskInfo{Path: path, Name: name, Size: size, Used: use})
	}

	return disks, nil
}

func GetDiskSize(name string) (uint64, error) {
	utils.Debug("[STORAGE] Getting size of disk " + name)

	// Read the size file in /sys/block/<name>/size
	content, err := ioutil.ReadFile("/sys/block/" + name + "/size")
	if err != nil {
		return 0, err
	}

	// The size is reported in 512-byte sectors
	sizeInSectors, err := strconv.ParseUint(strings.TrimSpace(string(content)), 10, 64)
	if err != nil {
		return 0, err
	}

	// Convert size to bytes
	sizeInBytes := sizeInSectors * 512

	return sizeInBytes, nil
}

func GetDiskUsage(path string) (uint64, error) {
	utils.Debug("[STORAGE] Getting usage of disk " + path)

	// Get the disk usage using the du command
	cmd := exec.Command("du", "-s", path)

	// Run the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, err
	}

	// The output is in the format "<size> <path>"
	parts := strings.Fields(string(output))
	if len(parts) < 1 {
		return 0, fmt.Errorf("unexpected output: %s", string(output))
	}

	// Parse the size
	size, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return 0, err
	}

	return size, nil
}

func FormatDisk(diskPath string, filesystemType string) error {
	utils.Log("[STORAGE] Formatting disk " + diskPath + " with filesystem " + filesystemType)

	// check filesystem type
	supportedFilesystems := []string{"ext4", "xfs", "ntfs", "fat32", "exfat", "btrfs", "zfs", "ext3", "ext2"}
	isSupported := false
	for _, fs := range supportedFilesystems {
		if fs == filesystemType {
		isSupported = true
		break
		}
	}
	if !isSupported {
		return errors.New("unsupported filesystem type")
	}

	// check if the disk is mounted
	mounted, err := IsDiskMounted(diskPath)
	if err != nil {
		return err
	}
	if mounted {
		return errors.New("disk is mounted, please unmount it first")
	}

	// Example: mkfs.ext4 /dev/sdx - Make sure the disk path is correct!
	// WARNING: This will erase all data on the disk!
	cmd := exec.Command("mkfs", "-t", filesystemType, diskPath)

	// Run the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New("Format: " + string(output) + " " + err.Error())
	}

	return nil
}