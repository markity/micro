package options

import "context"

func WithCtx(ctx context.Context) Option {
	if ctx == nil {
		panic("unexpected nil")
	}
	return Option{
		F: func(o *Options) {
			o.Ctx = ctx
		},
	}
}
