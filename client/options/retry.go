package options

import (
	"time"
)

// 仅仅当network错误是重试
type Retry struct {
	// 如果小于0, 那么会重试无数次
	// 如果=0, 不重试
	// 如果大于0, 重试MaxRetryTimes次
	MaxRetryTimes int
	// 最大累积耗时
	MaxTotalDuration time.Duration
	// 同节点重试
	RetrySameNode bool
}

func WithRetry(policy Retry) Option {
	if policy.MaxTotalDuration <= 0 {
		panic("unexpected value: MaxTotalDuration")
	}
	return Option{
		F: func(o *Options) {
			o.RetryPolocy = policy
		},
	}
}
