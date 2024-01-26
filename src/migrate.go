package main 

import (
	"os"
	"strings"
	"context"
	// "fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/azukaar/cosmos-server/src/utils"
)

func MigratePre014Coll(collection string, from, to *mongo.Client) {
	name := os.Getenv("MONGODB_NAME"); if name == "" {
		name = "COSMOS"
	}

	utils.Log("Migrating collection " + collection + " from database " + name)

	applicationId := utils.GetRootAppId()
	
	// Debug("Getting collection " + applicationId + "_" + collection + " from database " + name)
	
	cf := from.Database(name).Collection(applicationId + "_" + collection)
	ct := to.Database(name).Collection(applicationId + "_" + collection)

	// get all documents
	cur, err := cf.Find(context.Background(), bson.D{})
	if err != nil {
		utils.Error("Error getting documents from " + collection + " collection", err)
		return
	}

	// iterate over documents
	for cur.Next(context.Background()) {
		// create a value into which the single document can be decoded
		var elem bson.D
		err := cur.Decode(&elem)
		if err != nil {
			utils.Error("Error decoding document from " + collection + " collection", err)
			// return
		}


		// insert into new collection
		_, err = ct.InsertOne(context.Background(), elem)
		if err != nil {
			// fmt.Println("moving document", elem)
			utils.Error("Error inserting document into " + collection + " collection", err)
			// return/
		}
	}

	if err := cur.Err(); err != nil {
		utils.Error("Error iterating over documents from " + collection + " collection", err)
		return
	}

	// Close the cursor once finished
	cur.Close(context.Background())
}

func MigratePre014() {
	config := utils.GetMainConfig()

	// check if COSMOS.db does NOT exist
	if _, err := os.Stat(utils.CONFIGFOLDER + "COSMOS.sqlite"); err != nil && config.MongoDB != "" {
		// connect to MongoDB
		utils.Log("Connecting to MongoDB...")
		
		
		mongoURL := utils.GetMainConfig().MongoDB

		var err error

		opts := options.Client().ApplyURI(mongoURL).SetRetryWrites(true).SetWriteConcern(writeconcern.New(writeconcern.WMajority()))
		
		if os.Getenv("HOSTNAME") == "" || utils.IsHostNetwork {
			hostname := opts.Hosts[0]
			// split port
			hostnameParts := strings.Split(hostname, ":")
			hostname = hostnameParts[0]
			port := "27017" 

			if len(hostnameParts) > 1 {
				port = hostnameParts[1]
			}

			utils.Log("Getting Mongo DB IP from name : " + hostname + " (port " + port + ")")

			ip, _ := utils.GetContainerIPByName(hostname)
			if ip != "" {
				// IsDBaContainer = true
				opts.SetHosts([]string{ip + ":" + port})
				utils.Log("Mongo DB IP : " + ip)
			}
		}
	
		client, err := mongo.Connect(context.TODO(), opts)

		if err != nil {
			panic(err)
		}

		// Ping the primary
		if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
			panic(err)
		}

		utils.Log("Successfully connected to the database.")
		
		clientTo, err := utils.GetDBClient()
		
		MigratePre014Coll("users", client, clientTo)
		MigratePre014Coll("devices", client, clientTo)
		MigratePre014Coll("events", client, clientTo)
		MigratePre014Coll("notifications", client, clientTo)
		MigratePre014Coll("metrics", client, clientTo)
	}
}