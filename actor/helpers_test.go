package actor_test

func drainC(c <-chan any, count int) {
	for i := 0; i < count; i++ {
		<-c
	}
}
