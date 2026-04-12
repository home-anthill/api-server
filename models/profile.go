package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
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

// DbGithubUserTestmock mock object
var DbGithubUserTestmock = GitHub{
	ID:        123456,
	Login:     "Test",
	Name:      "Test Test",
	Email:     "test@test.com",
	AvatarURL: "https://avatars.githubusercontent.com/u/123456?v=4",
}

// Profile struct
type Profile struct {
	ID         bson.ObjectID   `json:"id" bson:"_id"`
	Github     GitHub          `json:"github" bson:"github"`
	APIToken   string          `json:"apiToken" bson:"apiToken"`
	Devices    []bson.ObjectID `json:"devices" bson:"devices"`
	Homes      []bson.ObjectID `json:"homes" bson:"homes"`
	CreatedAt  time.Time       `json:"createdAt" bson:"createdAt"`
	ModifiedAt time.Time       `json:"modifiedAt" bson:"modifiedAt"`
	// as recommended by official FCM documentation, we save both token and timestamp
	// More info at https://firebase.google.com/docs/cloud-messaging/manage-tokens
	FCMToken          string    `json:"fcmToken" bson:"fcmToken"`
	FCMTokenTimestamp time.Time `json:"fcmTokenTimestamp" bson:"fcmTokenTimestamp"`
}
