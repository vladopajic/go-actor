package actor

// OptOnStart adds function to actor which will be executed
// before first worker's iteration.
// This functions is executed in actor's gorutine.
func OptOnStart(f func()) Option {
	return func(o options) options {
		o.OnStartFunc = f
		return o
	}
}

// OptOnStop adds function to actor which will be executed
// after last worker's iteration.
// This functions is executed in actor's gorutine.
func OptOnStop(f func()) Option {
	return func(o options) options {
		o.OnStopFunc = f
		return o
	}
}

type Option func(o options) options

type options struct {
	OnStartFunc func()
	OnStopFunc  func()
}

func newOptions(opts []Option) options {
	return applyOptions(options{}, opts)
}

func applyOptions(base options, opts []Option) options {
	for _, opt := range opts {
		base = opt(base)
	}

	return base
}
