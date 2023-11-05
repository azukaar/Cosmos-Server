package metrics 

import (
	"strings"
	"regexp"
	"fmt"
	"time"

	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/azukaar/cosmos-server/src/docker"
)

func CheckAlerts(TrackingMetric string, Period string, metric utils.AlertMetricTrack, Value int) {
	config := utils.GetMainConfig()
	ActiveAlerts := config.MonitoringAlerts
	
	alerts := []utils.Alert{}
	ok := false

	// if tracking metric contains a wildcard
	if strings.Contains(TrackingMetric, "*") {
		regexPattern := "^" + strings.ReplaceAll(TrackingMetric, "*", ".*?")
		regex, _ := regexp.Compile(regexPattern)

		// Iterate over the map to find a match
		for _, val := range ActiveAlerts {
			if regex.MatchString(val.TrackingMetric) && val.Period == Period {
				alerts = append(alerts, val)
				ok = true
			}
		}
	} else {
		for _, val := range ActiveAlerts {
			if val.TrackingMetric == TrackingMetric  && val.Period == Period {
				alerts = append(alerts, val)
				ok = true
				break 
			}
		}
	}

	if !ok {
		return
	}

	for _, alert := range alerts {
		if !alert.Enabled {
			continue
		}

		if alert.Throttled && alert.LastTriggered.Add(time.Hour * 24).After(time.Now()) {
			continue
		}
		
		ValueToTest := Value 

		if alert.Condition.Percent {
			ValueToTest = int(float64(Value) / float64(metric.Max) * 100)

			utils.Debug(fmt.Sprintf("Alert %s: %d / %d = %d%%", alert.Name, Value, metric.Max, ValueToTest))
		}

		// Check if the condition is met
		if alert.Condition.Operator == "gt" {
			if ValueToTest > alert.Condition.Value {
				ExecuteAllActions(alert, alert.Actions, metric)
			}
		} else if alert.Condition.Operator == "lt" {
			if ValueToTest < alert.Condition.Value {
				ExecuteAllActions(alert, alert.Actions, metric)
			}
		} else if alert.Condition.Operator == "eq" {
			if ValueToTest == alert.Condition.Value {
				ExecuteAllActions(alert, alert.Actions, metric)
			}
		}
	}
}

func ExecuteAllActions(alert utils.Alert, actions []utils.AlertAction, metric utils.AlertMetricTrack) {
	utils.Debug("Alert triggered: " + alert.Name)
	for _, action := range actions {
		ExecuteAction(alert, action, metric)
	}

	// set LastTriggered to now
	alert.LastTriggered = time.Now()

	// update alert in config
	config := utils.GetMainConfig()
	for i, val := range config.MonitoringAlerts {
		if val.Name == alert.Name {
			config.MonitoringAlerts[i] = alert
			break
		}
	}

	utils.SetBaseMainConfig(config)
}

func ExecuteAction(alert utils.Alert, action utils.AlertAction, metric utils.AlertMetricTrack) {
	utils.Log("Executing action " + action.Type + " on " + metric.Key + " " + metric.Object	)

	if action.Type == "email" {
		utils.Debug("Sending email to " + action.Target)

		if utils.GetMainConfig().EmailConfig.Enabled {
			users := utils.ListAllUsers("admin")
			for _, user := range users {
				if user.Email != "" {
					utils.SendEmail([]string{user.Email}, "Alert Triggered: " + alert.Name,
					fmt.Sprintf(`<h1>Alert Triggered [%s]</h1>
You are recevining this email because you are admin on a Cosmos
server where an Alert has been subscribed to.<br />
You can manage your subscriptions in the Monitoring tab.<br />
Alert triggered on %s. Please refer to the Monitoring tab for
more information.<br />`, alert.Severity, metric.Key))
				}
			}
		} else {
			utils.Warn("Alert triggered but Email is not enabled")
		}

	} else if action.Type == "webhook" {
		utils.Debug("Calling webhook " + action.Target)

	} else if action.Type == "stop" {
		utils.Debug("Stopping application")

		parts := strings.Split(metric.Object, "@")

		if len(parts) > 1 {
			object := parts[0]
			objectName := strings.Join(parts[1:], "@")

			if object == "container" {
				docker.StopContainer(objectName)
			} else if object == "route" {
				config := utils.ReadConfigFromFile()

				objectIndex := -1
				for i, route := range config.HTTPConfig.ProxyConfig.Routes {
					if route.Name == objectName {
						objectIndex = i
						break
					}
				}

				if objectIndex != -1 {
					config.HTTPConfig.ProxyConfig.Routes[objectIndex].Disabled = true		

					utils.SetBaseMainConfig(config)
				} else {
					utils.Warn("No route found, for " + objectName)
				}

				utils.RestartHTTPServer()
			}
		} else {
			utils.Warn("No object found, for " + metric.Object)
		}

	} else if action.Type == "notification" {
		utils.WriteNotification(utils.Notification{
			Recipient: "admin",
			Title: "Alert triggered",
			Message: "The alert \"" + alert.Name + "\" was triggered.",
			Level: alert.Severity,
			Link: "/cosmos-ui/monitoring",
		})

	} else if action.Type == "script" {
		utils.Debug("Executing script")
	}
}