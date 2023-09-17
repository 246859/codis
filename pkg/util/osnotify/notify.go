package osnotify

import (
	"os"
	"os/signal"
)

type Notification struct {
	sigs []os.Signal
	ch   chan os.Signal
}

func (n *Notification) Notify(cb func(os.Signal)) {
	if cb == nil {
		return
	}
	signal.Notify(n.ch, n.sigs...)
	go func() {
		sig := <-n.ch
		for _, s := range n.sigs {
			if s == sig {
				cb(s)
			}
		}
	}()
}

func Signals(sigs ...os.Signal) *Notification {
	return &Notification{sigs: sigs, ch: make(chan os.Signal)}
}
