package utils 

import (
	"net/http"
	"encoding/json"
	"time"
	"fmt"
	"strings"
	
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/bson"
)

type NotificationActions struct {
	Text string
	Link string
}

type Notification struct {
	ID primitive.ObjectID      `bson:"_id,omitempty"`
	Title string
	Message string
	Vars string
	Icon string
	Link string
	Date time.Time
	Level string
	Read bool
	Recipient string
	Actions []NotificationActions
}

func NotifGet(w http.ResponseWriter, req *http.Request) {
	_from := req.URL.Query().Get("from")
	from, _ := primitive.ObjectIDFromHex(_from)
	
	if LoggedInOnly(w, req) != nil {
		return
	}
	
	nickname := req.Header.Get("x-cosmos-user")

	if(req.Method == "GET") {
		c, errCo := GetCollection(GetRootAppId(), "notifications")
		if errCo != nil {
				Error("Database Connect", errCo)
				HTTPError(w, "Database", http.StatusInternalServerError, "DB001")
				return
		}

		Debug("Notifications: Get notif for " + nickname)

		notifications := []Notification{}

		reqdb := map[string]interface{}{
			"Recipient": nickname, 
		}

		if from != primitive.NilObjectID {
			reqdb = map[string]interface{}{
				// nickname or role 
				"Recipient": nickname, 
				// get notif before from
				"_id": map[string]interface{}{
					"$lt": from,
				},
			}
		}

		limit := int64(20)

		cursor, err := c.Find(nil, reqdb, &options.FindOptions{
			Sort: map[string]interface{}{
				"Date": -1,
			},
			Limit: &limit,
		})
		defer cursor.Close(nil)

		if err != nil {
			Error("Notifications: Error while getting notifications", err)
			HTTPError(w, "notifications Get Error", http.StatusInternalServerError, "UD001")
			return
		}

		
		if err = cursor.All(nil, &notifications); err != nil {
			Error("Notifications: Error while decoding notifications", err)
			HTTPError(w, "notifications Get Error", http.StatusInternalServerError, "UD002")
			return
		}
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data": notifications,
		})
	} else {
		Error("Notifications: Method not allowed" + req.Method, nil)
		HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

func MarkAsRead(w http.ResponseWriter, req *http.Request) {
	if(req.Method == "GET") {
		if LoggedInOnly(w, req) != nil {
				return
		}

		notificationIDs := []primitive.ObjectID{}
		nickname := req.Header.Get("x-cosmos-user")

		notificationIDsRawRunes := req.URL.Query().Get("ids")

		notificationIDsRaw := strings.Split(notificationIDsRawRunes, ",")

		Debug(fmt.Sprintf("Marking %v notifications as read",notificationIDsRaw))

		for _, notificationIDRaw := range notificationIDsRaw {
			notificationID, err := primitive.ObjectIDFromHex(notificationIDRaw)

			if err != nil {
					HTTPError(w, "Invalid notification ID " + notificationIDRaw, http.StatusBadRequest, "InvalidID")
					return
			}

			notificationIDs = append(notificationIDs, notificationID)
		}


		c, errCo := GetCollection(GetRootAppId(), "notifications")
		if errCo != nil {
				Error("Database Connect", errCo)
				HTTPError(w, "Database connection error", http.StatusInternalServerError, "DB001")
				return
		}

		filter := bson.M{"_id": bson.M{"$in": notificationIDs}, "Recipient": nickname}
		update := bson.M{"$set": bson.M{"Read": true}}
		result, err := c.UpdateMany(nil, filter, update)
		if err != nil {
				Error("Notifications: Error while marking notification as read", err)
				HTTPError(w, "Error updating notification", http.StatusInternalServerError, "UpdateError")
				return
		}

		if result.MatchedCount == 0 {
				HTTPError(w, "No matching notification found", http.StatusNotFound, "NotFound")
				return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "OK",
				"message": "Notification marked as read",
		})
	} else {
		Error("Notifications: Method not allowed" + req.Method, nil)
		HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}


func WriteNotification(notification Notification) {
	notification.Date = time.Now()

	notification.Read = false

	if notification.Recipient == "all" || notification.Recipient == "admin" || notification.Recipient == "user" {
		// list all users
		users := ListAllUsers(notification.Recipient)

		Debug("Notifications: Sending notification to " + string(len(users)) + " users")

		for _, user := range users {
			BufferedDBWrite("notifications", map[string]interface{}{
				"Title": notification.Title,
				"Message": notification.Message,
				"Vars": notification.Vars,
				"Icon": notification.Icon,
				"Link": notification.Link,
				"Date": notification.Date,
				"Level": notification.Level,
				"Read": notification.Read,
				"Recipient": user.Nickname,
				"Actions": notification.Actions,
			})
		}
	} else {
		BufferedDBWrite("notifications", map[string]interface{}{
			"Title": notification.Title,
			"Message": notification.Message,
			"Vars": notification.Vars,
			"Icon": notification.Icon,
			"Link": notification.Link,
			"Date": notification.Date,
			"Level": notification.Level,
			"Read": notification.Read,
			"Recipient": notification.Recipient,
			"Actions": notification.Actions,
		})
	}
}