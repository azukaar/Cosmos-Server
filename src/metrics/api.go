package metrics 

import (
	"net/http"
	"encoding/json"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	
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

	c, errCo = utils.GetCollection(utils.GetRootAppId(), "events")
	if errCo != nil {
		utils.Error("MetricsReset: Database error" , errCo)
		utils.HTTPError(w, "Database error ", http.StatusMethodNotAllowed, "HTTP001")
		return
	}

	// delete all metrics from database
	_, err = c.DeleteMany(nil, map[string]interface{}{})
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

type MetricList struct {
	Key string
	Label string
}

func ListMetrics(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}
	
	if(req.Method == "GET") {
		c, errCo := utils.GetCollection(utils.GetRootAppId(), "metrics")
		if errCo != nil {
				utils.Error("Database Connect", errCo)
				utils.HTTPError(w, "Database", http.StatusInternalServerError, "DB001")
				return
		}

		metrics := []MetricList{}

		cursor, err := c.Find(nil, map[string]interface{}{}, options.Find().SetProjection(bson.M{"Key": 1, "Label":1, "_id": 0}))

		if err != nil {
			utils.Error("metrics: Error while getting metrics", err)
			utils.HTTPError(w, "metrics Get Error", http.StatusInternalServerError, "UD001")
			return
		}

		defer cursor.Close(nil)

		if err = cursor.All(nil, &metrics); err != nil {
			utils.Error("metrics: Error while decoding metrics", err)
			utils.HTTPError(w, "metrics decode Error", http.StatusInternalServerError, "UD002")
			return
		}

		// Extract the names into a string slice
		metricNames := map[string]string{}

		for _, metric := range metrics {
				metricNames[metric.Key] = metric.Label
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "OK",
				"data":   metricNames,
		})
	} else {
		utils.Error("metrics: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}