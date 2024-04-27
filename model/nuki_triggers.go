package model

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
