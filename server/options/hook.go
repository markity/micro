package options

func WithDoAfterRunHook(f func()) Option {
	return Option{
		F: func(o *Options) {
			o.DoAfterRunHook = f
		},
	}
}

func WithDoAfterStopHook(f func()) Option {
	return Option{
		F: func(o *Options) {
			o.DoAfterStopHook = f
		},
	}
}
