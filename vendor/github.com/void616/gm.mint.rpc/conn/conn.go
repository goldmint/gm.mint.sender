package conn

import (
	"context"
	"fmt"
	"math"
	"net"
	"sync"
	"time"
)

var terminator byte = 0

// Conn is a Mint node client connection
type Conn struct {
	logger func(string, string)
	id     string

	conn net.Conn

	serveLock sync.Mutex
	stopOnce  sync.Once
	stopper   chan struct{}

	rpcRequestCounter *uint32
	recvLock          sync.Mutex
	recvContext       context.Context
}

// New instance
func New(addr string, opts Options) (*Conn, error) {

	// connect
	conn, err := net.DialTimeout("tcp", addr, time.Duration(math.Max(float64(time.Second), float64(opts.ConnTimeout))))
	if err != nil {
		return nil, err
	}

	c := &Conn{
		logger:            opts.Logger,
		id:                opts.ID,
		conn:              conn,
		rpcRequestCounter: new(uint32),
		stopper:           make(chan struct{}),
		recvContext:       nil,
	}
	return c, nil
}

// Close instance and release resources
func (c *Conn) Close() {
	c.log("Closing")
	defer c.log("Closed")
	// stop routine in case it's not stopped
	c.markStopping()
	// close connection in case it's not closed
	c.conn.Close()
	// wait routine termination
	c.serveLock.Lock()
	defer c.serveLock.Unlock()
}

// Stopping checks connections is stopping and have to be closed
func (c *Conn) Stopping() bool {
	select {
	case _, ok := <-c.stopper:
		return !ok
	default:
		return false
	}
}

func (c *Conn) markStopping() {
	c.stopOnce.Do(func() {
		close(c.stopper)
		c.log("Stopping")
	})
}

// Serve serves the connection (call is blocking) until failure or Close() invokation
func (c *Conn) Serve() error {
	c.log("Routine started")
	defer c.log("Routine stopped")

	c.serveLock.Lock()
	defer c.serveLock.Unlock()

	// keep first error from subroutines
	var fe error
	var feOnce sync.Once
	var setErr = func(err error) {
		feOnce.Do(func() {
			fe = err
		})
	}

	wg := sync.WaitGroup{}

	// receiving subroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer c.conn.Close()
		defer c.markStopping()

		// EOF is not an error here
		if err := c.recv(c.conn); err != nil {
			c.log(fmt.Sprintf("Recv error: %v", err))
			setErr(err)
			return
		}
	}()

	wg.Wait()
	return fe
}

func (c *Conn) log(msg string) {
	if c.logger != nil {
		c.logger(c.id, msg)
	}
}
