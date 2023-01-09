package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// Github struct
// no need to define json, because we set this value manually
// we need onlt bson, because we want to store it into the db as part of Profile
type Github struct {
	ID        int64  `json:"id" bson:"id"`
	Login     string `json:"login" bson:"login"`
	Name      string `json:"name" bson:"name"`
	Email     string `json:"email" bson:"email"`
	AvatarURL string `json:"avatarURL" bson:"avatarURL"`
}

type Profile struct {
	ID         primitive.ObjectID   `json:"id" bson:"_id"`
	Github     Github               `json:"github" bson:"github"`
	ApiToken   string               `json:"apiToken" bson:"apiToken"`
	Devices    []primitive.ObjectID `json:"devices" bson:"devices"`
	Homes      []primitive.ObjectID `json:"homes" bson:"homes"`
	CreatedAt  time.Time            `json:"createdAt" bson:"createdAt"`
	ModifiedAt time.Time            `json:"modifiedAt" bson:"modifiedAt"`
}
