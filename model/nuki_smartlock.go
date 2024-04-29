package model

import "time"

type SmartLockState struct {
	Name                      string `json:"name"`
	BatteryCritical           bool   `json:"battery_critical"`
	BatteryCharge             int32  `json:"battery_remaining_percent"`
	KeypadBatteryCritical     bool   `json:"keypad_battery_critical"`
	DoorsensorBatteryCritical bool   `json:"doorsensor_battery_critical"`
}

type SmartlockResponse struct {
	SmartlockId int64  `json:"smartlockId"`
	AccountId   int    `json:"accountId"`
	Type        int    `json:"type"`
	LmType      int    `json:"lmType"`
	AuthId      int    `json:"authId"`
	Name        string `json:"name"`
	Favorite    bool   `json:"favorite"`
	Config      struct {
		Name               string  `json:"name"`
		Latitude           float64 `json:"latitude"`
		Longitude          float64 `json:"longitude"`
		AutoUnlatch        bool    `json:"autoUnlatch"`
		LiftUpHandle       bool    `json:"liftUpHandle"`
		PairingEnabled     bool    `json:"pairingEnabled"`
		ButtonEnabled      bool    `json:"buttonEnabled"`
		LedEnabled         bool    `json:"ledEnabled"`
		LedBrightness      int     `json:"ledBrightness"`
		TimezoneOffset     int     `json:"timezoneOffset"`
		DaylightSavingMode int     `json:"daylightSavingMode"`
		FobPaired          bool    `json:"fobPaired"`
		FobAction1         int     `json:"fobAction1"`
		FobAction2         int     `json:"fobAction2"`
		FobAction3         int     `json:"fobAction3"`
		SingleLock         bool    `json:"singleLock"`
		AdvertisingMode    int     `json:"advertisingMode"`
		KeypadPaired       bool    `json:"keypadPaired"`
		Keypad2Paired      bool    `json:"keypad2Paired"`
		HomekitState       int     `json:"homekitState"`
		MatterState        int     `json:"matterState"`
		TimezoneId         int     `json:"timezoneId"`
		DeviceType         int     `json:"deviceType"`
		WifiEnabled        bool    `json:"wifiEnabled"`
	} `json:"config"`
	AdvancedConfig struct {
		TotalDegrees                            int  `json:"totalDegrees"`
		SingleLockedPositionOffsetDegrees       int  `json:"singleLockedPositionOffsetDegrees"`
		UnlockedToLockedTransitionOffsetDegrees int  `json:"unlockedToLockedTransitionOffsetDegrees"`
		UnlockedPositionOffsetDegrees           int  `json:"unlockedPositionOffsetDegrees"`
		LockedPositionOffsetDegrees             int  `json:"lockedPositionOffsetDegrees"`
		DetachedCylinder                        bool `json:"detachedCylinder"`
		BatteryType                             int  `json:"batteryType"`
		AutoLock                                bool `json:"autoLock"`
		AutoLockTimeout                         int  `json:"autoLockTimeout"`
		AutoUpdateEnabled                       bool `json:"autoUpdateEnabled"`
		LngTimeout                              int  `json:"lngTimeout"`
		SingleButtonPressAction                 int  `json:"singleButtonPressAction"`
		DoubleButtonPressAction                 int  `json:"doubleButtonPressAction"`
		AutomaticBatteryTypeDetection           bool `json:"automaticBatteryTypeDetection"`
		UnlatchDuration                         int  `json:"unlatchDuration"`
	} `json:"advancedConfig"`
	WebConfig struct {
		BatteryWarningPerMailEnabled bool `json:"batteryWarningPerMailEnabled"`
	} `json:"webConfig"`
	State struct {
		Mode                      int   `json:"mode"`
		State                     int   `json:"state"`
		Trigger                   int   `json:"trigger"`
		LastAction                int   `json:"lastAction"`
		BatteryCritical           bool  `json:"batteryCritical"`
		BatteryCharging           bool  `json:"batteryCharging"`
		BatteryCharge             int32 `json:"batteryCharge"`
		KeypadBatteryCritical     bool  `json:"keypadBatteryCritical"`
		DoorsensorBatteryCritical bool  `json:"doorsensorBatteryCritical"`
		DoorState                 int   `json:"doorState"`
		RingToOpenTimer           int   `json:"ringToOpenTimer"`
		NightMode                 bool  `json:"nightMode"`
	} `json:"state"`
	FirmwareVersion     int       `json:"firmwareVersion"`
	HardwareVersion     int       `json:"hardwareVersion"`
	ServerState         int       `json:"serverState"`
	AdminPinState       int       `json:"adminPinState"`
	VirtualDevice       bool      `json:"virtualDevice"`
	CreationDate        time.Time `json:"creationDate"`
	UpdateDate          time.Time `json:"updateDate"`
	CurrentSubscription struct {
		Type         string    `json:"type"`
		State        string    `json:"state"`
		CreationDate time.Time `json:"creationDate"`
	} `json:"currentSubscription"`
}

func (s SmartlockResponse) ToSmartlockState() SmartLockState {
	return SmartLockState{
		Name:                      s.Name,
		BatteryCritical:           s.State.BatteryCritical,
		BatteryCharge:             s.State.BatteryCharge,
		KeypadBatteryCritical:     s.State.KeypadBatteryCritical,
		DoorsensorBatteryCritical: s.State.DoorsensorBatteryCritical,
	}
}
