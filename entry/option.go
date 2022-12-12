package entry

import (
	"time"
)

type Logger interface {
	Printf(format string, args ...interface{})
}

type Option func(config *Options)

//WithTTL
func WithTTL(ttl time.Duration) Option {
	return func(c *Options) {
		c.TTL = ttl
	}
}

//WithTimeout set lock max time
func WithTimeout(timeout time.Duration) Option {
	return func(c *Options) {
		c.Timeout = timeout
	}
}

//WithNoSpinMode
func WithNoSpinMode() Option {
	return func(c *Options) {
		c.NoSpin = true
	}
}

func WithLogger(l Logger) Option {
	return func(c *Options) {
		c.Logger = l
	}
}
