package backups

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"context"
	"time"

	"github.com/creack/pty"
	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/azukaar/cosmos-server/src/cron"
	"github.com/azukaar/cosmos-server/src/docker"
	"encoding/json"
)

var stripAnsi = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]|\x1b\]0;[^\a]*\a|\r`)

// SplitJSONObjects splits a string containing multiple JSON objects, respecting quotes
func SplitJSONObjects(input string) string {
	var objects []string
	inQuotes := false
	escapeNext := false
	start := -1
	braceCount := 0
	
	for i, char := range input {
			if escapeNext {
					escapeNext = false
					continue
			}
			
			switch char {
			case '"':
					if !escapeNext {
							inQuotes = !inQuotes
					}
			case '\\':
					escapeNext = true
			case '{':
					if !inQuotes {
							if braceCount == 0 {
									start = i
							}
							braceCount++
					}
			case '}':
					if !inQuotes {
							braceCount--
							if braceCount == 0 && start != -1 {
									objects = append(objects, input[start:i+1])
									start = -1
							}
					}
			}
	}
	
	return "[" + strings.Join(objects, ",") + "]"
}

func prependResticArgs(args []string) []string {
	rcloneFilePath := utils.CONFIGFOLDER + "rclone.conf"
	
	execPath, _ := os.Executable()

	prependArgs := []string{"-o", "rclone.program=" + execPath, "-o", "rclone.args=rclone serve --config=" + rcloneFilePath + " restic --stdio --b2-hard-delete --verbose"}

	args = append(prependArgs, args...)

	return args
}

// execResticRaw executes a restic command with already-prepended arguments
func execResticRaw(args []string, env []string) (string, error) {
	cmd := exec.Command("./restic",  args...)

	utils.Debug("[Restic] Executing command: restic " + strings.Join(cmd.Args, " "))

	cmd.Env = append(os.Environ(), env...)

	// Start the command with a pseudo-terminal
	f, err := pty.Start(cmd)
	if err != nil {
		return "", fmt.Errorf("[Restic] failed to start command: %w", err)
	}
	defer f.Close()

	filters := []string{"unable to open cache: unable to locate cache directory: neither $XDG_CACHE_HOME nor $HOME are defined"}

	// Create a buffer to store all output
	var output bytes.Buffer
	var tempBuffer bytes.Buffer
	done := make(chan error, 1)

	// Start a goroutine to read the output
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := f.Read(buf)
			if err != nil {
				done <- err
				return
			}

			// output.Write(buf[:n])
			tempBuffer.Write(buf[:n])

			// if not in filters, write to output
			if strings.TrimSpace(tempBuffer.String()) != filters[0] {
				output.Write(buf[:n])
				b := string(buf[:n])
				b = stripAnsi.ReplaceAllString(b, "")
				b = strings.TrimSuffix(b, "\n")
			}

			utils.Debug("[Restic] " + tempBuffer.String())
		}
	}()

	// Wait for the command to finish
	err = cmd.Wait()
	if err != nil {
		return output.String(), fmt.Errorf("[Restic] command failed: %w\nOutput: %s", err, output.String())
	}

	return output.String(), nil
}

// isLockError checks if the output indicates a repository lock error
func isLockError(output string) bool {
	return strings.Contains(output, "repository is already locked")
}

// staleLockThreshold is the minimum age of a lock before it is automatically removed
var staleLockThreshold = 30 * 24 * time.Hour // 1 month

var lockAgeRegex = regexp.MustCompile(`lock was created at .+ \(([0-9hms.]+) ago\)`)

// parseLockAge extracts the lock age from restic error output.
// Returns 0 if the age cannot be parsed.
func parseLockAge(output string) time.Duration {
	matches := lockAgeRegex.FindStringSubmatch(output)
	if len(matches) < 2 {
		return 0
	}
	d, err := time.ParseDuration(matches[1])
	if err != nil {
		return 0
	}
	return d
}

type LockInfo struct {
	Time      string `json:"time"`
	Exclusive bool   `json:"exclusive"`
	Hostname  string `json:"hostname"`
	Username  string `json:"username"`
	PID       int    `json:"pid"`
}

// GetLocks returns lock details for a repository
func GetLocks(repository, password string) []LockInfo {
	args := []string{"list", "locks", "--no-lock", "--repo", repository}
	env := []string{fmt.Sprintf("RESTIC_PASSWORD=%s", password)}

	output, err := ExecRestic(args, env)
	if err != nil {
		return nil
	}

	ids := strings.Fields(strings.TrimSpace(output))
	if len(ids) == 0 {
		return nil
	}

	var locks []LockInfo
	for _, id := range ids {
		catArgs := []string{"cat", "lock", id, "--no-lock", "--repo", repository}
		lockJSON, err := ExecRestic(catArgs, env)
		if err != nil {
			continue
		}
		var lock LockInfo
		if err := json.Unmarshal([]byte(lockJSON), &lock); err != nil {
			continue
		}
		locks = append(locks, lock)
	}
	return locks
}

// UnlockRepository removes locks from a restic repository
func UnlockRepository(repository, password string) error {
	utils.Log("[Restic] Unlocking repository: " + repository)
	args := prependResticArgs([]string{"unlock", "--repo", repository})
	env := []string{fmt.Sprintf("RESTIC_PASSWORD=%s", password)}
	_, err := execResticRaw(args, env)
	return err
}

// ExecRestic executes a restic command with the given arguments and environment variables.
// If the command fails due to a stale repository lock (older than 1 month), it automatically
// unlocks and retries. For newer locks, the error is returned as-is so the user can decide
// whether to unlock manually.
func ExecRestic(args []string, env []string) (string, error) {
	args = prependResticArgs(args)

	output, err := execResticRaw(args, env)

	if err != nil && isLockError(output) {
		lockAge := parseLockAge(output)
		if lockAge >= staleLockThreshold {
			// Lock is very old, safe to auto-unlock
			repo := ""
			for i, arg := range args {
				if arg == "--repo" && i+1 < len(args) {
					repo = args[i+1]
					break
				}
			}
			if repo != "" {
				utils.Log(fmt.Sprintf("[Restic] Repository has a stale lock (%s old), auto-unlocking...", lockAge))
				unlockArgs := prependResticArgs([]string{"unlock", "--repo", repo})
				execResticRaw(unlockArgs, env)
				// Retry the original command
				output, err = execResticRaw(args, env)
			}
		} else {
			utils.Warn(fmt.Sprintf("[Restic] Repository is locked (lock age: %s). Use the unlock button in the UI to remove it if the lock is stale.", lockAge))
		}
	}

	return output, err
}

// CreateRepository initializes a new Restic repository
func CreateRepository(repository, password string) error {
	args := []string{"init", "--repo", repository}
	env := []string{fmt.Sprintf("RESTIC_PASSWORD=%s", password)}

	output, err := ExecRestic(args, env)
	if err != nil {
		return err
	}

	if !strings.Contains(output, "created restic repository") {
		return fmt.Errorf("[Restic] failed to create repository: %s", output)
	}
	return nil
}

// DeleteRepository removes a Restic repository
func DeleteRepository(repository string) error {
	// if repo starts with rclone:
	if strings.HasPrefix(repository, "rclone:") {
		// remove rclone config
		rcloneConfig := strings.TrimPrefix(repository, "rclone:")

		//extract rclone config name
		rcloneConfigArr := strings.Split(rcloneConfig, ":")
		rcloneConfigName := rcloneConfigArr[0]
		rcPath := rcloneConfigArr[1]

		repository = "/mnt/cosmos-storage-" + rcloneConfigName + "/" + rcPath
	}
		
	// First, check if the repository exists
	if _, err := os.Stat(repository); os.IsNotExist(err) {
		return fmt.Errorf("[Restic] repository does not exist: %s", repository)
	}

	// list content, see if it matches a restic repository
	files := []string{"data", "index", "keys", "locks", "snapshots", "config"}

	for _, file := range files {
		if _, err := os.Stat(repository + "/" + file); os.IsNotExist(err) {
			return fmt.Errorf("[Restic] repository does not contain required files: %s", repository)
		}
	}

	// Remove the repository directory
	err := os.RemoveAll(repository)
	if err != nil {
		return fmt.Errorf("[Restic] failed to delete repository: %w", err)
	}
	return nil
}

// EditRepositoryPassword changes the password of an existing repository
func EditRepositoryPassword(repository, currentPassword, newPassword string) error {
	args := []string{"key", "change", "--repo", repository}
	env := []string{
		fmt.Sprintf("RESTIC_PASSWORD=%s", currentPassword),
		fmt.Sprintf("RESTIC_NEW_PASSWORD=%s", newPassword),
	}

	output, err := ExecRestic(args, env)
	if err != nil {
		return err
	}

	if !strings.Contains(output, "changed password") {
		return fmt.Errorf("[Restic] failed to change password: %s", output)
	}
	return nil
}

// CheckRepository verifies if a repository exists and is valid
func CheckRepository(repository, password string) error {
	args := []string{"check", "--no-lock", "--repo", repository}
	env := []string{fmt.Sprintf("RESTIC_PASSWORD=%s", password)}

	_, err := ExecRestic(args, env)
	return err
}

func StatsRepository(repository, password string) (string,error) {
	args := []string{"stats", "--no-lock", "--repo", repository, "--json", "--mode", "raw-data"}
	env := []string{fmt.Sprintf("RESTIC_PASSWORD=%s", password)}

	output, err := ExecRestic(args, env)
	return output,err
}

func StatsRepositorySubfolder(repository, password, snapshot, path string) (string,error) {
	args := []string{"stats", "--no-lock", "--mode", "restore-size", "--repo", repository, "--path", path, snapshot,  "--json"}
	env := []string{fmt.Sprintf("RESTIC_PASSWORD=%s", password)}

	output, err := ExecRestic(args, env)
	return output,err
}

type BackupConfig struct {
	Repository string
	Password   string
	Source     string
	Name       string
	Tags       []string
	Exclude    []string
	Retention	 string
	AutoStopContainers bool
}

// CreateBackupJob creates a backup job configuration
func CreateBackupJob(config BackupConfig, crontab string) {
	utils.Log("Creating backup job for " + config.Name + " with crontab " + crontab)

	args := []string{"backup", "--repo", config.Repository, config.Source}

	// Add tags if specified
	for _, tag := range config.Tags {
		args = append(args, "--tag", tag)
	}

	// Add exclude patterns if specified
	for _, exclude := range config.Exclude {
		args = append(args, "--exclude", exclude)
	}

	env := []string{
		fmt.Sprintf("RESTIC_PASSWORD=%s", config.Password),
	}

	cron.RegisterJob(cron.ConfigJob{
		Scheduler:   "Restic",
		Name:       fmt.Sprintf("Restic backup %s", config.Name),
		Cancellable: true,
		Job:  func(OnLog func(string), OnFail func(error), OnSuccess func(), ctx context.Context, cancel context.CancelFunc) {
			var containers []string
			var err error

			if config.AutoStopContainers {
				containers, err = docker.GetContainersUsingPath(config.Source)
				if err != nil {
					OnFail(err)
					return
				}

				OnLog("Found container(s) using path: " + strings.Join(containers, ", "))


				// Stop all containers
				err = docker.StopContainers(containers)
				if err != nil {
					docker.StartContainers(containers)
					OnFail(err)
					return
				}

				OnLog("Stopped containers, starting backup")
			}

			cron.JobFromCommandWithEnv(env, "./restic", prependResticArgs(args)...)(OnLog, OnFail, OnSuccess, ctx, cancel)

			if config.AutoStopContainers {
				// Start all containers
				err = docker.StartContainers(containers)
				if err != nil {
					OnFail(err)
					return
				}
			}
		},
		Crontab: 		 crontab,
		Resource:   "backup@" + config.Name,
	})
}

func CreateForgetJob(config BackupConfig, crontab string) {
	utils.Log("Creating forget job for " + config.Name + " with crontab " + crontab)

	args := []string{"forget", "--repo", config.Repository, "--prune"}

	ret := config.Retention
	if ret == "" {
		ret = "--keep-last 3 --keep-daily 7 --keep-weekly 8 --keep-yearly 3"
	}

	retArr := strings.Split(ret, " ")
	args = append(args, retArr...)

	// Add tags if specified
	for _, tag := range config.Tags {
		args = append(args, "--tag", tag)
	}

	env := []string{
		fmt.Sprintf("RESTIC_PASSWORD=%s", config.Password),
	}

	cron.RegisterJob(cron.ConfigJob{
		Scheduler:   "Restic",
		Name:       fmt.Sprintf("Restic forget %s", config.Name),
		Cancellable: true,
		Job:        cron.JobFromCommandWithEnv(env, "./restic", prependResticArgs(args)...),
		Crontab: 		 crontab,
		Resource:   "backup@" + config.Name,
	})
}

type RestoreConfig struct {
	Repository  string
	Password    string
	SnapshotID  string
	Target      string
	Name        string
	Include     []string
	OriginalSource string
	AutoStopContainers bool
}

// CreateRestoreJob creates a restore job configuration
func CreateRestoreJob(config RestoreConfig) {
	args := []string{
		"restore",
		"--repo", config.Repository,
		config.SnapshotID + ":/" + config.OriginalSource,
		"--target", config.Target,
	}

	// Add include patterns if specified
	for _, include := range config.Include {
		//remove OriginalSource from include path
		include := strings.TrimPrefix(include, config.OriginalSource)
		utils.Debug("[RESTIC] Restore includes: " + include)
		args = append(args, "--include", include)
	}

	env := []string{
		fmt.Sprintf("RESTIC_PASSWORD=%s", config.Password),
	}

	go (func() {
		cron.RunOneTimeJob(cron.ConfigJob{
			Scheduler:    "Restic",
			Name:         fmt.Sprintf("Restic restore %s", config.Name),
			Cancellable:  true,
			Job:            func(OnLog func(string), OnFail func(error), OnSuccess func(), ctx context.Context, cancel context.CancelFunc) {
				var containers []string
				var err error

				if config.AutoStopContainers {
					containers, err = docker.GetContainersUsingPath(config.OriginalSource)
					if err != nil {
						OnFail(err)
						return
					}

					OnLog("Found container(s) using path: " + strings.Join(containers, ", "))


					// Stop all containers
					err = docker.StopContainers(containers)
					if err != nil {
						docker.StartContainers(containers)
						OnFail(err)
						return
					}

					OnLog("Stopped containers, starting restore")
				}

				cron.JobFromCommandWithEnv(env, "./restic", prependResticArgs(args)...)(OnLog, OnFail, OnSuccess, ctx, cancel)

				if config.AutoStopContainers {
					// Start all containers
					err = docker.StartContainers(containers)
					if err != nil {
						OnFail(err)
						return
					}
				}
			},
			Resource: "backup@" + config.Name,
		})
	})()
}

// ListSnapshots returns a list of all snapshots in the repository
func ListSnapshots(repository, password string) (string, error) {
	args := []string{
		"snapshots",
		"--no-lock",
		"--repo", repository,
		"--json",
	}
	env := []string{fmt.Sprintf("RESTIC_PASSWORD=%s", password)}

	output, err := ExecRestic(args, env)
	if err != nil {
		return "", fmt.Errorf("[Restic] failed to list snapshots: %w", err)
	}

	return output, nil
}

// ListSnapshotsWithFilters returns a filtered list of snapshots
func ListSnapshotsWithFilters(repository, password string, tags []string, host string, path string) (string, error) {
	args := []string{
		"snapshots",
		"--no-lock",
		"--repo", repository,
		"--json",
	}

	// Add optional filters
	for _, tag := range tags {
		args = append(args, "--tag", tag)
	}
	if host != "" {
		args = append(args, "--host", host)
	}
	if path != "" {
		args = append(args, "--path", path)
	}

	env := []string{fmt.Sprintf("RESTIC_PASSWORD=%s", password)}

	output, err := ExecRestic(args, env)
	if err != nil {
		return "", fmt.Errorf("[Restic] failed to list filtered snapshots: %w", err)
	}

	return output, nil
}

// ListDirectory lists the contents of a directory in a specific snapshot
func ListDirectory(repository, password, snapshotID, path string) (string, error) {
	args := []string{
		"ls",
		"--no-lock",
		"--repo", repository,
		snapshotID,
		path,
		"--json",
	}
	env := []string{fmt.Sprintf("RESTIC_PASSWORD=%s", password)}

	output, err := ExecRestic(args, env)
	if err != nil {
		return "", fmt.Errorf("[Restic] failed to list directory contents: %w", err)
	}

	return output, nil
}

// ListDirectoryWithFilters lists directory contents with additional filters
func ListDirectoryWithFilters(repository, password, snapshotID, path string, recursive bool, longFormat bool) (string, error) {
	args := []string{
		"ls",
		"--no-lock",
		"--repo", repository,
		snapshotID,
		path,
		"--json",
	}

	if recursive {
		args = append(args, "--recursive")
	}
	if longFormat {
		args = append(args, "--long")
	}

	env := []string{fmt.Sprintf("RESTIC_PASSWORD=%s", password)}

	output, err := ExecRestic(args, env)
	if err != nil {
		return "", fmt.Errorf("[Restic] failed to list directory contents with filters: %w", err)
	}

	return output, nil
}

func DeleteByTag(repository string, password string, tag string) error {
	// First get all snapshots with this tag
	output, err := ListSnapshotsWithFilters(repository, password, []string{tag}, "", "")
	if err != nil {
			return fmt.Errorf("[Restic] failed to list snapshots for deletion: %w", err)
	}
	
	// Parse the JSON output to get snapshot IDs
	var snapshots []map[string]interface{}
	if err := json.Unmarshal([]byte(output), &snapshots); err != nil {
			return fmt.Errorf("[Restic] failed to parse snapshots: %w", err)
	}
	
	// Extract all snapshot IDs
	var ids []string
	for _, snap := range snapshots {
			ids = append(ids, snap["id"].(string))
	}

	if len(ids) == 0 {
			return nil // No snapshots to delete
	}
	
	// Create forget command with all snapshot IDs
	args := append([]string{"forget", "--repo", repository, "--prune"}, ids...)
	env := []string{fmt.Sprintf("RESTIC_PASSWORD=%s", password)}
	
	// Execute the forget command
	_, err = ExecRestic(args, env)
	if err != nil {
			return fmt.Errorf("[Restic] failed to delete snapshots: %w", err)
	}
	
	return nil
}

func ForgetSnapshot(repository, password, snapshot string) error {
	args := []string{"forget", snapshot, "--prune", "--repo", repository}
	env := []string{fmt.Sprintf("RESTIC_PASSWORD=%s", password)}

	_, err := ExecRestic(args, env)
	if err != nil {
		return fmt.Errorf("[Restic] failed to forget snapshot: %w", err)
	}

	return nil
}
