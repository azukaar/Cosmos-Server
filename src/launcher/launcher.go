package main

import (
	"fmt"
	"os"
	"crypto/md5"
	"os/exec"
	"encoding/hex"
	"io"
	"archive/zip"
	"runtime"
	"path/filepath"
	"strings"
	// "os/exec"
)

// func run(binaryPath string, args ...string) error {
// 	// Create the command
// 	cmd := exec.Command(binaryPath, args...)

// 	// Redirect STDIN, STDOUT, and STDERR to the current process
// 	cmd.Stdin = os.Stdin
// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr

// 	// Run the command
// 	err := cmd.Run()
// 	if err != nil {
// 			return fmt.Errorf("error running command: %w", err)
// 	}
// 	return nil
// }

func unzip(src string, dest string) error {
    // Open the zip file
    r, err := zip.OpenReader(src)
    if err != nil {
        return err
    }
    defer r.Close()

    // Determine the root directory name in the zip
    var rootFolder string
    if len(r.File) > 0 {
        firstFile := r.File[0].Name
        rootFolder = strings.Split(firstFile, "/")[0] + "/"
    }

    // Loop through each file in the archive
    for _, f := range r.File {
        // Skip the root folder from the file path
        fpath := filepath.Join(dest, strings.TrimPrefix(f.Name, rootFolder))

        // Check if the file matches "cosmos-launcher" or "cosmos-launcher-arm64" for renaming
        baseName := filepath.Base(fpath)
				
				if baseName == "start.sh" {
					// skip
					continue
				}

        if baseName == "cosmos-launcher" || baseName == "cosmos-launcher-arm64" {
            fpath = filepath.Join(filepath.Dir(fpath), baseName+".updated")
        }

        // Check for directory and create it if necessary
        if f.FileInfo().IsDir() {
            os.MkdirAll(fpath, os.ModePerm)
            continue
        }

        // Ensure that directories are created if necessary
        if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
            return err
        }

        // Open the file inside the zip archive
        rc, err := f.Open()
        if err != nil {
            return err
        }
        defer rc.Close()

        // Create the destination file
        outFile, err := os.Create(fpath)
        if err != nil {
            return err
        }
        defer outFile.Close()

        // Copy the content of the file
        _, err = io.Copy(outFile, rc)
        if err != nil {
            return err
        }
    }
    return nil
}


func main() {
	fmt.Println("-- Cosmos Cloud Launcher --")
	fmt.Println("Checking for updates to install...")
	
	// killall cosmos procedss before updating
	cmd := exec.Command("killall", "cosmos")
	cmd.Run()


	execPath, err := os.Executable()
	if err != nil {
		fmt.Println("checkUpdatesAvailable", err)
		return
	}
	
	currentFolder := filepath.Dir(execPath)

	dlPath := currentFolder + "/cosmos-update.zip"
	betaFile := currentFolder + "/.BETA"

	// if there's no updates
	if _, err := os.Stat(dlPath); err != nil {
		fmt.Println("No updates to install, starting Cosmos...")
		return
	}

	isBeta := false
	if _, err := os.Stat(betaFile); err == nil {
		isBeta = true
	}
	
	v, err := GetLatestVersion(isBeta)

	if err != nil {
		fmt.Println(err)
		return
	}

	hash := v.AMDMD5
	if runtime.GOARCH == "arm64" {
		hash = v.ARMMD5
	}
	
	// check md5
	if hash != "" {
		fmt.Println("Update found, checking MD5 hash...")
		file, err := os.Open(dlPath)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		
		hasher := md5.New()
		if _, err := io.Copy(hasher, file); err != nil {
			fmt.Println(err)
			return
		}
		
		if hash != hex.EncodeToString(hasher.Sum(nil)) {
			fmt.Println("[ERROR] MD5 mismatch - aborting update!")
			// delete the file
			os.Remove(dlPath)
			return 
		}

		// extract the file
		fmt.Println("MD5 hash matches, extracting update...")
		err = unzip(dlPath, currentFolder)
		if err != nil {
			fmt.Println(err)
			return
		}

		// delete the file
		os.Remove(dlPath)

		fmt.Println("Update complete, starting Cosmos...")
	}
}