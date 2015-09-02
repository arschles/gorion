package mq

import (
	"fmt"
	"strconv"

	"github.com/arschles/gorion/Godeps/_workspace/src/golang.org/x/net/context"
)

// Timeout is the number of seconds until a message reservation times out
type Timeout uint32

// String converts a timeout value to a printable string
func (t Timeout) String() string {
	return strconv.Itoa(int(t))
}

// Wait is the number of seconds to wait for a message
type Wait uint16

// String converts a wait value to a printable string
func (w Wait) String() string {
	return strconv.Itoa(int(w))
}

func waitInRange(w Wait) bool {
	return w <= MaxWait && w >= MinWait
}

func timeoutInRange(t Timeout) bool {
	return t <= MaxTimeout && t >= MinTimeout
}

const (
	// MinTimeout is the minimum value for a Timeout
	MinTimeout Timeout = 30
	// MaxTimeout is the maximum value for a Timeout
	MaxTimeout Timeout = 86400
	// MinWait is the minimum value for a Wait
	MinWait Wait = 0
	// MaxWait is the maximum value for a wait
	MaxWait Wait = 30
)

var (
	// ErrTimeoutOutOfRange is returned when a Timeout is given that's out of the [MinTimeout, MaxTimeout] range
	ErrTimeoutOutOfRange = fmt.Errorf("timeout out of range [%d, %d]", MinTimeout, MaxTimeout)
	// ErrWaitOutOfRange is returned when a Wait is given that's out of the [MinWait, MaxTimeout] range
	ErrWaitOutOfRange = fmt.Errorf("wait out of range [%d, %d]", MinWait, MaxWait)
)

// Enqueued is the the result of the Enqueue interface function
type Enqueued struct {
	// IDs are the IDs of the enqueued messages
	IDs []string `json:"ids"`
	// Msg is the resulting status of the enqueue operation
	Msg string `json:"msg"`
}

// Client is an interface for communicating with the IronMQ service.
type Client interface {
	// Enqueue enqueues msgs onto qName. if ctx.Done() receives before the enqueue
	// operation completes, the client must attempt to cancel the enqueue operation and
	// return no messages and a non-nil error.
	//
	// Note that clients need not roll back a partially applied enqueue operation if
	// ctx.Done() received before it completely finished
	Enqueue(ctx context.Context, qName string, msgs []NewMessage) (*Enqueued, error)

	// Dequeue dequeues at most num messages from qName or until wait expires.
	// Each dequeued message's reservation will expire after timeout. If delete is
	// false, all dequeued messages will be put back onto the queue after the
	// reservation expires, otherwise they will never go back onto the queue.
	//
	// Returns an empty slice of dequeued messages and an error if ctx.Done() receives
	// before the dequeue operation succeeds or any other error occurred. Also returns
	// errors if either timeout or wait are out of range
	//
	// Note that clients need not roll back a partially applied dequeue operation
	// if ctx.Done() received before it completely finished.
	Dequeue(ctx context.Context, qName string, num int, timeout Timeout, wait Wait, delete bool) ([]DequeuedMessage, error)
}
