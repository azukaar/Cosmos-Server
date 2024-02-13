package storage

import (
<<<<<<< HEAD
=======
	"io/ioutil"
>>>>>>> 092a95a ([release] v0.15.0-unstable1)
	"os/exec"
	"os"
	"errors"
	"k8s.io/utils/mount"
	"strings"
<<<<<<< HEAD
	"bufio"
	"io"
	"fmt"
=======
>>>>>>> 092a95a ([release] v0.15.0-unstable1)

	"github.com/azukaar/cosmos-server/src/utils"
)

<<<<<<< HEAD
type MountPoint struct {
	Path string `json:"path"`
	Permenant bool `json:"permenant"`
	Device string `json:"device"`
	Type string `json:"type"`
	Opts []string `json:"opts"`
}

// ListMounts lists all the mount points on the system
func ListMounts() ([]MountPoint, error) {
=======
// ListMounts lists all the mount points on the system
func ListMounts() ([]mount.MountPoint, error) {
>>>>>>> 092a95a ([release] v0.15.0-unstable1)
	utils.Log("[STORAGE] Listing all mount points")

	// Create a new mounter
	mounter := mount.New("")

	// List all mounted filesystems
	mountPoints, err := mounter.List()
	if err != nil {
		return nil, err
	}

	// filter out the mount points that are not disks
<<<<<<< HEAD
	finalMountPoints := []MountPoint{}
	for i := 0; i < len(mountPoints); i++ {
		path := mountPoints[i].Path
		if strings.HasPrefix(path, "/mnt/host") && os.Getenv("HOSTNAME") != "" {
			// remove the host path
			path = strings.Replace(path, "/mnt/host", "", 1)
		}

		// if not proc or sys or dev or run
		if !strings.HasPrefix(path, "/mnt") ||
			 !strings.HasPrefix(path, "/var/mnt") {
			continue
		}

		isPermenant, err := isMountPointInFstab(path)
		if err != nil {
			return nil, err
		}
		
		finalMountPoints = append(finalMountPoints, MountPoint{
			Device: mountPoints[i].Device,
			Path: path,
			Type: mountPoints[i].Type,
			Opts: mountPoints[i].Opts,
			Permenant: isPermenant,
		})

		// finalMountPoints = append(finalMountPoints, MountPoint{
		// 	MountPoint: mountPoints[i],
		// 	Path: path,
		// 	Permenant: isPermenant,
		// })
=======
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
>>>>>>> 092a95a ([release] v0.15.0-unstable1)
	}

	// use df -h to get the disk usage
	// TODO

	return finalMountPoints, nil
}


<<<<<<< HEAD

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

		_, err := dir.Readdirnames(1) // Or use Readdir to get FileInfo
		if err != io.EOF {
			return errors.New("mountpoint is not empty")
		}
	}

	// chown the mountpoint
	if chown != "" {
		utils.Log("[STORAGE] Chowning " + mountpoint + " to " + chown)
		cmd := exec.Command("chown", chown, mountpoint)
		if err := cmd.Run(); err != nil {
=======
// mount disk
func MountDisk(diskPath string, mountPath string) error {
	utils.Log("[STORAGE] Mounting disk " + diskPath + " to " + mountPath)

	// Create the mount point if it doesn't exist
	if _, err := ioutil.ReadDir(mountPath); os.IsNotExist(err) {
		if err := os.MkdirAll(mountPath, 0755); err != nil {
>>>>>>> 092a95a ([release] v0.15.0-unstable1)
			return err
		}
	}

<<<<<<< HEAD
	// Execute the mount command
	cmd := exec.Command("mount", path, mountpoint)
	if err := cmd.Run(); err != nil {
		return err
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
		fstabEntry := fmt.Sprintf("%s %s auto defaults 0 0\n", path, mountpoint)

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
=======
	// Mount the disk
	cmd := exec.Command("mount", diskPath, mountPath)

	// Run the command
	output, err := cmd.CombinedOutput()

	if err != nil {
		return errors.New("Mount: " + string(output) + " " + err.Error())
	}

>>>>>>> 092a95a ([release] v0.15.0-unstable1)

	return nil
}

<<<<<<< HEAD

// Unmount unmounts the filesystem at 'mountpoint'.
func Unmount(mountpoint string, permanent bool) error {
	utils.Log("[STORAGE] Unmounting " + mountpoint)

	// Execute the umount command
	cmd := exec.Command("umount", mountpoint)
	if err := cmd.Run(); err != nil {
		return err
	}

	if permanent {
		utils.Log("[STORAGE] Removing mountpoint from /etc/fstab")

		// Remove the mountpoint
		if err := os.Remove(mountpoint); err != nil {
			return err
		}

		// Remove entry from /etc/fstab if it exists
		if err := removeFstabEntry(mountpoint); err != nil {
			return err
		}
	}
=======
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

	
>>>>>>> 092a95a ([release] v0.15.0-unstable1)

	return nil
}

<<<<<<< HEAD

// isMountPointInFstab checks if the given mountpoint is already in /etc/fstab.
func isMountPointInFstab(mountpoint string) (bool, error) {
	file, err := os.Open("/etc/fstab")
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), mountpoint) {
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
	file, err := os.Open("/etc/fstab")
	if err != nil {
		return err
	}
	defer file.Close()

	lines := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, mountpoint) {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return os.WriteFile("/etc/fstab", []byte(strings.Join(lines, "\n")), 0644)
}

=======
>>>>>>> 092a95a ([release] v0.15.0-unstable1)
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