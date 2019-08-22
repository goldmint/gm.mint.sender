package conn

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// Conn is a connection to Sumus node
type Conn struct {
	logger       func(string)
	addr         string
	conn         net.Conn
	connTimeout  time.Duration
	recvTimeout  time.Duration
	sendTimeout  time.Duration
	readerBuffer int
	chanEvent    chan *RPCEvent
	hasSub       *int32
	closeLock    sync.Mutex
	closeFlag    *int32
}

// Options are options
type Options struct {
	ConnTimeout uint32
	RecvTimeout uint32
	SendTimeout uint32
	Logger      func(string)
}

// RPCEvent is receiving event: a valid message or an error
type RPCEvent struct {
	Message []byte
	Error   error
}

// New opens new RPC connection and runs serving i/o routines
func New(addr string, opts Options) (*Conn, error) {
	ret := &Conn{
		logger:       opts.Logger,
		addr:         addr,
		conn:         nil,
		connTimeout:  time.Second * 15,
		recvTimeout:  time.Second * 15,
		sendTimeout:  time.Second * 15,
		readerBuffer: 2048,
		chanEvent:    make(chan *RPCEvent),
		hasSub:       new(int32),
		closeLock:    sync.Mutex{},
		closeFlag:    new(int32),
	}

	// opts
	if opts.ConnTimeout > 0 {
		ret.connTimeout = time.Second * time.Duration(opts.ConnTimeout)
	}
	if opts.RecvTimeout > 0 {
		ret.recvTimeout = time.Second * time.Duration(opts.RecvTimeout)
	}
	if opts.SendTimeout > 0 {
		ret.sendTimeout = time.Second * time.Duration(opts.SendTimeout)
	}

	// connect
	conn, err := net.DialTimeout("tcp", ret.addr, ret.connTimeout)
	if err != nil {
		return nil, err
	}
	ret.conn = conn

	// serve
	go func() {
		ret.serve()
	}()

	return ret, nil
}

// Close closes RPC connection
func (r *Conn) Close() {
	if r.logger != nil {
		r.logger("Close")
	}

	// close
	atomic.StoreInt32(r.closeFlag, 1)

	// empty event chan
	select {
	case <-r.chanEvent:
	default:
	}

	r.closeLock.Lock()
	defer r.closeLock.Unlock()

	if r.chanEvent != nil {
		if r.logger != nil {
			r.logger("Closed")
		}
		close(r.chanEvent)
		r.chanEvent = nil
	}
}

// Closing checks connections should be closed
func (r *Conn) Closing() bool {
	return atomic.LoadInt32(r.closeFlag) != 0
}

// Subscribe signals that this connection has subscriber
func (r *Conn) Subscribe() <-chan *RPCEvent {
	atomic.StoreInt32(r.hasSub, 1)
	return r.chanEvent
}

// Unsubscribe signals that this connection has no subscriber
func (r *Conn) Unsubscribe() {
	atomic.StoreInt32(r.hasSub, 0)
	select {
	case _ = <-r.chanEvent:
	default:
	}
}

// serve serves the connection io
func (r *Conn) serve() {
	r.closeLock.Lock()
	defer r.closeLock.Unlock()

	wg := sync.WaitGroup{}
	wg.Add(2)

	// close
	go func() {
		defer wg.Done()
		for {
			if r.Closing() {
				// close tcp connection to unblock reading op
				r.conn.Close()
				return
			}
			time.Sleep(time.Millisecond * 100)
		}
	}()

	// receiver
	go func() {
		defer wg.Done()
		for !r.Closing() {

			data, err := r.receiveMessage()
			if err != nil {
				return
			}

			if r.logger != nil {
				r.logger("Received:\n" + hex.Dump(data))
			}

			r.tryPushEvent(data, nil)
		}
	}()

	wg.Wait()
	r.conn.Close()
}

// sendMessage sends message
func (r *Conn) sendMessage(data []byte) (err error) {

	err = nil
	defer func() {
		if err != nil {
			if !r.Closing() {
				if r.logger != nil {
					r.logger("I/O failure (sending): " + err.Error())
				}
			}
			// close it
			atomic.StoreInt32(r.closeFlag, 1)
		}
	}()

	// set timeout
	if err = r.conn.SetWriteDeadline(time.Now().Add(r.sendTimeout)); err != nil {
		return
	}

	// write data bytes
	var n int
	n, err = r.conn.Write(data)
	if err != nil {
		return
	}
	if n != len(data) {
		err = fmt.Errorf("failed to write exact amount of data bytes, written %v, expected %v", n, len(data))
		return
	}

	// write \0
	n, err = r.conn.Write([]byte{0})
	if err != nil {
		return
	}
	if n != 1 {
		err = fmt.Errorf("failed to write terminator")
		return
	}

	if r.logger != nil {
		r.logger("Sent:\n" + hex.Dump(data))
	}

	return
}

// receiveMessage receives message (blocking)
func (r *Conn) receiveMessage() (data []byte, err error) {

	data = nil
	err = nil

	defer func() {
		if err != nil {
			if !r.Closing() {
				if r.logger != nil {
					r.logger("I/O failure (reading): " + err.Error())
				}
				r.tryPushEvent(nil, err)
			}
			// close it
			atomic.StoreInt32(r.closeFlag, 1)
		}
	}()

	// read until next \0
	reader := bufio.NewReaderSize(r.conn, r.readerBuffer)
	data, err = reader.ReadBytes(0)
	if err != nil {
		data = nil
		return
	}

	data = data[:len(data)-1]
	return
}

// tryPushEvent makes an attempt to send event to the subscriber
func (r *Conn) tryPushEvent(msg []byte, err error) {
	if atomic.LoadInt32(r.hasSub) != 0 {
		select {
		case r.chanEvent <- &RPCEvent{
			Message: msg,
			Error:   err,
		}:
		case <-time.After(time.Second * 30):
		}
	}
}
