package utils

import (
	"context"
	"log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/imdario/mergo"
)

type GucoConfiguration struct {
	Test string
	Caca string
}

var defaultConfig = GucoConfiguration{
	Test: "test",
	Caca: "prout",
}

func GetConfigs() GucoConfiguration {
	c := GetCollection(GetRootAppId(), "Configurations")
	config := GucoConfiguration{}
	err := c.FindOne(context.TODO(), bson.M{"_id": GetRootAppId()}).Decode(&config)
	if err == mongo.ErrNoDocuments {
    log.Println("Record does not exist")
	} else if err != nil {
			log.Fatal(err)
	}
	
	mergo.Merge(&config, defaultConfig)

	return config;
}

func SetConfig(config GucoConfiguration) {
	currentConfig := GetConfigs()
	
	mergo.Merge(&config, currentConfig)

	c := GetCollection(GetRootAppId(), "Configurations")

	opts := options.Update().SetUpsert(true)
	filter := bson.D{{"_id", GetRootAppId()}}
	update := bson.D{{"$set", config}}

	_, err := c.UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		log.Fatal(err)
	}
}