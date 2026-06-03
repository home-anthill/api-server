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

// OnlineDeviceStatus includes the online status with its owning device and feature.
type OnlineDeviceStatus struct {
	CreatedAt   time.Time `json:"createdAt"`
	ModifiedAt  time.Time `json:"modifiedAt"`
	CurrentTime time.Time `json:"currentTime"`
	Device      Device    `json:"device"`
	Feature     Feature   `json:"feature"`
}
