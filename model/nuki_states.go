package model

import "github.com/enescakir/emoji"

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
	case NukiStateTooRecent:
		return emoji.Warning.String()
	case NukiStateWrongKeypadCode:
	case NukiStateRejectedNightMode:
	case NukiStateRejected:
		return emoji.RedCircle.String()
	case NukiStateOtherError:
	case NukiStateUnknownError:
	case NukiStateClutchFailure:
	case NukiStateMotorPowerFailure:
		return emoji.ExclamationMark.String()
	case NukiStateCanceled:
	case NukiStateIncomplete:
		return emoji.HollowRedCircle.String()
	case NukiStateSuccess:
		return emoji.GreenCircle.String()
	}
	return n.String()
}
