package cron

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/azukaar/cosmos-server/src/utils"
)

func ListJobs(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	jobs := getJobsList()
	newJobs := map[string]map[string]ConfigJob{}

	// remove logs 
	for schedulerName, sched := range jobs {
		if newJobs[schedulerName] == nil {
			newJobs[schedulerName] = map[string]ConfigJob{}
		}

		for _, job := range sched {

			newJobs[job.Scheduler][job.Name] = ConfigJob{
				Disabled: job.Disabled,
				Scheduler: job.Scheduler,
				Cancellable: job.Cancellable,
				Name: job.Name,
				Crontab: job.Crontab,
				Running: job.Running,
				LastStarted: job.LastStarted,
				LastRun: job.LastRun,
				LastRunSuccess: job.LastRunSuccess,
			}
		}
	}

	if req.Method == "GET" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data":   newJobs,
		})
		return
	} else {
		utils.Error("Listjobs: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

type JobRequestJSON struct {
	Scheduler string `validate:"required,min=3,max=32,alphanum"`
	Name string `validate:"required,min=3,max=32,alphanum"`
}


func StopJobRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "POST" {
		var request JobRequestJSON
		err1 := json.NewDecoder(req.Body).Decode(&request)
		if err1 != nil {
			utils.Error("StopJobRoute: Invalid Request", err1)
			utils.HTTPError(w, "StopJobRoute Error", 
				http.StatusInternalServerError, "UC001")
			return 
		}
		scheduler := request.Scheduler
		job := request.Name

		err := CancelJob(scheduler, job)
		if err != nil {
			utils.Error("StopJob: " + err.Error(), nil)
			utils.HTTPError(w, err.Error(), http.StatusInternalServerError, "HTTP002")
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
		})
		return
	} else {
		utils.Error("Listjobs: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

func DeleteJobRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "POST" {
		var request JobRequestJSON
		err1 := json.NewDecoder(req.Body).Decode(&request)
		if err1 != nil {
			utils.Error("DeleteJobRoute: Invalid Request", err1)
			utils.HTTPError(w, "DeleteJobRoute Error", 
				http.StatusInternalServerError, "UC001")
			return 
		}
		job := request.Name
		
		config := utils.ReadConfigFromFile()
		
		// remove config.CRON[job]
		delete(config.CRON, job)	

		utils.SetBaseMainConfig(config)

		go (func() {
			InitJobs()
			InitScheduler()
		})()

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
		})
		return
	} else {
		utils.Error("Listjobs: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

func RunJobRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "POST" {
		var request JobRequestJSON
		err1 := json.NewDecoder(req.Body).Decode(&request)
		if err1 != nil {
			utils.Error("RunJobRoute: Invalid Request", err1)
			utils.HTTPError(w, "RunJobRoute Error", 
				http.StatusInternalServerError, "UC001")
			return 
		}
		scheduler := request.Scheduler
		job := request.Name

		err := ManualRunJob(scheduler, job)

		time.Sleep(1 * time.Second)

		if err != nil {
			utils.Error("RunJob: " + err.Error(), nil)
			utils.HTTPError(w, err.Error(), http.StatusInternalServerError, "HTTP002")
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
		})
		return
	} else {
		utils.Error("Listjobs: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

func GetJobRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "POST" {
		var request JobRequestJSON
		err1 := json.NewDecoder(req.Body).Decode(&request)
		if err1 != nil {
			utils.Error("GetJobRoute: Invalid Request", err1)
			utils.HTTPError(w, "GetJobRoute Error", 
				http.StatusInternalServerError, "UC001")
			return 
		}

		scheduler := request.Scheduler
		job := request.Name
		
		jobs := getJobsList()
		
		if _, ok := jobs[scheduler]; !ok {
			utils.Error("GetJob: Scheduler not found", nil)
			utils.HTTPError(w, "Scheduler not found", http.StatusNotFound, "HTTP002")
			return
		}

		if _, ok := jobs[scheduler][job]; !ok {
			utils.Error("GetJob: Job not found", nil)
			utils.HTTPError(w, "Job not found", http.StatusNotFound, "HTTP002")
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data":   jobs[scheduler][job],
		})

		return
	} else {
		utils.Error("Listjobs: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

func GetRunningJobsRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "GET" {
		jobs := RunningJobs()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data":   jobs,
		})
		return
	} else {
		utils.Error("Listjobs: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}