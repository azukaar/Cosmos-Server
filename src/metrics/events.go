package metrics 

import (
	"net/http"
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/azukaar/cosmos-server/src/utils"
)

type Event struct {
	Id primitive.ObjectID `json:"id" bson:"_id"`
	Label string `json:"label" bson:"label"`
	Application string `json:"application" bson:"application"`
	EventId string `json:"eventId" bson:"eventId"`
	Date time.Time `json:"date" bson:"date"`
	Level string `json:"level" bson:"level"`
	Data map[string]interface{} `json:"data" bson:"data"`
	Object string `json:"object" bson:"object"`
}

func API_ListEvents(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}
	
	if(req.Method == "GET") {

		query := req.URL.Query()
		from, errF := time.Parse("2006-01-02T15:04:05Z", query.Get("from"))
		if errF != nil {
			utils.Error("events: Error while parsing from date", errF)
		}
		to, errF := time.Parse("2006-01-02T15:04:05Z", query.Get("to"))
		if errF != nil {
			utils.Error("events: Error while parsing from date", errF)
		}

		logLevel := query.Get("logLevel")
		if logLevel == "" {
			logLevel = "info"
		}
		search := query.Get("search")
		dbQuery := query.Get("query")

		page := query.Get("page")
		var pageId primitive.ObjectID
		if page != "" {
			pageId, _ = primitive.ObjectIDFromHex(page)
		}

		// decode to bson
		dbQueryBson := bson.M{}
		if dbQuery != "" {
			err := bson.UnmarshalExtJSON([]byte(dbQuery), true, &dbQueryBson)
			if err != nil {
				utils.Error("events: Error while parsing query " + dbQuery, err)
				utils.HTTPError(w, "events Get Error", http.StatusInternalServerError, "UD001")
				return
			}
		} else if search != "" {
			dbQueryBson["$text"] = bson.M{
				"$search": search,
			}
		}

	
		// merge date query into dbQueryBson
		if dbQueryBson["date"] == nil {
			dbQueryBson["date"] = bson.M{}
		}
		dbQueryBson["date"].(bson.M)["$gte"] = from
		dbQueryBson["date"].(bson.M)["$lte"] = to

		if logLevel != "" {
			if dbQueryBson["level"] == nil {
				dbQueryBson["level"] = bson.M{}
			}
			levels := []string{"error"}
			if logLevel == "debug" {
				levels = []string{"debug", "info", "warning", "error", "important", "success"}
			} else if logLevel == "info" {
				levels = []string{"info", "warning", "error", "important", "success"}
			} else if logLevel == "success" {
				levels = []string{"warning", "error", "important", "success"}
			} else if logLevel == "warning" {
				levels = []string{"warning", "error", "important"}
			} else if logLevel == "important" {
				levels = []string{"important", "error"}
			}
			dbQueryBson["level"].(bson.M)["$in"] = levels
		}

		if pageId != primitive.NilObjectID {
			dbQueryBson["_id"] = bson.M{
				"$lt": pageId,
			}
		}
		
		c, errCo := utils.GetCollection(utils.GetRootAppId(), "events")
		if errCo != nil {
				utils.Error("Database Connect", errCo)
				utils.HTTPError(w, "Database", http.StatusInternalServerError, "DB001")
				return
		}

		events := []Event{}

		limit := int64(50)
		opts := options.Find().SetLimit(limit).SetSort(bson.D{{"date", -1}})
		// .SetProjection(bson.D{{"_id", 1}, {"eventId", 1}, {"date", 1}, {"level", 1}, {"data", 1}})

		cursor, err := c.Find(nil, dbQueryBson, opts)

		if err != nil {
			utils.Error("events: Error while getting events", err)
			utils.HTTPError(w, "events Get Error", http.StatusInternalServerError, "UD001")
			return
		}

		defer cursor.Close(nil)

		if err = cursor.All(nil, &events); err != nil {
			utils.Error("events: Error while decoding events", err)
			utils.HTTPError(w, "events decode Error", http.StatusInternalServerError, "UD002")
			return
		}

		totalCount, err := c.CountDocuments(nil, dbQueryBson)
		if err != nil {
			utils.Error("events: Error while counting events", err)
			utils.HTTPError(w, "events count Error", http.StatusInternalServerError, "UD003")
			return
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "OK",
				"total": totalCount,
				"data":   events,
		})
	} else {
		utils.Error("events: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}