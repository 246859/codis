package syncx

// Wait when the f() is completed successfully, the channel will be written with struct{}
func Wait(f func()) chan struct{} {
	done := make(chan struct{})
	go func() {
		f()
		done <- struct{}{}
	}()
	return done
}
