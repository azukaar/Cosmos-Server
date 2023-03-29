package utils

import (
	"context"
	"os"
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)


var client *mongo.Client

func DB() error {
	if(GetBaseMainConfig().DisableUserManagement)	{
		return errors.New("User Management is disabled")
	}

	uri := MainConfig.MongoDB + "/?retryWrites=true&w=majority"

	if(client != nil && client.Ping(context.TODO(), readpref.Primary()) == nil) {
		return nil
	}

	Log("(Re) Connecting to the database...")

	var err error

	client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}
	defer func() {
	}()

	// Ping the primary
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		return err
	}

	Log("Successfully connected to the database.")
	return nil
}

func Disconnect() {
	if err := client.Disconnect(context.TODO()); err != nil {
		Fatal("DB", err)
	}
}

func GetCollection(applicationId string, collection string) (*mongo.Collection, error) {
	if client == nil {
		errCo := DB()
		if errCo != nil {
			return nil, errCo
		}
	}
	
	name := os.Getenv("MONGODB_NAME"); if name == "" {
		name = "COSMOS"
	}
	
	Debug("Getting collection " + applicationId + "_" + collection + " from database " + name)
	
	c := client.Database(name).Collection(applicationId + "_" + collection)
	
	return c, nil
}

// func query(q string) (*sql.Rows, error) {
// 	return db.Query(q)
// }