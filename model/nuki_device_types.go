package model

type NukiDeviceType int32

var (
	NukiDeviceTypeSmartlock = NukiDeviceType(0)
	NukiDeviceTypeOpener    = NukiDeviceType(2)
	NukiDeviceTypeSmartdoor = NukiDeviceType(3)
	NukiDeviceTypes         = map[NukiDeviceType]string{
		NukiDeviceTypeSmartlock: "smartlock",
		NukiDeviceTypeOpener:    "opener",
		NukiDeviceTypeSmartdoor: "smartdoor",
	}
)
