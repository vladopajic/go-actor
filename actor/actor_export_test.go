package actor

func NewContext() *contextImpl {
	return newContext()
}

func (c *contextImpl) SignalEnd() {
	c.signalEnd()
}
