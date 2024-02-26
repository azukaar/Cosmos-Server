package storage 

import (
	"os/exec"
	"os"
	"errors"
	"strings"
	"io"
	"fmt"

	"github.com/azukaar/cosmos-server/src/utils"
)

// Mount mounts a filesystem located at 'path' to 'mountpoint'.
func MountMergerFS(paths []string, mountpoint string, opts string, permanent bool, chown string) error {
	utils.Log("[STORAGE] Merging " + strings.Join(paths, ":") + " into " + mountpoint)

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
			return err
		}
	}

	if len(opts) > 0 && opts[0] != ',' {
		opts = "," + opts
	}

	// Execute the mount command
	cmd := exec.Command("mergerfs", "-o", "use_ino,cache.files=partial,dropcacheonclose=true,allow_other,category.create=mfs" + opts, strings.Join(paths, ":"), mountpoint)
	if err := cmd.Run(); err != nil {
		return err
	}

	utils.Log("[STORAGE] command: mergerfs -o use_ino,cache.files=partial,dropcacheonclose=true,allow_other,category.create=mfs" + opts + " " + strings.Join(paths, ":") + " " + mountpoint)

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
		fstabEntry := fmt.Sprintf("%s %s mergerfs use_ino,cache.files=partial,dropcacheonclose=true,allow_other,category.create=mfs%s 0 0\n", strings.Join(paths, ":"), mountpoint, opts)

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
