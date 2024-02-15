package storage

import (
	"fmt"
	// "io/ioutil"
	"strings"
	"strconv"
	"os/exec"
	"errors"

	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/dell/csi-baremetal/pkg/base/linuxutils/lsblk"
	"github.com/sirupsen/logrus"
	"github.com/anatol/smart.go"
)


type BlockDevice struct {
	lsblk.BlockDevice
	Children []BlockDevice `json:"children"`
	Usage uint64 `json:"usage"`
	SMART smart.GenericAttributes `json:"smart"` // Add SMART data field
}

func ListDisks() ([]BlockDevice, error) {
	// Create a new logrus Logger
	logger := logrus.New()

	// Initialize lsblk with the logger
	lsblkExecutor := lsblk.NewLSBLK(logger)

	devices, err := lsblkExecutor.GetBlockDevices("")
	if err != nil {
			return nil, err
	}

	return GetRecursiveDiskUsageAndSMARTInfo(devices)
}

func GetRecursiveDiskUsageAndSMARTInfo(devices []lsblk.BlockDevice) ([]BlockDevice, error) {
	devicesF := make([]BlockDevice, len(devices))
	for i, device := range devices {
		used, _, err := GetDiskUsage(device.Name)
		if err != nil {
			return nil, err
		}

		// Retrieve SMART information
		if device.Type == "disk" {
			dev, err := smart.Open(device.Name)
			if err != nil {
				return nil, err
			}

			defer dev.Close()

			smartInfo, err := dev.ReadGenericAttributes()
			if err != nil {
				return nil, err
			}

			devicesF[i].SMART = *smartInfo

			if err != nil {
				return nil, err
			}
		}

		devicesF[i].BlockDevice = device
		devicesF[i].Usage = used

		// Get usage and SMART info for children
		devicesF[i].Children, err = GetRecursiveDiskUsageAndSMARTInfo(device.Children)
		if err != nil {
			return nil, err
		}
	}

	return devicesF, nil
}
func GetDiskUsage(path string) (used uint64, size uint64, err error) {
	fmt.Println("[STORAGE] Getting usage of disk " + path)

	// Get the disk usage using the df command
	cmd := exec.Command("df", "-k", path)

	// Run the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, 0, err
	}

	// Split the output into lines
	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return 0, 0, fmt.Errorf("unexpected output: %s", string(output))
	}

	// The output is in the format "Filesystem 1K-blocks Used Available Use% Mounted on"
	// We are interested in the second line
	parts := strings.Fields(lines[1])
	if len(parts) < 5 {
		return 0, 0, fmt.Errorf("unexpected output: %s", string(output))
	}

	// Parse the size (1K-blocks)
	size, err = strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	// Parse the used space
	used, err = strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	return used * 512, size * 512, nil
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