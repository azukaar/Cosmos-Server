package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strings"
)

type ReleaseAsset struct {
    Name        string `json:"name"`
    DownloadURL string `json:"browser_download_url"`
}

type Release struct {
    TagName string         `json:"tag_name"`
    PreRelease bool       `json:"prerelease"`
    Assets  []ReleaseAsset `json:"assets"`
}

type VersionInfo struct {
    Version     string
    AMDURL      string
    ARMURL      string
    AMDMD5      string
    ARMMD5      string
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
        return nil, fmt.Errorf("no suitable release found")
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
                info.AMDMD5 = md5String
            }
            if strings.Contains(name, "arm64") {
                info.ARMMD5 = md5String
            }
        }
    }

    return info, nil
}
