package actor

func NewOptions(opts []Option) options {
	return newOptions(opts)
}

func NewZeroOptions() options {
	return options{}
}
