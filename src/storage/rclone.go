package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/accounting"
	"github.com/rclone/rclone/fs/config"
	"github.com/rclone/rclone/fs/config/configfile"
	"github.com/rclone/rclone/vfs"
	"github.com/rclone/rclone/vfs/vfscommon"
	"github.com/rclone/rclone/cmd/mountlib"
	"github.com/rclone/rclone/cmd/serve/nfs"
	"github.com/rclone/rclone/cmd/serve/s3"
	"github.com/rclone/rclone/cmd/serve/sftp"
	"github.com/rclone/rclone/cmd/serve/webdav"
	"github.com/rclone/rclone/cmd/serve/proxy"
	_ "github.com/rclone/rclone/backend/all"
	_ "github.com/rclone/rclone/cmd/config"

	"github.com/rclone/rclone/cmd"

	"github.com/azukaar/cosmos-server/src/utils"
)

// RunRClone runs rclone with the given arguments (CLI passthrough)
func RunRClone(args []string) {
	configLocation := utils.CONFIGFOLDER + "rclone.conf"
	config.SetConfigPath(configLocation)

	cmd.Root.SetArgs(args)
	cmd.Main()
}

// ServeHandle is the interface that serve servers implement
type ServeHandle interface {
	Serve() error
	Shutdown() error
}

var (
	rcloneMutex       sync.Mutex
	liveMounts        = make(map[string]*mountlib.MountPoint) // For VFS stats only
	liveServers       = make(map[string]ServeHandle)
	liveCancels       = make(map[string]context.CancelFunc)
	mountsMutex       sync.RWMutex
	serversMutex      sync.RWMutex
	rcloneInitialized bool
	rcloneCtx         context.Context
	rcloneCancelAll   context.CancelFunc
)

func initRCloneLibrary(configLocation string) error {
	rcloneMutex.Lock()
	defer rcloneMutex.Unlock()

	if rcloneInitialized {
		return nil
	}

	if err := config.SetConfigPath(configLocation); err != nil {
		return fmt.Errorf("error setting config path: %w", err)
	}

	// Load the config file
	configfile.Install()

	rcloneCtx, rcloneCancelAll = context.WithCancel(context.Background())
	rcloneInitialized = true
	utils.Log("[RemoteStorage] RClone library initialized")
	return nil
}

func StopAllRCloneProcess(forever bool) {
	rcloneMutex.Lock()
	defer rcloneMutex.Unlock()

	utils.Log("[RemoteStorage] Restarting Samba service to remove shares")
	utils.Exec("smbcontrol", "all", "reload-config")

	utils.Log("[RemoteStorage] Stopping all RClone mounts and servers")
	stopAllServers()
	rcloneUnmountAll()

	if forever && rcloneCancelAll != nil {
		rcloneCancelAll()
		rcloneInitialized = false
	}
}

func stopAllServers() {
	serversMutex.Lock()
	defer serversMutex.Unlock()

	for key, server := range liveServers {
		utils.Log(fmt.Sprintf("[RemoteStorage] Stopping server %s", key))
		if cancel, ok := liveCancels[key]; ok {
			cancel()
		}
		if server != nil {
			server.Shutdown()
		}
	}
	liveServers = make(map[string]ServeHandle)
	liveCancels = make(map[string]context.CancelFunc)
}

func rcloneUnmountAll() error {
	utils.Log("[RemoteStorage] Unmounting all remote storages")

	// Get storage list from config file
	storageList, err := getStorageList()
	if err != nil {
		utils.Error("[RemoteStorage] Error getting storage list for unmounting", err)
		return err
	}

	baseDir := "/mnt/cosmos-storage-"
	if utils.IsInsideContainer {
		baseDir = "/mnt/host/mnt/cosmos-storage-"
	}

	for _, storage := range storageList {
		mountPath := baseDir + storage.Name
		utils.Log(fmt.Sprintf("[RemoteStorage] Unmounting %s", mountPath))
		Unmount(mountPath, false)
	}
	return nil
}

func rcloneUnmount(mountPath string, silent bool) error {
	utils.Exec("fusermount", "-u", "-z", mountPath) // in case of stale mount
	return Unmount(mountPath, silent)
}

func Restart() {
	if !utils.FBL.LValid {
		return
	}

	StopAllRCloneProcess(false)

	configLocation := utils.CONFIGFOLDER + "rclone.conf"
	if err := initRCloneLibrary(configLocation); err != nil {
		utils.MajorError("[RemoteStorage] Failed to reinitialize RClone library", err)
		return
	}

	utils.Log("[RemoteStorage] RClone restarted and ready!")
	remountAll()
}

type RemoteStorage struct {
	Name  string
	Chown string
}

func mountRemoteStorage(remoteStorage RemoteStorage) error {
	baseDir := "/mnt/cosmos-storage-"
	if utils.IsInsideContainer {
		baseDir = "/mnt/host/mnt/cosmos-storage-"
	}

	mountPoint := baseDir + remoteStorage.Name

	if _, err := os.Stat(mountPoint); !os.IsNotExist(err) {
		rcloneUnmount(mountPoint, true)
	}

	if _, err := os.Stat(mountPoint); os.IsNotExist(err) {
		if err := os.MkdirAll(mountPoint, 0755); err != nil {
			return fmt.Errorf("[RemoteStorage] error creating mount point directory: %w", err)
		}
	}

	files, err := os.ReadDir(mountPoint)
	if err != nil {
		return fmt.Errorf("[RemoteStorage] error reading mount point directory: %w", err)
	}

	if len(files) > 0 {
		utils.Warn(fmt.Sprintf("[RemoteStorage] Mount point %s is not empty. Moving to backup location.", mountPoint))
		backupPath := mountPoint + ".old"
		counter := 1
		for {
			if _, err := os.Stat(backupPath); os.IsNotExist(err) {
				break
			}
			backupPath = fmt.Sprintf("%s.old%d", mountPoint, counter)
			counter++
		}
		if err := os.Rename(mountPoint, backupPath); err != nil {
			return fmt.Errorf("[RemoteStorage] error moving mount point to backup: %w", err)
		}
		if err := os.MkdirAll(mountPoint, 0755); err != nil {
			return fmt.Errorf("[RemoteStorage] error recreating mount point directory: %w", err)
		}
	}

	chown := remoteStorage.Chown
	if chown != "" {
		utils.Exec("chown", chown, mountPoint)
	}

	uid, gid := 1000, 1000
	if chown != "" {
		parts := strings.Split(chown, ":")
		if len(parts) != 2 {
			return fmt.Errorf("[RemoteStorage] invalid chown value: %s", chown)
		}
		uid, _ = strconv.Atoi(parts[0])
		gid, _ = strconv.Atoi(parts[1])
	}

	remotePath := fmt.Sprintf("%s:", remoteStorage.Name)
	ctx, cancel := context.WithCancel(rcloneCtx)

	f, err := fs.NewFs(ctx, remotePath)
	if err != nil {
		cancel()
		return fmt.Errorf("[RemoteStorage] error creating filesystem for %s: %w", remotePath, err)
	}

	// Configure VFS options
	vfsOpt := vfscommon.Opt
	vfsOpt.CacheMode = vfscommon.CacheModeFull
	vfsOpt.CacheMaxAge = fs.Duration(24 * time.Hour)
	vfsOpt.ChunkSize = 10 * fs.Mebi
	vfsOpt.ChunkSizeLimit = 100 * fs.Mebi

	// TODO: Improve this 
	vfsOpt.DirPerms = 0770
	vfsOpt.FilePerms = 0660
	vfsOpt.UID = uint32(uid)
	vfsOpt.GID = uint32(gid)

	mountOpt := mountlib.Opt
	mountOpt.AllowOther = true

	mountFn := mountlib.GetMountFn("mount")
	if mountFn == nil {
		cancel()
		return fmt.Errorf("[RemoteStorage] mount function not available")
	}

	mp := mountlib.NewMountPoint(mountFn, mountPoint, f, &mountOpt, &vfsOpt)

	// Mount runs in background
	_, err = mp.Mount()
	if err != nil {
		cancel()
		return fmt.Errorf("[RemoteStorage] error mounting remote storage: %w", err)
	}

	// Give it a moment to start
	time.Sleep(1000 * time.Millisecond)

	// Track mount for VFS stats only (not used for unmounting)
	mountsMutex.Lock()
	liveMounts[mountPoint] = mp
	mountsMutex.Unlock()

	utils.Log(fmt.Sprintf("[RemoteStorage] Successfully mounted %s to %s", remoteStorage.Name, mountPoint))
	return nil
}

var CachedRemoteStorageList []StorageInfo

func getStorageList() ([]RemoteStorage, error) {
	result := config.DumpRcBlob()
	CachedRemoteStorageList = []StorageInfo{}

	var storageList []RemoteStorage
	for key, value := range result {
		if key == "install_id" || key == "client_id" || key == "client_secret" {
			continue
		}
		utils.Debug("[RemoteStorage] Found storage: " + key)
		storage := RemoteStorage{Name: key}

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

	return storageList, nil
}

func startServeInstance(protocol, target, addr string, settings map[string]string) error {
	ctx, cancel := context.WithCancel(rcloneCtx)

	f, err := fs.NewFs(ctx, target)
	if err != nil {
		cancel()
		return fmt.Errorf("[RemoteStorage] error creating filesystem for serve: %w", err)
	}

	vfsOpt := vfscommon.Opt
	proxyOpt := proxy.Opt // proxy options needed by serve functions

	var server ServeHandle

	switch protocol {
	case "sftp":
		opt := sftp.DefaultOpt
		opt.ListenAddr = addr
		if v, ok := settings["user"]; ok {
			opt.User = v
		}
		if v, ok := settings["pass"]; ok {
			opt.Pass = v
		}
		if v, ok := settings["authorized-keys"]; ok {
			opt.AuthorizedKeys = v
		}
		server, err = sftp.NewServer(ctx, f, &opt, &vfsOpt, &proxyOpt)

	case "webdav":
		opt := webdav.DefaultOpt
		opt.HTTP.ListenAddr = []string{addr}
		if v, ok := settings["user"]; ok {
			opt.Auth.BasicUser = v
		}
		if v, ok := settings["pass"]; ok {
			opt.Auth.BasicPass = v
		}
		server, err = webdav.NewServer(ctx, f, &opt, &vfsOpt, &proxyOpt)

	case "nfs":
		opt := nfs.DefaultOpt
		opt.ListenAddr = addr
		vfsLayer := vfs.New(f, &vfsOpt)
		server, err = nfs.NewServer(ctx, vfsLayer, &opt)

	case "s3":
		opt := s3.DefaultOpt
		opt.HTTP.ListenAddr = []string{addr}
		if v, ok := settings["auth-key"]; ok {
			opt.AuthKey = []string{v}
		}
		server, err = s3.NewServer(ctx, f, &opt, &vfsOpt, &proxyOpt)

	default:
		cancel()
		return fmt.Errorf("[RemoteStorage] unsupported serve protocol: %s", protocol)
	}

	if err != nil {
		cancel()
		return fmt.Errorf("[RemoteStorage] error creating %s server: %w", protocol, err)
	}

	go func() {
		utils.Log(fmt.Sprintf("[RemoteStorage] Starting %s server on %s", protocol, addr))
		if err := server.Serve(); err != nil {
			select {
			case <-ctx.Done():
			default:
				utils.Error(fmt.Sprintf("[RemoteStorage] %s server error", protocol), err)
			}
		}
	}()

	serverKey := fmt.Sprintf("%s:%s", protocol, addr)
	serversMutex.Lock()
	liveServers[serverKey] = server
	liveCancels[serverKey] = cancel
	serversMutex.Unlock()

	return nil
}

type StorageRoutes struct {
	Name        string
	Protocol    string
	Source      string
	Target      string
	SmartShield bool
}

var StorageRoutesList []StorageRoutes

func remountAll() {
	utils.WaitForAllJobs()
	StorageRoutesList = []StorageRoutes{}

	storageList, err := getStorageList()
	if err != nil {
		utils.MajorError("[RemoteStorage] Error getting remote storage list for mounting", err)
		return
	}

	for _, remoteStorage := range storageList {
		utils.Log(fmt.Sprintf("[RemoteStorage] Mounting %s", remoteStorage.Name))
		if err := mountRemoteStorage(remoteStorage); err != nil {
			utils.MajorError("[RemoteStorage] Error mounting remote storage", err)
			return
		}
	}

	shares := utils.GetMainConfig().RemoteStorage.Shares
	cleanupCosmosSambaShares()

	var sambaShares []utils.LocationRemoteStorageConfig
	for _, share := range shares {
		if share.Protocol == "smb" || share.Protocol == "samba" {
			if err := startSambaShare(share); err != nil {
				utils.MajorError("[RemoteStorage] [SAMBA] Error setting up Samba user", err)
				continue
			}
			sambaShares = append(sambaShares, share)
		} else {
			urlN, _ := url.Parse(share.Route.Target)
			addr := "127.0.0.1:" + urlN.Port()
			if err := startServeInstance(share.Protocol, share.Target, addr, share.Settings); err != nil {
				utils.MajorError(fmt.Sprintf("[RemoteStorage] Error starting %s server", share.Protocol), err)
			}
		}
	}

	if len(sambaShares) > 0 {
		writeCosmosSambaShares(sambaShares)
	}
	utils.Exec("smbcontrol", "all", "reload-config")
}

func API_Rclone_remountAll(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "GET" {
		Restart()
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "OK"})
	} else {
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
	}
}

var signalHandlerSetup bool

func setupSignalHandler() {
	if signalHandlerSetup {
		return
	}
	signalHandlerSetup = true

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-sigChan
		utils.Log(fmt.Sprintf("[RemoteStorage] Received signal %v, unmounting all storages...", sig))
		rcloneUnmountAll()
		utils.Log("[RemoteStorage] All storages unmounted, exiting...")
		os.Exit(0)
	}()
}

// Called multiuple times on config change
func InitRemoteStorage() bool {
	utils.StopAllRCloneProcess = StopAllRCloneProcess
	configLocation := utils.CONFIGFOLDER + "rclone.conf"

	if !utils.FBL.LValid {
		utils.Warn("RemoteStorage: No valid licence found, not starting module.")
		return false
	}

	if err := initRCloneLibrary(configLocation); err != nil {
		utils.MajorError("[RemoteStorage] Failed to initialize RClone library", err)
		return false
	}

	setupSignalHandler()

	go func() {
		time.Sleep(1 * time.Second)
		utils.Log("[RemoteStorage] RClone library ready!")
		remountAll()
	}()

	return true
}

type RcloneStatsObj struct {
	Bytes  float64
	Errors float64
}

func RCloneStats() (RcloneStatsObj, error) {
	if utils.FBL == nil || !utils.FBL.LValid || !rcloneInitialized {
		return RcloneStatsObj{}, nil
	}

	stats := accounting.GlobalStats()
	params, err := stats.RemoteStats(false)
	if err != nil {
		return RcloneStatsObj{}, fmt.Errorf("error getting rclone stats: %w", err)
	}

	bytes, _ := params["bytes"].(float64)
	errors, _ := params["errors"].(float64)
	return RcloneStatsObj{bytes, errors}, nil
}
