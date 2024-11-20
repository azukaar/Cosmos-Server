package main

import (
    "encoding/json"
    "fmt"
    "io"
    "os"
    "net/http"
    "path/filepath"
    "bufio"
    "strings"
    "syscall"
)

type ReleaseAsset struct {
    Name        string `json:"name"`
    DownloadURL string `json:"browser_download_url"`
}

type Release struct {
    TagName string         `json:"tag_name"`
    PreRelease bool        `json:"prerelease"`
    Assets  []ReleaseAsset `json:"assets"`
}

type VersionInfo struct {
    Version     string
    AMDURL      string
    ARMURL      string
    AMDMD5      string
    ARMMD5      string
}

func parseMD5File(content string) string {
	// Read first line only (there should only be one anyway)
	scanner := bufio.NewScanner(strings.NewReader(content))
	if scanner.Scan() {
			// Split by whitespace and take first part (the hash)
			parts := strings.Fields(scanner.Text())
			if len(parts) > 0 {
					return parts[0]
			}
	}
	return ""
}

func GetLatestVersion(includeBeta bool) (*VersionInfo, error) {
    // Fetch releases from GitHub API
    resp, err := http.Get("https://api.github.com/repos/azukaar/cosmos-server/releases")
    if err != nil {
        return nil, fmt.Errorf("failed to fetch releases: %v", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response body: %v", err)
    }

    var releases []Release
    if err := json.Unmarshal(body, &releases); err != nil {
        return nil, fmt.Errorf("failed to parse JSON: %v", err)
    }

    // Find latest release that matches criteria
    var latestRelease *Release
    for _, release := range releases {
        if !includeBeta && release.PreRelease {
            continue
        }
        latestRelease = &release
        break
    }

    if latestRelease == nil {
        return nil, nil
    }

    // Initialize version info
    info := &VersionInfo{
        Version: latestRelease.TagName,
    }

    // Parse assets to find URLs and MD5s
    for _, asset := range latestRelease.Assets {
        name := strings.ToLower(asset.Name)
        
        // Match AMD64 binary
        if strings.Contains(name, "amd64") && !strings.HasSuffix(name, ".md5") {
            info.AMDURL = asset.DownloadURL
        }
        // Match ARM64 binary
        if strings.Contains(name, "arm64") && !strings.HasSuffix(name, ".md5") {
            info.ARMURL = asset.DownloadURL
        }
        // Match MD5 files
        if strings.HasSuffix(name, ".md5") {
            // Fetch MD5 content
            md5Resp, err := http.Get(asset.DownloadURL)
            if err != nil {
                continue
            }
            md5Content, err := io.ReadAll(md5Resp.Body)
            md5Resp.Body.Close()
            if err != nil {
                continue
            }
            
            md5String := strings.TrimSpace(string(md5Content))
            
            if strings.Contains(name, "amd64") {
                info.AMDMD5 = parseMD5File(md5String)
            }
            if strings.Contains(name, "arm64") {
                info.ARMMD5 = parseMD5File(md5String)
            }
        }
    }

    return info, nil
}

func cleanUpUpdateFiles() {
    execPath, err := os.Executable()
    if err != nil {
        return
    }
    
    currentFolder := filepath.Dir(execPath)

    dlPath := currentFolder + "/cosmos-update.zip"

    if _, err := os.Stat(dlPath); err == nil {
        os.Remove(dlPath)
    }

    // if cosmos-laumcher.updated exists, rename it to cosmos-launcher
    updatedPath := currentFolder + "/cosmos-launcher.updated"
    if _, err := os.Stat(updatedPath); err == nil {
        // get old permissions
        var perms os.FileMode
        if info, err := os.Stat(currentFolder + "/cosmos-launcher"); err == nil {
            perms = info.Mode()
        } else {
            fmt.Println("Update: Failed to get old permissions:", err)
        }
        // get old owner
        var owner int
        if info, err := os.Stat(currentFolder + "/cosmos-launcher"); err == nil {
            owner = int(info.Sys().(*syscall.Stat_t).Uid)
        } else {
            fmt.Println("Update: Failed to get old owner:", err)
        }

        err := os.Rename(updatedPath, currentFolder + "/cosmos-launcher")
        if err != nil {
            fmt.Println("Update: Failed to rename cosmos-launcher.updated:", err)
        }

        // set old permissions
        if perms != 0 {
            err = os.Chmod(currentFolder + "/cosmos-launcher", perms)
            if err != nil {
                fmt.Println("Update: Failed to set old permissions:", err)
            }
        }
        // set old owner
        if owner != 0 {
            err = os.Chown(currentFolder + "/cosmos-launcher", owner, -1)
            if err != nil {
                fmt.Println("Update: Failed to set old owner:", err)
            }
        }
    }
}