package mq

import (
	"testing"
	"time"

	"github.com/arschles/gorion/Godeps/_workspace/src/github.com/arschles/assert"
	"github.com/arschles/gorion/Godeps/_workspace/src/github.com/arschles/synctest"
	"github.com/arschles/gorion/Godeps/_workspace/src/github.com/pivotal-golang/timer/fake_timer"
)

func TestReleaseReservedMsg(t *testing.T) {
	fakeTmr := fake_timer.NewFakeTimer(time.Now())
	lckr := synctest.NewNotifyingLocker()
	cl := MemClient{tmr: fakeTmr, reserved: make(map[string]memMsg), queues: make(map[string][]memMsg), lck: lckr}
	msg := cl.newMemMsg(NewMessage{Body: "abc", Delay: 1, PushHeaders: make(map[string]string)})
	cl.reserved[msg.ReservationID] = msg
	go cl.releaseReservedMsg(projID, qName, msg.ReservationID, Timeout(2))
	lockCh := lckr.NotifyLock()
	fakeTmr.Elapse(3 * time.Second)
	<-lockCh // wait for the goroutine to get the lock and do its thing
	cl.lck.Lock()
	defer cl.lck.Unlock()
	assert.Equal(t, 1, len(cl.queues[qKey(projID, qName)]), "queue length")
	assert.Equal(t, 0, len(cl.reserved), "reserved length")
}

func TestDeferEnqueue(t *testing.T) {
	fakeTmr := fake_timer.NewFakeTimer(time.Now())
	lckr := synctest.NewNotifyingLocker()
	cl := MemClient{tmr: fakeTmr, lck: lckr, queues: make(map[string][]memMsg)}
	msg := cl.newMemMsg(NewMessage{Body: "abc", Delay: 1, PushHeaders: make(map[string]string)})
	go cl.deferEnqueue(projID, qName, msg)
	cl.lck.Lock()
	assert.Equal(t, 0, len(cl.queues[qKey(projID, qName)]), "queue length")
	assert.Equal(t, 0, len(cl.reserved), "reserved length")
	cl.lck.Unlock()
	lockCh := lckr.NotifyLock()
	fakeTmr.Elapse(2 * time.Second)
	<-lockCh // wait for the goroutine to get the lock and do its thing
	cl.lck.Lock()
	assert.Equal(t, 1, len(cl.queues[qKey(projID, qName)]), "queue length")
	assert.Equal(t, 0, len(cl.reserved), "reserved length")
	cl.lck.Unlock()
}

func TestMemQueueOperations(t *testing.T) {
	cl := NewMemClient()
	err := qOperations(cl)
	assert.NoErr(t, err)
}
