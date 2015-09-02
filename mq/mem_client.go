package mq

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/arschles/gorion/Godeps/_workspace/src/code.google.com/p/go-uuid/uuid"
	"github.com/arschles/gorion/Godeps/_workspace/src/github.com/pivotal-golang/timer"
	"github.com/arschles/gorion/Godeps/_workspace/src/golang.org/x/net/context"
)

type memMsg struct {
	NewMessage
	DequeuedMessage
}

type memClient struct {
	lck sync.Locker
	tmr timer.Timer
	// the counter for message IDs
	ctr uint64
	// the live queue
	q []memMsg
	// the map from reservation ID to the message
	reserved map[string]memMsg
}

// NewMemClient returns a purely in-memory Client implementation that can be used
// for testing
func NewMemClient() Client {
	mtx := sync.Mutex{}
	return &memClient{
		lck:      &mtx,
		tmr:      timer.NewTimer(),
		ctr:      0,
		q:        nil,
		reserved: make(map[string]memMsg),
	}
}

func (m *memClient) newMemMsg(n NewMessage) memMsg {
	id := atomic.AddUint64(&m.ctr, 1)
	return memMsg{
		NewMessage: n,
		DequeuedMessage: DequeuedMessage{
			ID:            int(id),
			Body:          n.Body,
			ReservedCount: 0,
			ReservationID: "",
		},
	}
}

func (m *memClient) Enqueue(ctx context.Context, qName string, msgs []NewMessage) (*Enqueued, error) {
	ret := &Enqueued{}
	m.lck.Lock()
	defer m.lck.Unlock()
	for _, msg := range msgs {
		mmsg := m.newMemMsg(msg)
		if mmsg.Delay > 0 {
			go m.deferEnqueue(mmsg)
		} else {
			m.q = append(m.q, mmsg)
		}
		ret.IDs = append(ret.IDs, string(mmsg.ID))
	}
	ret.Msg = "Messages put on queue"
	return ret, nil
}

func (m *memClient) Dequeue(ctx context.Context, qName string, num int, timeout Timeout, wait Wait, delete bool) ([]DequeuedMessage, error) {
	ch := make(chan memMsg)

	go func() {
		m.lck.Lock()
		defer m.lck.Unlock()
		timeCh := m.tmr.After(time.Duration(int(wait)) * time.Second)
		for {
			select {
			case <-timeCh:
				close(ch)
				return
			default:
				if len(m.q) <= 0 {
					m.tmr.Sleep(100 * time.Millisecond)
				} else {
					msg := m.q[0]
					m.q = m.q[1:]
					msg.ReservedCount++
					msg.ReservationID = uuid.New()
					if !delete {
						m.reserved[msg.ReservationID] = msg
						go m.releaseReservedMsg(msg.ReservationID, timeout)
					}
					ch <- msg
				}
			}
		}
	}()

	var ret []DequeuedMessage
	for r := range ch {
		ret = append(ret, r.DequeuedMessage)
	}
	return ret, nil
}

func (m *memClient) releaseReservedMsg(resID string, timeout Timeout) {
	m.tmr.Sleep(time.Duration(int(timeout)) * time.Second)
	m.lck.Lock()
	defer m.lck.Unlock()
	msg, ok := m.reserved[resID]
	if !ok {
		return
	}
	delete(m.reserved, resID)
	m.q = append(m.q, msg)
}

func (m *memClient) deferEnqueue(msg memMsg) {
	m.tmr.Sleep(time.Duration(int(msg.Delay)) * time.Second)
	m.lck.Lock()
	defer m.lck.Unlock()
	m.q = append(m.q, msg)
}
