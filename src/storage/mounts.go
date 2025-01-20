package storage

import (
	"os"
	"errors"
	"k8s.io/utils/mount"
	"strings"
	"bufio"
	"fmt"

	"github.com/azukaar/cosmos-server/src/utils"
)

type MountPoint struct {
	Path string `json:"path"`
	Permenant bool `json:"permenant"`
	Device string `json:"device"`
	Type string `json:"type"`
	Opts []string `json:"opts"`
}

// ListMounts lists all the mount points on the system
func ListMounts() ([]MountPoint, error) {
	utils.Log("[STORAGE] Listing all mount points")

	// Create a new mounter
	mounter := mount.New("")

	// List all mounted filesystems
	mountPoints, err := mounter.List()
	if err != nil {
		return nil, err
	}

	// filter out the mount points that are not disks in a Set
	finalMountPoints := map[string]MountPoint{}
	for i := 0; i < len(mountPoints); i++ {
		path := mountPoints[i].Path
		if strings.HasPrefix(path, "/mnt/host") && utils.IsInsideContainer {
			// remove the host path
			path = strings.Replace(path, "/mnt/host", "", 1)
		}

		utils.Debug("[STORAGE] Checking if " + path + " is a disk")

		// if not proc or sys or dev or run
		if !strings.HasPrefix(path, "/mnt") &&
			 !strings.HasPrefix(path, "/var/mnt") {
			continue
		}

		isPermenant, err := isMountPointInFstab(path)
		if err != nil {
			return nil, err
		}
		
		finalMountPoints[path] = MountPoint{
			Device: mountPoints[i].Device,
			Path: path,
			Type: mountPoints[i].Type,
			Opts: mountPoints[i].Opts,
			Permenant: isPermenant,
		}

		// finalMountPoints = append(finalMountPoints, MountPoint{
		// 	MountPoint: mountPoints[i],
		// 	Path: path,
		// 	Permenant: isPermenant,
		// })
	}

	// use df -h to get the disk usage
	// TODO

	return utils.Values(finalMountPoints), nil
}



// Mount mounts a filesystem located at 'path' to 'mountpoint'.
func Mount(path, mountpoint string, permanent bool, chown string) error {
	utils.Log("[STORAGE] Mounting " + path + " to " + mountpoint)

	// Check if mountpoint exists
	if _, err := os.Stat(mountpoint); os.IsNotExist(err) {
		utils.Log("[STORAGE] Mountpoint does not exist, creating " + mountpoint)
		// Create the mountpoint if it does not exist
		if err := os.Mkdir(mountpoint, 0750); err != nil {
			return err
		}
	} else {
		utils.Log("[STORAGE] Mountpoint exists, checking if it is empty")
		// Check if mountpoint is empty
		dir, _ := os.Open(mountpoint)
		defer dir.Close()

		files, _ := dir.Readdirnames(1) // Or use Readdir to get FileInfo
		if len(files) > 0 {
			return errors.New("mountpoint is not empty")
		}
	}

	// Execute the mount command
	out, err := utils.Exec("mount", path, mountpoint)
	utils.Debug(out)
	if err != nil {
		return err
	}
	
	if chown != "" {
		utils.Log("[STORAGE] Chowning " + mountpoint + " to " + chown)
		out, err := utils.Exec("chown", chown, mountpoint)
		utils.Debug(out)
		if err != nil {
			return err
		}
	}

	if permanent {
		utils.Log("[STORAGE] Adding mountpoint to /etc/fstab")
		
		// Check if mountpoint is already in /etc/fstab
		exists, err := isMountPointInFstab(mountpoint)
		if err != nil {
			return err
		}
		if exists {
			return nil
		}

		// Format the fstab entry
		fstabEntry := fmt.Sprintf("\n%s %s auto defaults 0 0\n", path, mountpoint)

		// Append to /etc/fstab
		file, err := os.OpenFile("/etc/fstab", os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer file.Close()

		if _, err := file.WriteString(fstabEntry); err != nil {
			return err
		}

		utils.Log("[STORAGE] Added mountpoint to /etc/fstab")
	}

	return nil
}


// Unmount unmounts the filesystem at 'mountpoint'.
func Unmount(mountpoint string, permanent bool) error {
	utils.Log("[STORAGE] Unmounting " + mountpoint)

	// Execute the umount command
	_, err := utils.Exec("umount", mountpoint)
	if err != nil {
		return err
	}

	isPermanent, _ := isMountPointInFstab(mountpoint)
	if !isPermanent || permanent {
		utils.Log("[STORAGE] Mountpoint is not in /etc/fstab or unmount is permanent, removing mountpoint")
		// Remove the mountpoint
		if err := os.Remove(mountpoint); err != nil {
			return err
		}
	}

	if permanent {
		utils.Log("[STORAGE] Removing mountpoint from /etc/fstab")

		// Remove entry from /etc/fstab if it exists
		if err := removeFstabEntry(mountpoint); err != nil {
			return err
		}
	}

	return nil
}


// isMountPointInFstab checks if the given mountpoint is already in /etc/fstab.
func isMountPointInFstab(mountpoint string) (bool, error) {
	mountpoint = strings.Replace(mountpoint, "/var/mnt/", "/mnt/", 1)

	file, err := os.Open("/etc/fstab")
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), " " + mountpoint + " ") {
			return true, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	return false, nil
}

// removeFstabEntry removes an entry for the specified mountpoint from /etc/fstab.
func removeFstabEntry(mountpoint string) error {
	mountpoint = strings.Replace(mountpoint, "/var/mnt/", "/mnt/", 1)

	file, err := os.Open("/etc/fstab")
	if err != nil {
		return err
	}
	defer file.Close()

	lines := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		
		isComment := strings.HasPrefix(line, "#")
		isEmpty := len(strings.TrimSpace(line)) == 0

		if !isEmpty && (!strings.Contains(line, " " + mountpoint + " ") || isComment) {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	utils.Debug("[STORAGE] Writing new /etc/fstab")

	return os.WriteFile("/etc/fstab", []byte(strings.Join(lines, "\n")), 0644)
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