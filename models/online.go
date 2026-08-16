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

type OnlineStatus string

const (
	OnlineStatusOnline  OnlineStatus = "online"
	OnlineStatusOffline OnlineStatus = "offline"
	OnlineStatusUnknown OnlineStatus = "unknown"
)

// OnlineDeviceStatus is the client-facing status for one enabled online feature.
type OnlineDeviceStatus struct {
	DeviceID    string       `json:"deviceId"`
	FeatureUUID string       `json:"featureUuid"`
	Status      OnlineStatus `json:"status"`
	CreatedAt   *time.Time   `json:"createdAt"`
	ModifiedAt  *time.Time   `json:"modifiedAt"`
	CurrentTime time.Time    `json:"currentTime"`
}
