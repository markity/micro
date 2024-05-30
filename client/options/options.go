package options

import (
	"time"

	"github.com/markity/micro/plugin/discovery"
	loadbalance "github.com/markity/micro/plugin/load_balance"
)

type Options struct {
	Discovery   discovery.Discovery     // 用户必须设置一个discovery, 否则就会报错
	LoadBalance loadbalance.LoadBalance // 如果不设置LoadBalance, 默认为RoundRobin
	Timeout     time.Duration           // 如果不设置, 默认值3s
	RetryPolocy Retry                   // 重试策略, 默认不开启重试
}

type Option struct {
	F func(*Options)
}
