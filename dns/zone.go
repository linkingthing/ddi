package dns

import ()

type Zone struct {
	ID           string
	ZoneName     string
	ZoneFileName string
	RRList       map[string]RR
}
