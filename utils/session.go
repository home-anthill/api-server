package utils

import (
	"api-server/models"
	"fmt"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
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

// GetProfileFromContext returns the authenticated identity for JWT-protected
// handlers. These routes must pass through JWTMiddleware, which stores the
// already-validated claims in the Gin context.
func GetProfileFromContext(c *gin.Context) (SessionProfile, error) {
	value, exists := c.Get("jwt_claims")
	if !exists {
		return SessionProfile{}, fmt.Errorf("jwt claims not found in context")
	}

	claims, ok := value.(*JWTClaims)
	if !ok || claims == nil {
		return SessionProfile{}, fmt.Errorf("invalid jwt claims in context")
	}

	profileID, err := bson.ObjectIDFromHex(claims.ProfileID)
	if err != nil {
		return SessionProfile{}, fmt.Errorf("invalid profile in jwt claims: %w", err)
	}
	if claims.ID == 0 {
		return SessionProfile{}, fmt.Errorf("missing github id in jwt claims")
	}

	return SessionProfile{
		ID:       profileID,
		GithubID: claims.ID,
	}, nil
}

// GetLoggedProfileFromContext loads the current profile from MongoDB using the
// identity stored in JWT claims.
func GetLoggedProfileFromContext(c *gin.Context, collection *mongo.Collection) (models.Profile, error) {
	profileSession, err := GetProfileFromContext(c)
	if err != nil {
		return models.Profile{}, err
	}

	var profile models.Profile
	err = collection.FindOne(c.Request.Context(), bson.M{
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
