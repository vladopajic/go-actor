package actor

// OptOnStart adds function to Actor which will be executed
// before first worker's iteration.
// This functions is executed in actor's gorutine.
func OptOnStart(f func()) Option {
	return func(o *options) {
		o.Actor.OnStartFunc = f
	}
}

// OptOnStop adds function to Actor which will be executed
// after last worker's iteration.
// This functions is executed in actor's gorutine.
func OptOnStop(f func()) Option {
	return func(o *options) {
		o.Actor.OnStopFunc = f
	}
}

// OptCapacity sets initial Mailbox queue capacity.
// Value must be power of 2.
func OptCapacity(capacity int) Option {
	return func(o *options) {
		o.Mailbox.Capacity = capacity
	}
}

// OptMinCapacity sets minimum Mailbox queue capacity.
// Value must be power of 2.
func OptMinCapacity(minCapacity int) Option {
	return func(o *options) {
		o.Mailbox.MinCapacity = minCapacity
	}
}

// OptMailbox sets all Mailbox capacity options at once.
func OptMailbox(capacity, minCapacity int) Option {
	return func(o *options) {
		o.Mailbox.Capacity = capacity
		o.Mailbox.MinCapacity = minCapacity
	}
}

// OptUsingChan makes Mailbox to function as wrapper for
// native go channel.
func OptUsingChan(usingChan bool) Option {
	return func(o *options) {
		o.Mailbox.UsingChan = usingChan
	}
}

type Option func(o *options)

type options struct {
	Actor struct {
		OnStartFunc func()
		OnStopFunc  func()
	}

	Mailbox struct {
		UsingChan   bool
		Capacity    int
		MinCapacity int
	}
}

func (o *options) apply(opts []Option) {
	for _, opt := range opts {
		opt(o)
	}
}

func newOptions(opts []Option) options {
	o := &options{}
	o.apply(opts)

	return *o
}
