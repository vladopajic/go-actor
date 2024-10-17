package actor

// OptOnStart adds a function to the Actor that will be executed
// before the first iteration of the Worker.
//
// This allows for any necessary initialization or setup to occur
// before the Worker begins processing its tasks.
//
// Note:
//  1. The provided function will be executed within the
//     Actor's goroutine, ensuring safe access to the Actor's resources.
//  2. If the Actor implements the StoppableWorker interface, the
//     specified function will be called after the corresponding
//     method from this interface has been invoked.
func OptOnStart(f func(Context)) Option {
	return func(o *options) {
		o.Actor.OnStartFunc = f
	}
}

// OptOnStop adds a function to the Actor that will be executed
// after the last iteration of the Worker.
//
// This allows for any necessary cleanup or resource release
// to occur once the Worker has completed its tasks.
//
// Note:
//  1. The provided function will be executed within the
//     Actor's goroutine, ensuring safe access to the Actor's resources.
//  2. If the Actor implements the StoppableWorker interface, the
//     specified function will be called after the corresponding
//     method from this interface has been invoked.
func OptOnStop(f func()) Option {
	return func(o *options) {
		o.Actor.OnStopFunc = f
	}
}

// OptCapacity sets the queue capacity for the Mailbox.
//
// This option allows you to specify the initial capacity of the
// Mailbox queue. Setting an appropriate capacity can help
// optimize performance by minimizing unnecessary capacity
// increases as messages are queued. A well-defined capacity
// may prevent the Mailbox from having to dynamically resize
// its internal structures, which can lead to performance
// improvements in high-throughput scenarios.
//
// When used in conjunction with the `OptAsChan` option, the
// specified capacity value is applied to the underlying
// Go channel, allowing the Mailbox to mimic the behavior of
// a standard channel with the defined capacity (buffering).
func OptCapacity(capacity int) MailboxOption {
	return func(o *options) {
		o.Mailbox.Capacity = capacity
	}
}

// OptAsChan transforms the Mailbox into a wrapper for a native Go channel.
//
// When this option is applied, the Mailbox will behave like a
// standard Go channel, allowing seamless integration with code
// that expects channel behavior. This includes support for
// blocking and non-blocking sends and receives.
//
// Note: When using `OptAsChan`, ensure that the Mailbox is
// configured appropriately, especially regarding capacity,
// to align with the expected behavior of Go channels.
func OptAsChan() MailboxOption {
	return func(o *options) {
		o.Mailbox.AsChan = true
	}
}

// OptStopAfterReceivingAll configures the Mailbox to close its
// ReceiveC channel only after all messages have been received.
//
// By default, when the Mailbox is stopped, it will immediately
// close the ReceiveC channel, signaling to Actors that no new
// messages can be received. However, there are scenarios
// where it is necessary to process all queued messages before
// stopping the Mailbox. This option ensures that all messages
// in the queue are handled before the channel is closed,
// preventing potential data loss and ensuring that all
// Actors have finished processing.
func OptStopAfterReceivingAll() MailboxOption {
	return func(o *options) {
		o.Mailbox.StopAfterReceivingAll = true
	}
}

// OptStopTogether ensures that all combined Actors are stopped
// when any single Actor within the group is stopped.
//
// This option is useful in scenarios where you want to maintain
// a coordinated shutdown process across multiple Actors.
// When applied, stopping any one of the combined Actors will
// trigger the stop process for all Actors, ensuring they
// all cease execution simultaneously.
func OptStopTogether() CombinedOption {
	return func(o *options) {
		o.Combined.StopTogether = true
	}
}

// OptOnStopCombined registers a function to be executed after
// all combined Actors have been stopped.
//
// This option allows you to specify a cleanup or finalization
// function that will be called once the stop process for
// all Actors in the combination is complete. This is useful
// for performing any necessary resource cleanup, logging, or
// final state updates after the Actors have ceased execution.
func OptOnStopCombined(f func()) CombinedOption {
	return func(o *options) {
		o.Combined.OnStopFunc = f
	}
}

// OptOnStartCombined registers a function to be executed before
// the start of all combined Actors.
//
// This option allows you to specify an initialization function
// that will be called prior to the execution of any combined
// Actors. This is useful for performing any necessary setup,
// such as resource allocation before the Actors begin their work.
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
