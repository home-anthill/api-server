package db

import (
	"context"
	"fmt"
	"os"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.uber.org/zap"
)

var client *mongo.Client

// Collections struct
type Collections struct {
	Profiles      *mongo.Collection
	Homes         *mongo.Collection
	Devices       *mongo.Collection
	AppLoginCodes *mongo.Collection
	RefreshTokens *mongo.Collection
}

// InitDb function
func InitDb(logger *zap.SugaredLogger) *mongo.Client {
	mongoDBUrl := os.Getenv("MONGODB_URL")
	logger.Info("InitDb - connecting to MongoDB URL = [redacted]")

	// connect to DB
	var err error
	client, err = mongo.Connect(options.Client().ApplyURI(mongoDBUrl))
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

	if err = ensureIndexes(client, logger); err != nil {
		logger.Fatalf("Cannot ensure MongoDB indexes: %s", err)
		panic("Cannot ensure MongoDB indexes")
	}

	return client
}

// GetCollections function
func GetCollections(client *mongo.Client) *Collections {
	return &Collections{
		Profiles:      client.Database(getDbName()).Collection("profiles"),
		Homes:         client.Database(getDbName()).Collection("homes"),
		Devices:       client.Database(getDbName()).Collection("devices"),
		AppLoginCodes: client.Database(getDbName()).Collection("app_login_codes"),
		RefreshTokens: client.Database(getDbName()).Collection("refresh_tokens"),
	}
}

func ensureIndexes(client *mongo.Client, logger *zap.SugaredLogger) error {
	colls := GetCollections(client)

	_, err := colls.AppLoginCodes.Indexes().CreateMany(context.TODO(), []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "code", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("app_login_code_code_unique"),
		},
		{
			Keys:    bson.D{{Key: "expiresAt", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0).SetName("app_login_code_expires_ttl"),
		},
	})
	if err != nil {
		return fmt.Errorf("cannot create app_login_codes indexes: %w", err)
	}

	_, err = colls.RefreshTokens.Indexes().CreateMany(context.TODO(), []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "tokenHash", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("refresh_token_hash_unique"),
		},
		{
			Keys:    bson.D{{Key: "familyId", Value: 1}},
			Options: options.Index().SetName("refresh_token_family"),
		},
		{
			Keys:    bson.D{{Key: "expiresAt", Value: 1}},
			Options: options.Index().SetName("refresh_token_expires"),
		},
	})
	if err != nil {
		return fmt.Errorf("cannot create refresh_tokens indexes: %w", err)
	}

	logger.Info("MongoDB indexes ensured")
	return nil
}

// getDbName function
func getDbName() string {
	if os.Getenv("ENV") == "testing" {
		return "api-server-test"
	} else {
		return "api-server"
	}
}
