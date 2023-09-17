package coco

import (
	"context"
	errors2 "errors"
	"github.com/246859/codis/pkg/util/syncx"
	"github.com/pkg/errors"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Config struct {
	MaxConn      uint64        `yaml:"maxConn"`
	Timeout      time.Duration `yaml:"timeout"`
	CloseTimeout time.Duration `yaml:"closeTimeout"`
	Retry        time.Duration `yaml:"retry"`
}

type Option func(cfg *Config)

func (o Option) apply(cfg *Config) {
	o(cfg)
}

func WithTimeout(timeout time.Duration) Option {
	return func(cfg *Config) {
		cfg.Timeout = timeout
	}
}

func WithMaxConn(maxConn uint64) Option {
	return func(cfg *Config) {
		cfg.MaxConn = maxConn
	}
}

func WithRetry(retry time.Duration) Option {
	return func(cfg *Config) {
		cfg.Retry = retry
	}
}

// NewServer create a new tcp server
func NewServer(ctx context.Context, opts ...Option) *Server {

	newServer := new(Server)

	newServer.ctx = ctx

	for _, opt := range opts {
		opt.apply(&newServer.cfg)
	}

	if newServer.cfg.MaxConn == 0 {
		newServer.cfg.MaxConn = 128
	}

	if newServer.cfg.Timeout == 0 {
		newServer.cfg.Timeout = 10 * time.Second
	}

	if newServer.cfg.Retry == 0 {
		newServer.cfg.Retry = 2 * time.Second
	}

	if newServer.ctx == nil {
		newServer.ctx = context.Background()
	}

	newServer.listeners = make(map[*net.Listener]*Handler, newServer.cfg.MaxConn)

	return newServer
}

type Server struct {
	cfg Config

	ctx context.Context

	// record the number of active client connections on the server
	connCount uint64

	closed atomic.Bool

	listeners map[*net.Listener]*Handler

	lngroups sync.WaitGroup

	mu sync.Mutex

	closeCh chan struct{}
}

func (s *Server) isShutdown() bool {
	return s.closed.Load()
}

func (s *Server) Shutdown() error {
	if s.isShutdown() {
		return ErrServerStopped
	}
	s.closed.Store(true)

	s.mu.Lock()
	// close listeners and handlers
	lnerr := s.closeListeners()
	s.mu.Unlock()

	done := syncx.Wait(func() {
		s.lngroups.Wait()
	})
	defer close(done)

	timer := time.NewTimer(s.cfg.CloseTimeout)
	defer timer.Stop()

	select {
	case <-done:
		return lnerr
	case <-timer.C:
		return ErrClosedTimeout
	}
}

func (s *Server) closeListeners() error {
	var err error
	for l, h := range s.listeners {
		// close the listener to refuse new connection
		lerr := (*l).Close()
		// how to close the exists connections depends on the handler's implementation
		herr := (*h).Close()
		err = errors2.Join(herr, lerr)
	}
	return err
}

func (s *Server) trackListener(lis net.Listener, handler Handler, add bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listeners == nil {
		s.listeners = make(map[*net.Listener]*Handler, 8)
	}

	if add {
		if s.isShutdown() {
			return false
		}
		s.listeners[&lis] = &handler
		s.lngroups.Add(1)
	} else {
		delete(s.listeners, &lis)
		s.lngroups.Done()
	}

	return true
}

// Serve start to accept tcp connections
func (s *Server) Serve(lis net.Listener, handler Handler) error {

	if lis == nil {
		return errors.Wrap(ErrInvalidListener, "nil")
	}

	if handler == nil {
		return errors.Wrap(ErrInvalidHandler, "nil")
	}

	if lis.Addr().Network() != protocol {
		handler.Close()
		lis.Close()
		return errors.Wrap(ErrUnsupportedProtocol, lis.Addr().Network())
	}

	timeC := 0

	if !s.trackListener(lis, handler, true) {
		return ErrServerStopped
	}
	defer s.trackListener(lis, handler, false)

	log.Printf("tcp server is listening on %s\n", lis.Addr().String())

	// handle connection
	for {

		conn, err := lis.Accept()

		if err != nil {
			// while listener is closed, accept will return error immediately
			if s.isShutdown() {
				return ErrServerStopped
			}

			var neterr net.Error
			if errors.As(err, &neterr) && neterr != nil && neterr.Timeout() && timeC < 5 {
				log.Println("connect timeout, will be retry in 2 seconds")
				time.Sleep(s.cfg.Retry)
				timeC++
				continue
			}
			return err
		}

		s.connCount++
		log.Printf("[%d]connection established: %s\n", s.connCount, conn.RemoteAddr())

		go func() {
			defer func() {
				s.mu.Lock()
				s.connCount--
				s.mu.Unlock()
			}()
			handler.Handle(s.ctx, conn)
		}()
	}
}
