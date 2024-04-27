package model

type NukiSource int32

var (
	NukiSourceDefault     = NukiSource(0)
	NukiSourceKeypadCode  = NukiSource(1)
	NukiSourceFingerprint = NukiSource(2)
	NukiSources           = map[NukiSource]string{
		NukiSourceDefault:     "Default",
		NukiSourceKeypadCode:  "Keypad code",
		NukiSourceFingerprint: "Fingerprint",
	}
)
