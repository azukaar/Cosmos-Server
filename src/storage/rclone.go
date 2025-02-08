package storage 

import (
	"os"
	"os/exec"
	"io"
	"bufio"
	"bytes"
	"net/http"
	"net/url"
	"time"
	"strings"
	"fmt"
	"encoding/json"
	"io/ioutil"
	"strconv"
	"sync"
	"crypto/md5"
	"syscall"
	
	"github.com/rclone/rclone/cmd"
	"github.com/fsnotify/fsnotify"	
	_ "github.com/rclone/rclone/backend/all"
	_ "github.com/rclone/rclone/cmd/all"
	_ "github.com/rclone/rclone/lib/plugin"

	"github.com/azukaar/cosmos-server/src/utils"
)


var rcloneMutex = &sync.Mutex{}

type RCloneProcess struct {
	RcloneCmd     *exec.Cmd
	RcloneRestart chan bool
	RestartCount  int
	LastRestart   time.Time
	Main bool
}
func (rp *RCloneProcess) SetRcloneRestartCount(v int) {
	rp.RestartCount = v
}
func (rp *RCloneProcess) SetLastRestart(v time.Time) {
	rp.LastRestart = v
}

var rcloneProcesses = make(map[int]*RCloneProcess)

func RunRClone(args []string) {
	cmd.Root.SetArgs(args)
	cmd.Main()
}

func RunRCloneCommand(command []string) (*exec.Cmd, io.WriteCloser, *bytes.Buffer, *bytes.Buffer) {
	utils.Log("[RemoteStorage] Running Rclone process")
	utils.Debug("[RemoteStorage] Running Rclone process with command: " + strings.Join(command, " "))

	args := []string{"rclone"}
	args = append(args, command...)
	cmd := exec.Command(os.Args[0], args...)

	var stdoutBuf, stderrBuf bytes.Buffer
	
	stdin, err := cmd.StdinPipe()
	if err != nil {
		utils.Error("[RemoteStorage] Error creating stdin pipe", err)
		return nil, nil, nil, nil
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		utils.Error("[RemoteStorage] Error creating stdout pipe", err)
		return nil, nil, nil, nil
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		utils.Error("[RemoteStorage] Error creating stderr pipe", err)
		return nil, nil, nil, nil
	}

	err = cmd.Start()
	if err != nil {
		utils.Error("[RemoteStorage] Error starting rclone command", err)
		return nil, nil, nil, nil
	}

	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			line := scanner.Text()
			utils.Debug("[RemoteStorage] " + line)
			stdoutBuf.WriteString(line + "\n")
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			line := scanner.Text()
			utils.Error("[RemoteStorage] " + line, nil)
			stderrBuf.WriteString(line + "\n")
		}
	}()

	return cmd, stdin, &stdoutBuf, &stderrBuf
}

func monitorRCloneProcess(cmd *exec.Cmd, args ...string) {
	rcloneProcesses[cmd.Process.Pid].RcloneRestart = make(chan bool)

	err := cmd.Wait()
	if err != nil {
		utils.Error("[RemoteStorage] RClone process exited with error", err)
	} else {
		utils.Log("[RemoteStorage] RClone process exited")
	}

	rcloneMutex.Lock()
	// stop monitoring if the process is not in the map
	if _, ok := rcloneProcesses[cmd.Process.Pid]; !ok {
		rcloneMutex.Unlock()
		return
	}
	rcloneMutex.Unlock()

	// Monitor and restart RClone process if needed
	go func() {
		for {
			select {
			case <-rcloneProcesses[cmd.Process.Pid].RcloneRestart:
				utils.Debug("[RemoteStorage] RcloneRestart signal received")
				startRCloneProcess(args...)
				return
			}
		}
	}()

	now := time.Now()
	if now.Sub(rcloneProcesses[cmd.Process.Pid].LastRestart) < 10*time.Second {
		rcloneProcesses[cmd.Process.Pid].SetRcloneRestartCount(rcloneProcesses[cmd.Process.Pid].RestartCount + 1)
	} else {
		rcloneProcesses[cmd.Process.Pid].SetRcloneRestartCount(1)
	}
	rcloneProcesses[cmd.Process.Pid].SetLastRestart(now)

	if rcloneProcesses[cmd.Process.Pid].RestartCount <= 3 {
		utils.Log("[RemoteStorage] Restarting RClone process")
		rcloneProcesses[cmd.Process.Pid].RcloneRestart <- true
	} else {
		utils.MajorError("[RemoteStorage] RClone process restarted too many times in a short period. Stopping automatic restarts.", nil)
	}
}

func startRCloneProcess(args ...string) {
	rcloneMutex.Lock()
	defer rcloneMutex.Unlock()

	utils.Log("[RemoteStorage] Starting RClone process")

	configLocation := utils.CONFIGFOLDER + "rclone.conf"

	if len(args) == 0 {
		rcloneCmd, _, _, _ := RunRCloneCommand([]string{"rcd", "--rc-user=" + utils.ProxyRCloneUser, "--rc-pass=" + utils.ProxyRClonePwd, "--config=" + configLocation, "--rc-baseurl=/cosmos/rclone"})
		rcloneProcesses[rcloneCmd.Process.Pid] = &RCloneProcess{
			RcloneCmd: rcloneCmd,
			Main: true,
		}
		go monitorRCloneProcess(rcloneCmd)
	} else {
		rcloneCmd, _, _, _ := RunRCloneCommand(append(args, "--config=" + configLocation))
		rcloneProcesses[rcloneCmd.Process.Pid] = &RCloneProcess{
			RcloneCmd: rcloneCmd,
		}
		go monitorRCloneProcess(rcloneCmd, args...)
	}
}

var isWaitingToStop = false
func StopAllRCloneProcess(forever bool) {
	rcloneMutex.Lock()
	defer rcloneMutex.Unlock()
	
	if isWaitingToStop {
		return
	}

	isWaitingToStop = true

	// TODO: wait for backups to finish

	for _, process := range rcloneProcesses {
		if process.Main {
			unmountAll()
		}

		if process.RcloneCmd != nil && process.RcloneCmd.Process != nil {
			utils.Log("[RemoteStorage] Stopping RClone process")

			err := process.RcloneCmd.Process.Signal(syscall.SIGTERM)
			if err != nil {
				utils.Error("[RemoteStorage] Error stopping RClone process", err)
			}
			// Wait for the process to exit
			_, _ = process.RcloneCmd.Process.Wait()

			if !process.Main || forever {
				delete(rcloneProcesses, process.RcloneCmd.Process.Pid)
			}
		}
	}

	isWaitingToStop = false
}

func unmountAll() error {
	utils.Log("[RemoteStorage] Unmounting all remote storages")

	// Get the list of current mounts
	response, err := runRDC("/mount/listmounts")
	if err != nil {
		return fmt.Errorf("error getting mount list: %w", err)
	}

	var mountList struct {
		Mounts []struct {
			MountPoint string `json:"MountPoint"`
		} `json:"mounts"`
	}

	if err := json.Unmarshal(response, &mountList); err != nil {
		return fmt.Errorf("error parsing mount list response: %w", err)
	}

	// Unmount each mount point
	for _, mount := range mountList.Mounts {
		utils.Log(fmt.Sprintf("[RemoteStorage] Unmounting %s", mount.MountPoint))

		unmountPayload := map[string]string{
			"mountPoint": mount.MountPoint,
		}

		payloadBytes, err := json.Marshal(unmountPayload)
		if err != nil {
			utils.Error(fmt.Sprintf("[RemoteStorage] Error marshaling unmount payload for %s", mount.MountPoint), err)
			continue
		}

		_, err = runRDC("/mount/unmount", "json", string(payloadBytes))
		if err != nil {
			utils.Error(fmt.Sprintf("[RemoteStorage] Error unmounting %s", mount.MountPoint), err)
			continue
		}

		utils.Log(fmt.Sprintf("[RemoteStorage] Successfully unmounted %s", mount.MountPoint))
	}

	return nil
}


func Restart() {
	if !utils.FBL.LValid {
		return
	}

	StopAllRCloneProcess(false)

	// wait for rclone to start
	go func() {
		retries := 0
		for {
			_, err := runRDC("/core/version")
			if err == nil { break }

			time.Sleep(2 * time.Second)
			if retries > 5 {
				utils.MajorError("[RemoteStorage] Failed to reach RClone, check the port 5572 is free", nil)
				return
			}
			retries++
		}

		utils.Log("[RemoteStorage] RClone started and ready!")

		remountAll()
	}()
}

func runRDC(path string, params ...string) ([]byte, error) {
	baseURL := "http://localhost:5572/cosmos/rclone"
	fullURL := fmt.Sprintf("%s%s", baseURL, path)

	utils.Debug("[RemoteStorage] Sending request to RClone server: " + fullURL)

	var req *http.Request
	var err error

	if len(params) == 2 && params[0] == "json" {
		// If a JSON payload is provided
		req, err = http.NewRequest("POST", fullURL, strings.NewReader(params[1]))
		if err != nil {
			return nil, fmt.Errorf("[RemoteStorage] error creating request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		utils.Debug("[RemoteStorage] Request payload: " + params[1])
	} else {
		// For regular key-value params
		data := url.Values{}
		for i := 0; i < len(params); i += 2 {
			if i+1 < len(params) {
				data.Set(params[i], params[i+1])
			}
		}
		req, err = http.NewRequest("POST", fullURL, strings.NewReader(data.Encode()))
		if err != nil {
			return nil, fmt.Errorf("[RemoteStorage] error creating request: %w", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		utils.Debug("[RemoteStorage] Request payload: " + data.Encode())
	}

	req.SetBasicAuth(utils.ProxyRCloneUser, utils.ProxyRClonePwd)

	// Create an HTTP client with a timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Check for non-200 status code
	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error != "" {
			return nil, fmt.Errorf("RClone server error: %s", errorResp.Error)
		}
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return body, nil
}

type RemoteStorage struct {
	Name string
	Chown string
}

func mountRemoteStorage(remoteStorage RemoteStorage) error {
	// Create the base directory if it doesn't exist
	baseDir := "/mnt/cosmos-storage-"

	if utils.IsInsideContainer {
		baseDir = "/mnt/host/mnt/cosmos-storage-"
	}
	
	// Create the storage-specific directory
	mountPoint := baseDir + remoteStorage.Name

	if _, err := os.Stat(mountPoint); !os.IsNotExist(err) {
		// ensure the mount point is unmounted
		Unmount(mountPoint, true)
	}

	if _, err := os.Stat(mountPoint); os.IsNotExist(err) {
		if err := os.MkdirAll(mountPoint, 0755); err != nil {
			return fmt.Errorf("[RemoteStorage] error creating mount point directory: %w", err)
		}
	}

	// if there are files in the directory, error 
	files, err := ioutil.ReadDir(mountPoint)
	if err != nil {
		return fmt.Errorf("[RemoteStorage] error reading mount point directory: %w", err)
	}

	if len(files) > 0 {
		return fmt.Errorf("[RemoteStorage] mount point directory is not empty")
	}

	chown := remoteStorage.Chown
	if chown != "" {
		utils.Log("[STORAGE] Chowning " + mountPoint + " to " + chown)
		out, err := utils.Exec("chown", chown, mountPoint)
		utils.Debug(out)
		if err != nil {
			return err
		}
	}

	// Prepare the mount command
	remotePath := fmt.Sprintf("%s:", remoteStorage.Name) // Assuming the remote name in rclone config matches the storageName
	
	uid, gid := 1000, 1000 // Default to user 1000:1000
	
	if chown != "" {
		if len(strings.Split(chown, ":")) != 2 {
			return fmt.Errorf("[RemoteStorage] invalid chown value: %s", chown)
		} else {
			uids, gids := strings.Split(chown, ":")[0], strings.Split(chown, ":")[1]
			uid, _ = strconv.Atoi(uids)
			gid, _ = strconv.Atoi(gids)
		}
	}

	payload := map[string]interface{}{
		"fs":         remotePath,
		"mountPoint": mountPoint,
		"mountType":  "mount",
		"vfsOpt": map[string]interface{}{
			"CacheMode":         "full",
			"CacheMaxAge":       "24h",
			"ReadChunkSize":     "10M",
			"ReadChunkSizeLimit": "100M",
			"UID":               uid,
			"GID":               gid,
			"Umask":             077, // This sets permissions to 700 for directories and 600 for files
		},
		"mountOpt": map[string]interface{}{
			// "AllowNonEmpty": true,
			"AllowOther":    true,
		},
	}
	
	// Convert payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling payload: %w", err)
	}
	
	// Execute the mount command using runRDC
	_, err = runRDC("/mount/mount", "json", string(payloadBytes))

	if err != nil {
		return fmt.Errorf("[RemoteStorage] error mounting remote storage: %w", err)
	}

	utils.Log(fmt.Sprintf("[RemoteStorage] Successfully mounted %s to %s", remoteStorage.Name, mountPoint))
	return nil
}

var CachedRemoteStorageList []StorageInfo

func getStorageList() ([]RemoteStorage, error) {
	response, err := runRDC("/config/dump")
	if err != nil {
		return nil, fmt.Errorf("error getting config dump: %w", err)
	}

	CachedRemoteStorageList = []StorageInfo{}

	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return nil, fmt.Errorf("error parsing config dump response: %w", err)
	}

	// Extract the storage names and chown values from the config
	var storageList []RemoteStorage
	for key, value := range result {
		utils.Debug("[RemoteStorage] Found storage: " + key)
		// Exclude the special keys that are not storage names
		if key != "install_id" && key != "client_id" && key != "client_secret" {
			storage := RemoteStorage{
				Name: key,
			}

			// Extract the cosmos-chown value if it exists
			if storageConfig, ok := value.(map[string]interface{}); ok {
				if chown, exists := storageConfig["cosmos-chown"]; exists {
					if chownStr, ok := chown.(string); ok {
						storage.Chown = chownStr
					}
				}
			}

			storageList = append(storageList, storage)
			
			CachedRemoteStorageList = append(CachedRemoteStorageList, StorageInfo{
				Name: storage.Name,
				Path: "/mnt/cosmos-storage-" + storage.Name,
			})
		}
	}

	return storageList, nil
}

func watchConfigFile(configLocation string, restart func()) {
	utils.Log("[RemoteStorage] Watching config file for changes")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		utils.Error("[RemoteStorage] Error creating file watcher, falling back to polling only", err)
		watchConfigFilePollingOnly(configLocation, restart)
		return
	}
	defer watcher.Close()

	lastHash := getFileHash(configLocation)
	lastModTime := getFileModTime(configLocation)
	var mutex sync.Mutex
	normalMechanismDetectedChange := false

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					mutex.Lock()
					if fileChanged(configLocation, &lastHash, &lastModTime) {
						utils.Log("[RemoteStorage] Config file modified (detected by watcher). Restarting...")
						normalMechanismDetectedChange = true
						restart()
					}
					mutex.Unlock()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				utils.Error("[RemoteStorage] Error watching config file", err)
			}
		}
	}()

	err = watcher.Add(configLocation)
	if err != nil {
		utils.Error("[RemoteStorage] Error adding config file to watcher, falling back to polling only", err)
		watchConfigFilePollingOnly(configLocation, restart)
		return
	}

	// Fallback polling mechanism
	for {
		time.Sleep(5 * time.Second)
		mutex.Lock()
		if !normalMechanismDetectedChange && fileChanged(configLocation, &lastHash, &lastModTime) {
			utils.Log("[RemoteStorage] Config file modified (detected by polling). Restarting...")
			restart()
		}
		normalMechanismDetectedChange = false
		mutex.Unlock()
	}
}
func watchConfigFilePollingOnly(configLocation string, restart func()) {
	utils.Log("[RemoteStorage] Using polling method to watch config file")
	lastHash := getFileHash(configLocation)
	lastModTime := getFileModTime(configLocation)

	for {
		time.Sleep(5 * time.Second)
		if fileChanged(configLocation, &lastHash, &lastModTime) {
			utils.Log("[RemoteStorage] Config file modified (detected by polling). Restarting...")
			restart()
		}
	}
}

func getFileHash(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		utils.Error("[RemoteStorage] Error opening file for hashing", err)
		return ""
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		utils.Error("[RemoteStorage] Error calculating file hash", err)
		return ""
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
}

func getFileModTime(filePath string) time.Time {
	info, err := os.Stat(filePath)
	if err != nil {
		utils.Error("[RemoteStorage] Error getting file info", err)
		return time.Time{}
	}
	return info.ModTime()
}

func fileChanged(filePath string, lastHash *string, lastModTime *time.Time) bool {
	currentHash := getFileHash(filePath)
	currentModTime := getFileModTime(filePath)

	if currentHash != *lastHash || !currentModTime.Equal(*lastModTime) {
		*lastHash = currentHash
		*lastModTime = currentModTime
		return true
	}

	return false
}

type StorageRoutes struct {
	Name string
	Protocol string
	Source string
	Target string
	SmartShield bool
}

var StorageRoutesList []StorageRoutes

func remountAll() {
	utils.WaitForAllJobs() 
	
	StorageRoutesList = []StorageRoutes{}

	// Mount remote storages
	storageList, err := getStorageList()
	if err != nil {
		utils.MajorError("[RemoteStorage] Error getting remote storage list for mounting", err)
		return
	}

	for _, remoteStorage := range storageList {
		utils.Log(fmt.Sprintf("[RemoteStorage] Mounting %s", remoteStorage))
		if err := mountRemoteStorage(remoteStorage); err != nil {
			utils.MajorError("[RemoteStorage] Error mounting remote storage", err)
			return
		}
	}

	shares := utils.GetMainConfig().RemoteStorage.Shares
	for _, share := range shares {
		utils.Log("[RemoteStorage] Sharing " + share.Target)

		argsShare := []string{"serve"}
		// addr, err := utils.GetNextAvailableLocalPort(12000)
		// if err != nil {
		// 	utils.MajorError("[RemoteStorage] Error: cannot find a free port to share on network", err)
		// 	return
		// }
		urlN, _ := url.Parse(share.Route.Target)
		addr := "127.0.0.1:" + urlN.Port()
		// remote scehme

		utils.Debug("[RemoteStorage] Sharing on port " + addr)

		argsShare = append(argsShare, "--addr="+addr)
		argsShare = append(argsShare, share.Protocol)
		for k,v := range share.Settings {
			argsShare = append(argsShare, "--"+k+"="+v)
		}
		argsShare = append(argsShare, share.Target)

		startRCloneProcess(argsShare...)
	}
}

func API_Rclone_remountAll(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "GET" {
		Restart()

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
		})
	} else {
		utils.Error("API_Rclone_remountAll: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

func InitRemoteStorage() bool {
	utils.StopAllRCloneProcess = StopAllRCloneProcess
	
	configLocation := utils.CONFIGFOLDER + "rclone.conf"
	utils.ProxyRCloneUser = utils.GenerateRandomString(8)
	utils.ProxyRClonePwd = utils.GenerateRandomString(16)
	
	if !utils.FBL.LValid {
		utils.Warn("RemoteStorage: No valid licence found, not starting module.")
		return false
	}
	
	if _, err := os.Stat(configLocation); os.IsNotExist(err) {
		utils.Log("[RemoteStorage] Creating rclone config file")
		file, err := os.Create(configLocation)
		if err != nil {
			utils.Error("[RemoteStorage] Error creating rclone config file", err)
			return false
		}
		file.Close()
	}

	utils.Log("[RemoteStorage] Initializing remote storage")
	startRCloneProcess()

	// Start watching the config file
	go watchConfigFile(configLocation, Restart)

	// wait for rclone to start
	go func() {
		retries := 0
		for {
			_, err := runRDC("/core/version")
			if err == nil { break }

			time.Sleep(2 * time.Second)
			if retries > 5 {
				utils.MajorError("[RemoteStorage] Failed to reach RClone, check the port 5572 is free", nil)
				return
			}
			retries++
		}

		utils.Log("[RemoteStorage] RClone started and ready!")

		remountAll()
	}()

	return true
}

type RcloneStatsObj struct {
	Bytes float64
	Errors float64 
}

func RCloneStats() (RcloneStatsObj, error) {
	utils.Debug("[RemoteStorage] Getting rclone stats")
	
	if utils.FBL == nil || !utils.FBL.LValid {
		return RcloneStatsObj{0, 0}, nil
	}

	response, err := runRDC("/core/stats")
	if err != nil {
		return RcloneStatsObj{0, 0}, fmt.Errorf("error getting rclone stats: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return RcloneStatsObj{0, 0}, fmt.Errorf("error parsing rclone stats response: %w", err)
	}

	bytes, _ := result["bytes"].(float64)
	errors, _ := result["errors"].(float64)

	return RcloneStatsObj{bytes, errors}, nil
}