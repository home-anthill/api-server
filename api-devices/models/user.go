package models

type User struct {
	ID      int64  `json:"id" bson:"id"`
	Login   string `json:"login" bson:"login"`
	Name    string `json:"name" bson:"name"`
	Email   string `json:"email" bson:"email"`
	Company string `json:"company" bson:"company"`
	URL     string `json:"url" bson:"url"`
}
