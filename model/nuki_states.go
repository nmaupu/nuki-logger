package model

type NukiState int32

var (
	NukiStateSuccess           = NukiState(0)
	NukiStateMotorBlocked      = NukiState(1)
	NukiStateCanceled          = NukiState(2)
	NukiStateTooRecent         = NukiState(3)
	NukiStateBusy              = NukiState(4)
	NukiStateLowMotorVoltage   = NukiState(5)
	NukiStateClutchFailure     = NukiState(6)
	NukiStateMotorPowerFailure = NukiState(7)
	NukiStateIncomplete        = NukiState(8)
	NukiStateRejected          = NukiState(9)
	NukiStateRejectedNightMode = NukiState(10)
	NukiStateWrongKeypadCode   = NukiState(224)
	NukiStateOtherError        = NukiState(254)
	NukiStateUnknownError      = NukiState(255)
	NukiStates                 = map[NukiState]string{
		NukiStateSuccess:           "Success",
		NukiStateMotorBlocked:      "Motor blocked",
		NukiStateCanceled:          "Canceled",
		NukiStateTooRecent:         "Too recent",
		NukiStateBusy:              "Busy",
		NukiStateLowMotorVoltage:   "Low motor voltage",
		NukiStateClutchFailure:     "Clutch failure",
		NukiStateMotorPowerFailure: "Motor power failure",
		NukiStateIncomplete:        "Incomplete",
		NukiStateRejected:          "Rejected",
		NukiStateRejectedNightMode: "Rejected night mode",
		NukiStateWrongKeypadCode:   "Wrong keypad code",
		NukiStateOtherError:        "Other error",
		NukiStateUnknownError:      "Unknown error",
	}
)

func (n NukiState) GetEmoji() string {
	switch n {
	case NukiStateLowMotorVoltage:
	case NukiStateMotorBlocked:
	case NukiStateBusy:
		return "⚠️"
	case NukiStateWrongKeypadCode:
		return "⛔️"
	case NukiStateOtherError:
	case NukiStateUnknownError:
	case NukiStateRejectedNightMode:
	case NukiStateRejected:
	case NukiStateIncomplete:
	case NukiStateCanceled:
	case NukiStateClutchFailure:
	case NukiStateTooRecent:
	case NukiStateMotorPowerFailure:
		return "❌"
	case NukiStateSuccess:
		return "✅"
	}
	return n.String()
}
