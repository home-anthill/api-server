package db

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
	"os"
)

var client *mongo.Client

// InitDb function
func InitDb(ctx context.Context, logger *zap.SugaredLogger) (*mongo.Collection, *mongo.Collection, *mongo.Collection) {
	mongoDBUrl := os.Getenv("MONGODB_URL")
	logger.Info("InitDb - connecting to MongoDB URL = " + mongoDBUrl)

	// connect to DB
	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoDBUrl))
	if err != nil {
		logger.Fatalf("Cannot connect to MongoDB: %s", err)
		panic("Cannot connect to MongoDB")
	}
	if os.Getenv("ENV") != "prod" {
		if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
			logger.Fatalf("Cannot ping MongoDB: %s", err)
			panic("Cannot ping MongoDB")
		}
	}
	logger.Info("Connected to MongoDB")

	var dbName string
	if os.Getenv("ENV") == "testing" {
		dbName = "api-server-test"
	} else {
		dbName = "api-server"
	}
	collectionProfiles := client.Database(dbName).Collection("profiles")
	collectionHomes := client.Database(dbName).Collection("homes")
	collectionDevices := client.Database(dbName).Collection("devices")
	return collectionProfiles, collectionHomes, collectionDevices
}
