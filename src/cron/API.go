package cron

import (
	"encoding/json"
	"net/http"

	"github.com/azukaar/cosmos-server/src/utils"
)

func ListJobs(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "GET" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data":   getJobsList(),
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