package utils

import (
	"context"
	"log"
	"os"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)


var client *mongo.Client

func DB() {
	log.Println("Connecting to the database...")

	uri := os.Getenv("MONGODB") + "/?retryWrites=true&w=majority"
	
	var err error

	client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
	}()

	// Ping the primary
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}

	log.Println("Successfully connected to the database.")
}

func Disconnect() {
	if err := client.Disconnect(context.TODO()); err != nil {
		log.Fatal(err)
	}
}

func GetCollection(applicationId string, collection string) *mongo.Collection {
	name := os.Getenv("MONGODB_NAME"); if name == "" {
		name = "GUCO"
	}
	log.Println("Getting collection " + applicationId + "_" + collection + " from database " + name)
	c := client.Database(name).Collection(applicationId + "_" + collection)
	return c
}

// func query(q string) (*sql.Rows, error) {
// 	return db.Query(q)
// }