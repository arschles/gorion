package mq

import (
	"testing"
	"time"

	"github.com/arschles/gorion/Godeps/_workspace/src/github.com/arschles/assert"
	"github.com/arschles/gorion/Godeps/_workspace/src/github.com/arschles/synctest"
	"github.com/arschles/gorion/Godeps/_workspace/src/github.com/pivotal-golang/timer/fake_timer"
	"github.com/arschles/gorion/Godeps/_workspace/src/golang.org/x/net/context"
)

func TestReleaseReservedMsg(t *testing.T) {
	fakeTmr := fake_timer.NewFakeTimer(time.Now())
	lckr := synctest.NewNotifyingLocker()
	cl := memClient{tmr: fakeTmr, reserved: make(map[string]memMsg), q: nil, lck: lckr}
	msg := cl.newMemMsg(NewMessage{Body: "abc", Delay: 1, PushHeaders: make(map[string]string)})
	cl.reserved[msg.ReservationID] = msg
	go cl.releaseReservedMsg(msg.ReservationID, Timeout(2))
	lockCh := lckr.NotifyLock()
	fakeTmr.Elapse(3 * time.Second)
	<-lockCh // wait for the goroutine to get the lock and do its thing
	cl.lck.Lock()
	assert.Equal(t, 1, len(cl.q), "queue length")
	assert.Equal(t, 0, len(cl.reserved), "reserved length")
}

func TestDeferEnqueue(t *testing.T) {
	fakeTmr := fake_timer.NewFakeTimer(time.Now())
	lckr := synctest.NewNotifyingLocker()
	cl := memClient{tmr: fakeTmr, lck: lckr}
	msg := cl.newMemMsg(NewMessage{Body: "abc", Delay: 1, PushHeaders: make(map[string]string)})
	go cl.deferEnqueue(msg)
	cl.lck.Lock()
	assert.Equal(t, 0, len(cl.q), "queue length")
	assert.Equal(t, 0, len(cl.reserved), "reserved length")
	cl.lck.Unlock()
	lockCh := lckr.NotifyLock()
	fakeTmr.Elapse(2 * time.Second)
	<-lockCh // wait for the goroutine to get the lock and do its thing
	cl.lck.Lock()
	assert.Equal(t, 1, len(cl.q), "queue length")
	assert.Equal(t, 0, len(cl.reserved), "reserved length")
	cl.lck.Unlock()
}

func TestEnqueueDequeue(t *testing.T) {
	cl := NewMemClient()
	enq, err := cl.Enqueue(context.Background(), "q", []NewMessage{
		{Body: "test1", Delay: 0, PushHeaders: make(map[string]string)},
	})
	assert.NoErr(t, err)
	assert.Equal(t, len(enq.IDs), 1, "number of message IDs")
}