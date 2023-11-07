package utils

import (
	"time"
	"strconv"
	"context"

	"go.mongodb.org/mongo-driver/bson"
)

type CleanupObject struct {
	Date time.Time
}

func CleanupByDate(collectionName string) {
	c, errCo := GetCollection(GetRootAppId(), collectionName)
	if errCo != nil {
		MajorError("Database Cleanup", errCo)
			return
	}
	
	del, err := c.DeleteMany(context.Background(), bson.M{"Date": bson.M{"$lt": time.Now().AddDate(0, -1, 0)}})

	if err != nil {
		MajorError("Database Cleanup", err)
		return
	}

	Log("Cleanup: " + collectionName + " " + strconv.Itoa(int(del.DeletedCount)) + " objects deleted")
	
	TriggerEvent(
		"cosmos.database.cleanup",
		"Database Cleanup of " + collectionName,
		"success",
		"",
		map[string]interface{}{
			"collection": collectionName,
			"deleted": del.DeletedCount,
	})
}
