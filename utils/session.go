package utils

import (
	"api-server/models"
	"context"
	"fmt"

	"github.com/gin-contrib/sessions"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// SessionProfile is the minimal identity reconstructed from primitive session
// values. The session stores profileID and githubID separately to avoid gob
// encoding custom structs into the cookie.
type SessionProfile struct {
	ID       bson.ObjectID `json:"id"`
	GithubID int64         `json:"githubId"`
}

// GetProfileFromSession returns the minimal logged-in profile identity stored
// in the current session.
func GetProfileFromSession(session sessions.Session) (SessionProfile, error) {
	profileIDHex, ok := session.Get("profileID").(string)
	if !ok || profileIDHex == "" {
		return SessionProfile{}, fmt.Errorf("cannot find profile in session")
	}

	profileID, err := bson.ObjectIDFromHex(profileIDHex)
	if err != nil {
		return SessionProfile{}, fmt.Errorf("invalid profile in session: %w", err)
	}

	githubID, ok := session.Get("githubID").(int64)
	if !ok || githubID == 0 {
		return SessionProfile{}, fmt.Errorf("cannot find profile in session")
	}

	return SessionProfile{
		ID:       profileID,
		GithubID: githubID,
	}, nil
}

// GetLoggedProfile loads the current profile from MongoDB using the profile ID
// stored in session, ensuring handlers use fresh server-side profile data.
func GetLoggedProfile(ctx context.Context, session sessions.Session, collection *mongo.Collection) (models.Profile, error) {
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

// Contains reports whether s ObjectID list contains objToFind ObjectID.
func Contains(s []bson.ObjectID, objToFind bson.ObjectID) bool {
	for _, v := range s {
		if v == objToFind {
			return true
		}
	}
	return false
}
