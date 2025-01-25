package models

import (
	"time"
)

// Online struct
type Online struct {
	CreatedAt   time.Time `json:"createdAt"`
	ModifiedAt  time.Time `json:"modifiedAt"`
	CurrentTime time.Time `json:"currentTime"`
}
