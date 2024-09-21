package utils

import (
	"context"
	"os"
	"errors"
	"sync"
	"time"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"go.mongodb.org/mongo-driver/bson" 

	lungo "github.com/256dpi/lungo"
)


var client *mongo.Client
var DBContainerName string
var DBStatus bool

func DB() error {
	config := GetMainConfig()

	if(config.DisableUserManagement)	{
		DBStatus = false
		return errors.New("User Management is disabled")
	}

	mongoURL := config.MongoDB
	
	if(client != nil && client.Ping(context.TODO(), readpref.Primary()) == nil) {
		DBStatus = true
		return nil
	}

	Log("(Re) Connecting to the database...")
	
	DBContainerName = ""
	
	isPuppetMode := config.Database.PuppetMode
	puppetHostname := config.Database.Hostname
	username := config.Database.Username
	password := config.Database.Password
	
	if isPuppetMode {
		mongoURL = "mongodb://" + username + ":" + password + "@" + puppetHostname + ":27017"
		DBContainerName = puppetHostname
	}

	if mongoURL == "" {
		DBStatus = false
		return errors.New("MongoDB URL is not set, cannot connect to the database.")
	}

	var err error

	opts := options.Client().
	  SetConnectTimeout(7 * time.Second).
		ApplyURI(mongoURL).
		SetRetryWrites(true).
		SetWriteConcern(writeconcern.New(writeconcern.WMajority()))
	
	hostname := ""
	port := "27017"

	if !isPuppetMode {
		if len(opts.Hosts) == 0 {
			// TODO
			return errors.New("Cannot establish DB host, please check config")
		}

		hostname = opts.Hosts[0]
		// split port
		hostnameParts := strings.Split(hostname, ":")
		hostname = hostnameParts[0]

		if len(hostnameParts) > 1 {
			port = hostnameParts[1]
		}
	} else {
		hostname = puppetHostname
	}

	if !IsInsideContainer || IsHostNetwork {
		Log("Getting Mongo DB IP from name : " + hostname + " (port " + port + ")")

		ip, _ := GetContainerIPByName(hostname)
		if ip != "" {
			DBContainerName = hostname
			opts.SetHosts([]string{ip + ":" + port})
			Log("Mongo DB IP : " + ip)
		}
	} else {
		if DoesContainerExist(hostname) {
			Debug("MongoDB is running in a container")
			DBContainerName = hostname
		}
	}
	
	client, err = mongo.Connect(context.TODO(), opts)

	if err != nil {
		DBStatus = false
		return err
	}

	// Ping the primary
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		DBStatus = false
		return err
	}

	initDB()

	DBStatus = true

	Log("Successfully connected to the database.")
	return nil
}

func DisconnectDB() {
	if client == nil {
		return
	}
	if err := client.Disconnect(context.TODO()); err != nil {
		Error("DB", err)
	}
	client = nil
}

var embeddedClient lungo.IClient
var embeddedClientClose func()

func CloseEmbeddedDB() {
	if embeddedClientClose != nil {
		embeddedClientClose()
		embeddedClientClose = nil
	}
	embeddedClient = nil
}

func GetEmbeddedCollection(applicationId string, collection string) (lungo.ICollection, func(), error) {
	opts := lungo.Options{
		Store: lungo.NewFileStore(CONFIGFOLDER + "database", 0700),
	}
	
	name := os.Getenv("MONGODB_NAME"); if name == "" {
		name = "COSMOS"
	}

	if embeddedClient != nil {
		c := embeddedClient.Database(name).Collection(applicationId + "_" + collection)
		return c, func() {
			//engine.Close()
		}, nil
	}

	// open database
	client, engine, err := lungo.Open(nil, opts)
	if err != nil {
		return nil, nil, err
	}

	// ensure engine is closed
	// defer engine.Close()
	
	c := client.Database(name).Collection(applicationId + "_" + collection)

	embeddedClient = client
	embeddedClientClose = engine.Close

	return c, func() {
		//engine.Close()
	}, nil
}

func GetCollection(applicationId string, collection string) (*mongo.Collection, error) {
	if client == nil {
		errCo := DB()
		if errCo != nil {
			return nil, errCo
		}
	}

	DBStatus = true
	
	name := os.Getenv("MONGODB_NAME"); if name == "" {
		name = "COSMOS"
	}
	
	// Debug("Getting collection " + applicationId + "_" + collection + " from database " + name)
	
	c := client.Database(name).Collection(applicationId + "_" + collection)
	
	return c, nil
}

// func query(q string) (*sql.Rows, error) {
// 	return db.Query(q)
// }

var (
	bufferLock     sync.Mutex
	writeBuffer    = make(map[string][]map[string]interface{})
	bufferTicker   = time.NewTicker(1 * time.Minute)
	bufferCapacity = 100
)

func GetWriteBuffer() map[string][]map[string]interface{} {
	return writeBuffer
}

func GetWriteBufferLength() int {
	result := 0
	for _, objects := range writeBuffer {
		buflen := 0
		for _, _ = range objects {
			buflen++
		}
		result += len(objects)
	}
	return result
}

func InitDBBuffers() {
	go func() {
		for {
			select {
			case <-bufferTicker.C:
				flushAllBuffers()
			}
		}
	}()
}

func flushBuffer(collectionName string) {
	bufferLock.Lock()
	defer bufferLock.Unlock()
	objects, exists := writeBuffer[collectionName]
	if exists && len(objects) > 0 {
		collection, errG := GetCollection(GetRootAppId(), collectionName)
		if errG != nil {
			Error("BulkDBWritter: Error getting collection", errG)
		} else {
			if err := WriteToDatabase(collection, objects); err != nil {
				Error("BulkDBWritter: Error writing to database", err)
			}
			writeBuffer[collectionName] = make([]map[string]interface{}, 0)
		}
	}
}

func flushAllBuffers() {
	bufferLock.Lock()
	defer bufferLock.Unlock()
	for collectionName, objects := range writeBuffer {
		if len(objects) > 0 {
			collection, errG := GetCollection(GetRootAppId(), collectionName)
			if errG != nil {
				Error("BulkDBWritter: Error getting collection", errG)
			} else {
				if err := WriteToDatabase(collection, objects); err != nil {
					Error("BulkDBWritter: Error writing to database: ", err)
				}
				writeBuffer[collectionName] = make([]map[string]interface{}, 0)
			}
		}
	}
}

func BufferedDBWrite(collectionName string, object map[string]interface{}) {
	bufferLock.Lock()
	writeBuffer[collectionName] = append(writeBuffer[collectionName], object)
	
	if len(writeBuffer[collectionName]) >= bufferCapacity {
		bufferLock.Unlock()
		flushBuffer(collectionName)
	} else {
		bufferLock.Unlock()
	}
}

func WriteToDatabase(collection *mongo.Collection, objects []map[string]interface{}) error {
	if len(objects) == 0 {
		return nil // Nothing to write
	}

	// Convert to a slice of interface{} for insertion
	interfaceSlice := make([]interface{}, len(objects))
	for i, v := range objects {
		interfaceSlice[i] = v
	}

	_, err := collection.InsertMany(context.Background(), interfaceSlice)
	if err != nil {
		return err
	}

	return nil
}

func ListAllUsers(role string) []User { 
	// list all users
	c, closeDb, errCo := GetEmbeddedCollection(GetRootAppId(), "users")
  defer closeDb()
	if errCo != nil {
			Error("Database Connect", errCo)
			return []User{}
	}

	users := []User{}

	condition := map[string]interface{}{}

	if role == "admin" {
		condition = map[string]interface{}{
			"Role": 2,
		}
	} else if role == "user" {
		condition = map[string]interface{}{
			"Role": 1,
		}
	}

	cursor, err := c.Find(nil, condition)
	defer cursor.Close(nil)

	if err != nil {
		Error("Database: Error while getting users", err)
		return []User{}
	}
	
	if err = cursor.All(nil, &users); err != nil {
		Error("Database: Error while decoding users", err)
		return []User{}
	}

	return users
}

func initDB() {
	c, errCo := GetCollection(GetRootAppId(), "events")
	if errCo != nil {
		Error("Metrics - Database Connect", errCo)
	} else {
		// Create a text index on the _search field
		model := mongo.IndexModel{
			Keys: bson.M{"_search": "text"},
		}

		// Creating the index
		_, err := c.Indexes().CreateOne(context.Background(), model)
		if err != nil {
			Error("Metrics - Create Index", err)
			return // Handle error appropriately
		}
		
		// Create a date index
		model = mongo.IndexModel{
			Keys: bson.M{"Date": -1},
		}

		// Creating the index
		_, err = c.Indexes().CreateOne(context.Background(), model)
		if err != nil {
			Error("Metrics - Create Index", err)
			return // Handle error appropriately
		}
	}

	c, errCo = GetCollection(GetRootAppId(), "metrics")
	if errCo != nil {
		Error("Metrics - Database Connect", errCo)
	} else {
		// create search index on metrics key
		model := mongo.IndexModel{
			Keys: bson.M{"Key": 1},
		}

		// Creating the index
		_, err := c.Indexes().CreateOne(context.Background(), model)
		if err != nil {
			Error("Metrics - Create Index", err)
			return // Handle error appropriately
		}
	}
}
