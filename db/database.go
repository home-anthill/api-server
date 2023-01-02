package db

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
	"os"
)

const DbName = "api-server"

var client *mongo.Client

func InitDb(ctx context.Context, logger *zap.SugaredLogger) (*mongo.Collection, *mongo.Collection, *mongo.Collection) {
	mongoDBUrl := os.Getenv("MONGODB_URL")
	logger.Info("InitDb - connecting to MongoDB URL = " + mongoDBUrl)

	// connect to DB
	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoDBUrl))
	if os.Getenv("ENV") != "prod" {
		if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
			logger.Fatalf("Cannot connect to MongoDB: %s", err)
			panic("Cannot connect to MongoDB")
		}
	}
	logger.Info("Connected to MongoDB")

	// define DB collections
	var collNameProfiles string
	var collNameHomes string
	var collNameDevices string
	if os.Getenv("ENV") == "testing" {
		collNameProfiles = "profiles_test"
		collNameHomes = "homes_test"
		collNameDevices = "devices_test"
	} else {
		collNameProfiles = "profiles"
		collNameHomes = "homes"
		collNameDevices = "devices"
	}
	collectionProfiles := client.Database(DbName).Collection(collNameProfiles)
	collectionHomes := client.Database(DbName).Collection(collNameHomes)
	collectionDevices := client.Database(DbName).Collection(collNameDevices)
	return collectionProfiles, collectionHomes, collectionDevices
}
