package options

import (
	"time"
)

type RetryBackoffPolicyType int

const (
	// 不等待
	RetryBackoffPolicyTypeNoWait RetryBackoffPolicyType = iota
	// Range时间内的随机等待时长, 如果两个时间一样, 则为固定时长
	RetryBackoffPolicyTypeRangeTime
)

// 重试等待策略
type RetryBackoffPolicy struct {
	Type RetryBackoffPolicyType
	// 下面两个字段仅仅在Type==RetryBackoffPolicyTypeRangeTime有意义
	TimeMin time.Duration
	TimeMax time.Duration
}

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
	// 重试等待策略
	BackoffPolocy RetryBackoffPolicy
}

func WithRetry(policy Retry) Option {
	if policy.MaxTotalDuration <= 0 {
		panic("unexpected value: MaxTotalDuration")
	}
	if policy.BackoffPolocy.Type == RetryBackoffPolicyTypeRangeTime &&
		policy.BackoffPolocy.TimeMin > policy.BackoffPolocy.TimeMax {
		panic("BackoffPolocy time unexpected: TimeMin should be less than TimeMax")
	}
	return Option{
		F: func(o *Options) {
			o.RetryPolocy = policy
		},
	}
}
