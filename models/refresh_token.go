package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type RefreshToken struct {
	ID             bson.ObjectID `bson:"_id,omitempty"`
	ProfileID      bson.ObjectID `bson:"profileId"`
	TokenHash      string        `bson:"tokenHash"`
	FamilyID       string        `bson:"familyId"`
	ClientType     string        `bson:"clientType"`
	CreatedAt      time.Time     `bson:"createdAt"`
	ExpiresAt      time.Time     `bson:"expiresAt"`
	RevokedAt      *time.Time    `bson:"revokedAt,omitempty"`
	ReplacedByHash string        `bson:"replacedByHash,omitempty"`
	LastUsedAt     *time.Time    `bson:"lastUsedAt,omitempty"`
}
