package coco

import (
	"bufio"
	"context"
	"errors"
	"github.com/246859/codis/pkg/logger"
	"io"
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

	// lazy initialization
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
			logger.Warn(ctx.Err())
			return
		default:
			readString, err := reader.ReadString('\n')
			if err != nil {
				if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
					return
				}
				logger.Error("connection error: ", err)
				return
			}
			logger.Info("server receive: ", readString)
			_, err = conn.Write([]byte(readString))
			if err != nil {
				logger.Error("server write error: ", err)
				return
			}
			logger.Info("server send: ", readString)
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
