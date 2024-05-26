package options

// 会在服务注册之后调用
func WithDoAfterRunHook(f func()) Option {
	return Option{
		F: func(o *Options) {
			o.DoAfterRunHook = f
		},
	}
}

// 会在服务解除注册之后调用
func WithDoAfterStopHook(f func()) Option {
	return Option{
		F: func(o *Options) {
			o.DoAfterStopHook = f
		},
	}
}
