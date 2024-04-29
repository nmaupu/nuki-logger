package model

import "time"

type NukiReservationResponse struct {
	ID                      string    `json:"id"`
	AddressID               int       `json:"addressId"`
	AccountID               int       `json:"accountId"`
	Email                   string    `json:"email"`
	Name                    string    `json:"name"`
	Guests                  int       `json:"guests"`
	GuestsIssued            int       `json:"guestsIssued"`
	KeypadIssued            bool      `json:"keypadIssued"`
	State                   string    `json:"state"`
	ServiceId               string    `json:"serviceId"`
	Reference               string    `json:"reference"`
	Automation              int       `json:"automation"`
	CheckedIn               bool      `json:"checkedIn"`
	StartDate               time.Time `json:"startDate"`
	EndDate                 time.Time `json:"endDate"`
	UpdateDate              time.Time `json:"updateDate"`
	IsCurrentlyIssuingAuth  bool      `json:"isCurrentlyIssuingAuth"`
	IsCurrentlyRevokingAuth bool      `json:"isCurrentlyRevokingAuth"`
	HasCustomAccessTimes    bool      `json:"hasCustomAccessTimes"`
	CurrentlyIssuingAuth    bool      `json:"currentlyIssuingAuth"`
	CurrentlyRevokingAuth   bool      `json:"currentlyRevokingAuth"`
}
