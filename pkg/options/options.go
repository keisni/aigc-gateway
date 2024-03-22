package options

import (
	"github.com/spf13/pflag"
	"time"
)

const (
	defaultCheckInterval = time.Second * 5
	defaultIdleLimit     = time.Duration(0)
)

type ServerOption struct {
	RedisAddress  string
	CheckInterval time.Duration
	IdleLimit     time.Duration
}

func NewServerOption() *ServerOption {
	return &ServerOption{}
}

func (s *ServerOption) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.RedisAddress, "redis-addr", "", "Redis address")
	fs.DurationVar(&s.CheckInterval, "check-interval", defaultCheckInterval, "Check Interval")
	fs.DurationVar(&s.IdleLimit, "idle-limit", defaultIdleLimit, "Idle limit")
}
