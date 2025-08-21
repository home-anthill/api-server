package utils

import (
	"api-server/models"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/gin-contrib/sessions"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
)

// GetProfileFromSession retrieve current profile ID from session
func GetProfileFromSession(session *sessions.Session) (models.Profile, error) {
	profileSession, ok := (*session).Get("profile").(models.Profile)
	if !ok {
		return models.Profile{}, fmt.Errorf("cannot find profile in session")
	}
	return profileSession, nil
}

func GetLoggedProfile(ctx context.Context, session *sessions.Session, collection *mongo.Collection) (models.Profile, error) {
	profileSession, err := GetProfileFromSession(session)
	if err != nil {
		return models.Profile{}, err
	}
	// search the current profile in DB
	// This is required to get fresh data from db, because data in session could be outdated
	var profile models.Profile
	err = collection.FindOne(ctx, bson.M{
		"_id": profileSession.ID,
	}).Decode(&profile)
	return profile, err
}

func Contains(s []primitive.ObjectID, objToFind primitive.ObjectID) bool {
	for _, v := range s {
		if v.Hex() == objToFind.Hex() {
			return true
		}
	}
	return false
}

func RandToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}
