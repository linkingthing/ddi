package postgres

import (
	"time"
)

type Lease4 struct {
	Address       int
	Hwaddr        []byte
	ClientId      []byte
	ValidLifetime int32
	Expire        time.Time
	SubnetId      int64
	FqdnFwd       bool
	FqdnRev       bool
	Hostname      string
	State         int64
	UserContext   string
}
