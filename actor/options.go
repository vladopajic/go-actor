package actor

// OptOnStart adds function to Actor which will be executed
// before first worker's iteration.
// If Actor implements StartableWorker interface then function
// will be called after calling respective method on this interface.
// This functions is executed in actor's goroutine.
func OptOnStart(f func(Context)) Option {
	return func(o *options) {
		o.Actor.OnStartFunc = f
	}
}

// OptOnStop adds function to Actor which will be executed
// after last worker's iteration.
// If Actor implements StoppableWorker interface then function
// will be called after calling respective method on this interface.
// This functions is executed in actor's goroutine.
func OptOnStop(f func()) Option {
	return func(o *options) {
		o.Actor.OnStopFunc = f
	}
}

// OptCapacity sets initial Mailbox queue capacity.
// Value must be power of 2.
func OptCapacity(capacity int) MailboxOption {
	return func(o *options) {
		o.Mailbox.Capacity = capacity
	}
}

// OptMinCapacity sets minimum Mailbox queue capacity.
// Value must be power of 2.
func OptMinCapacity(minCapacity int) MailboxOption {
	return func(o *options) {
		o.Mailbox.MinCapacity = minCapacity
	}
}

// OptMailbox sets all Mailbox capacity options at once.
func OptMailbox(capacity, minCapacity int) MailboxOption {
	return func(o *options) {
		o.Mailbox.Capacity = capacity
		o.Mailbox.MinCapacity = minCapacity
	}
}

// OptAsChan makes Mailbox to function as wrapper for
// native go channel.
func OptAsChan() MailboxOption {
	return func(o *options) {
		o.Mailbox.AsChan = true
	}
}

// OptStopTogether will stop all actors when any of combined
// actors is stopped.
func OptStopTogether() CombinedOption {
	return func(o *options) {
		o.Combined.StopTogether = true
	}
}

type (
	option func(o *options)

	Option         option
	MailboxOption  option
	CombinedOption option
)

type options struct {
	Actor struct {
		OnStartFunc func(Context)
		OnStopFunc  func()
	}

	Combined struct {
		StopTogether bool
	}

	Mailbox struct {
		AsChan      bool
		Capacity    int
		MinCapacity int
	}
}

func newOptions[T ~func(o *options)](opts []T) options {
	o := &options{}

	for _, opt := range opts {
		opt(o)
	}

	return *o
}
