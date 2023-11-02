package metrics

import (
	"time"

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