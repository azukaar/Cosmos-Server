package main

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"github.com/azukaar/cosmos-server/src/utils"
)

func RunBackup() {
	utils.Log("Backup: Running backup")

	inputDir := utils.CONFIGFOLDER
	outputDir := utils.CONFIGFOLDER

	if utils.GetMainConfig().BackupOutputDir != "" {
		outputDir = utils.GetMainConfig().BackupOutputDir
		if os.Getenv("HOSTNAME") != "" {
			outputDir = "/mnt/host/" + os.Getenv("HOSTNAME")
		}
	}

	// Recursive file listing helper
	var fileList []string
	err := filepath.Walk(inputDir, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() && filepath.Ext(f.Name()) != ".zip" {
			relPath, err := filepath.Rel(inputDir, path)
			if err != nil {
				return err
			}
			fileList = append(fileList, relPath)
		}
		return nil
	})
	if err != nil {
		utils.MajorError("Backup: Error reading directory", err)
		return
	}

	// create a new zip file
	zipfile, err := os.Create(outputDir + "/cosmos-backup" + ".temp.zip")
	if err != nil {
		utils.MajorError("Backup: Error creating zip file", err)
		return
	}
	defer zipfile.Close()

	// create a new zip archive
	zipw := zip.NewWriter(zipfile)
	defer zipw.Close()

	// loop through the files
	for _, file := range fileList {
		f, err := os.Open(filepath.Join(inputDir, file))
		if err != nil {
			utils.MajorError("Backup: Error opening file", err)
			return
		}
		defer f.Close()

		// add file to zip
		zipf, err := zipw.Create(file)
		if err != nil {
			utils.MajorError("Backup: Error adding file to zip", err)
			return
		}

		_, err = io.Copy(zipf, f)
		if err != nil {
			utils.MajorError("Backup: Error copying file", err)
			return
		}
	}

	// rename the file
	err = os.Rename(outputDir+"/cosmos-backup.temp.zip", outputDir+"/cosmos-backup.zip")
	if err != nil {
		utils.MajorError("Backup: Error renaming file", err)
		return
	}

	utils.Log("Backup: Backup complete")

	utils.TriggerEvent(
		"cosmos.backup",
		"Cosmos Backup Succesful",
		"success",
		"",
		map[string]interface{}{},
	)
}
