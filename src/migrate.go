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
func MigratePre014Coll(collection string, from *mongo.Client) {
	name := os.Getenv("MONGODB_NAME")
	if name == "" {
			name = "COSMOS"
	}

	utils.Log("Migrating collection " + collection + " from database " + name)

	applicationId := utils.GetRootAppId()

	cf := from.Database(name).Collection(applicationId + "_" + collection)
	ct, closeDb, err := utils.GetEmbeddedCollection(utils.GetRootAppId(), collection)
	if err != nil {
			utils.Error("Error getting collection " + applicationId + "_" + collection + " from database " + name, err)
			return
	}
	defer closeDb()

	// get all documents
	opts := options.Find()
	cur, err := cf.Find(context.Background(), bson.D{}, opts)
	if err != nil {
			utils.Error("Error getting documents from " + collection + " collection", err)
			return
	}
	defer cur.Close(context.Background())

	var batch []interface{}
	batchSize := 100 // Define a suitable batch size

	for cur.Next(context.Background()) {
			var elem bson.D
			if err := cur.Decode(&elem); err != nil {
					utils.Error("Error decoding document from " + collection + " collection", err)
					continue
			}

			batch = append(batch, elem)

			if len(batch) >= batchSize {
					if _, err := ct.InsertMany(context.Background(), batch); err != nil {
							utils.Error("Error inserting batch into " + collection + " collection", err)
					}
					batch = batch[:0] // Clear the batch
			}
	}

	// Insert any remaining documents
	if len(batch) > 0 {
			if _, err := ct.InsertMany(context.Background(), batch); err != nil {
					utils.Error("Error inserting remaining documents into " + collection + " collection", err)
			}
	}

	if err := cur.Err(); err != nil {
			utils.Error("Error iterating over documents from " + collection + " collection", err)
	}
}


func MigratePre014() {
	config := utils.GetMainConfig()

	// check if COSMOS.db does NOT exist
	if _, err := os.Stat(utils.CONFIGFOLDER + "database"); err != nil && config.MongoDB != "" {
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
		
		ct, closeDb, err := utils.GetEmbeddedCollection(utils.GetRootAppId(), "events")
		if err != nil {
				return
		}
		defer closeDb()

		// Create a date index
		model := mongo.IndexModel{
			Keys: bson.M{"Date": -1},
		}

		// Creating the index
		_, err = ct.Indexes().CreateOne(context.Background(), model)
		if err != nil {
			utils.Error("Metrics - Create Index", err)
			return // Handle error appropriately
		}
		
		MigratePre014Coll("users", client)
		MigratePre014Coll("devices", client)
		// MigratePre014Coll("events", client)
		// MigratePre014Coll("notifications", client)
		// MigratePre014Coll("metrics", client)
	}
}