package utils 

import (
	"time"
	"encoding/json"
)

func TriggerEvent(eventId string, label string, level string, object string, data map[string]interface{}) {
	Debug("Triggering event " + eventId)


	// Marshal the data map into a JSON string
	dataAsBytes, err := json.Marshal(data)
	if err != nil {
		Error("Error marshaling data: %v\n", err)
		return
	}
	dataAsString := string(dataAsBytes)

	BufferedDBWrite("events", map[string]interface{}{
		"eventId": eventId,
		"label": label,
		"application": "Cosmos",
		"level": level,
		"date": time.Now(),
		"data": data,
		"object": object,
		"_search": eventId + " " + dataAsString,
	})
}

