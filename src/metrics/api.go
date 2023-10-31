package metrics 

import (
	"net/http"
	"encoding/json"
	"strings"
	
	"github.com/azukaar/cosmos-server/src/utils"
)

func API_GetMetrics(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	//get query string "metrics"
	query := req.URL.Query()
	metrics := query.Get("metrics")

	// split by comma
	metricsList := []string{}
	if metrics != "" {
		metricsList = strings.Split(metrics, ",")
	}


	if(req.Method == "GET") {
		w.Header().Set("Content-Type", "application/json")
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data": AggloMetrics(metricsList),
		})
	} else {
		utils.Error("MetricsGet: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

func API_ResetMetrics(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	c, errCo := utils.GetCollection(utils.GetRootAppId(), "metrics")
	if errCo != nil {
		utils.Error("MetricsReset: Database error" , errCo)
		utils.HTTPError(w, "Database error ", http.StatusMethodNotAllowed, "HTTP001")
		return
	}

	// delete all metrics from database
	_, err := c.DeleteMany(nil, map[string]interface{}{})
	if err != nil {
		utils.Error("MetricsReset: Database error ", err)
		utils.HTTPError(w, "Database error ", http.StatusMethodNotAllowed, "HTTP001")
		return
	}

	if(req.Method == "GET") {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
		})
	} else {
		utils.Error("SettingGet: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}