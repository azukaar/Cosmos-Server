package cron

import (
	"strconv"
	"time"
	"context"
	"errors"
	"io"
	"bufio"
	"os/exec"
	"os"
	"fmt"
	"strings"

	"github.com/go-co-op/gocron/v2"
	"github.com/creack/pty"

	"github.com/azukaar/cosmos-server/src/docker"
	"github.com/azukaar/cosmos-server/src/utils"

	"github.com/docker/docker/api/types"
)

var scheduler gocron.Scheduler
var CRONLock = make(chan bool, 1)
var RunningLock = make(chan bool, 1)

type ConfigJob struct {
	Disabled bool
	Scheduler string
	Name string
	Cancellable bool
	Job func(OnLog func(string), OnFail func(error), OnSuccess func(),  ctx context.Context, cancel context.CancelFunc) `json:"-"`
	Crontab string
	Running bool
	LastStarted time.Time
	LastRun time.Time
	LastRunSuccess bool
	Logs []string
	Ctx context.Context `json:"-"`
	CancelFunc context.CancelFunc `json:"-"`
	Container string
	Timeout time.Duration
	MaxLogs int
	Resource string
}

var jobsList = map[string]map[string]ConfigJob{}
var wasInit = false
var InternalProcessTracker = NewProcessTracker()

func GetJobsList() map[string]map[string]ConfigJob {
	return getJobsList()
}

func Init() {
	utils.Log("Initializing CRON...")

	var err error
	scheduler, err = gocron.NewScheduler()
	if err != nil {
		utils.Fatal("CRON initialization", err)
	}
	
	wasInit = true
	InitJobs()
	InitScheduler()
}

func streamLogs(r io.Reader, OnLog func(string)) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
			OnLog(scanner.Text())
	}
}

func getJobsList() map[string]map[string]ConfigJob {
	return jobsList
}

func RunningJobs() []string {
	runningJobs := []string{}
	for _, schedulerList := range jobsList {
		for _, job := range schedulerList {
			if job.Running {
				runningJobs = append(runningJobs, job.Name)
			}
		}
	}
	return runningJobs
}

func JobFromCommand(command string, args ...string) func(OnLog func(string), OnFail func(error), OnSuccess func(), ctx context.Context, cancel context.CancelFunc) {
	return JobFromCommandWithEnv([]string{}, command, args...)
}

func JobFromCommandWithEnv(env []string, command string, args ...string) func(OnLog func(string), OnFail func(error), OnSuccess func(), ctx context.Context, cancel context.CancelFunc) {
	return func(OnLog func(string), OnFail func(error), OnSuccess func(), ctx context.Context, cancel context.CancelFunc) {
			done := make(chan bool, 1)
			var cmdErr error

			go func() {
					defer func() {
							if r := recover(); r != nil {
									OnFail(fmt.Errorf("panic in command execution: %v", r))
							}
							done <- true
					}()

					cmd := exec.CommandContext(ctx, command, args...)
					cmd.Env = append(os.Environ(), env...)
					
					// Create a pseudo-terminal (PTY)
					ptmx, err := pty.Start(cmd)
					if err != nil {
							cmdErr = fmt.Errorf("failed to start pty: %v", err)
							return
					}

					// Create a single channel for log completion
					logsDone := make(chan bool, 1)

					// Stream both stdout and stderr through the PTY
					go func() {
							buf := make([]byte, 1024)
							for {
									n, err := ptmx.Read(buf)
									if err != nil {
											// if err != io.EOF {
											// 		cmdErr = fmt.Errorf("pty read error: %v", err)
											// }
											logsDone <- true
											return
									}
									if n > 0 {
											OnLog(string(buf[:n]))
									}
							}
					}()

					// Wait for log streaming to complete
					<-logsDone

					// Wait for the command to complete
					if err := cmd.Wait(); err != nil {
							ptmx.Close()
							cmdErr = fmt.Errorf("command failed: %v", err)
							return
					}

					ptmx.Close()
			}()

			// Handle completion and context cancellation
			select {
			case <-done:
					if cmdErr != nil {
							OnFail(cmdErr)
					} else {
							OnSuccess()
					}
			case <-ctx.Done():
					OnFail(ctx.Err())
			}
	}
}

// Optional: Add a helper function to set terminal size if needed
func setTerminalSize(ptmx *os.File, rows, cols uint16) error {
	return pty.Setsize(ptmx, &pty.Winsize{
			Rows: rows,
			Cols: cols,
	})
}

type ExecuterFn func(OnLog func(string), OnFail func(error), OnSuccess func(), ctx context.Context, cancel context.CancelFunc)

func JobFromContainerCommand(containerID string, command string, args ...string) func(OnLog func(string), OnFail func(error), OnSuccess func(), ctx context.Context, cancel context.CancelFunc) {
	return func(OnLog func(string), OnFail func(error), OnSuccess func(), ctx context.Context, cancel context.CancelFunc) {
			// Connect to Docker
			err := docker.Connect()
			if err != nil {
					OnFail(err)
					return
			}

			// Check if the container is running
			containerJSON, err := docker.DockerClient.ContainerInspect(ctx, containerID)
			if err != nil {
					OnFail(err)
					return
			}

			if !containerJSON.State.Running {
					OnFail(fmt.Errorf("Container %s is not running", containerID))
					return
			}

			// Create exec configuration
			execConfig := types.ExecConfig{
					Cmd:          append([]string{command}, args...),
					AttachStdout: true,
					AttachStderr: true,
			}

			// Create exec instance
			execID, err := docker.DockerClient.ContainerExecCreate(ctx, containerID, execConfig)
			if err != nil {
					OnFail(err)
					return
			}

			// Attach to exec instance
			execAttach, err := docker.DockerClient.ContainerExecAttach(ctx, execID.ID, types.ExecStartCheck{})
			if err != nil {
					OnFail(err)
					return
			}
			defer execAttach.Close()

			// Stream logs from exec
			go streamLogs(execAttach.Reader, OnLog)

			// Inspect exec process to wait for completion
			for {
					execInspect, err := docker.DockerClient.ContainerExecInspect(ctx, execID.ID)
					if err != nil {
							OnFail(err)
							return
					}

					if !execInspect.Running {
							if execInspect.ExitCode == 0 {
									OnSuccess()
							} else {
									OnFail(fmt.Errorf("exec process exited with code %d", execInspect.ExitCode))
							}
							break
					}

					// Don't spam the API
					time.Sleep(500 * time.Millisecond)
			}
	}
}

func InitJobs() {
	CRONLock <- true
	defer func() { <-CRONLock }()

	utils.Log("Initializing custom jobs from config...")

	// copy jobsList
	resetScheduler("Custom")
	
	config := utils.GetMainConfig()
	configJobsList := config.CRON

	for _, job := range configJobsList {
		cmd := JobFromCommand("sh", "-c", job.Command)

		if job.Container != "" {
			cmd = JobFromContainerCommand(job.Container, "sh", "-c", job.Command)
		}

		j := ConfigJob{
			Scheduler: "Custom",
			Name: job.Name,
			Job: cmd,
			Crontab: job.Crontab,
			Running: false,
			Cancellable: true,
			LastRun: time.Time{},
			LastStarted: time.Time{},
			Logs: []string{},
			Disabled: !job.Enabled,
			Container: job.Container,
		}

		if CustomScheduler, ok := jobsList["Custom"]; ok {
			if old, ok := CustomScheduler[job.Name]; ok {
				j.Logs = old.Logs
				j.Running = old.Running
				j.LastStarted = old.LastStarted
				j.LastRun = old.LastRun
				j.LastRunSuccess = old.LastRunSuccess
				j.Ctx = old.Ctx
				j.CancelFunc = old.CancelFunc
			}
		}

		registerJob(j)
	}
}

func jobRunner(schedulerName, jobName string) func(OnLog func(string), OnFail func(error), OnSuccess func()) {
	return func(OnLog func(string), OnFail func(error), OnSuccess func()) {
			RunningLock <- true
			defer func() { <-RunningLock }()

			CRONLock <- true
			
			var job ConfigJob
			var ok bool
			
			if job, ok = jobsList[schedulerName][jobName]; !ok {
					utils.Error("Scheduler: job "+jobName+" not found", nil)
					<-CRONLock
					return
			}

			if job.Running {
					utils.Error("Scheduler: job "+job.Name+" is already running", nil)
					<-CRONLock
					return
			}

			// Create context with timeout
			// if job.Timeout == 0 {
			// 		job.Timeout = 240 * time.Hour
			// }

			// ctx, cancel := context.WithTimeout(context.Background(), )
			ctx, cancel := context.WithCancel(context.Background())
			job.Ctx = ctx
			job.LastStarted = time.Now()
			job.LastRunSuccess = false
			job.Running = true
			job.Logs = []string{}
			job.CancelFunc = cancel
			jobsList[job.Scheduler][job.Name] = job
			<-CRONLock

			// Ensure cleanup happens
			defer func() {
					CRONLock <- true
					if j, ok := jobsList[schedulerName][jobName]; ok {
							j.Running = false
							j.CancelFunc = nil
							jobsList[schedulerName][jobName] = j
					}
					<-CRONLock
					cancel()
			}()

			triggerJobEvent(job, "started", "CRON job " + job.Name + " started", "info", map[string]interface{}{})

			triggerJobUpdated("start", job.Name)
			
			InternalProcessTracker.StartProcess()
			
			job.Job(OnLog, OnFail, OnSuccess, ctx, cancel)
	}
}

func jobRunner_OnLog(schedulerName, jobName string) func(log string) {
	return func(log string) {
			CRONLock <- true
			if job, ok := jobsList[schedulerName][jobName]; ok {
					// Implement circular buffer for logs
					maxlog := job.MaxLogs
					if maxlog == 0 {
							maxlog = 5000
					}
					if maxlog != 0 && len(job.Logs) >= maxlog {
							job.Logs = job.Logs[1:]
					}
					job.Logs = append(job.Logs, log)
					jobsList[job.Scheduler][job.Name] = job
					utils.Debug(log)
					triggerJobUpdated("log", job.Name, log)
			}
			<-CRONLock
	}
}

func jobRunner_OnFail(schedulerName, jobName string) func(err error) {
	return func(err error) {
		CRONLock <- true
		if job, ok := jobsList[schedulerName][jobName]; ok {
			job.Logs = append(job.Logs, err.Error())
			job.LastRunSuccess = false
			job.LastRun = time.Now()
			job.Running = false
			jobsList[job.Scheduler][job.Name] = job
			triggerJobUpdated("fail", job.Name, err.Error())
			utils.MajorError("CRON job " + job.Name + " failed", err)
			
			InternalProcessTracker.EndProcess()

			triggerJobEvent(job, "fail", "CRON job " + job.Name + " failed", "error", map[string]interface{}{
				"error": err.Error(),
			})
		}
		<-CRONLock
	}
}
func jobRunner_OnSuccess(schedulerName, jobName string) func() {
	return func() {
		CRONLock <- true
		if job, ok := jobsList[schedulerName][jobName]; ok {
			job.LastRunSuccess = true
			job.LastRun = time.Now()
			job.Running = false
			jobsList[job.Scheduler][job.Name] = job
			triggerJobUpdated("success", job.Name)
			utils.Log("CRON job " + job.Name + " finished")
			
			InternalProcessTracker.EndProcess()

			triggerJobEvent(job, "success", "CRON job " + job.Name + " finished", "success", map[string]interface{}{})
		}
		<-CRONLock
	}
}

func WaitForAllJobs() {
	utils.Log("Waiting for " + strconv.Itoa(InternalProcessTracker.count) + " jobs to finish...")
	InternalProcessTracker.WaitForZero()
}

func triggerJobEvent(job ConfigJob, eventId, eventTitle, eventLevel string, extra map[string]interface{}) {
	utils.TriggerEvent(
		"cosmos.cron." + strings.Replace(job.Scheduler, ".", "_", -1) + "." + strings.Replace(job.Name, ".", "_", -1) + "." + eventId,
		eventTitle,
		eventLevel,
		"job@" + job.Scheduler + "@" + job.Name,
		extra,
	)

	if job.Resource != "" {
		utils.TriggerEvent(
			"cosmos.cron-resource." + strings.Replace(job.Scheduler, ".", "_", -1) + "." + strings.Replace(job.Name, ".", "_", -1) + "." + eventId,
			eventTitle,
			eventLevel,
			job.Resource,
			extra,
		)
	}

	if job.Container != "" {
		utils.TriggerEvent(
			"cosmos.cron-container." + strings.Replace(job.Scheduler, ".", "_", -1) + "." + strings.Replace(job.Name, ".", "_", -1) + "." + eventId,
			eventTitle,
			eventLevel,
			"container@" + job.Container,
			extra,
		)
	}
}

func InitScheduler() {
	utils.WaitForAllJobs = WaitForAllJobs

	var err error
	
	if !wasInit {
		return
	}

	utils.Log("Initializing CRON jobs to scheduler...")
	
	CRONLock <- true
	defer func() { <-CRONLock }()

	// clear
	scheduler.Shutdown()
	scheduler, err = gocron.NewScheduler()
	if err != nil {
		utils.Fatal("CRON initialization", err)
	}

	count := 0
	for _, schedulerList := range jobsList {
		for _, job := range schedulerList {
			if job.Disabled {
				continue
			}

			_, err := scheduler.NewJob(
				gocron.CronJob(job.Crontab, true),
				gocron.NewTask(
					jobRunner(job.Scheduler, job.Name),
					jobRunner_OnLog(job.Scheduler, job.Name),
					jobRunner_OnFail(job.Scheduler, job.Name),
					jobRunner_OnSuccess(job.Scheduler, job.Name),
				),
			)

			if err != nil {
				utils.MajorError("CRON job scheduling", err)
			} else {
				count++
			}
		}
	}

	// start
	
	go func() {
		scheduler.Start()
		utils.Log("CRON scheduler started with " + strconv.Itoa(count) + " jobs")
	}()
}

func cancelJob(scheduler string, jobName string) error {
	if job, ok := jobsList[scheduler][jobName]; ok {
		if !job.Running {
			return errors.New("Job is not running")
		}
		if !job.Cancellable {
			return errors.New("Job is not cancellable")
		}
		job.CancelFunc()
	}
	return nil
}

func CancelJob(scheduler string, jobName string) error {
	CRONLock <- true
	defer func() { <-CRONLock }()

	utils.Log("Canceling CRON job: " + jobName)
	
	return cancelJob(scheduler, jobName)
}

func registerJob(job ConfigJob) {
	// create scheduler if not exists
	if _, ok := jobsList[job.Scheduler]; !ok {
		jobsList[job.Scheduler] = map[string]ConfigJob{}
	}

	utils.Log("Registering CRON job: " + job.Name)

	jobsList[job.Scheduler][job.Name] = job
}

func RegisterJob(job ConfigJob) {
	CRONLock <- true
	
	registerJob(job)

	<-CRONLock

	InitScheduler()
}

func resetScheduler(scheduler string) {
	jobsList[scheduler] = map[string]ConfigJob{}

	utils.Log("Resetting CRON scheduler: " + scheduler)
}
func ResetScheduler(scheduler string) {
	CRONLock <- true
	
	resetScheduler(scheduler)

	<-CRONLock

	InitScheduler()
}

func DeregisterJob(scheduler string, name string) {
	CRONLock <- true

	utils.Log("Deregistering CRON job: " + name)

	delete(jobsList[scheduler], name)

	<-CRONLock

	InitScheduler()
}

func ManualRunJob(scheduler string, name string) error {
	CRONLock <- true

	if job, ok := jobsList[scheduler][name]; ok {
		utils.Log("Manually start CRON job: " + name)
		<-CRONLock
		// Execute the job
		go (func() {
			jobRunner(job.Scheduler, job.Name)(jobRunner_OnLog(job.Scheduler, job.Name), jobRunner_OnFail(job.Scheduler, job.Name), jobRunner_OnSuccess(job.Scheduler, job.Name))
		})()
	} else {
		<-CRONLock
		utils.Error("CRON job " + name + " not found", nil)
		return errors.New("CRON job not found")
	}
	
	return nil
}


func RunOneTimeJob(job ConfigJob) {
	CRONLock <- true

	job.Scheduler = "__OT__" + job.Scheduler
	job.Name = job.Name + " #" + strconv.FormatInt(time.Now().Unix(), 10)
	
	// create scheduler if not exists
	if _, ok := jobsList[job.Scheduler]; !ok {
		jobsList[job.Scheduler] = map[string]ConfigJob{}
	}

	utils.Log("Registering one time CRON job: " + job.Name)

	jobsList[job.Scheduler][job.Name] = job

	<-CRONLock

	// Execute the job
	jobRunner(job.Scheduler, job.Name)(jobRunner_OnLog(job.Scheduler, job.Name), jobRunner_OnFail(job.Scheduler, job.Name), jobRunner_OnSuccess(job.Scheduler, job.Name))
}

func AddJobConfig(job utils.CRONConfig) {
	utils.Log("Adding CRON job to config...")

	config := utils.ReadConfigFromFile()
	config.CRON[job.Name] = job
	utils.SetBaseMainConfig(config)
	
	utils.TriggerEvent(
		"cosmos.settings",
		"Settings updated",
		"success",
		"",
		map[string]interface{}{
			"from": "Added CRON job",
	})

	InitJobs()
	InitScheduler()
}

func RemoveJobConfig(name string) {
	utils.Log("Removing CRON job from config...")

	config := utils.ReadConfigFromFile()
	delete(config.CRON, name)
	utils.SetBaseMainConfig(config)
	
	utils.TriggerEvent(
		"cosmos.settings",
		"Settings updated",
		"success",
		"",
		map[string]interface{}{
			"from": "Removed CRON job",
	})

	InitJobs()
	InitScheduler()
}
