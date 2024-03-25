package cron

import (
	"strconv"
	"time"
	"context"
	"errors"
	"io"
	"bufio"
	"os/exec"

	"github.com/go-co-op/gocron/v2"

	"github.com/azukaar/cosmos-server/src/utils"
)

var scheduler gocron.Scheduler
var CRONLock = make(chan bool, 1)
var RunningLock = make(chan bool, 1)

type ConfigJob struct {
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
}

var jobsList = map[string]map[string]ConfigJob{}
var wasInit = false

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

func JobFromCommand(command string, args ...string) func(OnLog func(string), OnFail func(error), OnSuccess func(), ctx context.Context, cancel context.CancelFunc) {
	return func(OnLog func(string), OnFail func(error), OnSuccess func(), ctx context.Context, cancel context.CancelFunc) {
		// Create a command that respects the provided context
		cmd := exec.CommandContext(ctx, command, args...)

		// Getting the pipe for standard output
		stdout, err := cmd.StdoutPipe()
		if err != nil {
				OnFail(err)
				return
		}

		// Getting the pipe for standard error
		stderr, err := cmd.StderrPipe()
		if err != nil {
				OnFail(err)
				return
		}

		utils.Debug("Running command: " + cmd.String())
		
		// Start the command
		if err := cmd.Start(); err != nil {
				OnFail(err)
				return
		}

		// Concurrently read from stdout and stderr
		go streamLogs(stdout, OnLog)
		go streamLogs(stderr, OnLog)

		// Wait for the command to finish
		err = cmd.Wait()
		if err != nil {
				OnFail(err)
				return
		}

		OnSuccess()
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
		j := ConfigJob{
			Scheduler: "Custom",
			Name: job.Name,
			Job: JobFromCommand("sh", "-c", job.Command),
			Crontab: job.Crontab,
			Running: false,
			Cancellable: true,
			LastRun: time.Time{},
			LastStarted: time.Time{},
			Logs: []string{},
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
	return func (OnLog func(string), OnFail func(error), OnSuccess func()) {
		CRONLock <- true
		if job, ok := jobsList[schedulerName][jobName]; ok {
			utils.Log("Starting CRON job: " + job.Name)

			if job.Running {
				utils.Error("Scheduler: job " + job.Name + " is already running", nil)
				<-CRONLock
				return
			}

			ctx, cancel := context.WithCancel(context.Background())
			job.Ctx = ctx
			job.LastStarted = time.Now()
			job.LastRunSuccess = false
			job.Running = true
			job.Logs = []string{}
			job.CancelFunc = cancel
			jobsList[job.Scheduler][job.Name] = job
			<-CRONLock

			triggerJobUpdated("start", job.Name)

			select {
				case <-ctx.Done():
					OnFail(errors.New("Scheduler: job was canceled."))
					return
				default:
					job.Job(OnLog, OnFail, OnSuccess, ctx, cancel)
			}
		} else {
			utils.Error("Scheduler: job " + jobName + " not found", nil)
			<-CRONLock
		}
	}

	return nil
}
func jobRunner_OnLog(schedulerName, jobName string) func(log string) {
	return func(log string) {
		CRONLock <- true
		if job, ok := jobsList[schedulerName][jobName]; ok {
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
		}
		<-CRONLock
	}
}

func InitScheduler() {
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
		jobRunner(job.Scheduler, job.Name)(jobRunner_OnLog(job.Scheduler, job.Name), jobRunner_OnFail(job.Scheduler, job.Name), jobRunner_OnSuccess(job.Scheduler, job.Name))
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
