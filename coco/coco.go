package coco

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
)

const (
	protocol = "tcp"
)

var (
	ErrServerStopped       = errors.New("coco: server already stopped")
	ErrInvalidListener     = errors.New("coco: invalid net listener")
	ErrInvalidHandler      = errors.New("coco: invalid handler")
	ErrUnsupportedProtocol = errors.New("coco: unsupported protocol")
	ErrClosedTimeout       = errors.New("coco: server closed connection timeout")
)

type Handler interface {
	Handle(ctx context.Context, conn net.Conn)
	Close() error
}

// CocoHandler Example handler, just for testing
type CocoHandler struct {
	closing atomic.Bool
	conns   map[*net.Conn]struct{}
	mu      sync.Mutex
}

func (c *CocoHandler) Handle(ctx context.Context, conn net.Conn) {
	if c.closing.Load() {
		conn.Close()
		return
	}

	c.mu.Lock()

	if c.conns == nil {
		c.conns = make(map[*net.Conn]struct{})
	}
	c.conns[&conn] = struct{}{}

	defer func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		delete(c.conns, &conn)
	}()

	c.mu.Unlock()

	var (
		reader = bufio.NewReader(conn)
	)

	for {
		select {
		// context canceled
		case <-ctx.Done():
			log.Println(ctx.Err())
			return
		default:
			readString, err := reader.ReadString('\n')
			if err != nil {
				if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
					return
				}
				log.Println("connection error", err)
				return
			}
			log.Println("receive", readString)
			n, err := conn.Write([]byte(readString))
			if err != nil {
				log.Println(err)
				return
			}
			log.Println("write", readString, n)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func (c *CocoHandler) Close() error {
	if c.closing.Load() {
		return errors.New("connection already closed")
	}
	c.closing.Store(true)

	c.mu.Lock()
	defer c.mu.Unlock()

	var closeErr error
	for conn, _ := range c.conns {
		closeErr = errors.Join(closeErr, (*conn).Close())
	}
	return closeErr
}
