package model

type NukiAction int32

var (
	NukiActionUnlock                    = NukiAction(1)
	NukiActionLock                      = NukiAction(2)
	NukiActionUnlatch                   = NukiAction(3)
	NukiActionLockNGo                   = NukiAction(4)
	NukiActionLockNGoWithUnlatch        = NukiAction(5)
	NukiActionDoorWarningAjar           = NukiAction(208)
	NukiActionDoorWarningStatusMismatch = NukiAction(209)
	NukiActionDoorbellRecognition       = NukiAction(224)
	NukiActionDoorOpened                = NukiAction(240)
	NukiActionDoorClosed                = NukiAction(241)
	NukiActionDoorSensorJammed          = NukiAction(242)
	NukiActionFirmwareUpdate            = NukiAction(243)
	NukiActionDoorLogEnabled            = NukiAction(250)
	NukiActionDoorLogDisabled           = NukiAction(251)
	NukiActionInitialization            = NukiAction(252)
	NukiActionCalibration               = NukiAction(253)
	NukiActionLogEnabled                = NukiAction(254)
	NukiActionLogDisabled               = NukiAction(255)
	NukiActions                         = map[NukiAction]string{
		NukiActionUnlock:                    "unlock",
		NukiActionLock:                      "lock",
		NukiActionUnlatch:                   "unlatch",
		NukiActionLockNGo:                   "lock'n'go",
		NukiActionLockNGoWithUnlatch:        "lock'n'go with unlatch",
		NukiActionDoorWarningAjar:           "door warning ajar",
		NukiActionDoorWarningStatusMismatch: "door warning status mismatch",
		NukiActionDoorbellRecognition:       "doorbell recognition (only Opener)",
		NukiActionDoorOpened:                "door opened",
		NukiActionDoorClosed:                "door closed",
		NukiActionDoorSensorJammed:          "door sensor jammed",
		NukiActionFirmwareUpdate:            "firmware update",
		NukiActionDoorLogEnabled:            "door log enabled",
		NukiActionDoorLogDisabled:           "door log disabled",
		NukiActionInitialization:            "initialization",
		NukiActionCalibration:               "calibration",
		NukiActionLogEnabled:                "log enabled",
		NukiActionLogDisabled:               "log disabled",
	}
)
