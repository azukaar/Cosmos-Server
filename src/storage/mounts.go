package storage

import (
	"io/ioutil"
	"os/exec"
	"os"
	"errors"
	"k8s.io/utils/mount"
	"strings"

	"github.com/azukaar/cosmos-server/src/utils"
)

// ListMounts lists all the mount points on the system
func ListMounts() ([]mount.MountPoint, error) {
	utils.Log("[STORAGE] Listing all mount points")

	// Create a new mounter
	mounter := mount.New("")

	// List all mounted filesystems
	mountPoints, err := mounter.List()
	if err != nil {
		return nil, err
	}

	// filter out the mount points that are not disks
	finalMountPoints := []mount.MountPoint{}
	for i := 0; i < len(mountPoints); i++ {
		// if not proc or sys or dev or run
		if strings.HasPrefix(mountPoints[i].Path, "/proc") ||
		   strings.HasPrefix(mountPoints[i].Path, "/sys") ||
			 strings.HasPrefix(mountPoints[i].Path, "/dev") || 
			 strings.HasPrefix(mountPoints[i].Path, "/run") ||
			 mountPoints[i].Type == "tmpfs" {
			continue
		}
		finalMountPoints = append(finalMountPoints, mountPoints[i])
	}

	// use df -h to get the disk usage
	// TODO

	return finalMountPoints, nil
}


// mount disk
func MountDisk(diskPath string, mountPath string) error {
	utils.Log("[STORAGE] Mounting disk " + diskPath + " to " + mountPath)

	// Create the mount point if it doesn't exist
	if _, err := ioutil.ReadDir(mountPath); os.IsNotExist(err) {
		if err := os.MkdirAll(mountPath, 0755); err != nil {
			return err
		}
	}

	// Mount the disk
	cmd := exec.Command("mount", diskPath, mountPath)

	// Run the command
	output, err := cmd.CombinedOutput()

	if err != nil {
		return errors.New("Mount: " + string(output) + " " + err.Error())
	}


	return nil
}

// unmount disk
func UnmountDisk(mountPath string) error {
	utils.Log("[STORAGE] Unmounting disk " + mountPath)

	// Unmount the disk
	cmd := exec.Command("umount", mountPath)

	// Run the command
	output, err := cmd.CombinedOutput()

	if err != nil {
		return errors.New("Unmount: " + string(output) + " " + err.Error())
	}

	

	return nil
}

// check if disk is mounted
func IsDiskMounted(diskPath string) (bool, error) {
	utils.Log("[STORAGE] Checking if disk " + diskPath + " is mounted using kubernetes/mount-utils")

	// Create a new mounter
	mounter := mount.New("")

	// List all mounted filesystems
	mountPoints, err := mounter.List()
	if err != nil {
		return false, err
	}

	// Check if the disk is in the list of mounted filesystems
	for _, mp := range mountPoints {
		if mp.Path == diskPath {
			return true, nil
		}
	}

	return false, nil
}