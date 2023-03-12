package utils

import (
	"context"
	"os"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)


var client *mongo.Client

func DB() {
	Log("Connecting to the database...")

	uri := MainConfig.MongoDB + "/?retryWrites=true&w=majority"
	
	var err error

	client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		Fatal("DB", err)
	}
	defer func() {
	}()

	// Ping the primary
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		Fatal("DB", err)
	}

	Log("Successfully connected to the database.")
}

func Disconnect() {
	if err := client.Disconnect(context.TODO()); err != nil {
		Fatal("DB", err)
	}
}

func GetCollection(applicationId string, collection string) *mongo.Collection {
	if client == nil {
		DB()
	}
	
	name := os.Getenv("MONGODB_NAME"); if name == "" {
		name = "COSMOS"
	}
	
	Debug("Getting collection " + applicationId + "_" + collection + " from database " + name)
	
	c := client.Database(name).Collection(applicationId + "_" + collection)
	
	return c
}

// func query(q string) (*sql.Rows, error) {
// 	return db.Query(q)
// }