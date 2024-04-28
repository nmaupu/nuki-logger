package model

import "github.com/enescakir/emoji"

type NukiTrigger int32

var (
	NukiTriggerSystem    = NukiTrigger(0)
	NukiTriggerManual    = NukiTrigger(1)
	NukiTriggerButton    = NukiTrigger(2)
	NukiTriggerAutomatic = NukiTrigger(3)
	NukiTriggerWeb       = NukiTrigger(4)
	NukiTriggerApp       = NukiTrigger(5)
	NukiTriggerAutoLock  = NukiTrigger(6)
	NukiTriggerAccessory = NukiTrigger(7)
	NukiTriggerKeypad    = NukiTrigger(255)
	NukiTriggers         = map[NukiTrigger]string{
		NukiTriggerSystem:    "system",
		NukiTriggerManual:    "manual",
		NukiTriggerButton:    "button",
		NukiTriggerAutomatic: "automatic",
		NukiTriggerWeb:       "web",
		NukiTriggerApp:       "app",
		NukiTriggerAutoLock:  "auto lock",
		NukiTriggerAccessory: "accessory",
		NukiTriggerKeypad:    "keypad",
	}
)

func (n NukiTrigger) GetEmoji() string {
	switch n {
	case NukiTriggerSystem:
		return emoji.Gear.String()
	case NukiTriggerKeypad:
		return emoji.InputNumbers.String()
	case NukiTriggerApp:
		return emoji.MobilePhone.String()
	case NukiTriggerWeb:
		return emoji.GlobeShowingEuropeAfrica.String()
	case NukiTriggerButton:
		return emoji.RadioButton.String()
	case NukiTriggerManual:
		return emoji.HandWithFingersSplayed.String()
	}
	return n.String()
}
