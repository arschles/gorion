package mq

import (
	"sync"
	"sync/atomic"
	"time"

	"code.google.com/p/go-uuid/uuid"
	"github.com/pivotal-golang/timer"
	"golang.org/x/net/context"
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
	queues map[string][]memMsg
	// the map from reservation ID to the message
	reserved map[string]memMsg
}

// NewMemClient returns a purely in-memory Client implementation that can be used
// for testing. Note that funcs with in-memory client receivers do not pay attention
// to the context.Context parameters that are passed to them.
func NewMemClient() Client {
	mtx := sync.Mutex{}
	return &memClient{
		lck:      &mtx,
		tmr:      timer.NewTimer(),
		ctr:      0,
		queues:   make(map[string][]memMsg),
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

func qKey(projID, qName string) string {
	return projID + "|" + qName
}

func (m *memClient) Enqueue(ctx context.Context, token, projID, qName string, msgs []NewMessage) (*Enqueued, error) {
	ret := &Enqueued{}
	m.lck.Lock()
	defer m.lck.Unlock()
	for _, msg := range msgs {
		mmsg := m.newMemMsg(msg)
		if mmsg.Delay > 0 {
			go m.deferEnqueue(projID, qName, mmsg)
		} else {
			q, _ := m.queues[qKey(projID, qName)]
			q = append(q, mmsg)
			m.queues[qKey(projID, qName)] = q
		}
		ret.IDs = append(ret.IDs, string(mmsg.ID))
	}
	ret.Msg = "Messages put on queue"
	return ret, nil
}

func (m *memClient) Dequeue(ctx context.Context, token, projID, qName string, num int, timeout Timeout, wait Wait, delete bool) ([]DequeuedMessage, error) {
	ch := make(chan memMsg)

	go func() {
		m.lck.Lock()
		defer m.lck.Unlock()
		timeCh := m.tmr.After(time.Duration(int(wait)) * time.Second)
		q := m.queues[qKey(projID, qName)]
		for {
			select {
			case <-timeCh:
				close(ch)
				return
			default:
				if len(q) <= 0 {
					m.tmr.Sleep(100 * time.Millisecond)
				} else {
					msg := q[0]
					q = q[1:]
					msg.ReservedCount++
					msg.ReservationID = uuid.New()
					if !delete {
						m.reserved[msg.ReservationID] = msg
						go m.releaseReservedMsg(projID, qName, msg.ReservationID, timeout)
					}
					ch <- msg
				}
			}
		}
		m.queues[qKey(projID, qName)] = q
	}()

	var ret []DequeuedMessage
	for r := range ch {
		ret = append(ret, r.DequeuedMessage)
	}
	return ret, nil
}

func (m *memClient) DeleteReserved(ctx context.Context, token, projID, qName string, messageID int, reservationID string) (*Deleted, error) {
	m.lck.Lock()
	defer m.lck.Unlock()
	msg, ok := m.reserved[reservationID]
	if !ok {
		return nil, ErrNoSuchReservation
	}
	if msg.ID != messageID {
		return nil, ErrNoSuchMessage
	}
	return &Deleted{Msg: "deleted"}, nil
}

func (m *memClient) releaseReservedMsg(projID, qName, resID string, timeout Timeout) {
	m.tmr.Sleep(time.Duration(int(timeout)) * time.Second)
	m.lck.Lock()
	defer m.lck.Unlock()
	msg, ok := m.reserved[resID]
	if !ok {
		return
	}
	delete(m.reserved, resID)
	m.queues[qKey(projID, qName)] = append(m.queues[qKey(projID, qName)], msg)
}

func (m *memClient) deferEnqueue(projID, qName string, msg memMsg) {
	m.tmr.Sleep(time.Duration(int(msg.Delay)) * time.Second)
	m.lck.Lock()
	defer m.lck.Unlock()
	m.queues[qKey(projID, qName)] = append(m.queues[qKey(projID, qName)], msg)
}
