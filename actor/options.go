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

// OptCapacity sets Mailbox queue capacity.
// When `OptAsChan` is used together with `OptCapacity` capacity value
// will set be set to underlaying channel.
func OptCapacity(capacity int) MailboxOption {
	return func(o *options) {
		o.Mailbox.Capacity = capacity
	}
}

// OptAsChan transforms the Mailbox into a wrapper for the native Go channel.
func OptAsChan() MailboxOption {
	return func(o *options) {
		o.Mailbox.AsChan = true
	}
}

// OptStopAfterReceivingAll will close ReceiveC channel of Mailbox
// after all messages have been received from this channel.
func OptStopAfterReceivingAll() MailboxOption {
	return func(o *options) {
		o.Mailbox.StopAfterReceivingAll = true
	}
}

// OptStopTogether will stop all actors when any of combined
// actors is stopped.
func OptStopTogether() CombinedOption {
	return func(o *options) {
		o.Combined.StopTogether = true
	}
}

// OptOnStopCombined is called after all combined actors are stopped.
func OptOnStopCombined(f func()) CombinedOption {
	return func(o *options) {
		o.Combined.OnStopFunc = f
	}
}

// OptOnStartCombined is called before all.
func OptOnStartCombined(f func(Context)) CombinedOption {
	return func(o *options) {
		o.Combined.OnStartFunc = f
	}
}

type (
	option func(o *options)

	Option         option
	MailboxOption  option
	CombinedOption option
)

type options struct {
	Actor    optionsActor
	Combined optionsCombined
	Mailbox  optionsMailbox
}

type optionsActor struct {
	OnStartFunc func(Context)
	OnStopFunc  func()
}

type optionsCombined struct {
	StopTogether bool
	OnStopFunc   func()
	OnStartFunc  func(Context)
}

type optionsMailbox struct {
	AsChan                bool
	Capacity              int
	StopAfterReceivingAll bool
}

func newOptions[T ~func(o *options)](opts []T) options {
	o := &options{}

	for _, opt := range opts {
		opt(o)
	}

	return *o
}
