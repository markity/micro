package options

import (
	"time"
)

// 如果不设置, 默认为3s
func WithTimeout(duration time.Duration) Option {
	if duration <= 0 {
		panic("unexpected value: duration")
	}
	return Option{
		F: func(o *Options) {
			o.Timeout = duration
		},
	}
}
