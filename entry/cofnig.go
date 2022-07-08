package entry

import (
	"time"
)

var GlobalPrefix string


type OptionConfig struct {
	TTL       time.Duration
	Timeout   time.Duration
	NoSpin    bool
}

type Config struct {
	Endpoints   []string
	Password    string
	DBIndex     int
	MaxConns    int
	IdleTimeout time.Duration
}
