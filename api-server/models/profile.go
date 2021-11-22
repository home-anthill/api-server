package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// no need to define json, because we set this value manually
// we need onlt bson, because we want to store it into the db as part of Profile
type Github struct {
	ID        int64  `bson:"id"`
	Login     string `bson:"login"`
	Name      string `bson:"name"`
	Email     string `bson:"email"`
	AvatarURL string `bson:"avatarURL"`
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
