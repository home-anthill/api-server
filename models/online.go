package models

import (
	"time"
)

// Online struct
type Online struct {
	UUID       string    `json:"uuid"`
	APIToken   string    `json:"apiToken"`
	CreatedAt  time.Time `json:"createdAt"`
	ModifiedAt time.Time `json:"modifiedAt"`
}
