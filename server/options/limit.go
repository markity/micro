package options

func WithQPSLimit(maxQPS int) Option {
	if maxQPS <= 0 {
		panic("should be greater than 0")
	}
	return Option{
		func(o *Options) {
			o.QPSLimit = &maxQPS
		},
	}
}
