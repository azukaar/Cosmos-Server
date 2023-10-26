package metrics 

import (
	"net/http"
	"encoding/json"
	
	"github.com/azukaar/cosmos-server/src/utils"
)

func API_GetMetrics(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if(req.Method == "GET") {
		c, errCo := utils.GetCollection(utils.GetRootAppId(), "metrics")
		if errCo != nil {
				utils.Error("Metrics - Database Connect", errCo)
				return
		}

		// get all metrics from database
		var metrics []DataDefDB
		cursor, err := c.Find(nil, map[string]interface{}{})
		if err != nil {
			utils.Error("Metrics: Error fetching metrics", err)
			return
		}
		defer cursor.Close(nil)
		
		if err = cursor.All(nil, &metrics); err != nil {
			utils.Error("Metrics: Error decoding metrics", err)
			return
		}
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data": metrics,
		})
	} else {
		utils.Error("SettingGet: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}