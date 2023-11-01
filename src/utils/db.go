package utils

import (
	"context"
	"os"
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern" 
)


var client *mongo.Client

func DB() error {
	if(GetMainConfig().DisableUserManagement)	{
		return errors.New("User Management is disabled")
	}

	mongoURL := GetMainConfig().MongoDB
	
	if(client != nil && client.Ping(context.TODO(), readpref.Primary()) == nil) {
		return nil
	}

	Log("(Re) Connecting to the database...")

	if mongoURL == "" {
		return errors.New("MongoDB URL is not set, cannot connect to the database.")
	}

	var err error

	opts := options.Client().ApplyURI(mongoURL).SetRetryWrites(true).SetWriteConcern(writeconcern.New(writeconcern.WMajority()))
	
	client, err = mongo.Connect(context.TODO(), opts)

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

func DisconnectDB() {
	if err := client.Disconnect(context.TODO()); err != nil {
		Fatal("DB", err)
	}
	client = nil
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