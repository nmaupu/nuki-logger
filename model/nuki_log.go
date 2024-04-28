package model

import (
	"slices"
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

func (n NukiSmartlockLogResponse) Equals(n2 NukiSmartlockLogResponse) bool {
	return n.ID == n2.ID
}

// Diff returns new NukiSmartlockLogResponse from new not present in old
func Diff(new, old []NukiSmartlockLogResponse) []NukiSmartlockLogResponse {
	if len(new) == 0 {
		return []NukiSmartlockLogResponse{}
	}
	if len(old) == 0 {
		return new
	}

	if new[0].Equals(old[0]) { // no new logs
		return []NukiSmartlockLogResponse{}
	}

	var diff []NukiSmartlockLogResponse
	for _, r := range new {
		// While not equals the first old, adding this new log
		if r.Equals(old[0]) {
			break
		}
		diff = append(diff, r)
	}

	slices.Reverse(diff)
	return diff
}
