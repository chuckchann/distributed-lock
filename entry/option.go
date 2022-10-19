package entry

import (
	"time"
)

type Option func(config *OptionConfig)

//WithTTL
func WithTTL(ttl time.Duration) Option {
	return func(c *OptionConfig) {
		c.TTL = ttl
	}
}

//WithTimeout set lock max time
func WithTimeout(timeout time.Duration) Option {
	return func(c *OptionConfig) {
		c.Timeout = timeout
	}
}

//WithNoSpinMode
func WithNoSpinMode() Option {
	return func(c *OptionConfig) {
		c.NoSpin = true
	}
}
