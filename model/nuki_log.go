package model

import (
	"time"
)

type NukiSmartlockLogResponse struct {
	ID            string         `json:"id"`
	SmartLockID   int64          `json:"smartlockId"`
	DeviceType    NukiDeviceType `json:"deviceType"`
	AccountUserID int32          `json:"accountUserId"`
	AuthID        string         `json:"authId"`
	Name          string         `json:"name"`
	Action        NukiAction     `json:"action"`
	Trigger       NukiTrigger    `json:"trigger"`
	State         NukiState      `json:"state"`
	AutoUnlock    bool           `json:"autoUnlock"`
	Date          time.Time      `json:"date"`
	Source        NukiSource     `source:"source"`
}

func (n NukiDeviceType) String() string {
	str, ok := NukiDeviceTypes[n]
	if !ok {
		return "unknown"
	}
	return str
}

func (n NukiAction) String() string {
	str, ok := NukiActions[n]
	if !ok {
		return "unknown"
	}
	return str
}

func (n NukiTrigger) String() string {
	str, ok := NukiTriggers[n]
	if !ok {
		return "unknown"
	}
	return str
}

func (n NukiState) String() string {
	str, ok := NukiStates[n]
	if !ok {
		return "unknown"
	}
	return str
}

func (n NukiSource) String() string {
	str, ok := NukiSources[n]
	if !ok {
		return "unknown"
	}
	return str
}
