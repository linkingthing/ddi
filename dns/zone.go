package dns

import ()

type Zone struct {
	ID       string
	ZoneName string
	RRList   map[string]RR
}
