package metrics

import (
	"time"
	"net/http"
	"encoding/json"

	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/azukaar/cosmos-server/src/utils"
)


func PushRequestMetrics(route utils.ProxyRouteConfig, statusCode int, TimeStarted time.Time, size int64) error {
	responseTime := time.Since(TimeStarted)

	if !utils.GetMainConfig().MonitoringDisabled {
		if statusCode >= 400 {
			PushSetMetric("proxy.all.error", 1, DataDef{
				Max: 0,
				Period: time.Second * 30,
				Label: "Global Request Errors",
				AggloType: "sum",
				SetOperation: "sum",
			})
			PushSetMetric("proxy.route.error."+route.Name, 1, DataDef{
				Max: 0,
				Period: time.Second * 30,
				Label: "Request Errors " + route.Name,
				AggloType: "sum",
				SetOperation: "sum",
				Object: "route@" + route.Name,
			})
		} else {
			PushSetMetric("proxy.all.success", 1, DataDef{
				Max: 0,
				Period: time.Second * 30,
				Label: "Global Request Success",
				AggloType: "sum",
				SetOperation: "sum",
			})
			PushSetMetric("proxy.route.success."+route.Name, 1, DataDef{
				Max: 0,
				Period: time.Second * 30,
				Label: "Request Success " + route.Name,
				AggloType: "sum",
				SetOperation: "sum",
				Object: "route@" + route.Name,
			})
		}

		PushSetMetric("proxy.all.time", int(responseTime.Milliseconds()), DataDef{
			Max: 0,
			Period: time.Second * 30,
			Label: "Global Request Time",
			AggloType: "sum",
			SetOperation: "sum",
			Unit: "ms",
		})

		PushSetMetric("proxy.route.time."+route.Name, int(responseTime.Milliseconds()), DataDef{
			Max: 0,
			Period: time.Second * 30,
			Label: "Response Request " + route.Name,
			AggloType: "sum",
			SetOperation: "sum",
			Unit: "ms",
			Object: "route@" + route.Name,
		})

		PushSetMetric("proxy.all.bytes", int(size), DataDef{
			Max: 0,
			Period: time.Second * 30,
			Label: "Global Transfered Bytes",
			AggloType: "sum",
			SetOperation: "sum",
			Unit: "B",
		})

		PushSetMetric("proxy.route.bytes."+route.Name, int(size), DataDef{
			Max: 0,
			Period: time.Second * 30,
			Label: "Transfered Bytes " + route.Name,
			AggloType: "sum",
			SetOperation: "sum",
			Unit: "B",
			Object: "route@" + route.Name,
		})
	}

	return nil
}

func PushShieldMetrics(reason string) {
	reasonStr := map[string]string{
		"bots": "Bots",
		"geo": "By Geolocation",
		"referer": "By Referer",
		"hostname": "By Hostname",
		"ip-whitelists": "By IP Whitelists",
		"smart-shield": "Smart Shield",
	}

	PushSetMetric("proxy.blocked."+reason, 1, DataDef{
		Max: 0,
		Period: time.Second * 30,
		Label: "Blocked " + reasonStr[reason],
		AggloType: "sum",
		SetOperation: "sum",
	})
	
	PushSetMetric("proxy.all.blocked", 1, DataDef{
		Max: 0,
		Period: time.Second * 30,
		Label: "Global Blocked Requests",
		AggloType: "sum",
		SetOperation: "sum",
	})
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