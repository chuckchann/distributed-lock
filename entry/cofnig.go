package entry

import (
	"time"
)

var GlobalPrefix string

type Options struct {
	TTL           time.Duration
	Timeout       time.Duration
	NoSpin        bool
	Logger        Logger
	RenewalTime   time.Duration
	MaxRetryTimes int
}

type Config struct {
	Endpoints   []string
	Password    string
	DBIndex     int
	MaxConns    int
	IdleTimeout time.Duration
}
