package model

import "time"

type SmartlockAuthResponse struct {
	Id            string `json:"id"`
	SmartlockId   int    `json:"smartlockId"`
	AccountUserId int    `json:"accountUserId"`
	AuthId        int    `json:"authId"`
	Code          int    `json:"code"`
	Fingerprints  struct {
		AdditionalProp1 string `json:"additionalProp1"`
		AdditionalProp2 string `json:"additionalProp2"`
		AdditionalProp3 string `json:"additionalProp3"`
	} `json:"fingerprints"`
	Type             int       `json:"type"`
	Name             string    `json:"name"`
	Enabled          bool      `json:"enabled"`
	RemoteAllowed    bool      `json:"remoteAllowed"`
	LockCount        int       `json:"lockCount"`
	AllowedFromDate  time.Time `json:"allowedFromDate"`
	AllowedUntilDate time.Time `json:"allowedUntilDate"`
	AllowedWeekDays  int       `json:"allowedWeekDays"`
	AllowedFromTime  int       `json:"allowedFromTime"`
	AllowedUntilTime int       `json:"allowedUntilTime"`
	LastActiveDate   time.Time `json:"lastActiveDate"`
	CreationDate     time.Time `json:"creationDate"`
	UpdateDate       time.Time `json:"updateDate"`
	OperationId      struct {
		Timestamp         int       `json:"timestamp"`
		Counter           int       `json:"counter"`
		Time              int       `json:"time"`
		Date              time.Time `json:"date"`
		MachineIdentifier int       `json:"machineIdentifier"`
		ProcessIdentifier int       `json:"processIdentifier"`
		TimeSecond        int       `json:"timeSecond"`
	} `json:"operationId"`
	Error            string `json:"error"`
	AuthTypeAsString string `json:"authTypeAsString"`
}
