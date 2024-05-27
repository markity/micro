package options

import (
	"time"

	"github.com/markity/micro/plugin/discovery"
	loadbalance "github.com/markity/micro/plugin/load_balance"
)

type Options struct {
	Discovery   discovery.Discovery
	LoadBalance loadbalance.LoadBalance
	Timeout     time.Duration
	RetryPolocy Retry
}

type Option struct {
	F func(*Options)
}
