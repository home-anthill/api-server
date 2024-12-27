package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// GitHub struct
// no need to define json, because we set this value manually
// we need only bson, because we want to store it into the db as part of Profile
type GitHub struct {
	ID        int64  `json:"id" bson:"id"`
	Login     string `json:"login" bson:"login"`
	Name      string `json:"name" bson:"name"`
	Email     string `json:"email" bson:"email"`
	AvatarURL string `json:"avatarURL" bson:"avatarURL"`
}

// Profile struct
type Profile struct {
	ID         primitive.ObjectID   `json:"id" bson:"_id"`
	Github     GitHub               `json:"github" bson:"github"`
	APIToken   string               `json:"apiToken" bson:"apiToken"`
	FCMToken   string               `json:"fcmToken" bson:"fcmToken"`
	Devices    []primitive.ObjectID `json:"devices" bson:"devices"`
	Homes      []primitive.ObjectID `json:"homes" bson:"homes"`
	CreatedAt  time.Time            `json:"createdAt" bson:"createdAt"`
	ModifiedAt time.Time            `json:"modifiedAt" bson:"modifiedAt"`
}
