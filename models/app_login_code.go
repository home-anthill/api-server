package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// AppLoginCode is a short-lived, single-use code for exchanging a completed
// mobile OAuth login into app tokens over HTTPS.
type AppLoginCode struct {
	ID                  bson.ObjectID `json:"id" bson:"_id"`
	Code                string        `json:"code" bson:"code"`
	ProfileID           bson.ObjectID `json:"profileId" bson:"profileId"`
	PKCECodeChallenge   string        `json:"pkceCodeChallenge" bson:"pkceCodeChallenge"`
	PKCEChallengeMethod string        `json:"pkceChallengeMethod" bson:"pkceChallengeMethod"`
	ExpiresAt           time.Time     `json:"expiresAt" bson:"expiresAt"`
	UsedAt              *time.Time    `json:"usedAt,omitempty" bson:"usedAt,omitempty"`
	CreatedAt           time.Time     `json:"createdAt" bson:"createdAt"`
}
