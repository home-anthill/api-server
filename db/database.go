package db

import (
	"context"
	"errors"
	"fmt"
	"os"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.uber.org/zap"
)

// Collections struct
type Collections struct {
	Profiles      *mongo.Collection
	Homes         *mongo.Collection
	Devices       *mongo.Collection
	AppLoginCodes *mongo.Collection
	RefreshTokens *mongo.Collection
}

// InitDb connects to MongoDB and ensures the required indexes.
func InitDb(ctx context.Context, logger *zap.SugaredLogger) (*mongo.Client, error) {
	mongoDBUrl := os.Getenv("MONGODB_URL")
	logger.Info("InitDb - connecting to MongoDB URL = [redacted]")

	// connect to DB
	client, err := mongo.Connect(options.Client().ApplyURI(mongoDBUrl))
	if err != nil {
		return nil, fmt.Errorf("connect to MongoDB: %w", err)
	}
	if os.Getenv("ENV") != "prod" {
		if err = client.Ping(ctx, readpref.Primary()); err != nil {
			return nil, fmt.Errorf("ping MongoDB: %w", err)
		}
	}
	logger.Info("Connected to MongoDB")

	if err = ensureIndexes(ctx, client, logger); err != nil {
		return nil, fmt.Errorf("ensure MongoDB indexes: %w", err)
	}

	return client, nil
}

// GetCollections function
func GetCollections(client *mongo.Client) *Collections {
	database := client.Database(getDbName())
	return &Collections{
		Profiles:      database.Collection("profiles"),
		Homes:         database.Collection("homes"),
		Devices:       database.Collection("devices"),
		AppLoginCodes: database.Collection("app_login_codes"),
		RefreshTokens: database.Collection("refresh_tokens"),
	}
}

func ensureIndexes(ctx context.Context, client *mongo.Client, logger *zap.SugaredLogger) error {
	colls := GetCollections(client)

	_, err := colls.Profiles.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "github.id", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("profile_github_id_unique"),
	})
	if err != nil {
		return fmt.Errorf("cannot create profiles indexes: %w", err)
	}

	_, err = colls.AppLoginCodes.Indexes().CreateMany(ctx, []mongo.IndexModel{
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

	if err = colls.RefreshTokens.Indexes().DropOne(ctx, "refresh_token_expires"); err != nil {
		var commandErr mongo.CommandError
		if !errors.As(err, &commandErr) || (commandErr.Code != 26 && commandErr.Code != 27) {
			return fmt.Errorf("cannot drop old refresh_token_expires index: %w", err)
		}
	}

	_, err = colls.RefreshTokens.Indexes().CreateMany(ctx, []mongo.IndexModel{
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
			Options: options.Index().SetExpireAfterSeconds(0).SetName("refresh_token_expires_ttl"),
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
	}
	return "api-server"
}
